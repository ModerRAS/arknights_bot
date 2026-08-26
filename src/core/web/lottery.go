package web

import (
	"arknights_bot/utils/model"
	"arknights_bot/utils/repo"
	"fmt"
	"github.com/gin-gonic/gin"
)

type LotteryRenderProps struct {
	Cells []LotteryCellProps `json:"cells"`
}

type LotteryCellProps struct {
	Number     int    `json:"number"`
	State      string `json:"state"`
	UserName   string `json:"userName,omitempty"`
	UserNumber string `json:"userNumber,omitempty"`
}

func buildLotteryProps(details []model.GroupLotteryDetail) LotteryRenderProps {
	byNumber := make(map[int64]model.GroupLotteryDetail, len(details))
	for _, detail := range details {
		byNumber[detail.LotteryNumber] = detail
	}
	props := LotteryRenderProps{Cells: make([]LotteryCellProps, 100)}
	for number := range props.Cells {
		cell := LotteryCellProps{Number: number + 1, State: "empty"}
		if detail, ok := byNumber[int64(cell.Number)]; ok {
			cell.UserName = detail.UserName
			cell.UserNumber = fmt.Sprintf("ID:%d", detail.UserNumber)
			if detail.Status == 1 {
				cell.State = "winner"
			} else {
				cell.State = "selected"
			}
		}
		props.Cells[number] = cell
	}
	return props
}

func Lottery(r *gin.Engine) {
	r.GET("/lotteryDetail", func(c *gin.Context) {
		lotteryId := c.Query("lotteryId")
		var details []model.GroupLotteryDetail
		repo.GetLotteryDetails(lotteryId).Scan(&details)
		RenderSpec(c, "lottery", 982, 1111, buildLotteryProps(details))
	})
}
