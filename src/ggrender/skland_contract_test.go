package ggrender

// Freezes the skland→ggrender field-mapping diagnosis: which web/skland shapes
// ggrender takes verbatim, which ones diverge, and how many of the 16 frozen
// scenes have a skland feed. Imports neither plugins/skland nor core/web —
// those drag MySQL/Redis/gin/viper in, and the guard only needs the ggrender
// side. Red here means a shape moved: re-run the mapping, don't loosen the assert.

import (
	"reflect"
	"sort"
	"testing"
)

func sklandFieldNames(v any) []string {
	rt := reflect.TypeOf(v)
	names := make([]string, rt.NumField())
	for i := range names {
		names[i] = rt.Field(i).Name
	}
	sort.Strings(names)
	return names
}

func hasNoField(t *testing.T, v any, typ, field string) {
	t.Helper()
	if _, ok := reflect.TypeOf(v).FieldByName(field); ok {
		t.Errorf("%s 现在有 %s 了，skland 映射结论已过期，请复核", typ, field)
	}
}

// TestSklandContractAligns: shapes ggrender takes straight from skland.
func TestSklandContractAligns(t *testing.T) {
	for _, c := range []struct {
		typ  string
		v    any
		want []string
	}{
		{"Char", Char{}, []string{"CharId", "EvolvePhase", "FavorPercent", "Level", "Name", "PotentialRank", "Profession", "Rarity", "SkinId"}},
		{"Detail", Detail{}, []string{"Equips", "EvolvePhase", "Id", "Level", "Name", "PotentialRank", "Rarity", "Skills"}},
		{"Skill", Skill{}, []string{"Id", "Level"}},
		{"Equip", Equip{}, []string{"Id", "Level"}},
		{"BoxSummary", BoxSummary{}, []string{
			"AllCharCnt", "AllEquipStage1Cnt", "AllEquipStage2Cnt", "AllEquipStage3Cnt", "AllEvolvePhase2Cnt", "AllSkill10Cnt", "AllSkill8Cnt", "AllSkill9Cnt", "MissingChars", "Name",
			"Star4CharCnt", "Star4EquipStage1Cnt", "Star4EquipStage2Cnt", "Star4EquipStage3Cnt", "Star4EvolvePhase2Cnt", "Star4Skill10Cnt", "Star4Skill8Cnt", "Star4Skill9Cnt",
			"Star5CharCnt", "Star5EquipStage1Cnt", "Star5EquipStage2Cnt", "Star5EquipStage3Cnt", "Star5EvolvePhase2Cnt", "Star5Skill10Cnt", "Star5Skill8Cnt", "Star5Skill9Cnt",
			"Star6CharCnt", "Star6EquipStage1Cnt", "Star6EquipStage2Cnt", "Star6EquipStage3Cnt", "Star6EvolvePhase2Cnt", "Star6Skill10Cnt", "Star6Skill8Cnt", "Star6Skill9Cnt",
		}},
		{"DepotItem", DepotItem{}, []string{"Count", "Icon", "Name", "SortId"}},
		{"MissingChar", MissingChar{}, []string{"Name", "Profession", "Rarity", "SkinId"}},
		{"Cmd", Cmd{}, []string{"Cmd", "Desc", "IsBind", "Param"}},
	} {
		if got := sklandFieldNames(c.v); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s 字段集变了: got %v want %v（web/skland 侧映射已改动，请复核）", c.typ, got, c.want)
		}
	}
}

// TestSklandContractMismatches: shapes ggrender deliberately diverges from. Each
// asserts the divergence is still present, so the day the web layer catches up
// this goes red instead of drifting silently.
func TestSklandContractMismatches(t *testing.T) {
	// CardInfo: web carries AssistChars/Secretary/EquipOperatorCnt/EquipStage3Cnt/
	// NationList, and its Avatar is a {Type,Id} struct ggrender flattened to string.
	for _, f := range []string{"AssistChars", "Secretary", "EquipOperatorCnt", "EquipStage3Cnt", "NationList"} {
		hasNoField(t, CardInfo{}, "CardInfo", f)
	}
	if f, _ := reflect.TypeOf(CardInfo{}).FieldByName("Avatar"); f.Type.Kind() != reflect.String {
		t.Errorf("CardInfo.Avatar 不再是 string，web 侧 {Type,Id} 结构体的映射结论已过期")
	}

	// BaseInfo: web renames Cur→Current, SLevel→SpecializeLevel, Dorms→Dormitories,
	// and its Control.Chars is []BaseChar where ggrender flattened to []string.
	bi := reflect.TypeOf(BaseInfo{})
	for _, f := range []string{"Current", "SpecializeLevel", "Dormitories"} {
		hasNoField(t, BaseInfo{}, "BaseInfo", f)
	}
	ctl, _ := bi.FieldByName("Control")
	if chars, _ := ctl.Type.FieldByName("Chars"); chars.Type.Elem().Kind() != reflect.String {
		t.Errorf("BaseInfo.Control.Chars 不再是 []string（web 侧是 []BaseChar），丢 Avatar/AP 的结论已过期")
	}

	// HelpData: web uses PrivateCmds/PublicCmds/AdminCmds, ggrender renamed all three.
	for _, f := range []string{"PrivateCmds", "PublicCmds", "AdminCmds"} {
		hasNoField(t, HelpData{}, "HelpData", f)
	}

	// CalendarData: web returns one object carrying Close; ggrender dropped Close
	// and turned the payload into an array.
	hasNoField(t, CalendarData{}, "CalendarData", "Close")
	if e, _ := reflect.TypeOf(CalendarData{}).FieldByName("Entries"); e.Type.Kind() != reflect.Slice {
		t.Errorf("CalendarData.Entries 不再是数组，web 单对象→数组的结论已过期")
	}

	// StateInfo: web calls it Avatar, ggrender renamed it AvatarURL and bolted a
	// RecoverTs onto Recruitment/Trading/Manufacture that skland has no source for.
	hasNoField(t, StateInfo{}, "StateInfo", "Avatar")
	if _, ok := reflect.TypeOf(StateInfo{}).FieldByName("AvatarURL"); !ok {
		t.Error("StateInfo 不再把 Avatar 改名为 AvatarURL，请复核")
	}
	if _, ok := reflect.TypeOf(StateMeter{}).FieldByName("RecoverTs"); !ok {
		t.Error("StateMeter 不再有 RecoverTs（skland 侧无此字段），请复核")
	}
}

// TestSklandContractSceneCoverage: only 8 of the 16 frozen manifest scenes have a
// skland feed. gacha goes through MySQL, calendar through HTTP, headhunt through
// crypto/rand, the rest through plugins or local data. If this count moves, the
// scale-coefficient risk from the diagnosis moves with it.
func TestSklandContractSceneCoverage(t *testing.T) {
	withSklandPath := map[string]bool{
		"base": true, "box": true, "box-detail": true, "box-summary": true,
		"card": true, "depot": true, "missing": true, "state": true,
	}
	scenes := []string{
		"base", "box", "box-detail", "box-summary", "calendar", "card", "depot", "enemy",
		"gacha", "headhunt", "help", "lottery", "missing", "operator", "recruit", "state",
	}
	if len(scenes) != 16 {
		t.Fatalf("场景清单 = %d 个，期望 16", len(scenes))
	}
	known := make(map[string]bool, len(scenes))
	for _, s := range scenes {
		known[s] = true
	}
	for s := range withSklandPath {
		if !known[s] {
			t.Errorf("标记了通路但不在 16 场景清单里: %s", s)
		}
	}
	if len(withSklandPath) != 8 {
		t.Errorf("有 skland 通路的场景 = %d/16，期望 8", len(withSklandPath))
	}
}
