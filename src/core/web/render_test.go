package web

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRenderSpecContract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	RenderSpec(c, "card", 1280, 720, map[string]any{"id": "fixture"})

	if recorder.Code != 200 {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	var got map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 4 || got["component"] != "card" || got["width"] != float64(1280) || got["height"] != float64(720) {
		t.Fatalf("unexpected render spec: %#v", got)
	}
	props, ok := got["props"].(map[string]any)
	if !ok || props["id"] != "fixture" {
		t.Fatalf("unexpected props: %#v", got["props"])
	}
}

func TestRenderErrorContract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	RenderError(c, errContractTest{})

	if recorder.Code != 500 || recorder.Header().Get("Content-Type") == "" {
		t.Fatalf("unexpected response: code=%d content-type=%q", recorder.Code, recorder.Header().Get("Content-Type"))
	}
	var got struct {
		Error renderContractError `json:"error"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Error.Code != "render_error" || got.Error.Message != "contract failure" || got.Error.Retryable {
		t.Fatalf("unexpected render error: %#v", got.Error)
	}
}

func TestRenderSpecValidationContract(t *testing.T) {
	cases := []struct {
		name, component, code string
		width, height         int
	}{
		{"empty component", "", "invalid_component", 1280, 720},
		{"uppercase component", "Card", "invalid_component", 1280, 720},
		{"path component", "card/extra", "invalid_component", 1280, 720},
		{"unknown component", "unknown", "unknown_component", 1280, 720},
		{"zero width", "card", "invalid_dimensions", 0, 720},
		{"negative height", "card", "invalid_dimensions", 1280, -1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			RenderSpec(c, tc.component, tc.width, tc.height, nil)
			if recorder.Code != 400 {
				t.Fatalf("status = %d, want 400", recorder.Code)
			}
			var got struct {
				Error renderContractError `json:"error"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if got.Error.Code != tc.code || got.Error.Message == "" || got.Error.Retryable {
				t.Fatalf("unexpected error: %#v", got.Error)
			}
		})
	}
}

type errContractTest struct{}

func (errContractTest) Error() string { return "contract failure" }
