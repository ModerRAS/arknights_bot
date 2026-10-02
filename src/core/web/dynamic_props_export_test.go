package web

import (
	"encoding/json"
	"os"
	"testing"
	"time"

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

	statistic := legacyStateFixture()

	gacha := legacyGachaFixture()
	gachaProps, err := buildGachaProps(gacha, gachaNow, gachaTimestampMilliseconds)
	if err != nil {
		t.Fatal(err)
	}

	specs := []dynamicRenderSpec{
		{ID: "calendar", Component: "calendar", Width: 1920, Height: 1080, Scale: 1.5, Props: buildCalendarProps(calendarNow, legacyCalendarInfo())},
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
