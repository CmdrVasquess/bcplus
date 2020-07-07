package bcplus

import (
	"bytes"
	"crypto/subtle"
	"fmt"
	"math/rand"
	"net/http"
	"path/filepath"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"git.fractalqb.de/fractalqb/goxic"
	gxc "git.fractalqb.de/fractalqb/goxic/content"
	"git.fractalqb.de/fractalqb/goxic/html"
	"git.fractalqb.de/fractalqb/goxic/js"
	"git.fractalqb.de/fractalqb/nmconv"
	"git.fractalqb.de/fractalqb/qbsllm"
	"github.com/CmdrVasquess/bcplus/internal/wapp"
)

func webPIN(h http.HandlerFunc) http.HandlerFunc {
	if len(App.webPin) == 0 {
		return h
	}
	return func(wr http.ResponseWriter, rq *http.Request) {
		_, pass, ok := rq.BasicAuth()
		if ok && subtle.ConstantTimeCompare([]byte(pass), []byte(App.webPin)) != 1 {
			ok = false
			// ConstTimeComp still varies with length
			time.Sleep(time.Duration(300+rand.Intn(300)) * time.Millisecond)
			log.Warna("wrong web-pin from `remote`", rq.RemoteAddr)
		}
		if !ok {
			wr.Header().Set("WWW-Authenticate", `Basic realm="Password: BC+ Web PIN"`)
			http.Error(wr, "Unauthorized", http.StatusUnauthorized)
			return
		}
		h(wr, rq)
	}
}

type staticContent struct {
	fileSrv http.Handler
}

func (sc staticContent) ServeHTTP(wr http.ResponseWriter, rq *http.Request) {
	const cacheCtrl = "Cache-Control"
	wr.Header().Set(cacheCtrl, "max-age=86400")
	//wr.Header().Add(cacheCtrl, "public")
	sc.fileSrv.ServeHTTP(wr, rq)
}

func webRoutes() {
	htStatic := staticContent{
		fileSrv: http.FileServer(http.Dir(filepath.Join(App.assetDir, "s"))),
	}
	http.HandleFunc("/s/", webPIN(http.StripPrefix("/s", htStatic).ServeHTTP))
	http.HandleFunc("/ws/log", webPIN(logWs))
	http.HandleFunc("/ws/app", webPIN(appWs))
	for _, scrn := range wapp.Screens {
		http.Handle("/"+scrn.Key, scrn.Handler)
	}
	http.Handle("/", http.RedirectHandler("/travel", http.StatusSeeOther))
}

func screenTemplate(tld *tmplLoader) *goxic.Template {
	tmpls := tld.load("screen.html", "")
	return tmpls[""]
}

var goxicName = nmconv.Conversion{
	Norm:   nmconv.Uncamel,
	Xform:  nmconv.PerSegment(strings.ToLower),
	Denorm: nmconv.Sep(nmconv.Lisp),
}

func loadTemplates(lang string) {
	tabs := make([]*wapp.Screen, 0, len(wapp.Screens))
	for _, scrn := range wapp.Screens {
		tabs = append(tabs, scrn)
	}
	tmplLd := newTmplLoader()
	tmpls := tmplLd.load("screen.html", "")
	tmplScrn := tmpls[""]
	var bount goxic.BounT
	for key, scrn := range wapp.Screens {
		tmpls = tmplLd.load(key+".html", lang)
		tmplScrn.NewBounT(&bount)
		if sty := tmpls["style"]; sty == nil {
			bount.BindName(goxic.Empty, "style")
		} else {
			bount.BindName(sty.NewBounT(nil), "style")
		}
		main := tmpls["main"]
		if main == nil {
			log.Fatala("`screen` has no main template", key)
		}
		bount.BindName(main.NewBounT(nil), "main")
		fixt, err := bount.Fixate()
		if err != nil {
			log.Fatale(err)
		}
		if err = fixt.FlattenPhs(true); err != nil {
			log.Fatale(err)
		}
		fixt.NewBounT(&bount)
		bount.BindName(gxc.P(lang), "lang")
		if scrn.Title == "" {
			bount.BindName(gxc.P(scrn.Tab), "title")
		} else {
			bount.BindName(gxc.P(scrn.Title), "title")
		}
		bount.BindName(gxc.P(App.webTheme), "theme")
		bount.BindName(gxc.Json{V: tabs}, "tabs")
		bount.BindName(gxc.P(scrn.Key), "active-tab")
		fixt, err = bount.Fixate()
		if err != nil {
			log.Fatale(err)

		}
		fixt = fixt.Pack()
		fixt.Name = "screen:" + scrn.Key
		log.Debuga("`screen` templates `placeholders`", key, fixt.Phs())
		goxic.MustIndexMap(scrn.Handler, fixt, false, goxicName.Convert)
	}
}

func initWebUI() {
	lang := App.EDState.L10n.Lang
	if lang == "" {
		lang = "en"
	}
	for _, scrn := range wapp.Screens {
		scrn.Ext = &App.Extension
	}
	loadTemplates(lang)
}

func runWebUI() {
	webRoutes()
	keyf := filepath.Join(App.dataDir, keyFile)
	crtf := filepath.Join(App.dataDir, certFile)
	lstn := fmt.Sprintf("%s:%d", App.webAddr, App.WebPort)
	if App.webTLS {
		log.Infoa("run web ui on https `addr`", lstn)
		log.Fatale(http.ListenAndServeTLS(lstn, crtf, keyf, nil))
	} else {
		log.Infoa("run web ui on http `addr`", lstn)
		log.Fatale(http.ListenAndServe(lstn, nil))
	}
}

type tmplLoader struct {
	parser *goxic.Parser
	dir    string
}

func newTmplLoader() *tmplLoader {
	res := &tmplLoader{
		parser: html.NewParser(),
		dir:    filepath.Join(App.assetDir, "goxic"),
	}
	if strings.IndexRune(App.debugModes, 't') < 0 {
		res.parser.PrepLine = func(line []byte) []byte {
			if len(line) == 0 {
				return line
			}
			spc, spcLen := utf8.DecodeRune(line)
			spcStart := unicode.IsSpace(spc)
			trimd := bytes.TrimSpace(line)
			if spcStart {
				line = append(line[:spcLen], trimd...)
			} else {
				line = trimd
			}
			return line
		}
	}
	res.parser.CxfMap = map[string]goxic.CntXformer{
		"xml":   html.EscWrap,
		"jsStr": js.EscWrap,
	}
	return res
}

func (tld *tmplLoader) load(name, lang string) map[string]*goxic.Template {
	var fname string
	if len(lang) > 0 {
		fname = filepath.Join(tld.dir, lang, name)
	} else {
		fname = filepath.Join(tld.dir, name)
	}
	tname := name
	if ext := filepath.Ext(name); len(ext) > 0 {
		tname = name[:len(name)-len(ext)]
	}
	log.Debuga("load `template` from `file`", tname, fname)
	res := make(map[string]*goxic.Template)
	err := tld.parser.ParseFile(fname, tname, res)
	if err != nil {
		log.Fatale(err)
	}
	if log.Logs(qbsllm.Ltrace) {
		for nm, t := range res {
			if len(t.Phs()) == 0 {
				log.Tracea("`template` (`aka`) is static", nm, t.Name)
			} else {
				log.Tracea("`template` (`aka`) has `placeholders`",
					nm,
					t.Name,
					strings.Join(t.Phs(), ";"))
			}
		}
	}
	return res
}
