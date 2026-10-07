package web

import (
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"arknights_bot/plugins/skland"
	"arknights_bot/utils/model"
)

// legacyPage is one legacy-template fixture route: the route the Playwright
// capture client requests, the frozen template it renders, and the fixture data
// that template receives. The legacy handlers were recovered from 8873add^,
// where every route ended in c.HTML(http.StatusOK, "<Name>.tmpl", <rootType>);
// the root types are recorded per scene in baseline manifest.json "rootType".
type legacyPage struct {
	route    string
	template string
	props    any
}

func legacyPages(t *testing.T) []legacyPage {
	t.Helper()
	card := visualComplexCardFixture(t)
	card.RegisteredOn = registeredOn(card.RegTime)
	return []legacyPage{
		{"/base", "Base.tmpl", visualComplexBaseFixture(t)},
		{"/box", "Box.tmpl", BoxFixture()},
		{"/boxDetail", "BoxDetail.tmpl", BoxDetailFixture()},
		{"/boxSummary", "BoxSummary.tmpl", BoxSummaryFixture()},
		{"/calendar", "Calendar.tmpl", legacyCalendarHolidayMap()},
		{"/card", "Card.tmpl", card},
		{"/depot", "Depot.tmpl", DepotFixture()},
		{"/enemy", "Enemy.tmpl", visualComplexEnemyFixture(t)},
		{"/gacha", "Gacha.tmpl", legacyGachaFixture()},
		{"/headhunt", "Headhunt.tmpl", HeadhuntFixture()},
		{"/help", "Help.tmpl", HelpFixture()},
		{"/lotteryDetail", "Lottery.tmpl", legacyLotteryFixture()},
		{"/missing", "Missing.tmpl", MissingFixture()},
		{"/operator", "Operator.tmpl", visualComplexOperatorFixture(t)},
		{"/recruit", "Recruit.tmpl", RecruitFixture()},
		{"/state", "State.tmpl", legacyStateFixture()},
	}
}

func legacyLotteryFixture() []model.GroupLotteryDetail {
	return []model.GroupLotteryDetail{
		{LotteryNumber: 7, UserName: "中奖博士", UserNumber: 100000007, Status: 1},
		{LotteryNumber: 42, UserName: "占位博士", UserNumber: 100000042, Status: 0},
	}
}

func legacyStateFixture() skland.PlayerStatistic {
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
	return statistic
}

func legacyGachaFixture() GachaLog {
	return GachaLog{
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
}

func legacyCalendarInfo() []CalendarInfo {
	return []CalendarInfo{
		{Title: "测试活动A", Begin: "2025-01-14", End: "2025-01-15"},
		{Title: "测试活动B", Begin: "2025-01-15", End: "2025-01-16"},
		{Title: "测试活动C", Close: "2025-01-15"},
	}
}

// legacyCalendarHolidayMap reproduces the legacy /calendar handler verbatim,
// including its "关闭关卡" spacing asymmetry: the append branch emits a space
// before the title, the first-write branch does not.
func legacyCalendarHolidayMap() map[string]template.HTML {
	holiday := make(map[string]template.HTML)
	for _, info := range legacyCalendarInfo() {
		title := info.Title
		if _, has := holiday[info.Begin]; has {
			holiday[info.Begin] = template.HTML(fmt.Sprintf("%s<li>开始 %s</li>", holiday[info.Begin], title))
		} else {
			holiday[info.Begin] = template.HTML("<li>开始 " + title + "</li>")
		}
		if _, has := holiday[info.End]; has {
			holiday[info.End] = template.HTML(fmt.Sprintf("%s<li>结束 %s</li>", holiday[info.End], title))
		} else {
			holiday[info.End] = template.HTML("<li>结束 " + title + "</li>")
		}
		if info.Close != "" {
			if _, has := holiday[info.Close]; has {
				holiday[info.Close] = template.HTML(fmt.Sprintf("%s<li>关闭关卡 %s</li>", holiday[info.Close], title))
			} else {
				holiday[info.Close] = template.HTML("<li>关闭关卡" + title + "</li>")
			}
		}
	}
	return holiday
}

// TestLegacyFixtureService serves the frozen templates with isolated fixture
// data on 127.0.0.1:38126 — the service the Playwright capture client expects.
// It runs until <out>/fixture-service.stop appears so the capture run owns the
// process lifetime. Paths resolve relative to the repository root, never to an
// absolute or cross-worktree location.
func TestLegacyFixtureService(t *testing.T) {
	out := os.Getenv("LEGACY_FIXTURE_OUT")
	if out == "" {
		t.Skip("set LEGACY_FIXTURE_OUT to serve the legacy-template fixture service")
	}
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Static("/assets", filepath.Join(root, "assets"))
	router.Static("/template/js", filepath.Join(root, "template", "js"))
	// One parse for the whole set: gin's LoadHTMLFiles replaces the render
	// instance per call, so per-route loading would leave only the last template.
	router.LoadHTMLGlob(filepath.Join(root, "template", "*.tmpl"))
	for _, page := range legacyPages(t) {
		router.GET(page.route, func(c *gin.Context) {
			c.HTML(http.StatusOK, page.template, page.props)
		})
	}
	addr := os.Getenv("FIXTURE_ADDR")
	if addr == "" {
		addr = "127.0.0.1:38126"
	}
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	server := &http.Server{Handler: router}
	go func() { _ = server.Serve(listener) }()
	if err := os.WriteFile(filepath.Join(out, "fixture-service.ready"), []byte(addr+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	stop := filepath.Join(out, "fixture-service.stop")
	for {
		if _, err := os.Stat(stop); err == nil {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	_ = server.Close()
}
