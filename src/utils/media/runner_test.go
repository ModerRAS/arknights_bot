package media

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

type testWriteCloser struct{ io.Writer }

type repeatedByteReader struct{ remaining int }

func (r *repeatedByteReader) Read(p []byte) (int, error) {
	if r.remaining == 0 {
		return 0, io.EOF
	}
	n := len(p)
	if n > r.remaining {
		n = r.remaining
	}
	for i := 0; i < n; i++ {
		p[i] = 'x'
	}
	r.remaining -= n
	return n, nil
}

func (testWriteCloser) Close() error { return nil }

func pngFixture(t *testing.T, c color.NRGBA) []byte {
	return pngFixtureSize(t, 1, 1, c)
}

func pngFixtureSize(t *testing.T, width, height int, c color.NRGBA) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetNRGBA(x, y, c)
		}
	}
	var buf bytes.Buffer
	encoder := png.Encoder{CompressionLevel: png.NoCompression}
	if err := encoder.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func fakeRenderer(t *testing.T, response func(renderRequest) renderResponse) (*renderer, func()) {
	t.Helper()
	stdinReader, stdinWriter := io.Pipe()
	stdoutReader, stdoutWriter := io.Pipe()
	stopped := make(chan struct{})
	var stopOnce sync.Once
	stop := func() {
		stopOnce.Do(func() {
			_ = stdinReader.Close()
			_ = stdoutWriter.Close()
			close(stopped)
		})
	}
	go func() {
		scanner := bufio.NewScanner(stdinReader)
		for scanner.Scan() {
			var req renderRequest
			if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
				return
			}
			data, err := json.Marshal(response(req))
			if err != nil {
				return
			}
			data = append(data, '\n')
			if _, err := stdoutWriter.Write(data); err != nil {
				return
			}
		}
		stop()
	}()
	proc := &rendererProcess{
		stdin:  stdinWriter,
		stdout: stdoutReader,
		wait: func() error {
			<-stopped
			return nil
		},
	}
	r := newRenderer()
	r.start = func() (*rendererProcess, error) { return proc, nil }
	return r, stop
}

func testRequest() renderRequest {
	return renderRequest{ID: "test-id", Component: "card", Width: 1, Height: 1, Scale: 1, Props: json.RawMessage(`{"id":"fixture"}`)}
}

func TestRendererRoutesResponseAndPreservesPNG(t *testing.T) {
	pngBytes := pngFixture(t, color.NRGBA{R: 12, G: 34, B: 56, A: 255})
	r, stop := fakeRenderer(t, func(req renderRequest) renderResponse {
		return renderResponse{ID: req.ID, OK: true, MIME: "image/png", Width: 1, Height: 1, DataBase64: base64.StdEncoding.EncodeToString(pngBytes)}
	})
	defer stop()

	got, err := r.render(context.Background(), testRequest())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, pngBytes) {
		t.Fatal("renderer PNG was re-encoded")
	}
}

func TestRendererRoutesResponseLargerThanOneMiB(t *testing.T) {
	pngBytes := pngFixtureSize(t, 1024, 1024, color.NRGBA{R: 12, G: 34, B: 56, A: 255})
	dataBase64 := base64.StdEncoding.EncodeToString(pngBytes)
	if len(dataBase64) <= 1<<20 {
		t.Fatalf("fixture base64 length = %d, want > 1 MiB", len(dataBase64))
	}
	r, stop := fakeRenderer(t, func(req renderRequest) renderResponse {
		return renderResponse{ID: req.ID, OK: true, MIME: "image/png", Width: 1024, Height: 1024, DataBase64: dataBase64}
	})
	defer stop()

	got, err := r.render(context.Background(), testRequest())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, pngBytes) {
		t.Fatal("renderer PNG was re-encoded")
	}
}

func TestRendererRoutesMultipleResponsesAfterOneMiB(t *testing.T) {
	const responseCount = 80
	pngBytes := pngFixtureSize(t, 64, 64, color.NRGBA{R: 12, G: 34, B: 56, A: 255})
	dataBase64 := base64.StdEncoding.EncodeToString(pngBytes)
	if len(dataBase64)*responseCount <= 1<<20 {
		t.Fatalf("cumulative base64 length = %d, want > 1 MiB", len(dataBase64)*responseCount)
	}
	r, stop := fakeRenderer(t, func(req renderRequest) renderResponse {
		return renderResponse{ID: req.ID, OK: true, MIME: "image/png", Width: 64, Height: 64, DataBase64: dataBase64}
	})
	defer stop()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	results := make(chan error, responseCount)
	for i := 0; i < responseCount; i++ {
		request := testRequest()
		request.ID = "test-id-" + strconv.Itoa(i)
		request.Width = 64
		request.Height = 64
		go func(request renderRequest) {
			_, err := r.render(ctx, request)
			results <- err
		}(request)
	}
	for i := 0; i < responseCount; i++ {
		if err := <-results; err != nil {
			t.Fatalf("response %d failed: %v", i, err)
		}
	}
	r.mu.Lock()
	pending := len(r.pending)
	r.mu.Unlock()
	if pending != 0 {
		t.Fatalf("pending responses = %d, want 0", pending)
	}
}

func TestRendererOversizedResponseRejectsPending(t *testing.T) {
	wait := make(chan struct{})
	defer close(wait)
	r := newRenderer()
	r.start = func() (*rendererProcess, error) {
		return &rendererProcess{
			stdin:  testWriteCloser{io.Discard},
			stdout: io.NopCloser(&repeatedByteReader{remaining: maxResponseBytes + 1}),
			wait:   func() error { <-wait; return nil },
		}, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := r.render(ctx, testRequest())
	if err == nil || (!strings.Contains(err.Error(), "protocol_error") && !strings.Contains(err.Error(), "runner_exit")) {
		t.Fatalf("error = %v, want protocol_error or runner_exit", err)
	}
	r.mu.Lock()
	pending := len(r.pending)
	r.mu.Unlock()
	if pending != 0 {
		t.Fatalf("pending responses = %d after oversized response, want 0", pending)
	}
}

func TestRendererProtocolFailureRejectsPending(t *testing.T) {
	stdin := testWriteCloser{io.Discard}
	stdout := io.NopCloser(strings.NewReader(`{"id":"unknown","ok":true,"mime":"image/png","width":1,"height":1,"dataBase64":"bad"}` + "\n"))
	wait := make(chan struct{})
	defer close(wait)
	r := newRenderer()
	r.start = func() (*rendererProcess, error) {
		return &rendererProcess{stdin: stdin, stdout: stdout, wait: func() error { <-wait; return errors.New("fake exit") }}, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := r.render(ctx, testRequest())
	if err == nil || !strings.Contains(err.Error(), "protocol_error") {
		t.Fatalf("error = %v, want protocol_error", err)
	}
}

func TestRendererShutdownIsIdempotentAndRejectsNewRequests(t *testing.T) {
	r, stop := fakeRenderer(t, func(req renderRequest) renderResponse {
		return renderResponse{ID: req.ID, OK: true, MIME: "image/png", Width: 1, Height: 1, DataBase64: base64.StdEncoding.EncodeToString(pngFixture(t, color.NRGBA{A: 255}))}
	})
	stop()
	r.shutdown()
	r.shutdown()
	_, err := r.render(context.Background(), testRequest())
	if !errors.Is(err, ErrRendererClosed) {
		t.Fatalf("error = %v, want ErrRendererClosed", err)
	}
}

func TestPNGToJPEGUsesWhiteBackground(t *testing.T) {
	pngBytes := pngFixture(t, color.NRGBA{R: 255, A: 0})
	jpegBytes, err := PNGToJPEG(pngBytes)
	if err != nil {
		t.Fatal(err)
	}
	img, err := jpeg.Decode(bytes.NewReader(jpegBytes))
	if err != nil {
		t.Fatal(err)
	}
	r, g, b, _ := img.At(0, 0).RGBA()
	if r < 240<<8 || g < 240<<8 || b < 240<<8 {
		t.Fatalf("transparent pixel was not composited to white: %d,%d,%d", r>>8, g>>8, b>>8)
	}
}

func TestFetchRenderSpecRejectsRemoteAndOversizedDimensions(t *testing.T) {
	remote, err := fetchRenderSpec(context.Background(), "http://example.com:80/render")
	if err == nil || remote.Component != "" {
		t.Fatalf("remote URL result = %#v, error = %v", remote, err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"component":"card","width":5000,"height":1,"props":{}}`)
	}))
	defer server.Close()
	if _, err := fetchRenderSpec(context.Background(), server.URL); err == nil || !strings.Contains(err.Error(), "dimensions") {
		t.Fatalf("oversized spec error = %v", err)
	}
}

func TestScreenshotWaitParameterDoesNotSleepAndJPEGDefault(t *testing.T) {
	pngBytes := pngFixture(t, color.NRGBA{R: 10, G: 20, B: 30, A: 255})
	r, stop := fakeRenderer(t, func(req renderRequest) renderResponse {
		return renderResponse{ID: req.ID, OK: true, MIME: "image/png", Width: 1, Height: 1, DataBase64: base64.StdEncoding.EncodeToString(pngBytes)}
	})
	defer stop()
	old := defaultRenderer
	defaultRenderer = r
	defer func() { defaultRenderer = old }()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"component":"card","width":1,"height":1,"props":{}}`)
	}))
	defer server.Close()
	started := time.Now()
	jpegBytes, err := Screenshot(server.URL, 3600000, 1)
	if err != nil {
		t.Fatal(err)
	}
	if time.Since(started) > time.Second {
		t.Fatalf("waitTime caused a delay: %s", time.Since(started))
	}
	if _, err := jpeg.Decode(bytes.NewReader(jpegBytes)); err != nil {
		t.Fatalf("Screenshot did not return JPEG: %v", err)
	}
}

func TestRendererCommandIncludesNDJSONFlag(t *testing.T) {
	args := rendererCommandArgs("renderer/runner.mjs")
	want := []string{"node", "renderer/runner.mjs", "--ndjson"}
	if len(args) != len(want) {
		t.Fatalf("args = %#v, want %#v", args, want)
	}
	for i := range want {
		if args[i] != want[i] {
			t.Fatalf("args = %#v, want %#v", args, want)
		}
	}
}
