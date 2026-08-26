package web

import (
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
)

var canonicalRenderComponent = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

var renderComponents = map[string]struct{}{
	"base": {}, "box": {}, "box-detail": {}, "box-summary": {},
	"calendar": {}, "card": {}, "depot": {}, "enemy": {},
	"gacha": {}, "headhunt": {}, "help": {}, "lottery": {},
	"missing": {}, "operator": {}, "recruit": {}, "state": {},
}

type renderContractError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Retryable bool   `json:"retryable"`
}

func writeRenderError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"error": renderContractError{Code: code, Message: message}})
}

// RenderError writes the renderer-facing error contract.
func RenderError(c *gin.Context, err error) {
	message := "unknown render error"
	if err != nil {
		message = err.Error()
	}
	writeRenderError(c, http.StatusInternalServerError, "render_error", message)
}

// RenderSpec writes the renderer's stable, intentionally small input contract.
func RenderSpec(c *gin.Context, component string, width, height int, props any) {
	if !canonicalRenderComponent.MatchString(component) {
		writeRenderError(c, http.StatusBadRequest, "invalid_component", "component must be a canonical lowercase name")
		return
	}
	if _, ok := renderComponents[component]; !ok {
		writeRenderError(c, http.StatusBadRequest, "unknown_component", "component is not registered")
		return
	}
	if width <= 0 || height <= 0 {
		writeRenderError(c, http.StatusBadRequest, "invalid_dimensions", "width and height must be positive")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"component": component,
		"width":     width,
		"height":    height,
		"props":     props,
	})
}
