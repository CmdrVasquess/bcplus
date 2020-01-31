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

const insysTab = "insys"

func (s *insysScreen) ServeHTTP(wr http.ResponseWriter, rq *http.Request) {
	if offline.isOffline(insysTab, wr, rq) {
		return
	}
	var bt goxic.BounT
	var h WuiHdr
	s.init(&bt, &h, insysTab)
	bt.BindGen(s.Data, jsonContent(&inSysInfo))
	readState(noErr(func() {
		bt.Emit(wr)
	}))
}
