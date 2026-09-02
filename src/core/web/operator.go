package web

import (
	"arknights_bot/plugins/operator"
	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"html/template"
	"strings"
)

func operatorProps(value operator.Operator) operator.Operator {
	value.AttackRange = template.HTML(templateText(value.AttackRange))
	for i := range value.Talents {
		value.Talents[i].Name = template.HTML(templateText(value.Talents[i].Name))
		value.Talents[i].Desc = template.HTML(templateText(value.Talents[i].Desc))
	}
	for i := range value.Skills {
		value.Skills[i].Desc = template.HTML(templateText(value.Skills[i].Desc))
		value.Skills[i].SkillRange = template.HTML(templateText(value.Skills[i].SkillRange))
		for j := range value.Skills[i].SpType {
			value.Skills[i].SpType[j] = template.HTML(templateText(value.Skills[i].SpType[j]))
		}
	}
	return value
}

func templateText(value template.HTML) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(value)))
	if err != nil {
		return ""
	}
	return strings.Join(strings.Fields(doc.Text()), " ")
}

func Operator(r *gin.Engine) {
	r.GET("/operator", func(c *gin.Context) {
		name := c.Query("name")
		operatorInfo := operator.ParseOperator(name)
		RenderSpec(c, "operator", 1200, 800, operatorProps(operatorInfo))
	})
}
