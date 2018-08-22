package webui

import (
	"fmt"
	"path/filepath"
	"strings"

	gxc "git.fractalqb.de/fractalqb/goxic"
	gxw "git.fractalqb.de/fractalqb/goxic/html"
	l "git.fractalqb.de/fractalqb/qblog"
)

const (
	tmplDir         = "tmpl"
	tmplDefaultLang = "default"
	tmplExt         = ".html"
)

func loadTemplates(resDir, lang, version string) {
	page := loadTemplate(resDir, "page", lang)[""]
	btpl := page.NewBounT(nil)
	btpl.BindPName("version", version)
	page = btpl.Fixate()
	prepareOffline(page.NewBounT(btpl), resDir, lang)
}

func prepareOffline(btpl *gxc.BounT, resDir, lang string) {
	tSet := loadTemplate(resDir, "offline", lang)
	tmp, ok := lookupTmpl(tSet, "title").Static()
	if ok {
		btpl.BindName("title", gxc.Data(tmp))
	} else {
		btpl.BindName("title", gxc.Data("offline"))
	}
	tmp, _ = lookupTmpl(tSet, "body").Static()
	btpl.BindName("head", gxc.Empty)
	btpl.BindName("body", gxc.Data(tmp))
	pgOffline, ok = btpl.Fixate().Static()
	if !ok {
		log.Panic("cannot generate static offline page")
	}
}

// didn't find a file.MeCanRead() function, so do it the try-n-error way
func withTmpl(resDir, tmpl, lang string, do func(filename string) error) error {
	if len(lang) > 2 {
		fnm := filepath.Join(resDir, tmplDir, lang, tmpl+tmplExt)
		log.Logf(l.Ltrace, "check template file '%s'", fnm)
		if err := do(fnm); err == nil {
			log.Logf(l.Ldebug, "loaded: '%s'", fnm)
			return nil
		}
	}
	if len(lang) >= 2 {
		fnm := filepath.Join(resDir, tmplDir, lang[:2], tmpl+tmplExt)
		log.Logf(l.Ltrace, "check template file '%s'", fnm)
		if err := do(fnm); err == nil {
			log.Logf(l.Ldebug, "loaded: '%s'", fnm)
			return nil
		}
	}
	fnm := filepath.Join(resDir, tmplDir, tmplDefaultLang, tmpl+tmplExt)
	log.Logf(l.Ltrace, "check template file '%s'", fnm)
	if err := do(fnm); err == nil {
		log.Logf(l.Ldebug, "loaded: '%s'", fnm)
		return nil
	} else {
		return err
	}
}

func loadTemplate(resDir, tmpl, lang string) map[string]*gxc.Template {
	log.Logf(l.Ldebug, "load template '%s' (%s) from '%s'", tmpl, lang, resDir)
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
		log.Panic("no template '%s'", path)
	}
	return res
}
