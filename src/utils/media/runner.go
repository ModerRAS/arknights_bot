package media

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	frozenAssetManifest     = "src/utils/media/testdata/visual/baseline/resource-manifest.json"
	maxRenderWidth           = 4096
	maxRenderHeight          = 4096
	maxRenderPixels          = 16 * 1024 * 1024
	maxSpecBytes             = 1 << 20
	maxPNGBytes              = 32 << 20
	maxPropsBytes            = 768 << 10
	maxResponseEnvelopeBytes = 64 << 10
	renderSlots              = 4
	renderTimeout            = 30 * time.Second
	jpegQuality              = 95
)

// maxResponseBytes bounds one NDJSON line, not the lifetime of stdout. It
// covers the largest allowed base64 PNG plus a bounded JSON envelope.
var maxResponseBytes = base64.StdEncoding.EncodedLen(maxPNGBytes) + maxResponseEnvelopeBytes

var (
	ErrRendererClosed   = errors.New("media renderer is closed")
	ErrRendererExit     = errors.New("media renderer exited")
	ErrRendererProtocol = errors.New("media renderer protocol error")
	componentName       = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)
)

var requestSequence atomic.Uint64

// renderSpec is the stable HTTP contract served by core/web.RenderSpec.
type renderSpec struct {
	Component string          `json:"component"`
	Width     int             `json:"width"`
	Height    int             `json:"height"`
	Props     json.RawMessage `json:"props"`
}

type renderRequest struct {
	ID        string          `json:"id"`
	Component string          `json:"component"`
	Width     int             `json:"width"`
	Height    int             `json:"height"`
	Scale     float64         `json:"scale"`
	Props     json.RawMessage `json:"props"`
}

type renderResponse struct {
	ID         string               `json:"id"`
	OK         bool                 `json:"ok"`
	MIME       string               `json:"mime,omitempty"`
	Width      int                  `json:"width,omitempty"`
	Height     int                  `json:"height,omitempty"`
	DataBase64 string               `json:"dataBase64,omitempty"`
	Error      *renderResponseError `json:"error,omitempty"`
}

type renderResponseError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Retryable bool   `json:"retryable"`
}

type pendingRender struct {
	response chan renderResult
}

type renderResult struct {
	png []byte
	err error
}

type rendererProcess struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	done   chan error
	wait   func() error
}

type renderer struct {
	mu      sync.Mutex
	writeMu sync.Mutex
	proc    *rendererProcess
	pending map[string]*pendingRender
	slots   chan struct{}
	closed  bool
	start   func() (*rendererProcess, error)
}

func newRenderer() *renderer {
	r := &renderer{pending: make(map[string]*pendingRender), slots: make(chan struct{}, renderSlots)}
	r.start = startRendererProcess
	return r
}

var defaultRenderer = newRenderer()

// ScreenshotPNG renders the local RenderSpec endpoint and returns the renderer's PNG bytes unchanged.
func ScreenshotPNG(urlString string, waitTime float64, scale float64) ([]byte, error) {
	_ = waitTime
	ctx, cancel := context.WithTimeout(context.Background(), renderTimeout)
	defer cancel()
	spec, err := fetchRenderSpec(ctx, urlString)
	if err != nil {
		return nil, err
	}
	request := renderRequest{
		ID:        nextRequestID(),
		Component: spec.Component,
		Width:     spec.Width,
		Height:    spec.Height,
		Scale:     scale,
		Props:     spec.Props,
	}
	if err := validateRenderRequest(request); err != nil {
		return nil, err
	}
	return defaultRenderer.render(ctx, request)
}

// RenderPNG is the explicit PNG form of Screenshot. The renderer bytes are returned unchanged.
func RenderPNG(urlString string, waitTime float64, scale float64) ([]byte, error) {
	return ScreenshotPNG(urlString, waitTime, scale)
}

// PNGToJPEG composites a PNG onto white and encodes it with the global JPEG quality.
func PNGToJPEG(data []byte) ([]byte, error) {
	return pngToJPEG(data)
}

// Screenshot preserves the historical JPEG-returning API. The PNG is composited onto white before encoding.
func Screenshot(urlString string, waitTime float64, scale float64) ([]byte, error) {
	data, err := ScreenshotPNG(urlString, waitTime, scale)
	if err != nil {
		return nil, err
	}
	return PNGToJPEG(data)
}

// Shutdown stops the resident renderer. It is safe to call more than once.
func Shutdown() {
	defaultRenderer.shutdown()
}

func nextRequestID() string {
	return fmt.Sprintf("go-%d-%d", time.Now().UnixNano(), requestSequence.Add(1))
}

func (r *renderer) render(ctx context.Context, request renderRequest) ([]byte, error) {
	if err := validateRenderRequest(request); err != nil {
		return nil, err
	}
	select {
	case r.slots <- struct{}{}:
		defer func() { <-r.slots }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	result := &pendingRender{response: make(chan renderResult, 1)}
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return nil, ErrRendererClosed
	}
	if r.proc == nil {
		proc, err := r.start()
		if err != nil {
			r.mu.Unlock()
			return nil, fmt.Errorf("renderer_start: %w", err)
		}
		r.proc = proc
		r.startReadersLocked(proc)
	}
	r.pending[request.ID] = result
	proc := r.proc
	r.mu.Unlock()

	line, err := json.Marshal(request)
	if err == nil {
		line = append(line, '\n')
		r.writeMu.Lock()
		_, err = proc.stdin.Write(line)
		r.writeMu.Unlock()
	}
	if err != nil {
		r.removePending(request.ID)
		r.failProcess(proc, fmt.Errorf("runner_exit: write request: %w", err))
		return nil, fmt.Errorf("runner_exit: write request: %w", err)
	}

	select {
	case result := <-result.response:
		return result.png, result.err
	case <-ctx.Done():
		r.removePending(request.ID)
		return nil, ctx.Err()
	}
}

func (r *renderer) startReadersLocked(proc *rendererProcess) {
	go r.readResponses(proc)
	go func() {
		var err error
		if proc.wait != nil {
			err = proc.wait()
		} else if proc.cmd != nil {
			err = proc.cmd.Wait()
		} else {
			err = errors.New("process wait unavailable")
		}
		select {
		case proc.done <- err:
		default:
		}
		r.failProcess(proc, fmt.Errorf("runner_exit: %w", err))
	}()
}

func (r *renderer) readResponses(proc *rendererProcess) {
	scanner := bufio.NewScanner(proc.stdout)
	scanner.Buffer(make([]byte, 64<<10), maxResponseBytes)
	for scanner.Scan() {
		var response renderResponse
		if err := decodeOneJSON(scanner.Bytes(), &response); err != nil {
			r.failProcess(proc, fmt.Errorf("protocol_error: decode response: %w", err))
			return
		}
		if err := r.routeResponse(proc, response); err != nil {
			r.failProcess(proc, err)
			return
		}
	}
	if err := scanner.Err(); err != nil {
		r.failProcess(proc, fmt.Errorf("protocol_error: read response: %w", err))
		return
	}
	r.failProcess(proc, fmt.Errorf("runner_exit: stdout closed"))
}

func (r *renderer) routeResponse(proc *rendererProcess, response renderResponse) error {
	if response.ID == "" {
		return fmt.Errorf("protocol_error: response id is empty")
	}
	r.mu.Lock()
	pending, ok := r.pending[response.ID]
	if !ok || r.proc != proc {
		r.mu.Unlock()
		return fmt.Errorf("protocol_error: unknown response id %q", response.ID)
	}
	r.mu.Unlock()

	if !response.OK {
		if response.Error == nil || !validErrorCode(response.Error.Code) || response.Error.Message == "" {
			return fmt.Errorf("protocol_error: malformed renderer error")
		}
		r.mu.Lock()
		delete(r.pending, response.ID)
		r.mu.Unlock()
		pending.response <- renderResult{err: fmt.Errorf("%s: %s", response.Error.Code, response.Error.Message)}
		return nil
	}
	if response.MIME != "image/png" || response.Width <= 0 || response.Height <= 0 || response.DataBase64 == "" {
		return fmt.Errorf("protocol_error: malformed renderer success")
	}
	data, err := base64.StdEncoding.DecodeString(response.DataBase64)
	if err != nil || len(data) == 0 || len(data) > maxPNGBytes {
		return fmt.Errorf("protocol_error: invalid PNG payload")
	}
	config, err := png.DecodeConfig(bytes.NewReader(data))
	if err != nil || config.Width != response.Width || config.Height != response.Height {
		return fmt.Errorf("protocol_error: PNG dimensions do not match response")
	}
	if config.Width > maxRenderWidth || config.Height > maxRenderHeight || config.Width*config.Height > maxRenderPixels {
		return fmt.Errorf("protocol_error: PNG exceeds limits")
	}
	r.mu.Lock()
	delete(r.pending, response.ID)
	r.mu.Unlock()
	pending.response <- renderResult{png: data}
	return nil
}

func validErrorCode(code string) bool {
	switch code {
	case "invalid_request", "render_error", "asset_error", "timeout", "protocol_error", "runner_exit":
		return true
	default:
		return false
	}
}

func (r *renderer) removePending(id string) {
	r.mu.Lock()
	delete(r.pending, id)
	r.mu.Unlock()
}

func (r *renderer) failProcess(proc *rendererProcess, err error) {
	r.mu.Lock()
	if r.proc != proc {
		r.mu.Unlock()
		return
	}
	r.proc = nil
	pending := r.pending
	r.pending = make(map[string]*pendingRender)
	r.mu.Unlock()
	for _, item := range pending {
		item.response <- renderResult{err: err}
	}
	_ = proc.stdin.Close()
}

func (r *renderer) shutdown() {
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return
	}
	r.closed = true
	proc := r.proc
	r.proc = nil
	pending := r.pending
	r.pending = make(map[string]*pendingRender)
	r.mu.Unlock()
	for _, item := range pending {
		item.response <- renderResult{err: ErrRendererClosed}
	}
	if proc != nil {
		_ = proc.stdin.Close()
		if proc.cmd != nil && proc.cmd.Process != nil {
			_ = proc.cmd.Process.Kill()
		}
	}
}

func startRendererProcess() (*rendererProcess, error) {
	root, rendererDir, err := findRendererDir()
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(rendererCommandArgs(filepath.Join(rendererDir, "runner.mjs"))[0], rendererCommandArgs(filepath.Join(rendererDir, "runner.mjs"))[1:]...)
	cmd.Dir = root
	cmd.Env = rendererEnvironment(root)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		_ = stdin.Close()
		return nil, err
	}
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		_ = stdin.Close()
		_ = stdout.Close()
		return nil, err
	}
	return &rendererProcess{cmd: cmd, stdin: stdin, stdout: stdout, done: make(chan error, 1), wait: cmd.Wait}, nil
}

func rendererCommandArgs(entry string) []string {
	return []string{"node", entry, "--ndjson"}
}

func rendererEnvironment(root string) []string {
	manifest := os.Getenv("SATORI_ASSET_MANIFEST")
	if manifest == "" {
		manifest = filepath.Join(root, frozenAssetManifest)
	}
	env := make([]string, 0, len(os.Environ())+1)
	for _, value := range os.Environ() {
		if strings.HasPrefix(value, "SATORI_ASSET_MANIFEST=") {
			continue
		}
		env = append(env, value)
	}
	return append(env, "SATORI_ASSET_MANIFEST="+manifest)
}

func decodeOneJSON(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return errors.New("multiple JSON values")
		}
		return err
	}
	return nil
}

func findRendererDir() (string, string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", "", err
	}
	for i := 0; i < 6; i++ {
		candidate := filepath.Join(dir, "renderer")
		if _, err := os.Stat(filepath.Join(candidate, "runner.mjs")); err == nil {
			return dir, candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", "", errors.New("renderer directory not found")
}

func validateRenderRequest(request renderRequest) error {
	if !componentName.MatchString(request.Component) || request.Component == "" {
		return fmt.Errorf("invalid_request: invalid component")
	}
	if request.Width <= 0 || request.Width > maxRenderWidth || request.Height <= 0 || request.Height > maxRenderHeight || request.Width > maxRenderPixels/request.Height {
		return fmt.Errorf("invalid_request: dimensions exceed limits")
	}
	if request.Scale < 0.25 || request.Scale > 4 || request.Scale != request.Scale {
		return fmt.Errorf("invalid_request: scale exceeds limits")
	}
	if len(request.Props) == 0 || len(request.Props) > maxPropsBytes || !json.Valid(request.Props) {
		return fmt.Errorf("invalid_request: props exceed limits")
	}
	return nil
}

func fetchRenderSpec(ctx context.Context, rawURL string) (renderSpec, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil || !isLoopbackURL(parsed) {
		return renderSpec{}, fmt.Errorf("invalid_request: render URL must be loopback http(s)")
	}
	client := &http.Client{
		Timeout: renderTimeout,
		Transport: &http.Transport{
			Proxy: nil,
			DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
				return dialLoopback(ctx, network, address)
			},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if !isLoopbackURL(req.URL) {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return renderSpec{}, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return renderSpec{}, fmt.Errorf("render spec request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return renderSpec{}, fmt.Errorf("render spec request: status %d", resp.StatusCode)
	}
	contentType := strings.TrimSpace(strings.Split(resp.Header.Get("Content-Type"), ";")[0])
	if contentType != "application/json" {
		return renderSpec{}, errors.New("invalid_request: render spec content type must be application/json")
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxSpecBytes+1))
	if err != nil {
		return renderSpec{}, fmt.Errorf("render spec read: %w", err)
	}
	if len(body) > maxSpecBytes {
		return renderSpec{}, errors.New("invalid_request: render spec exceeds limits")
	}
	var spec renderSpec
	if err := decodeOneJSON(body, &spec); err != nil {
		return renderSpec{}, fmt.Errorf("invalid_request: malformed render spec: %w", err)
	}
	if err := validateRenderRequest(renderRequest{Component: spec.Component, Width: spec.Width, Height: spec.Height, Scale: 1, Props: spec.Props}); err != nil {
		return renderSpec{}, err
	}
	return spec, nil
}

func isLoopbackURL(parsed *url.URL) bool {
	if parsed == nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.User != nil || parsed.Hostname() == "" {
		return false
	}
	host := parsed.Hostname()
	if strings.EqualFold(host, "localhost") {
		return parsed.Port() != ""
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback() && parsed.Port() != ""
}

func dialLoopback(ctx context.Context, network, address string) (net.Conn, error) {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	if !strings.EqualFold(host, "localhost") {
		ip := net.ParseIP(host)
		if ip == nil || !ip.IsLoopback() {
			return nil, errors.New("non-loopback dial rejected")
		}
	}
	return (&net.Dialer{}).DialContext(ctx, network, address)
}

func pngToJPEG(data []byte) ([]byte, error) {
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode renderer PNG: %w", err)
	}
	bounds := img.Bounds()
	background := image.NewRGBA(bounds)
	draw.Draw(background, bounds, &image.Uniform{C: color.White}, image.Point{}, draw.Src)
	draw.Draw(background, bounds, img, bounds.Min, draw.Over)
	var out bytes.Buffer
	if err := jpeg.Encode(&out, background, &jpeg.Options{Quality: jpegQuality}); err != nil {
		return nil, fmt.Errorf("encode renderer JPEG: %w", err)
	}
	return out.Bytes(), nil
}
