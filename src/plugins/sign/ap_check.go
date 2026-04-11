package sign

import (
	bot "arknights_bot/config"
	"arknights_bot/plugins/account"
	"arknights_bot/plugins/skland"
	"arknights_bot/utils"
	"fmt"
	tgbotapi "github.com/ijnkawakaze/telegram-bot-api"
	"log"
)

// CheckApRemind 检查理智提醒
func CheckApRemind() {
	var users []UserSign
	res := utils.GetApRemindUsers().Scan(&users)
	if res.RowsAffected > 0 {
		go func() {
			log.Println("开始检查理智提醒...")
			for _, user := range users {
				checkUserAp(user)
			}
			log.Println("理智提醒检查完毕...")
		}()
	}
}

func checkUserAp(user UserSign) {
	var players []account.UserPlayer
	res := utils.GetPlayersByUserId(user.UserNumber).Scan(&players)
	if res.RowsAffected > 0 {
		for _, player := range players {
			var skAccount skland.Account
			var userAccount account.UserAccount
			res := utils.GetAccountByUid(user.UserNumber, player.Uid).Scan(&userAccount)
			if res.RowsAffected > 0 {
				skAccount.Hypergryph.Token = userAccount.HypergryphToken
				skAccount.Skland.Token = userAccount.SklandToken
				skAccount.Skland.Cred = userAccount.SklandCred

				playerData, _, err := skland.GetPlayerInfo(player.Uid, skAccount)
				if err != nil {
					log.Println("获取角色理智信息失败:", player.PlayerName, err)
					continue
				}

				currentAp := playerData.Status.Ap.Current
				maxAp := playerData.Status.Ap.Max
				if maxAp == 0 {
					continue
				}

				threshold := user.ApThreshold
				if threshold == 0 {
					threshold = 80
				}

				apPercent := currentAp * 100 / maxAp

				if apPercent >= threshold {
					// 理智达到阈值
					if user.ApNotified == 0 {
						// 未通知过，发送通知
						sendMessage := tgbotapi.NewMessage(user.UserNumber, fmt.Sprintf("⚡ 理智提醒\n角色 %s 当前理智：%d/%d (%d%%)\n已达到设定阈值 %d%%", player.PlayerName, currentAp, maxAp, apPercent, threshold))
						bot.Arknights.Send(sendMessage)
						// 标记已通知
						bot.DBEngine.Exec("update user_sign set ap_notified = 1 where user_number = ?", user.UserNumber)
					}
				} else {
					// 理智低于阈值，重置通知状态
					if user.ApNotified == 1 {
						bot.DBEngine.Exec("update user_sign set ap_notified = 0 where user_number = ?", user.UserNumber)
					}
				}
			}
		}
	}
}
