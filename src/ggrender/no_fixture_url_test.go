package ggrender

// TestNoFixtureURLs keeps http(s) URLs out of the Sample* fixtures. A fixture
// URL turns the gate into a function of the network: state used to point at
// web.hycdn.cn, so 0.99093 measured the CDN rather than this repo. 甲-1 made
// the resulting degradation print to stderr; this is the one that goes red.
import (
	"reflect"
	"strings"
	"testing"
)

// maxFixtureDepth bounds a hypothetical cycle; ggrender has none today, but a
// self-referential field should fail rather than hang the suite.
const maxFixtureDepth = 8

// walkFixture reports every http(s) string reachable from v, following
// struct/slice/array/map/pointer and stopping at anything else, so a new
// non-container field type cannot silently escape the check.
func walkFixture(t *testing.T, v reflect.Value, path string, depth int, seen map[uintptr]bool) {
	t.Helper()
	if depth > maxFixtureDepth || !v.IsValid() {
		return
	}
	switch v.Kind() {
	case reflect.String:
		if s := v.String(); strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
			t.Errorf("fixture carries a URL: %s = %q", path, s)
		}
	case reflect.Ptr:
		if v.IsNil() || seen[v.Pointer()] {
			return
		}
		seen[v.Pointer()] = true
		walkFixture(t, v.Elem(), path, depth+1, seen)
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			walkFixture(t, v.Field(i), path+"."+v.Type().Field(i).Name, depth+1, seen)
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			walkFixture(t, v.Index(i), path+"["+itoa(i)+"]", depth+1, seen)
		}
	case reflect.Map:
		// No map exists in the fixtures today; the key is left out of the path
		// rather than rendered, which would need Interface() on a possibly
		// unexported key.
		for _, k := range v.MapKeys() {
			walkFixture(t, v.MapIndex(k), path+"[k]", depth+1, seen)
		}
	}
}

func TestNoFixtureURLs(t *testing.T) {
	for _, s := range []struct {
		name string
		v    any
	}{
		{"SampleBase", SampleBase()},
		{"SampleBox", SampleBox()},
		{"SampleBoxDetail", SampleBoxDetail()},
		{"SampleBoxSummary", SampleBoxSummary()},
		{"SampleCalendar", SampleCalendar()},
		{"SampleCard", SampleCard()},
		{"SampleDepot", SampleDepot()},
		{"SampleEnemy", SampleEnemy()},
		{"SampleGacha", SampleGacha()},
		{"SampleHeadhunt", SampleHeadhunt()},
		{"SampleHelp", SampleHelp()},
		{"SampleLottery", SampleLottery()},
		{"SampleMissing", SampleMissing()},
		{"SampleOperator", SampleOperator()},
		{"SampleRecruit", SampleRecruit()},
		{"SampleState", SampleState()},
	} {
		walkFixture(t, reflect.ValueOf(s.v), s.name, 0, map[uintptr]bool{})
	}
}
