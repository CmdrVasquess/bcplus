package app

import (
	"net/http"

	"git.fractalqb.de/fractalqb/goxic"
	. "git.fractalqb.de/fractalqb/goxic/content"
)

type InSysBody struct {
	Id              int
	Name            string
	Dist            float32
	R, Grav, Temp   float32
	Volcano         string
	Land, TidalLock bool
	Disco, Mapd     bool
}

type InSysInfo struct {
	BodyNum int
	MiscNum int
	Sigs    map[string]int
	Bodies  []*InSysBody
	BdyDsp  string `json:"bdyDsp"`
}

var inSysInfo = InSysInfo{
	BodyNum: -1,
	MiscNum: -1,
	Sigs:    make(map[string]int),
	BdyDsp:  "c",
}

func (isi *InSysInfo) reset() {
	isi.BodyNum = -1
	isi.MiscNum = -1
	isi.Sigs = make(map[string]int)
	isi.Bodies = nil
}

func (isi *InSysInfo) addSignal(nm string) {
	num := isi.Sigs[nm]
	isi.Sigs[nm] = num + 1
}

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
	theCurrentTab = WuiUpInSys
	if offline.isOffline(insysTab, wr, rq) {
		return
	}
	var bt goxic.BounT
	var h WuiHdr
	s.init(&bt, &h, insysTab)
	bt.Bind(s.Data, Json{V: &inSysInfo})
	readState(noErr(func() {
		goxic.Must(bt.WriteTo(wr))
	}))
}
