package web

import (
	"arknights_bot/plugins/account"
	"arknights_bot/plugins/player"
	"arknights_bot/utils/repo"
	"arknights_bot/utils/search"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"strconv"
	"time"
)

type GachaLog struct {
	Name      string      `json:"name"`
	Total     int         `json:"total"`
	PoolCount []PoolCount `json:"poolCount"`
	Star6     int         `json:"star6"`
	Star5     int         `json:"star5"`
	Star4     int         `json:"star4"`
	Star3     int         `json:"star3"`
	Avg6      float64     `json:"avg6"`
	Avg5      float64     `json:"avg5"`
	Avg4      float64     `json:"avg4"`
	Avg3      float64     `json:"avg3"`
	Chars     []GachaChar `json:"chars"`
	Star6Info []Star6Info `json:"Star6Info"`
	BegTime   int64       `json:"begTime"`
	EndTime   int64       `json:"endTime"`
}

type GachaChar struct {
	PoolName string `json:"poolName"`
	CharName string `json:"charName"`
	Avatar   string `json:"avatar"`
	IsNew    bool   `json:"isNew"`
	Rarity   int64  `json:"rarity"`
	Ts       int64  `json:"ts"`
}
type Star6Info struct {
	Name      string `json:"name"`
	Avatar    string `json:"avatar"`
	Ts        int64  `json:"ts"`
	Count     int    `json:"count"`
	IsNew     bool   `json:"isNew"`
	PoolName  string `json:"poolName"`
	PoolOrder int    `json:"poolOrder"`
}

type PoolCount struct {
	PoolName  string `json:"poolName"`
	PoolCount int    `json:"count"`
}

type GachaRenderProps struct {
	Name      string              `json:"name"`
	Total     int                 `json:"total"`
	Period    string              `json:"period"`
	Today     string              `json:"today"`
	Rarities  []GachaRarityProps  `json:"rarities"`
	Averages  []GachaAverageProps `json:"averages"`
	Pools     []PoolCount         `json:"pools"`
	Chars     []GachaEntryProps   `json:"chars"`
	Star6Info []GachaEntryProps   `json:"star6Info"`
}

type GachaRarityProps struct {
	Label   string  `json:"label"`
	Count   int     `json:"count"`
	Percent float64 `json:"percent"`
	Color   string  `json:"color"`
}

type GachaAverageProps struct {
	Label string  `json:"label"`
	Count int     `json:"count"`
	Avg   float64 `json:"avg"`
}

type GachaEntryProps struct {
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
	IsNew  bool   `json:"isNew"`
	Date   string `json:"date"`
	Pool   string `json:"pool"`
	Count  int    `json:"count,omitempty"`
}

const gachaTimestampMilliseconds = "milliseconds"

func buildGachaProps(log GachaLog, now time.Time, timestampUnit string) (GachaRenderProps, error) {
	if timestampUnit != gachaTimestampMilliseconds {
		return GachaRenderProps{}, fmt.Errorf("unsupported gacha timestamp unit %q", timestampUnit)
	}
	props := GachaRenderProps{
		Name:   log.Name,
		Total:  log.Total,
		Period: fmt.Sprintf("%s——%s", gachaTimestamp(log.BegTime, now.Location(), true), gachaTimestamp(log.EndTime, now.Location(), true)),
		Today:  now.Format("2006年01月02日"),
		Pools:  log.PoolCount,
		Rarities: []GachaRarityProps{
			{Label: "6星", Count: log.Star6, Percent: gachaPercent(log.Star6, log.Total), Color: "rgba(244,110,30,1)"},
			{Label: "5星", Count: log.Star5, Percent: gachaPercent(log.Star5, log.Total), Color: "rgba(247,171,55,1)"},
			{Label: "4星", Count: log.Star4, Percent: gachaPercent(log.Star4, log.Total), Color: "rgba(161,53,246,1)"},
			{Label: "3星", Count: log.Star3, Percent: gachaPercent(log.Star3, log.Total), Color: "rgba(109,116,126,1)"},
		},
		Averages: []GachaAverageProps{
			{Label: "6星", Count: log.Star6, Avg: log.Avg6},
			{Label: "5星", Count: log.Star5, Avg: log.Avg5},
			{Label: "4星", Count: log.Star4, Avg: log.Avg4},
			{Label: "3星", Count: log.Star3, Avg: log.Avg3},
		},
	}
	props.Chars = make([]GachaEntryProps, len(log.Chars))
	for i, entry := range log.Chars {
		props.Chars[i] = GachaEntryProps{Name: entry.CharName, Avatar: entry.Avatar, IsNew: entry.IsNew, Date: gachaTimestamp(entry.Ts, now.Location(), false), Pool: entry.PoolName}
	}
	props.Star6Info = make([]GachaEntryProps, len(log.Star6Info))
	for i, entry := range log.Star6Info {
		props.Star6Info[i] = GachaEntryProps{Name: entry.Name, Avatar: entry.Avatar, IsNew: entry.IsNew, Date: gachaTimestamp(entry.Ts, now.Location(), false), Pool: entry.PoolName, Count: entry.Count}
	}
	return props, nil
}

func gachaPercent(count, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(count) * 100 / float64(total)
}

func gachaTimestamp(value int64, location *time.Location, withTime bool) string {
	format := "2006-01-02"
	if withTime {
		format = "2006-01-02 15:04:05"
	}
	return time.UnixMilli(value).In(location).Format(format)
}

func Gacha(r *gin.Engine) {
	r.GET("/gacha", func(c *gin.Context) {
		var gachaLog GachaLog
		var userGacha []player.UserGacha
		var gachaChars []GachaChar
		var star6Info []Star6Info
		var poolCount []PoolCount
		var PoolMap = make(map[string][]player.UserGacha)
		userId, _ := strconv.ParseInt(c.Query("userId"), 10, 64)
		uid := c.Query("uid")
		res := repo.GetUserGacha(userId, uid).Scan(&userGacha)
		if res.Error != nil {
			log.Println(res.Error)
			return
		}

		star6 := 0
		star5 := 0
		star4 := 0
		star3 := 0

		for i := range userGacha {
			var gachaChar GachaChar
			gachaChar.PoolName = userGacha[i].PoolName
			gachaChar.CharName = userGacha[i].CharName
			gachaChar.IsNew = userGacha[i].IsNew
			gachaChar.Rarity = userGacha[i].Rarity
			gachaChar.Ts = userGacha[i].Ts

			c := userGacha[len(userGacha)-i-1]
			operatorList := PoolMap[c.PoolName]
			PoolMap[c.PoolName] = append(operatorList, c)
			switch c.Rarity {
			case 5:
				star6++
				star6Info = append(star6Info, Star6Info{
					Name:      c.CharName,
					Count:     0,
					Ts:        c.Ts,
					IsNew:     c.IsNew,
					PoolName:  c.PoolName,
					PoolOrder: c.PoolOrder,
				})
			case 4:
				star5++
			case 3:
				star4++
			case 2:
				star3++
			}
			gachaChars = append(gachaChars, gachaChar)
		}

		total := len(userGacha)
		gachaLog.Total = total
		gachaLog.Star6 = star6
		gachaLog.Star5 = star5
		gachaLog.Star4 = star4
		gachaLog.Star3 = star3
		gachaLog.Avg6, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", float64(total)/float64(star6)), 64)
		gachaLog.Avg5, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", float64(total)/float64(star5)), 64)
		gachaLog.Avg4, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", float64(total)/float64(star4)), 64)
		gachaLog.Avg3, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", float64(total)/float64(star3)), 64)

		gachaLog.Chars = gachaChars
		if len(gachaChars) > 20 {
			gachaLog.Chars = gachaChars[:20]
		}
		for i := range gachaLog.Chars {
			gachaLog.Chars[i].Avatar = search.GetOperatorByName(gachaLog.Chars[i].CharName).Avatar
		}
		gachaLog.BegTime = userGacha[len(userGacha)-1].Ts
		gachaLog.EndTime = userGacha[0].Ts

		var userPlayer account.UserPlayer
		repo.GetPlayerByUserId(userId, uid).Scan(&userPlayer)
		gachaLog.Name = userPlayer.PlayerName

		count := 1
		for k, v := range PoolMap {
			count = 1
			for i, s := range v {
				if s.Rarity == 5 {
					PoolMap[k][i].Remark = strconv.Itoa(count)
					count = 1
					continue
				}
				count++
			}
		}

		for i, s6 := range star6Info {
			for _, m := range PoolMap[s6.PoolName] {
				if s6.Name == m.CharName && s6.Ts == m.Ts && s6.PoolOrder == m.PoolOrder {
					star6Info[i].Count, _ = strconv.Atoi(m.Remark)
				}
			}
		}
		for i, j := 0, len(star6Info)-1; i < j; i, j = i+1, j-1 {
			star6Info[i], star6Info[j] = star6Info[j], star6Info[i]
		}
		gachaLog.Star6Info = star6Info
		if len(star6Info) > 20 {
			gachaLog.Star6Info = star6Info[:20]
		}
		for i := range gachaLog.Star6Info {
			gachaLog.Star6Info[i].Avatar = search.GetOperatorByName(gachaLog.Star6Info[i].Name).Avatar
		}

		repo.GetUserPoolCount(userId, uid).Scan(&poolCount)
		gachaLog.PoolCount = poolCount
		if len(poolCount) > 10 {
			gachaLog.PoolCount = poolCount[len(poolCount)-10:]
		}

		if props, err := buildGachaProps(gachaLog, time.Now(), gachaTimestampMilliseconds); err != nil {
			renderError(c, err)
		} else {
			RenderSpec(c, "gacha", 1000, 882, props)
		}
	})
}
