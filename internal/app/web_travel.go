package app

import (
	"net/http"

	"git.fractalqb.de/fractalqb/goxic"
)

type travelScreen struct {
	Screen
	Data goxic.PhIdxs
}

var scrnTravel travelScreen

func (s *travelScreen) loadTmpl(page *WebPage) {
	ts := page.from("travel.html", App.Lang)
	goxic.MustIndexMap(s, ts[""], false, GxName.Convert)
}

const travelTab = "travel"

type travelData struct {
	JumpHist []FsdJump
	Bookms   []Bookmark
	DestBm   int
}

func (s *travelScreen) ServeHTTP(wr http.ResponseWriter, rq *http.Request) {
	theCurrentTab = WuiUpTrvl
	if offline.isOffline(travelTab, wr, rq) {
		return
	}
	var bt goxic.BounT
	var h WuiHdr
	s.init(&bt, &h, travelTab)
	readState(noErr(func() {
		data := travelData{
			JumpHist: cmdr.JumpHist,
			Bookms:   cmdr.Bookmarks,
			DestBm:   cmdr.DestBM,
		}
		bt.BindGen(s.Data, jsonContent(data))
		bt.Emit(wr)
	}))
}
