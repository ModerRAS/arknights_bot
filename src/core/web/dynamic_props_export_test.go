package web

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"arknights_bot/plugins/skland"
	"arknights_bot/utils/model"
)

type dynamicRenderSpec struct {
	ID        string  `json:"id"`
	Component string  `json:"component"`
	Width     int     `json:"width"`
	Height    int     `json:"height"`
	Scale     float64 `json:"scale"`
	Props     any     `json:"props"`
}

func TestExportDynamicRenderSpecs(t *testing.T) {
	path := os.Getenv("ARKNIGHTS_DYNAMIC_SPECS")
	if path == "" {
		t.Skip("set ARKNIGHTS_DYNAMIC_SPECS to export frozen dynamic RenderSpec NDJSON")
	}
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}
	calendarNow := time.Unix(1787070541, 0).In(location)
	stateNow := time.Unix(1787070552, 0).In(location)
	gachaNow := time.Unix(1787070547, 0).In(location)

	var statistic skland.PlayerStatistic
	statistic.PlayerName = "基线博士"
	statistic.Avatar = "char_002_amiya#1"
	statistic.Ap.Current, statistic.Ap.Max, statistic.Ap.RecoverTs = 95, 135, "1736951640"
	statistic.CheckedIn = true
	statistic.TowerLower.Current, statistic.TowerLower.Max, statistic.TowerLower.RecoverTs = 3, 6, "1737115200"
	statistic.TowerHigher.Current, statistic.TowerHigher.Max, statistic.TowerHigher.RecoverTs = 4, 8, "1737201601"
	statistic.Reward.Current, statistic.Reward.Max, statistic.Reward.RecoverTs = 1, 3, "1737028801"
	statistic.Recruitment.Current, statistic.Recruitment.Max = 2, 4
	statistic.Trading.Current, statistic.Trading.Max = 6, 10
	statistic.Manufacture.Current, statistic.Manufacture.Max = 7, 12
	statistic.TiredChars = 2
	statistic.Training.CharIcon = "https://web.hycdn.cn/arknights/game/assets/char_skin/avatar/char_1001_amiya2%232.png"
	statistic.Training.LeftSeconds = "93784"

	gacha := GachaLog{
		Name: "基线博士", Total: 10, Star6: 2, Star5: 2, Star4: 3, Star3: 3,
		Avg6: 5, Avg5: 5, Avg4: 3.33, Avg3: 3.33,
		BegTime: 1736856000000, EndTime: 1736942400000,
		PoolCount: []PoolCount{{PoolName: "测试卡池甲", PoolCount: 6}, {PoolName: "测试卡池乙", PoolCount: 4}},
		Chars: []GachaChar{
			{PoolName: "测试卡池甲", CharName: "阿米娅", Avatar: "https://media.prts.wiki/3/36/%E5%A4%B4%E5%83%8F_%E9%98%BF%E7%B1%B3%E5%A8%85.png?image_process=format,webp/quality,Q_90", IsNew: true, Rarity: 5, Ts: 1736856000000},
			{PoolName: "测试卡池乙", CharName: "能天使", Avatar: "https://media.prts.wiki/a/ad/%E5%A4%B4%E5%83%8F_%E8%83%BD%E5%A4%A9%E4%BD%BF.png?image_process=format,webp/quality,Q_90", IsNew: false, Rarity: 4, Ts: 1736942400000},
		},
		Star6Info: []Star6Info{
			{Name: "阿米娅", Avatar: "https://media.prts.wiki/3/36/%E5%A4%B4%E5%83%8F_%E9%98%BF%E7%B1%B3%E5%A8%85.png?image_process=format,webp/quality,Q_90", Ts: 1736856000000, Count: 5, IsNew: true, PoolName: "测试卡池甲", PoolOrder: 1},
			{Name: "能天使", Avatar: "https://media.prts.wiki/a/ad/%E5%A4%B4%E5%83%8F_%E8%83%BD%E5%A4%A9%E4%BD%BF.png?image_process=format,webp/quality,Q_90", Ts: 1736942400000, Count: 5, IsNew: false, PoolName: "测试卡池乙", PoolOrder: 1},
		},
	}
	gachaProps, err := buildGachaProps(gacha, gachaNow, gachaTimestampMilliseconds)
	if err != nil {
		t.Fatal(err)
	}

	specs := []dynamicRenderSpec{
		{ID: "calendar", Component: "calendar", Width: 1920, Height: 1080, Scale: 1.5, Props: buildCalendarProps(calendarNow, []CalendarInfo{{Title: "测试活动A", Begin: "2025-01-14", End: "2025-01-15"}, {Title: "测试活动B", Begin: "2025-01-15", End: "2025-01-16"}, {Title: "测试活动C", Close: "2025-01-15"}})},
		{ID: "state", Component: "state", Width: 1092, Height: 510, Scale: 1, Props: buildStateProps(&statistic, stateNow)},
		{ID: "lottery", Component: "lottery", Width: 982, Height: 1111, Scale: 1.5, Props: buildLotteryProps([]model.GroupLotteryDetail{{LotteryNumber: 7, UserName: "中奖博士", UserNumber: 100000007, Status: 1}, {LotteryNumber: 42, UserName: "占位博士", UserNumber: 100000042, Status: 0}})},
		{ID: "gacha", Component: "gacha", Width: 1000, Height: 882, Scale: 1.5, Props: gachaProps},
	}

	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	for _, spec := range specs {
		if err := encoder.Encode(spec); err != nil {
			t.Fatal(err)
		}
	}
}
