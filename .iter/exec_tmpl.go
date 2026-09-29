package main

// Execute template/Operator.tmpl with the frozen operator fixture props and
// write the HTML to repo root for Playwright measurement. Throwaway tool.

import (
	"encoding/json"
	"html/template"
	"os"
)

func main() {
	propsRaw, err := os.ReadFile(".iter/operator-props.json")
	if err != nil {
		panic(err)
	}
	var p map[string]any
	if err := json.Unmarshal(propsRaw, &p); err != nil {
		panic(err)
	}
	op := p["op"].(map[string]any)
	cache := "src/utils/media/testdata/visual/baseline/cache/"

	tplData := map[string]any{
		"Painting":  cache + "operator-painting-1024.png",
		"AttackRange": p["attackRange"],
		"OP": map[string]any{
			"HP": op["hp"], "ATK": op["atk"], "DEF": op["def"],
			"Res": op["res"], "Interval": op["interval"], "ReDeploy": op["reDeploy"],
			"Block": op["block"], "Cost": op["cost"], "Logo": op["logo"],
			"Tags": op["tags"], "Name": op["name"], "Code": op["code"],
			"NameEn": op["nameEn"], "Rarity": op["rarity"], "Profession": op["profession"],
		},
		"ProfessionBranch": map[string]any{"Name": p["professionBranch"].(map[string]any)["name"], "Desc": p["professionBranch"].(map[string]any)["desc"]},
		"Talents":          p["talents"],
	}
	skills := []map[string]any{}
	for _, item := range p["skills"].([]any) {
		m := item.(map[string]any)
		skills = append(skills, map[string]any{
			"Icon": cache + "operator-skill-128.png", "Name": m["name"], "Desc": m["desc"],
			"SkillRange": m["skillRange"], "SpType": m["spType"], "SpInit": m["spInit"],
			"SpCost": m["spCost"], "Duration": m["duration"],
		})
	}
	tplData["Skills"] = skills
	potentials := []map[string]any{}
	for _, item := range p["potentials"].([]any) {
		m := item.(map[string]any)
		potentials = append(potentials, map[string]any{"Rank": m["rank"], "Desc": m["desc"]})
	}
	tplData["Potentials"] = potentials
	talents := []map[string]any{}
	for _, item := range p["talents"].([]any) {
		m := item.(map[string]any)
		talents = append(talents, map[string]any{"Evolve": m["evolve"], "Name": m["name"], "Desc": m["desc"]})
	}
	tplData["Talents"] = talents
	bld := []map[string]any{}
	for _, item := range p["buildingSkills"].([]any) {
		m := item.(map[string]any)
		bld = append(bld, map[string]any{"Evolve": m["evolve"], "Icon": cache + "operator-building-36.png", "Name": m["name"], "Desc": m["desc"]})
	}
	tplData["BuildingSkills"] = bld

	tpl, err := template.ParseFiles("template/Operator.tmpl")
	if err != nil {
		panic(err)
	}
	out, err := os.Create("_measure_operator.html")
	if err != nil {
		panic(err)
	}
	defer out.Close()
	if err := tpl.Execute(out, tplData); err != nil {
		panic(err)
	}
	println("wrote _measure_operator.html")
}
