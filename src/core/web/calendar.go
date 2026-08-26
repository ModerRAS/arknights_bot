package web

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/tidwall/gjson"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type CalendarRenderProps struct {
	Date     string               `json:"date"`
	Weekday  string               `json:"weekday"`
	Resource string               `json:"resource"`
	Chip     string               `json:"chip"`
	Weekdays []string             `json:"weekdays"`
	Weeks    [][]CalendarDayProps `json:"weeks"`
}

type CalendarDayProps struct {
	Date           string   `json:"date"`
	Day            int      `json:"day"`
	InCurrentMonth bool     `json:"inCurrentMonth"`
	Weekend        bool     `json:"weekend"`
	Today          bool     `json:"today"`
	Events         []string `json:"events"`
}

func buildCalendarProps(now time.Time, events []CalendarInfo) CalendarRenderProps {
	location := now.Location()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, location)
	first := monthStart.AddDate(0, 0, -((int(monthStart.Weekday()) + 6) % 7))
	rows := 5
	if monthStart.AddDate(0, 1, 0).AddDate(0, 0, -1).Day()+((int(monthStart.Weekday())+6)%7) > 35 {
		rows = 6
	}

	labels := make(map[string][]string)
	for _, event := range events {
		labels[event.Begin] = append(labels[event.Begin], "开始 "+event.Title)
		labels[event.End] = append(labels[event.End], "结束 "+event.Title)
		if event.Close != "" {
			labels[event.Close] = append(labels[event.Close], "关闭关卡 "+event.Title)
		}
	}

	resource, chip := calendarDailyMaterials(now.Weekday())
	props := CalendarRenderProps{
		Date:     now.Format("2006年1月2日"),
		Weekday:  []string{"星期日", "星期一", "星期二", "星期三", "星期四", "星期五", "星期六"}[now.Weekday()],
		Resource: resource,
		Chip:     chip,
		Weekdays: []string{"周一", "周二", "周三", "周四", "周五", "周六", "周日"},
		Weeks:    make([][]CalendarDayProps, rows),
	}
	for row := range props.Weeks {
		props.Weeks[row] = make([]CalendarDayProps, 7)
		for column := range props.Weeks[row] {
			date := first.AddDate(0, 0, row*7+column)
			key := date.Format("2006-01-02")
			props.Weeks[row][column] = CalendarDayProps{
				Date:           key,
				Day:            date.Day(),
				InCurrentMonth: date.Month() == now.Month(),
				Weekend:        date.Weekday() == time.Saturday || date.Weekday() == time.Sunday,
				Today:          date.Year() == now.Year() && date.YearDay() == now.YearDay(),
				Events:         labels[key],
			}
		}
	}
	return props
}

func calendarDailyMaterials(day time.Weekday) (string, string) {
	switch day {
	case time.Sunday:
		return "经验书、技能书、钱、红票、", "近卫、特种、医疗、重装、辅助、先锋"
	case time.Monday:
		return "经验书、红票、碳", "术士、狙击、医疗、重装"
	case time.Tuesday:
		return "经验书、技能书、钱", "术士、狙击、近卫、特种"
	case time.Wednesday:
		return "经验书、技能书、碳", "近卫、特种、辅助、先锋"
	case time.Thursday:
		return "经验书、钱、红票", "医疗、重装、辅助、先锋"
	case time.Friday:
		return "经验书、技能书、碳", "术士、狙击、医疗、重装"
	default:
		return "经验书、钱、红票", "术士、狙击、近卫、特种、辅助、先锋"
	}
}

type CalendarInfo struct {
	Title string `json:"title"`
	Begin string `json:"begin"`
	Close string `json:"close"`
	End   string `json:"end"`
}

func Calendar(r *gin.Engine) {
	r.GET("/calendar", func(c *gin.Context) {
		resp, err := http.Get(viper.GetString("api.calendar"))
		if err != nil {
			log.Println(err)
			renderError(c, err)
			return
		}
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Println(err)
			renderError(c, err)
			return
		}
		text := doc.Text()
		begin := strings.Index(text, "{") + 1
		end := strings.Index(text, "}")
		reg := regexp.MustCompile("(\\[).*?(])")
		aaa := reg.FindAllStringSubmatch(text[begin:end], -1)
		var calendarInfo []CalendarInfo
		for _, a := range aaa {
			j := gjson.Parse(a[0])
			var c CalendarInfo
			c.Title = j.Get("0").String()
			c.Begin = fmt.Sprintf("%d-%02d-%02d", j.Get("1").Int(), j.Get("2").Int(), j.Get("3").Int())
			c.End = fmt.Sprintf("%d-%02d-%02d", j.Get("4").Int(), j.Get("5").Int(), j.Get("6").Int())
			if len(j.Array()) > 7 {
				c.Close = fmt.Sprintf("%d-%02d-%02d", j.Get("7").Int(), j.Get("8").Int(), j.Get("9").Int())
			}
			calendarInfo = append(calendarInfo, c)
		}
		defer resp.Body.Close()
		RenderSpec(c, "calendar", 1920, 1080, buildCalendarProps(time.Now(), calendarInfo))
	})
}
