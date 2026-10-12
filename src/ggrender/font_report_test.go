package ggrender

import (
	"os"
	"strings"
	"testing"

	"github.com/fogleman/gg"
)

// TestLoadDefaultFontReportsMissingFont pins the only check for the missing-font
// report: a total LoadDefaultFont failure must name the tried paths instead of
// rendering silently without a font.
func TestLoadDefaultFontReportsMissingFont(t *testing.T) {
	original := FontCandidates
	FontCandidates = []string{"assets/font/does-not-exist.ttf"}
	defer func() { FontCandidates = original }()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	origStderr := os.Stderr
	os.Stderr = w
	loadErr := LoadDefaultFont(gg.NewContext(10, 10), 12)
	w.Close()
	os.Stderr = origStderr

	if loadErr == nil {
		t.Fatal("LoadDefaultFont returned nil for a missing font file")
	}
	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	got := string(buf[:n])
	if !strings.Contains(got, "does-not-exist.ttf") {
		t.Fatalf("missing-font report does not name the path: %q", got)
	}
	t.Logf("missing-font report: %s", strings.TrimSpace(got))
}
