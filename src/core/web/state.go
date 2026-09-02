package web

import (
	"arknights_bot/plugins/player"
	"arknights_bot/plugins/skland"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"math"
	"net/url"
	"strconv"
	"time"
)

type StateRenderProps struct {
	Avatar      string              `json:"avatar"`
	PlayerName  string              `json:"playerName"`
	AP          StateMeterProps     `json:"ap"`
	TowerLower  StateMeterProps     `json:"towerLower"`
	TowerHigher StateMeterProps     `json:"towerHigher"`
	Reward      StateMeterProps     `json:"reward"`
	Recruitment StateMeterProps     `json:"recruitment"`
	Trading     StateMeterProps     `json:"trading"`
	Manufacture StateMeterProps     `json:"manufacture"`
	TiredChars  int                 `json:"tiredChars"`
	CheckedIn   bool                `json:"checkedIn"`
	Training    *StateTrainingProps `json:"training,omitempty"`
}

type StateMeterProps struct {
	Current int    `json:"current"`
	Max     int    `json:"max"`
	Label   string `json:"label,omitempty"`
}

type StateTrainingProps struct {
	Avatar string `json:"avatar"`
	Label  string `json:"label"`
}

func buildStateProps(statistic *skland.PlayerStatistic, now time.Time) StateRenderProps {
	props := StateRenderProps{
		Avatar:      stateAvatarURL(statistic.Avatar),
		PlayerName:  statistic.PlayerName,
		CheckedIn:   statistic.CheckedIn,
		TiredChars:  statistic.TiredChars,
		AP:          StateMeterProps{Current: statistic.Ap.Current, Max: statistic.Ap.Max},
		TowerLower:  StateMeterProps{Current: statistic.TowerLower.Current, Max: statistic.TowerLower.Max},
		TowerHigher: StateMeterProps{Current: statistic.TowerHigher.Current, Max: statistic.TowerHigher.Max},
		Reward:      StateMeterProps{Current: statistic.Reward.Current, Max: statistic.Reward.Max},
		Recruitment: StateMeterProps{Current: statistic.Recruitment.Current, Max: statistic.Recruitment.Max},
		Trading:     StateMeterProps{Current: statistic.Trading.Current, Max: statistic.Trading.Max},
		Manufacture: StateMeterProps{Current: statistic.Manufacture.Current, Max: statistic.Manufacture.Max},
	}
	props.AP.Label = stateRecoveryLabel(statistic.Ap.Current, statistic.Ap.Max, statistic.Ap.RecoverTs, now)
	props.TowerLower.Label = stateDaysLabel(statistic.TowerLower.RecoverTs, now)
	props.TowerHigher.Label = props.TowerLower.Label
	props.Reward.Label = stateDaysLabel(statistic.Reward.RecoverTs, now)
	if statistic.Training.CharIcon != "" {
		props.Training = &StateTrainingProps{Avatar: statistic.Training.CharIcon, Label: stateClockLabel(statistic.Training.LeftSeconds)}
	}
	return props
}

func stateAvatarURL(skinID string) string {
	return "https://web.hycdn.cn/arknights/game/assets/char_skin/avatar/" + url.QueryEscape(skinID) + ".png"
}

func stateRecoveryLabel(current, max int, raw string, now time.Time) string {
	if current >= max {
		return "理智已全部恢复"
	}
	seconds := stateUnix(raw) - now.Unix()
	hours := int(math.Floor(math.Mod(float64(seconds)/3600, 24)))
	minutes := int(math.Floor(math.Mod(float64(seconds)/60, 60)))
	return fmt.Sprintf("%d时%d分后恢复", hours, minutes)
}

func stateDaysLabel(raw string, now time.Time) string {
	return fmt.Sprintf("%d天", int(math.Ceil(float64(stateUnix(raw)-now.Unix())/86400)))
}

func stateClockLabel(raw string) string {
	seconds := stateUnix(raw)
	return fmt.Sprintf("%02d:%02d:%02d", seconds/3600, seconds%3600/60, seconds%60)
}

func stateUnix(raw string) int64 {
	value, _ := strconv.ParseInt(raw, 10, 64)
	return value
}

func State(r *gin.Engine) {
	r.GET("/state", func(c *gin.Context) {
		userId, _ := strconv.ParseInt(c.Query("userId"), 10, 64)
		uid := c.Query("uid")
		sklandId := c.Query("sklandId")
		playerData, userAccount, skAccount, err := player.GetPlayerData(userId, sklandId, uid)
		if err != nil {
			log.Println(err)
			renderError(c, err)
			return
		}
		playStatistic, _, err := skland.GetPlayerStatistic(uid, skAccount, userAccount.ServerName)
		if err != nil {
			log.Println(err)
			renderError(c, err)
			return
		}

		playStatistic.Avatar = playerData.Status.Secretary.SkinID
		RenderSpec(c, "state", 1092, 510, buildStateProps(playStatistic, time.Now()))
	})
}
