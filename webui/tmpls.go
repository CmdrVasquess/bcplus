package webui

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	gxc "git.fractalqb.de/fractalqb/goxic"
	gxw "git.fractalqb.de/fractalqb/goxic/html"
	"git.fractalqb.de/fractalqb/nmconv"
	l "git.fractalqb.de/fractalqb/qbsllm"
)

const (
	tmplDir         = "tmpl"
	tmplDefaultLang = "en"
	tmplExt         = ".html"
)

func jsonContent(j interface{}) func(io.Writer) int {
	return func(wr io.Writer) int {
		enc := json.NewEncoder(wr)
		enc.SetIndent("", "\t")
		err := enc.Encode(j)
		if err != nil {
			panic(err)
		}
		return 1
	}
}

var idxMapNames = nmconv.Conversion{
	Norm:   nmconv.Uncamel,
	Xform:  nmconv.PerSegment(strings.ToLower),
	Denorm: nmconv.Sep("-"),
}

func loadTemplates(resDir, lang, version string) {
	page := loadTemplate(resDir, "page", lang)[""]
	btpl := page.NewBounT(nil)
	btpl.BindPName("version", version)
	btpl.BindPName("theme", theTheme)
	page = btpl.Fixate()
	prepareOffline(page.NewBounT(btpl), resDir, lang)
	btTopic, btCTop, btOTop := prepareTopics(page.NewBounT(btpl), resDir, lang)
	var prepTmpls []*gxc.Template
	for _, tpc := range topics {
		t := prepareTopic(btTopic, tpc.key, resDir, lang)
		prepTmpls = append(prepTmpls, t)
	}
	for i := range topics {
		tmpl := finalizeNav(topics[i].key, prepTmpls[i], btCTop, btOTop)
		gxc.MustIndexMap(topics[i].gxt, tmpl, idxMapNames.Convert)
	}
}

func prepareOffline(btPage *gxc.BounT, resDir, lang string) {
	tSet := loadTemplate(resDir, "offline", lang)
	tmp, ok := lookupTmpl(tSet, "title").Static()
	if ok {
		btPage.BindName("title", gxc.Data(tmp))
	} else {
		btPage.BindName("title", gxc.Data("offline"))
	}
	tmp, _ = lookupTmpl(tSet, "body").Static()
	btPage.BindName("head", gxc.Empty)
	btPage.BindName("body", gxc.Data(tmp))
	pgOffline, ok = btPage.Fixate().Static()
	if !ok {
		log.Panic(l.Str("cannot generate static offline page"))
	}
}

func prepareTopics(btPage *gxc.BounT, resDir, lang string) (tpc, cur, oth *gxc.BounT) {
	tSet := loadTemplate(resDir, "topic", lang)
	tmpl := lookupTmpl(tSet, "head")
	if tmpl == nil {
		log.Panic(l.Str("cannot find head template for topics"))
	} else {
		btPage.BindName("head", tmpl.NewBounT(nil))
	}
	btPage.BindName("body", lookupTmpl(tSet, "body").NewBounT(nil))
	tmpl = btPage.Fixate()
	err := tmpl.XformPhs(false, gxc.StripPath)
	if err != nil {
		panic(err)
	}
	tpc = tmpl.NewBounT(nil)
	cur = lookupTmpl(tSet, "body/current-topic").NewBounT(nil)
	oth = lookupTmpl(tSet, "body/other-topic").NewBounT(nil)
	return tpc, cur, oth
}

func prepareTopic(btTopic *gxc.BounT, key, resDir, lang string) *gxc.Template {
	log.Debuga("prepare `topic`", key)
	tSet := loadTemplate(resDir, key, lang)
	title, ok := lookupTmpl(tSet, "title").Static()
	if ok {
		btTopic.BindName("title", gxc.Data(title))
	} else {
		log.Panica("no title for `topic`", key)
	}
	if tmp := tSet["head"]; tmp == nil {
		btTopic.BindName("head", gxc.Empty)
	} else if stat, ok := tmp.Static(); ok {
		btTopic.BindName("head", gxc.Data(stat))
	} else {
		log.Panica("not static content: `name`", tmp.Name)
	}
	if tmp, ok := lookupTmpl(tSet, "nav-name").Static(); ok {
		getTopic(key).nav = string(tmp)
	} else {
		getTopic(key).nav = string(title)
	}
	btTopic.BindName("main", lookupTmpl(tSet, "main").NewBounT(nil))
	if tmpl := tSet["script"]; tmpl == nil {
		btTopic.BindName("script", gxc.Empty)
	} else if raw, ok := tmpl.Static(); ok {
		btTopic.BindName("script", gxc.Data(raw))
	} else {
		btTopic.BindName("script", gxc.Empty)
	}
	res := btTopic.Fixate()
	res.XformPhs(false, gxc.StripPath)
	res.Name = key
	return res
}

func finalizeNav(key string, tmpl *gxc.Template, cTop, oTop *gxc.BounT) *gxc.Template {
	tgen := tmpl.NewBounT(nil)
	tgen.BindGenName("topics", func(wr io.Writer) (res int) {
		for _, navTpc := range topics {
			if navTpc.key == key {
				cTop.BindPName("title", navTpc.nav)
				res += cTop.Emit(wr)
			} else {
				oTop.BindPName("title", navTpc.nav)
				oTop.BindPName("path", navTpc.path)
				res += oTop.Emit(wr)
			}
		}
		return res
	})
	res := tgen.Fixate()
	res.XformPhs(false, gxc.StripPath)
	return res
}

// didn't find a file.MeCanRead() function, so do it the try-n-error way
func withTmpl(resDir, tmpl, lang string, do func(filename string) error) error {
	if len(lang) > 2 {
		fnm := filepath.Join(resDir, tmplDir, lang, tmpl+tmplExt)
		log.Tracea("check template `file`", fnm)
		if err := do(fnm); err == nil {
			log.Debuga("loaded `file`", fnm)
			return nil
		}
	}
	if len(lang) >= 2 {
		fnm := filepath.Join(resDir, tmplDir, lang[:2], tmpl+tmplExt)
		log.Tracea("check template `file`", fnm)
		if err := do(fnm); err == nil {
			log.Debuga("loaded `file`", fnm)
			return nil
		}
	}
	fnm := filepath.Join(resDir, tmplDir, tmplDefaultLang, tmpl+tmplExt)
	log.Tracea("check template `file`", fnm)
	if err := do(fnm); err == nil {
		log.Debuga("loaded `filie`", fnm)
		return nil
	} else {
		return err
	}
}

func loadTemplate(resDir, tmpl, lang string) map[string]*gxc.Template {
	log.Tracea("load `template` (`lang`) from `dir`", tmpl, lang, resDir)
	res := make(map[string]*gxc.Template)
	tpars := gxw.NewParser()
	tpars.PrepLine = strings.TrimSpace
	err := withTmpl(resDir, tmpl, lang, func(file string) error {
		return tpars.ParseFile(file, tmpl, res)
	})
	if err != nil {
		panic(fmt.Errorf("canot load `template` (`lang`) from `dir`: `err`",
			tmpl,
			lang,
			resDir,
			err))
	}
	return res
}

func lookupTmpl(tSet map[string]*gxc.Template, path string) *gxc.Template {
	res, ok := tSet[path]
	if !ok {
		log.Panica("no `template`", path)
	}
	return res
}
