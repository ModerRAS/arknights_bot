package sign

import (
	bot "arknights_bot/config"
	"arknights_bot/plugins/account"
	"arknights_bot/plugins/skland"
	"arknights_bot/utils"
	"fmt"
	tgbotapi "github.com/ijnkawakaze/telegram-bot-api"
	"log"
	"sync"
	"time"
)

const apRecoverySeconds = 360          // 1 AP per 6 minutes
const apFallbackCheckInterval = 2 * time.Hour // fallback polling when AP is at/above threshold

// apTimers holds active per-user timers to avoid duplicates and support cancellation.
var apTimers sync.Map // key: int64 (userNumber), value: *time.Timer

// InitApRemind is called once at startup to schedule timers for every user with AP remind enabled.
func InitApRemind() {
	var users []UserSign
	res := utils.GetApRemindUsers().Scan(&users)
	if res.RowsAffected > 0 {
		log.Println("初始化理智提醒定时器...")
		for _, user := range users {
			userCopy := user
			go performApCheckAndReschedule(userCopy.UserNumber)
		}
	}
}

// ScheduleNextApCheck starts (or restarts) AP monitoring for a user.
// Call this when the user enables AP remind or changes the threshold.
func ScheduleNextApCheck(userNumber int64) {
	go performApCheckAndReschedule(userNumber)
}

// CancelApTimer cancels the scheduled AP check for a user.
// Call this when the user disables AP remind.
func CancelApTimer(userNumber int64) {
	if v, ok := apTimers.Load(userNumber); ok {
		v.(*time.Timer).Stop()
		apTimers.Delete(userNumber)
	}
}

// performApCheckAndReschedule fetches current AP, notifies if threshold is reached,
// and schedules the next check at the calculated time.
func performApCheckAndReschedule(userNumber int64) {
	// Clean up any existing timer entry for this user.
	apTimers.Delete(userNumber)

	// Re-read user state from DB so we always work with the latest settings.
	var user UserSign
	if r := utils.GetAutoSignByUserId(userNumber).Scan(&user); r.RowsAffected == 0 || user.ApRemind == 0 {
		return
	}

	threshold := user.ApThreshold
	if threshold == 0 {
		threshold = 80
	}

	var players []account.UserPlayer
	utils.GetPlayersByUserId(user.UserNumber).Scan(&players)
	if len(players) == 0 {
		return
	}

	// Iterate all bound players; use the first successful AP fetch for scheduling.
	var scheduleDelay time.Duration = -1
	for _, player := range players {
		var skAccount skland.Account
		var userAccount account.UserAccount
		if r := utils.GetAccountByUid(user.UserNumber, player.Uid).Scan(&userAccount); r.RowsAffected == 0 {
			continue
		}
		skAccount.Hypergryph.Token = userAccount.HypergryphToken
		skAccount.Skland.Token = userAccount.SklandToken
		skAccount.Skland.Cred = userAccount.SklandCred

		playerData, _, err := skland.GetPlayerInfo(player.Uid, skAccount)
		if err != nil {
			log.Println("理智提醒：获取角色信息失败:", player.PlayerName, err)
			continue
		}

		ap := playerData.Status.Ap
		maxAp := ap.Max
		if maxAp == 0 {
			continue
		}

		thresholdAp := maxAp * threshold / 100

		// Compute current AP accounting for time elapsed since the last AP tick.
		now := time.Now().Unix()
		elapsed := now - int64(ap.LastApAddTime)
		elapsedAp := int(elapsed / apRecoverySeconds)
		currentAp := ap.Current + elapsedAp
		if currentAp > maxAp {
			currentAp = maxAp
		}

		apPercent := currentAp * 100 / maxAp

		if currentAp >= thresholdAp {
			if user.ApNotified == 0 {
				// Threshold reached – send notification.
				sendMessage := tgbotapi.NewMessage(user.UserNumber, fmt.Sprintf(
					"⚡ 理智提醒\n角色 %s 当前理智：%d/%d (%d%%)\n已达到设定阈值 %d%%",
					player.PlayerName, currentAp, maxAp, apPercent, threshold,
				))
				bot.Arknights.Send(sendMessage)
				bot.DBEngine.Exec("update user_sign set ap_notified = 1 where user_number = ?", user.UserNumber)
			}
			// AP is at or above threshold; poll again later to detect when it drops.
			if scheduleDelay < 0 {
				scheduleDelay = apFallbackCheckInterval
			}
		} else {
			if user.ApNotified == 1 {
				// AP dropped below threshold again – reset so the next crossing triggers a notification.
				bot.DBEngine.Exec("update user_sign set ap_notified = 0 where user_number = ?", user.UserNumber)
			}
			// Calculate exactly when AP will reach thresholdAp.
			// At time ap.LastApAddTime, AP was ap.Current; it increases by 1 every apRecoverySeconds.
			apNeeded := thresholdAp - ap.Current
			targetUnix := int64(ap.LastApAddTime) + int64(apNeeded)*apRecoverySeconds
			delay := time.Until(time.Unix(targetUnix, 0)) + 30*time.Second
			if delay < 0 {
				delay = 0
			}
			if scheduleDelay < 0 || delay < scheduleDelay {
				scheduleDelay = delay
			}
		}
	}

	if scheduleDelay < 0 {
		// Could not determine a schedule (all players failed) – retry in 30 minutes.
		scheduleDelay = 30 * time.Minute
	}

	if scheduleDelay == 0 {
		// Re-check immediately (threshold already passed).
		go performApCheckAndReschedule(userNumber)
	} else {
		log.Printf("理智提醒：用户 %d 下次检查时间：%s",
			userNumber, time.Now().Add(scheduleDelay).Format("2006-01-02 15:04:05"))
		timer := time.AfterFunc(scheduleDelay, func() {
			performApCheckAndReschedule(userNumber)
		})
		apTimers.Store(userNumber, timer)
	}
}

