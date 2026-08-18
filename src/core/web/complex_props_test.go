package web

import (
	"html/template"
	"testing"

	enemyplugin "arknights_bot/plugins/enemy"
	operatorplugin "arknights_bot/plugins/operator"
)

func TestRegisteredOnUsesFixedUTC8(t *testing.T) {
	if got, want := registeredOn(1704067200), "2024-01-01"; got != want {
		t.Fatalf("registeredOn() = %q, want %q", got, want)
	}
}

func TestOperatorPropsRemovesTemplateMarkup(t *testing.T) {
	props := operatorProps(operatorplugin.Operator{
		AttackRange: template.HTML(`<span>□</span>`),
		Talents: []operatorplugin.Talent{{
			Name: template.HTML(`<b>精神融合</b>`),
			Desc: template.HTML(`攻击力 <i>+10%</i>`),
		}},
		Skills: []operatorplugin.Skill{{
			Desc:       template.HTML(`<span>攻击力提升</span>`),
			SkillRange: template.HTML(`<img src="range.png">□`),
			SpType:     []template.HTML{template.HTML(`<em>自动回复</em>`)},
		}},
	})

	if got, want := string(props.AttackRange), "□"; got != want {
		t.Fatalf("AttackRange = %q, want %q", got, want)
	}
	if got, want := string(props.Talents[0].Name), "精神融合"; got != want {
		t.Fatalf("Talent.Name = %q, want %q", got, want)
	}
	if got, want := string(props.Talents[0].Desc), "攻击力 +10%"; got != want {
		t.Fatalf("Talent.Desc = %q, want %q", got, want)
	}
	if got, want := string(props.Skills[0].SkillRange), "□"; got != want {
		t.Fatalf("SkillRange = %q, want %q", got, want)
	}
	if got, want := string(props.Skills[0].SpType[0]), "自动回复"; got != want {
		t.Fatalf("SpType = %q, want %q", got, want)
	}
}

func TestEnemyPropsRemovesTemplateMarkup(t *testing.T) {
	props := enemyProps(enemyplugin.Enemy{
		Ability: template.HTML(`<span>免疫<strong>沉默</strong></span>`),
		Levels: []enemyplugin.Level{{
			Talent: template.HTML(`<b>攻击提升</b>`),
			Skills: []enemyplugin.Skill{{
				SpInit: template.HTML(`<span>0</span>`),
				SpCost: template.HTML(`<span>5</span>`),
				Desc:   template.HTML(`造成 <i>物理伤害</i>`),
			}},
		}},
	})

	if got, want := string(props.Ability), "免疫沉默"; got != want {
		t.Fatalf("Ability = %q, want %q", got, want)
	}
	if got, want := string(props.Levels[0].Talent), "攻击提升"; got != want {
		t.Fatalf("Talent = %q, want %q", got, want)
	}
	if got, want := string(props.Levels[0].Skills[0].Desc), "造成 物理伤害"; got != want {
		t.Fatalf("Skill.Desc = %q, want %q", got, want)
	}
}
