package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"git.fractalqb.de/fractalqb/ggja"
)

var (
	useLang     string
	defaultLang string
	indent      bool
)

func must(err error) {
	if err != nil {
		log.Panic(err)
	}
}

type Recipes struct {
	Synth []*Synth `xml:"synth"`
}

type Localized struct {
	Lang string `xml:"lang,attr,omitempty"`
	Text string `xml:",chardata"`
}

func getLocal(nms []Localized, lang string) string {
	for _, nm := range nms {
		if nm.Lang == lang {
			return nm.Text
		}
	}
	sep := strings.IndexRune(lang, '-')
	if sep > 0 {
		l := lang[:sep]
		for _, nm := range nms {
			if nm.Lang == l {
				return nm.Text
			}
		}
	}
	for _, nm := range nms {
		if nm.Lang == defaultLang {
			return nm.Text
		}
	}
	for _, nm := range nms {
		if nm.Lang == "" {
			return nm.Text
		}
	}
	log.Panicf("cannot localize '%s': %v", lang, nms)
	return fmt.Sprintf("<no-locale:%s>", lang)
}

type Synth struct {
	Id     string      `xml:"id,attr"`
	Name   []Localized `xml:"name"`
	Effect []Localized `xml:"effect"`
	Grades []Grade     `xml:"quality"`
}

func (s Synth) mats() (res []string) {
	collect := make(map[string]bool)
	for _, g := range s.Grades {
		for _, m := range g.Demand {
			collect[m.Key] = true
		}
	}
	for m := range collect {
		res = append(res, m)
	}
	sort.Strings(res)
	return res
}

type Grade struct {
	Level  int         `xml:"level,attr"`
	Bonus  []Localized `xml:"bonus"`
	Demand []Mat       `xml:"use"`
}

func (g Grade) demand(mat string) int {
	for _, m := range g.Demand {
		if m.Key == mat {
			if m.Num == 0 {
				return 1
			} else {
				return m.Num
			}
		}
	}
	return 0
}

type Mat struct {
	Key string `xml:"material,attr"`
	Num int    `xml:"count,attr"`
}

func wuiData(syn *Synth) map[string]interface{} {
	allMats := syn.mats()
	res := ggja.GenObj{
		"id":     syn.Id,
		"name":   getLocal(syn.Name, useLang),
		"effect": getLocal(syn.Effect, useLang),
		"mats":   allMats,
	}
	sort.Slice(syn.Grades, func(i, j int) bool {
		return syn.Grades[i].Level < syn.Grades[j].Level
	})
	var grades []ggja.GenObj
	for _, grade := range syn.Grades {
		jg := make(ggja.GenObj)
		if len(grade.Bonus) > 0 {
			jg["bonus"] = getLocal(grade.Bonus, useLang)
		} else {
			jg["bonus"] = nil
		}
		demand := make([]int, len(allMats))
		for i, m := range allMats {
			demand[i] = grade.demand(m)
		}
		jg["mats"] = demand
		grades = append(grades, jg)
	}
	res["grades"] = grades
	return res
}

func main() {
	flag.StringVar(&defaultLang, "std-lang", "en", "set default language")
	flag.StringVar(&useLang, "lang", "", "set output language")
	flag.BoolVar(&indent, "indent", false, "indent output")
	flag.Parse()
	rd, err := os.Open("synth.xml")
	if err != nil {
		log.Panic(err)
	}
	defer rd.Close()
	xdec := xml.NewDecoder(rd)
	var recipes Recipes
	must(xdec.Decode(&recipes))
	enc := json.NewEncoder(os.Stdout)
	if indent {
		enc.SetIndent("", "  ")
	}
	fmt.Println("store.state.synth = [")
	for i, s := range recipes.Synth {
		if i > 0 {
			fmt.Println(",")
		}
		wui := wuiData(s)
		must(enc.Encode(wui))
	}
	fmt.Println("];")
}
