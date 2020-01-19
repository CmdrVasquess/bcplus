package app

import (
	"net/http"

	"git.fractalqb.de/fractalqb/goxic"
)

type insysScreen struct {
	Screen
	Data goxic.PhIdxs
}

var scrnInSys insysScreen

func (s *insysScreen) loadTmpl(page *WebPage) {
	ts := page.from("insys.html", App.Lang)
	goxic.MustIndexMap(s, ts[""], false, GxName.Convert)
}

func (s *insysScreen) ServeHTTP(wr http.ResponseWriter, rq *http.Request) {
	if cmdr.isVoid() {
		http.NotFound(wr, rq)
		return
	}
	var bt goxic.BounT
	var h Head
	s.init(&bt, &h, "insys")
	bt.BindGen(s.Data, jsonContent(&inSysInfo))
	readState(noErr(func() {
		bt.Emit(wr)
	}))
}
