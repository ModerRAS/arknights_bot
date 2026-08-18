package web

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	enemyplugin "arknights_bot/plugins/enemy"
	operatorplugin "arknights_bot/plugins/operator"
)

type complexRenderSpec struct {
	ID        string  `json:"id"`
	Component string  `json:"component"`
	Width     int     `json:"width"`
	Height    int     `json:"height"`
	Scale     float64 `json:"scale"`
	Props     any     `json:"props"`
}

func decodeComplexFixture[T any](t *testing.T, payload string) T {
	t.Helper()
	var value T
	if err := json.Unmarshal([]byte(payload), &value); err != nil {
		t.Fatal(err)
	}
	return value
}

// These constructors mirror the approved Temp complex_fixtures.go payloads.
func visualComplexCardFixture(t *testing.T) PlayerCard {
	return decodeComplexFixture[PlayerCard](t, `{
		"name":"冻结博士","uid":"10000001","serverName":"官服","resume":"稳定基线签名","level":120,"regTime":1704067200,"mainStageProgress":"14-21",
		"avatar":"https://fixture-cache.invalid/card/player-avatar.png","secretaryName":"阿米娅","secretaryEnName":"Amiya","secretary":"https://fixture-cache.invalid/card/secretary-painting.png",
		"charCnt":289,"furnitureCnt":100,"skinCnt":88,"equipCnt":46,"equipOperatorCnt":23,"equipStage3Cnt":12,
		"nationList":[{"name":"rhodes","flag":1},{"name":"lungmen","flag":0},{"name":"yan","flag":-1}],
		"assistChars":[
			{"name":"阿米娅","charId":"char_002_amiya","skinId":"char_002_amiya#1","level":90,"evolvePhase":2,"potentialRank":5,"skillId":"skcom_amiya","mainSkillLvl":10,"specializeLevel":3,"isSpecMax":true,"equip":{"id":"uniequip_002_amiya","level":3,"name":"MIND OVER MATTER","typeIcon":"uniequip","shiningColor":"#fff"}},
			{"name":"能天使","charId":"char_103_angel","skinId":"char_103_angel#1","level":90,"evolvePhase":2,"potentialRank":1,"skillId":"skchr_angel_3","mainSkillLvl":7,"specializeLevel":0,"isSpecMax":false,"equip":{"id":"","level":0,"name":"","typeIcon":"","shiningColor":""}},
			{"name":"德克萨斯","charId":"char_102_texas","skinId":"char_102_texas#1","level":80,"evolvePhase":2,"potentialRank":3,"skillId":"skchr_texas_2","mainSkillLvl":10,"specializeLevel":3,"isSpecMax":true,"equip":{"id":"uniequip_102_texas","level":3,"name":"EXCALIBUR","typeIcon":"uniequip","shiningColor":"#fff"}}
		]
	}`)
}

func visualComplexBaseFixture(t *testing.T) PlayerBase {
	return decodeComplexFixture[PlayerBase](t, `{
		"labor":{"current":42,"total":99},
		"control":{"level":5,"chars":[{"name":"阿米娅","avatar":"char_002_amiya#1","AP":0}]},
		"dormitories":[{"level":5,"comfort":20000,"chars":[{"name":"能天使","avatar":"char_103_angel#1","AP":50}]}],
		"tradings":[{"level":3,"strategy":"贵金属订单","current":3,"total":3,"chars":[{"name":"德克萨斯","avatar":"char_102_texas#1","AP":100}]}],
		"manufactures":[{"level":3,"current":1,"total":3,"item":"赤金","speed":"130%","chars":[{"name":"推进之王","avatar":"char_112_siege#1","AP":0}]}],
		"powers":[{"level":3,"power":270,"chars":[{"name":"塞雷娅","avatar":"char_202_demkni#1","AP":50}]}],
		"meeting":{"level":3,"board":[1,7],"sharing":true,"chars":[{"name":"凯尔希","avatar":"char_003_kalts#1","AP":100}]},
		"hire":{"level":3,"refreshCount":3,"chars":[{"name":"银灰","avatar":"char_172_svrash#1","AP":0}]},
		"training":{"level":3,"skill":"skchr_amiya_3","specializeLevel":10,"chars":[{"name":"艾雅法拉","avatar":"char_180_amgoat#1","AP":50}]}
	}`)
}

func visualComplexOperatorFixture(t *testing.T) operatorplugin.Operator {
	return decodeComplexFixture[operatorplugin.Operator](t, `{
		"painting":"https://fixture-cache.invalid/operator/amiya-painting.png","attackRange":"<span>□</span>",
		"op":{"name":"阿米娅","nameEn":"Amiya","code":"R001","profession":"CASTER","rarity":5,"hp":"1742","atk":"699","def":"121","res":"10","interval":"1.6s","reDeploy":"70s","cost":"18","block":"1","logo":"罗德岛","tags":"远程位 输出","skins":[{"name":"默认服装","url":"https://fixture-cache.invalid/operator/amiya-default.png"},{"name":"见习联结者","url":"https://fixture-cache.invalid/operator/amiya-skin.png"}]},
		"professionBranch":{"name":"中坚术师","pic":"https://fixture-cache.invalid/operator/profession-branch.png","desc":"攻击造成法术伤害"},
		"potentials":[{"rank":1,"desc":"部署费用-1"},{"rank":2,"desc":"再部署时间-4秒"}],
		"talents":[{"evolve":"精英化1","name":"<b>精神融合</b>","desc":"攻击力 <i>+10%</i>"}],
		"buildingSkills":[{"evolve":"精英化0","icon":"https://fixture-cache.invalid/operator/building-skill-icon.png","name":"合作协议","desc":"控制中枢内线索搜集速度提升"}],
		"skills":[{"icon":"https://fixture-cache.invalid/operator/skill-icon.png","name":"战术咏唱","desc":"<span>攻击力提升</span>","skillRange":"<span>□</span>","spType":["<em>自动回复</em>"],"spInit":"0","spCost":"45","duration":"30"}]
	}`)
}

func visualComplexEnemyFixture(t *testing.T) enemyplugin.Enemy {
	return decodeComplexFixture[enemyplugin.Enemy](t, `{
		"name":"源石虫","pic":"https://media.prts.wiki/3/3e/%E5%A4%B4%E5%83%8F_%E6%95%8C%E4%BA%BA_%E6%BA%90%E7%9F%B3%E8%99%AB.png","desc":"拥有较高防御力的感染生物。","enemyRace":"感染生物","enemyLevel":"普通","attackType":"近战","motion":"地面","ability":"<span>免疫沉默</span>",
		"levels":[
			{"desc":"标准个体","attackType":"近战","motion":"地面","hpRecovery":"0","hp":"5000","atk":"400","def":"500","res":"20","ATKRadius":"1.1","weight":"2","moveSpeed":"0.8","interval":"1.5","damageRes":"0","elementRes":"0","ridicule":"0","point":"1","abnormal":"无","skills":[{"name":"啃噬","spInit":"<span>0</span>","spCost":"<span>5</span>","desc":"对目标造成 <i>物理伤害</i>"}],"talent":"<b>生命低于50%时防御提升</b>"},
			{"desc":"强化个体","attackType":"近战","motion":"地面","hpRecovery":"0","hp":"8000","atk":"600","def":"700","res":"30","ATKRadius":"1.1","weight":"3","moveSpeed":"0.8","interval":"1.5","damageRes":"10","elementRes":"0","ridicule":"1","point":"2","abnormal":"晕眩抗性","skills":[{"name":"啃噬+","spInit":"<span>0</span>","spCost":"<span>4</span>","desc":"造成更高 <i>物理伤害</i>"}],"talent":"<b>攻击力提升</b>"}
		]
	}`)
}

func complexRenderSpecs(t *testing.T) []complexRenderSpec {
	t.Helper()
	card := visualComplexCardFixture(t)
	card.RegisteredOn = registeredOn(card.RegTime)
	return []complexRenderSpec{
		{ID: "card", Component: "card", Width: 1280, Height: 720, Scale: 1, Props: card},
		{ID: "base", Component: "base", Width: 1110, Height: 612, Scale: 1.5, Props: visualComplexBaseFixture(t)},
		{ID: "operator", Component: "operator", Width: 1200, Height: 800, Scale: 1.5, Props: operatorProps(visualComplexOperatorFixture(t))},
		{ID: "enemy", Component: "enemy", Width: 656, Height: 318, Scale: 1.5, Props: enemyProps(visualComplexEnemyFixture(t))},
	}
}

func TestComplexRenderSpecFixtures(t *testing.T) {
	specs := complexRenderSpecs(t)
	if len(specs) != 4 || specs[0].Props.(PlayerCard).RegisteredOn != "2024-01-01" {
		t.Fatalf("unexpected card fixture: %#v", specs[0])
	}
	if base := specs[1].Props.(PlayerBase); len(base.Tradings) != 1 || base.Tradings[0].Chars[0].AP != 100 {
		t.Fatalf("unexpected base fixture: %#v", base)
	}
	operator := specs[2].Props.(operatorplugin.Operator)
	if len(operator.Skills) != 1 {
		t.Fatalf("operator skills = %#v, want one", operator.Skills)
	}
	if strings.Contains(string(operator.AttackRange), "<") || strings.Contains(string(operator.Skills[0].Desc), "<") {
		t.Fatalf("operator props retained template markup: %#v", operator)
	}
	enemy := specs[3].Props.(enemyplugin.Enemy)
	if len(enemy.Levels) != 0 || strings.Contains(string(enemy.Ability), "<") {
		t.Fatalf("unexpected frozen enemy props: %#v", enemy)
	}
}

func TestExportComplexRenderSpecs(t *testing.T) {
	path := os.Getenv("VISUAL_SPEC_OUT")
	if path == "" {
		t.Skip("set VISUAL_SPEC_OUT to export complex RenderSpec NDJSON")
	}
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	for _, spec := range complexRenderSpecs(t) {
		if err := encoder.Encode(spec); err != nil {
			t.Fatal(err)
		}
	}
}

func TestEnemyPropsRichFixtureIsNotBaselineExport(t *testing.T) {
	rich := decodeComplexFixture[enemyplugin.Enemy](t, `{
		"name":"源石虫","pic":"https://fixture-cache.invalid/enemy/originium-slug.png","ability":"<span>免疫沉默</span>",
		"level":[
			{"desc":"标准个体","attackType":"近战","motion":"地面","hpRecovery":"0","hp":"5000","atk":"400","def":"500","res":"20","ATKRadius":"1.1","weight":"2","moveSpeed":"0.8","interval":"1.5","damageRes":"0","elementRes":"0","ridicule":"0","point":"1","abnormal":"无","skills":[{"name":"啃噬","spInit":"<span>0</span>","spCost":"<span>5</span>","desc":"对目标造成 <i>物理伤害</i>"}],"talent":"<b>生命低于50%时防御提升</b>"},
			{"desc":"强化个体","attackType":"近战","motion":"地面","hpRecovery":"0","hp":"8000","atk":"600","def":"700","res":"30","ATKRadius":"1.1","weight":"3","moveSpeed":"0.8","interval":"1.5","damageRes":"10","elementRes":"0","ridicule":"1","point":"2","abnormal":"晕眩抗性","skills":[{"name":"啃噬+","spInit":"<span>0</span>","spCost":"<span>4</span>","desc":"造成更高 <i>物理伤害</i>"}],"talent":"<b>攻击力提升</b>"}
		]
	}`)
	props := enemyProps(rich)
	if len(props.Levels) != 2 || len(props.Levels[0].Skills) != 1 {
		t.Fatalf("unexpected rich enemy shape: %#v", props)
	}
	if strings.Contains(string(props.Levels[0].Talent), "<") || strings.Contains(string(props.Levels[1].Skills[0].Desc), "<") {
		t.Fatalf("rich enemy props retained template markup: %#v", props)
	}
}
