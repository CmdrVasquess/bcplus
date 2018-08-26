package webui

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	gxc "git.fractalqb.de/fractalqb/goxic"
	gxw "git.fractalqb.de/fractalqb/goxic/html"
	"git.fractalqb.de/fractalqb/nmconv"
)

const (
	tmplDir         = "tmpl"
	tmplDefaultLang = "default"
	tmplExt         = ".html"
)

var idxMapNames = nmconv.Conversion{
	Norm:   nmconv.Uncamel,
	Xform:  nmconv.PerSegment(strings.ToLower),
	Denorm: nmconv.Sep("-"),
}

func loadTemplates(resDir, lang, version string) {
	page := loadTemplate(resDir, "page", lang)[""]
	btpl := page.NewBounT(nil)
	btpl.BindPName("version", version)
	page = btpl.Fixate()
	prepareOffline(page.NewBounT(btpl), resDir, lang)
	tCTop, tOTop := prepareTopics(page.NewBounT(btpl), resDir, lang)
	tmplSyspop := prepareTopic(tkeySysPop, resDir, lang)
	// ^ more prepare<Topic>(key, resDir, lang) go here
	// All topics loaded => nav titles are known…
	tmplSyspop = finalizeNav(tkeySysPop, tmplSyspop, tCTop, tOTop)
	gxc.MustIndexMap(&gxtSysPop, tmplSyspop, idxMapNames.Convert)
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
		log.Panic("cannot generate static offline page")
	}
}

func prepareTopics(btPage *gxc.BounT, resDir, lang string) (cur, oth *gxc.BounT) {
	tSet := loadTemplate(resDir, "topic", lang)
	tmp, ok := lookupTmpl(tSet, "head").Static()
	if ok {
		btPage.BindName("head", gxc.Data(tmp))
	} else {
		log.Panic("cannot generate static topic head content")
	}
	btPage.BindName("body", lookupTmpl(tSet, "body").NewBounT(nil))
	tmpl := btPage.Fixate()
	err := tmpl.XformPhs(false, gxc.StripPath)
	if err != nil {
		panic(err)
	}
	gxc.MustIndexMap(&gxtTopic, tmpl, idxMapNames.Convert)
	cur = lookupTmpl(tSet, "body/current-topic").NewBounT(nil)
	oth = lookupTmpl(tSet, "body/other-topic").NewBounT(nil)
	return cur, oth
}

func prepareTopic(key, resDir, lang string) *gxc.Template {
	log.Debug("prepare topic " + key)
	tSet := loadTemplate(resDir, key, lang)
	tgen := gxtTopic.NewBounT(nil)
	title, ok := lookupTmpl(tSet, "title").Static()
	if ok {
		tgen.Bind(gxtTopic.Title, gxc.Data(title))
	} else {
		log.Panicf("no title for topic '%s'", key)
	}
	if tmp, ok := lookupTmpl(tSet, "nav-name").Static(); ok {
		getTopic(key).nav = string(tmp)
	} else {
		getTopic(key).nav = string(title)
	}
	tgen.Bind(gxtTopic.Main, lookupTmpl(tSet, "main").NewBounT(nil))
	res := tgen.Fixate()
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
		log.Tracef("check template file '%s'", fnm)
		if err := do(fnm); err == nil {
			log.Debugf("loaded: '%s'", fnm)
			return nil
		}
	}
	if len(lang) >= 2 {
		fnm := filepath.Join(resDir, tmplDir, lang[:2], tmpl+tmplExt)
		log.Tracef("check template file '%s'", fnm)
		if err := do(fnm); err == nil {
			log.Debugf("loaded: '%s'", fnm)
			return nil
		}
	}
	fnm := filepath.Join(resDir, tmplDir, tmplDefaultLang, tmpl+tmplExt)
	log.Tracef("check template file '%s'", fnm)
	if err := do(fnm); err == nil {
		log.Debugf("loaded: '%s'", fnm)
		return nil
	} else {
		return err
	}
}

func loadTemplate(resDir, tmpl, lang string) map[string]*gxc.Template {
	log.Tracef("load template '%s' (%s) from '%s'", tmpl, lang, resDir)
	res := make(map[string]*gxc.Template)
	tpars := gxw.NewParser()
	tpars.PrepLine = strings.TrimSpace
	err := withTmpl(resDir, tmpl, lang, func(file string) error {
		return tpars.ParseFile(file, tmpl, res)
	})
	if err != nil {
		panic(fmt.Errorf("canot load template '%s' (%s) from '%s': %s",
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
		log.Panic("no template '%s' ", path)
	}
	return res
}
