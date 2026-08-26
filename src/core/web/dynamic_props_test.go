package web

import (
	"testing"
	"time"

	"arknights_bot/plugins/skland"
	"arknights_bot/utils/model"
)

func TestBuildCalendarProps(t *testing.T) {
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}
	props := buildCalendarProps(time.Date(2025, time.January, 15, 20, 0, 0, 0, location), []CalendarInfo{
		{Title: "测试活动A", Begin: "2025-01-14", End: "2025-01-15", Close: "2025-01-15"},
		{Title: "测试活动B", Begin: "2025-01-15", End: "2025-01-16"},
	})
	if props.Date != "2025年1月15日" || props.Weekday != "星期三" || len(props.Weeks) != 5 {
		t.Fatalf("unexpected calendar header: %#v", props)
	}
	cell := props.Weeks[2][2]
	if !cell.Today || cell.Date != "2025-01-15" || len(cell.Events) != 3 {
		t.Fatalf("unexpected current day: %#v", cell)
	}
	if props.Resource != "经验书、技能书、碳" || props.Chip != "近卫、特种、辅助、先锋" {
		t.Fatalf("unexpected Wednesday materials: %q / %q", props.Resource, props.Chip)
	}
}

func TestBuildStateProps(t *testing.T) {
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}
	var statistic skland.PlayerStatistic
	statistic.PlayerName = "基线博士"
	statistic.Avatar = "char_002_amiya#1"
	statistic.Ap.Current, statistic.Ap.Max, statistic.Ap.RecoverTs = 95, 135, "1736951640"
	statistic.TowerLower.Current, statistic.TowerLower.Max, statistic.TowerLower.RecoverTs = 3, 6, "1737115200"
	statistic.TowerHigher.Current, statistic.TowerHigher.Max, statistic.TowerHigher.RecoverTs = 4, 8, "1737201601"
	statistic.Reward.Current, statistic.Reward.Max, statistic.Reward.RecoverTs = 1, 3, "1737028801"
	statistic.Training.CharIcon, statistic.Training.LeftSeconds = "https://example.invalid/training.png", "93784"
	props := buildStateProps(&statistic, time.Unix(1787070552, 0).In(location))
	if props.AP.Label != "-2时-56分后恢复" || props.TowerLower.Label != "-578天" || props.TowerHigher.Label != "-578天" || props.Reward.Label != "-579天" {
		t.Fatalf("unexpected state labels: %#v", props)
	}
	if props.Training == nil || props.Training.Label != "26:03:04" || props.Avatar != "https://web.hycdn.cn/arknights/game/assets/char_skin/avatar/char_002_amiya%231.png" {
		t.Fatalf("unexpected state assets: %#v", props)
	}
}

func TestBuildLotteryProps(t *testing.T) {
	props := buildLotteryProps([]model.GroupLotteryDetail{
		{LotteryNumber: 7, UserName: "中奖博士", UserNumber: 100000007, Status: 1},
		{LotteryNumber: 42, UserName: "占位博士", UserNumber: 100000042, Status: 0},
	})
	if len(props.Cells) != 100 || props.Cells[6].State != "winner" || props.Cells[41].State != "selected" || props.Cells[99].State != "empty" {
		t.Fatalf("unexpected lottery cells: %#v", props.Cells)
	}
	if props.Cells[6].UserNumber != "ID:100000007" || props.Cells[41].UserName != "占位博士" {
		t.Fatalf("unexpected lottery details: %#v", props.Cells)
	}
}

func TestBuildGachaProps(t *testing.T) {
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}
	log := GachaLog{
		Name: "基线博士", Total: 10, Star6: 2, Star5: 2, Star4: 3, Star3: 3,
		Avg6: 5, Avg5: 5, Avg4: 3.33, Avg3: 3.33,
		BegTime: 1736856000000, EndTime: 1736942400000,
		PoolCount: []PoolCount{{PoolName: "测试卡池甲", PoolCount: 6}, {PoolName: "测试卡池乙", PoolCount: 4}},
		Chars:     []GachaChar{{CharName: "阿米娅", Avatar: "https://example.invalid/amiya.webp", IsNew: true, PoolName: "测试卡池甲", Ts: 1736856000000}},
		Star6Info: []Star6Info{{Name: "能天使", Avatar: "https://example.invalid/exusiai.webp", PoolName: "测试卡池乙", Count: 5, Ts: 1736942400000}},
	}
	props, err := buildGachaProps(log, time.Unix(1736942400, 0).In(location), gachaTimestampMilliseconds)
	if err != nil {
		t.Fatal(err)
	}
	if props.Period != "2025-01-14 20:00:00——2025-01-15 20:00:00" || props.Today != "2025年01月15日" {
		t.Fatalf("unexpected gacha dates: %#v", props)
	}
	if len(props.Rarities) != 4 || props.Rarities[0].Percent != 20 || props.Chars[0].Date != "2025-01-14" || props.Star6Info[0].Date != "2025-01-15" {
		t.Fatalf("unexpected gacha props: %#v", props)
	}
	if _, err := buildGachaProps(log, time.Now(), "seconds"); err == nil {
		t.Fatal("expected unsupported timestamp unit error")
	}
}
