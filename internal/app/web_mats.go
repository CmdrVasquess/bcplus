package app

import (
	"net/http"

	"git.fractalqb.de/fractalqb/goxic"
)

type matsScreen struct {
	Screen
	MatNeed goxic.PhIdxs
	RcpNeed goxic.PhIdxs
}

var scrnMats matsScreen

func (s *matsScreen) loadTmpl(page *WebPage) {
	ts := page.from("mats.html", App.Lang)
	goxic.MustIndexMap(s, ts[""], false, GxName.Convert)
}

const matsTab = "mats"

func (s *matsScreen) ServeHTTP(wr http.ResponseWriter, rq *http.Request) {
	theCurrentTab = 0
	if offline.isOffline(matsTab, wr, rq) {
		return
	}
	var bt goxic.BounT
	var h WuiHdr
	s.init(&bt, &h, matsTab)
	bt.BindGen(s.MatNeed, jsonContent(&cmdr.Mats))
	bt.BindGen(s.RcpNeed, jsonContent(&cmdr.Rcps))
	readState(noErr(func() {
		bt.Emit(wr)
	}))
}
