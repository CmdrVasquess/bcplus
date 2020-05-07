package app

import (
	"net/http"

	"git.fractalqb.de/fractalqb/goxic"
)

type edpcScreen struct {
	Screen
	StoryList goxic.PhIdxs
}

var scrnEdpc edpcScreen

func (s *edpcScreen) loadTmpl(page *WebPage) {
	ts := page.from("edpc.html", App.Lang)
	goxic.MustIndexMap(s, ts[""], true, GxName.Convert)
}

const edpcTab = "edpc"

func (s *edpcScreen) ServeHTTP(wr http.ResponseWriter, rq *http.Request) {
	theCurrentTab = 0
	if offline.isOffline(edpcTab, wr, rq) {
		return
	}
	var bt goxic.BounT
	var h WuiHdr
	s.init(&bt, &h, edpcTab)
	sls, err := edpcStub.ListStories()
	if err != nil {
		panic(err)
	}
	// var frontStory *edpc.Story
	// for i := range sls {
	// 	if sls[i].Id == cmdr.EdpcStory {
	// 		frontStory = &sls[i]
	// 	}
	// }
	bt.BindGen(scrnEdpc.StoryList, jsonContent(sls))
	// if frontStory == nil {
	// 	bt.BindP(scrnEdpc.Story, "– No story selected –")
	// 	bt.BindP(scrnEdpc.Author, "–")
	// } else {
	// 	bt.BindP(scrnEdpc.Story, frontStory.Title)
	// 	bt.BindP(scrnEdpc.Author, frontStory.Author)
	// }
	readState(noErr(func() {
		bt.Emit(wr)
	}))
}
