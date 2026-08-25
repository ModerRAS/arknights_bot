package ggrender

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/fogleman/gg"
)

// Scenes canonical 16 must stay exact; harness fails if changed.
var Scenes = []string{
	"base", "box", "box-detail", "box-summary",
	"calendar", "card", "depot", "enemy",
	"gacha", "headhunt", "help", "lottery",
	"missing", "operator", "recruit", "state",
}

// SceneSet for validation.
var SceneSet = func() map[string]struct{} {
	m := make(map[string]struct{}, len(Scenes))
	for _, s := range Scenes {
		m[s] = struct{}{}
	}
	return m
}()

// RenderGG unified production renderer: scene -> image.
// data may be nil -> uses frozen Sample fixture for deterministic tests.
// ponytail: single dispatch, no extra abstraction.
func RenderGG(scene string, data interface{}) (image.Image, error) {
	dc, err := renderContext(scene, data)
	if err != nil {
		return nil, err
	}
	return dc.Image(), nil
}

// RenderGGContext returns gg Context directly (for EncodePNG convenience).
func RenderGGContext(scene string, data interface{}) (*gg.Context, error) {
	return renderContext(scene, data)
}

func normalizeScene(s string) string {
	// accepts "BoxDetail", "boxDetail", "box_detail", "box-detail", "box" etc -> canonical hyphen lower
	s = string([]rune(s))
	// simple lower
	lower := ""
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			r = r - 'A' + 'a'
		}
		lower += string(r)
	}
	// replace '_' with '-'
	rep := ""
	for _, r := range lower {
		if r == '_' {
			rep += "-"
		} else {
			rep += string(r)
		}
	}
	// handle camelCase boxDetail -> box-detail via known aliases
	alias := map[string]string{
		"boxdetail": "box-detail", "boxsummary": "box-summary", "lotterydetail": "lottery",
	}
	if v, ok := alias[rep]; ok {
		return v
	}
	// also handle "box-detail" already
	return rep
}

func renderContext(scene string, data interface{}) (*gg.Context, error) {
	scene = normalizeScene(scene)
	switch scene {
	case "base":
		if d, ok := data.(*BaseInfo); ok && d != nil {
			return RenderBase(d)
		}
		return RenderBase(SampleBase())
	case "box":
		if d, ok := data.(*BoxInfo); ok && d != nil {
			return RenderBox(d)
		}
		return RenderBox(SampleBox())
	case "box-detail":
		if d, ok := data.(*BoxDetailList); ok && d != nil {
			return RenderBoxDetail(d.Items)
		}
		if arr, ok := data.([]Detail); ok {
			return RenderBoxDetail(arr)
		}
		return RenderBoxDetail(SampleBoxDetail())
	case "box-summary":
		if d, ok := data.(*BoxSummary); ok && d != nil {
			return RenderBoxSummary(d)
		}
		return RenderBoxSummary(SampleBoxSummary())
	case "calendar":
		if d, ok := data.(*CalendarData); ok && d != nil {
			return RenderCalendar(d)
		}
		return RenderCalendar(SampleCalendar())
	case "card":
		if d, ok := data.(*CardInfo); ok && d != nil {
			return RenderCard(d)
		}
		return RenderCard(SampleCard())
	case "depot":
		if d, ok := data.(*DepotData); ok && d != nil {
			return RenderDepot(d)
		}
		return RenderDepot(SampleDepot())
	case "enemy":
		if d, ok := data.(*Enemy); ok && d != nil {
			return RenderEnemy(d)
		}
		return RenderEnemy(SampleEnemy())
	case "gacha":
		if d, ok := data.(*GachaData); ok && d != nil {
			return RenderGacha(d)
		}
		return RenderGacha(SampleGacha())
	case "headhunt":
		if d, ok := data.(*HeadhuntData); ok && d != nil {
			return RenderHeadhunt(d.Ops)
		}
		if arr, ok := data.([]HHOp); ok {
			return RenderHeadhunt(arr)
		}
		return RenderHeadhunt(SampleHeadhunt())
	case "help":
		if d, ok := data.(*HelpData); ok && d != nil {
			return RenderHelp(d)
		}
		return RenderHelp(SampleHelp())
	case "lottery":
		if d, ok := data.(*LotteryData); ok && d != nil {
			return RenderLottery(d)
		}
		return RenderLottery(SampleLottery())
	case "missing":
		if d, ok := data.(*MissingInfo); ok && d != nil {
			return RenderMissing(d)
		}
		return RenderMissing(SampleMissing())
	case "operator":
		if d, ok := data.(*OperatorInfo); ok && d != nil {
			return RenderOperator(d)
		}
		return RenderOperator(SampleOperator())
	case "recruit":
		return RenderRecruit(SampleRecruit())
	case "state":
		if d, ok := data.(*StateInfo); ok && d != nil {
			return RenderState(d)
		}
		return RenderState(SampleState())
	default:
		return nil, fmt.Errorf("unknown scene %q", scene)
	}
}

// ---------- shared sample/fixture types ----------

// Char mirrors Box.
var _ = math.Pi
var _ = color.RGBA{}
