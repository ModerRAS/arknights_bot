package web

import (
	"arknights_bot/utils/model"
	"encoding/json"
	"os"
	"testing"
)

const (
	operatorSkinID = "char_002_amiya#1"
	skillID        = "skcom_magic_rage[3]"
	skillIDTwo     = "skchr_amiya_2"
	skillIDThree   = "skchr_amiya_3"
	moduleTypeIcon = "original"

	operatorAvatarURL = "https://media.prts.wiki/3/36/%E5%A4%B4%E5%83%8F_%E9%98%BF%E7%B1%B3%E5%A8%85.png?image_process=format,webp/quality,Q_90"
	operatorThumbURL  = "https://media.prts.wiki/a/a0/%E5%8D%8A%E8%BA%AB%E5%83%8F_%E9%98%BF%E7%B1%B3%E5%A8%85_1.png?image_process=format,webp/quality,Q_90"
	materialURL       = "https://media.prts.wiki/thumb/6/6a/%E9%81%93%E5%85%B7_%E5%B8%A6%E6%A1%86_%E9%BE%99%E9%97%A8%E5%B8%81.png/75px-%E9%81%93%E5%85%B7_%E5%B8%A6%E6%A1%86_%E9%BE%99%E9%97%A8%E5%B8%81.png"
)

func repeatStaticFixture[T any](count int, value T) []T {
	values := make([]T, count)
	for i := range values {
		values[i] = value
	}
	return values
}

func HelpFixture() HelpCmd {
	return HelpCmd{
		PrivateCmds: []Cmd{
			{Cmd: "/bind", Desc: "绑定角色", Param: "", IsBind: false},
			{Cmd: "/unbind", Desc: "解绑角色", Param: "", IsBind: true},
			{Cmd: "/cancel", Desc: "取消操作", Param: "", IsBind: false},
			{Cmd: "/reset_token", Desc: "重设token", Param: "", IsBind: true},
			{Cmd: "/import_gacha", Desc: "导入抽卡记录", Param: "", IsBind: true},
			{Cmd: "/export_gacha", Desc: "导出抽卡记录", Param: "", IsBind: true},
		},
		PublicCmds: []Cmd{
			{Cmd: "/help", Desc: "使用说明", Param: "", IsBind: false},
			{Cmd: "/ping", Desc: "存活测试", Param: "", IsBind: false},
			{Cmd: "/tag", Desc: "自定义群标签", Param: "标签", IsBind: false},
			{Cmd: "/sign", Desc: "签到", Param: "", IsBind: true},
			{Cmd: "/sign", Desc: "开启自动签到", Param: "auto", IsBind: true},
			{Cmd: "/sign", Desc: "关闭自动签到", Param: "stop", IsBind: true},
			{Cmd: "/sign", Desc: "全部通知", Param: "notify_all", IsBind: true},
			{Cmd: "/sign", Desc: "仅失败时通知", Param: "notify_fail", IsBind: true},
			{Cmd: "/sign", Desc: "仅成功时通知", Param: "notify_success", IsBind: true},
			{Cmd: "/ap", Desc: "开启理智提醒", Param: "on", IsBind: true},
			{Cmd: "/ap", Desc: "关闭理智提醒", Param: "off", IsBind: true},
			{Cmd: "/ap", Desc: "设理智提醒阈值", Param: "thr [1-100]", IsBind: true},
			{Cmd: "/state", Desc: "当前状态", Param: "", IsBind: true},
			{Cmd: "/box", Desc: "我的干员(默认6星)", Param: "", IsBind: true},
			{Cmd: "/box", Desc: "所有干员", Param: "all", IsBind: true},
			{Cmd: "/box", Desc: "对应星级干员", Param: "5,6", IsBind: true},
			{Cmd: "/box_detail", Desc: "干员详情(默认6星)", Param: "", IsBind: true},
			{Cmd: "/box_detail", Desc: "对应星级干员", Param: "5", IsBind: true},
			{Cmd: "/box_summary", Desc: "干员信息汇总", Param: "", IsBind: true},
			{Cmd: "/missing", Desc: "未获取干员(默认6星)", Param: "", IsBind: true},
			{Cmd: "/missing", Desc: "所有未获取干员", Param: "all", IsBind: true},
			{Cmd: "/missing", Desc: "对应星级未获取干员", Param: "5,6", IsBind: true},
			{Cmd: "/card", Desc: "我的名片", Param: "", IsBind: true},
			{Cmd: "/base", Desc: "基建信息", Param: "", IsBind: true},
			{Cmd: "/gacha", Desc: "抽卡记录", Param: "", IsBind: true},
			{Cmd: "/operator", Desc: "干员查询", Param: "", IsBind: false},
			{Cmd: "/skin", Desc: "干员皮肤查询", Param: "", IsBind: false},
			{Cmd: "/enemy", Desc: "敌人查询", Param: "", IsBind: false},
			{Cmd: "/report", Desc: "举报", Param: "", IsBind: false},
			{Cmd: "/quiz", Desc: "云玩家检测", Param: "", IsBind: false},
			{Cmd: "/quiz", Desc: "云玩家检测(困难)", Param: "h", IsBind: false},
			{Cmd: "/redeem", Desc: "CDK兑换", Param: "[CDK]", IsBind: true},
			{Cmd: "/headhunt", Desc: "寻访模拟", Param: "", IsBind: false},
			{Cmd: "/recruit", Desc: "公招计算(图片附带)", Param: "", IsBind: false},
			{Cmd: "/calendar", Desc: "活动日历", Param: "", IsBind: false},
			{Cmd: "/depot", Desc: "我的仓库", Param: "", IsBind: true},
		},
		AdminCmds: []Cmd{
			{Cmd: "/news", Desc: "开启/关闭动态推送", Param: "", IsBind: false},
			{Cmd: "/birthday", Desc: "开启/关闭生日推送", Param: "", IsBind: false},
			{Cmd: "/request_mode", Desc: "切换群验证模式", Param: "", IsBind: false},
			{Cmd: "/quiz", Desc: "开启云玩家检测", Param: "on", IsBind: false},
			{Cmd: "/quiz", Desc: "关闭云玩家检测", Param: "off", IsBind: false},
			{Cmd: "/headhunt", Desc: "开启寻访模拟", Param: "on", IsBind: false},
			{Cmd: "/headhunt", Desc: "关闭寻访模拟", Param: "off", IsBind: false},
			{Cmd: "/reg", Desc: "回复消息设置为群规", Param: "", IsBind: false},
			{Cmd: "/welcome", Desc: "设置入群欢迎信息", Param: "文本", IsBind: false},
		},
	}
}

func BoxFixture() BoxInfo {
	return BoxInfo{
		Name: "测试博士",
		Chars: repeatStaticFixture(11, Char{
			CharId:        "char_002_amiya",
			SkinId:        operatorSkinID,
			Name:          "阿米娅",
			Level:         90,
			EvolvePhase:   2,
			PotentialRank: 5,
			FavorPercent:  100,
			Rarity:        5,
			Profession:    "WARRIOR",
		}),
	}
}

func BoxDetailFixture() []Detail {
	return []Detail{
		{
			Name:          "阿米娅",
			Id:            operatorSkinID,
			Rarity:        5,
			Level:         90,
			EvolvePhase:   2,
			PotentialRank: 5,
			Skills:        []Skill{{Id: skillID, Level: 10}},
			Equips:        []Equip{{Id: moduleTypeIcon, Level: 1}},
		},
		{
			Name:          "阿米娅",
			Id:            operatorSkinID,
			Rarity:        5,
			Level:         80,
			EvolvePhase:   2,
			PotentialRank: 4,
			Skills: []Skill{
				{Id: skillID, Level: 10},
				{Id: skillIDTwo, Level: 9},
				{Id: skillIDThree, Level: 8},
			},
			Equips: []Equip{
				{Id: moduleTypeIcon, Level: 2},
				{Id: moduleTypeIcon, Level: 3},
			},
		},
	}
}

func MissingFixture() MissingInfo {
	return MissingInfo{
		Name: "测试博士",
		Chars: repeatStaticFixture(11, MissingChar{
			SkinId:     operatorThumbURL,
			Name:       "阿米娅",
			Rarity:     5,
			Profession: "WARRIOR",
		}),
	}
}

func BoxSummaryFixture() BoxSummary {
	return BoxSummary{
		Name:                 "测试博士",
		AllCharCnt:           "60/100",
		AllEvolvePhase2Cnt:   40,
		AllSkill10Cnt:        30,
		AllSkill9Cnt:         20,
		AllSkill8Cnt:         10,
		AllEquipStage3Cnt:    12,
		AllEquipStage2Cnt:    8,
		AllEquipStage1Cnt:    4,
		Star6CharCnt:         "20/50",
		Star6EvolvePhase2Cnt: 18,
		Star6Skill10Cnt:      15,
		Star6Skill9Cnt:       10,
		Star6Skill8Cnt:       5,
		Star6EquipStage3Cnt:  6,
		Star6EquipStage2Cnt:  4,
		Star6EquipStage1Cnt:  2,
		Star5CharCnt:         "25/30",
		Star5EvolvePhase2Cnt: 15,
		Star5Skill10Cnt:      10,
		Star5Skill9Cnt:       8,
		Star5Skill8Cnt:       4,
		Star5EquipStage3Cnt:  4,
		Star5EquipStage2Cnt:  3,
		Star5EquipStage1Cnt:  1,
		Star4CharCnt:         "15/20",
		Star4EvolvePhase2Cnt: 7,
		Star4Skill10Cnt:      5,
		Star4Skill9Cnt:       2,
		Star4Skill8Cnt:       1,
		Star4EquipStage3Cnt:  2,
		Star4EquipStage2Cnt:  1,
		Star4EquipStage1Cnt:  1,
		MissingChars: repeatStaticFixture(24, MissingChar{
			SkinId:     operatorAvatarURL,
			Name:       "阿米娅",
			Rarity:     5,
			Profession: "WARRIOR",
		}),
	}
}

func RecruitFixture() []RecruitList {
	operator := model.Operator{
		Name:       "阿米娅",
		Profession: "WARRIOR",
		Rarity:     5,
		Avatar:     operatorAvatarURL,
	}
	return []RecruitList{
		{Tags: []string{"高级资深干员", "输出"}, Operators: repeatStaticFixture(12, operator)},
		{Tags: []string{"术师干员", "远程位"}, Operators: repeatStaticFixture(2, operator)},
	}
}

func HeadhuntFixture() []model.Operator {
	return repeatStaticFixture(10, model.Operator{
		Name:       "阿米娅",
		Profession: "WARRIOR",
		Rarity:     5,
		ThumbURL:   operatorThumbURL,
	})
}

func DepotFixture() []DepotItem {
	return repeatStaticFixture(11, DepotItem{
		Name:   "龙门币",
		Count:  "100000",
		Icon:   materialURL,
		SortId: 1,
	})
}

type staticRenderSpec struct {
	ID        string  `json:"id"`
	Component string  `json:"component"`
	Width     int     `json:"width"`
	Height    int     `json:"height"`
	Scale     float64 `json:"scale"`
	Props     any     `json:"props"`
}

func TestExportStaticVisualSpecs(t *testing.T) {
	output := os.Getenv("VISUAL_SPEC_OUT")
	if output == "" {
		t.Skip("VISUAL_SPEC_OUT is not set")
	}
	f, err := os.Create(output)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	specs := []staticRenderSpec{
		{ID: "help", Component: "help", Width: 660, Height: 1366, Scale: 1.5, Props: HelpFixture()},
		{ID: "box", Component: "box", Width: 700, Height: 357, Scale: 1.5, Props: BoxFixture()},
		{ID: "box-detail", Component: "box-detail", Width: 481, Height: 186, Scale: 1.5, Props: BoxDetailFixture()},
		{ID: "missing", Component: "missing", Width: 700, Height: 357, Scale: 1.5, Props: MissingFixture()},
		{ID: "box-summary", Component: "box-summary", Width: 900, Height: 482, Scale: 1.5, Props: BoxSummaryFixture()},
		{ID: "recruit", Component: "recruit", Width: 900, Height: 356, Scale: 1.5, Props: RecruitFixture()},
		{ID: "headhunt", Component: "headhunt", Width: 1049, Height: 576, Scale: 1, Props: HeadhuntFixture()},
		{ID: "depot", Component: "depot", Width: 850, Height: 156, Scale: 1.5, Props: DepotFixture()},
	}
	encoder := json.NewEncoder(f)
	for _, spec := range specs {
		if err := encoder.Encode(spec); err != nil {
			t.Fatal(err)
		}
	}
}
