package web

import (
	"arknights_bot/plugins/enemy"
	"github.com/gin-gonic/gin"
	"html/template"
)

func enemyProps(value enemy.Enemy) enemy.Enemy {
	value.Ability = template.HTML(templateText(value.Ability))
	for i := range value.Levels {
		value.Levels[i].Talent = template.HTML(templateText(value.Levels[i].Talent))
		for j := range value.Levels[i].Skills {
			value.Levels[i].Skills[j].SpInit = template.HTML(templateText(value.Levels[i].Skills[j].SpInit))
			value.Levels[i].Skills[j].SpCost = template.HTML(templateText(value.Levels[i].Skills[j].SpCost))
			value.Levels[i].Skills[j].Desc = template.HTML(templateText(value.Levels[i].Skills[j].Desc))
		}
	}
	return value
}

func Enemy(r *gin.Engine) {
	r.GET("/enemy", func(c *gin.Context) {
		name := c.Query("name")
		enemyInfo := enemy.ParseEnemy(name)
		RenderSpec(c, "enemy", 656, 318, enemyProps(enemyInfo))
	})
}
