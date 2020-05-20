package app

import (
	"net/http"

	"git.fractalqb.de/fractalqb/goxic"
	. "git.fractalqb.de/fractalqb/goxic/content"
	"github.com/CmdrVasquess/bcplus/internal/ship"
)

type shipsScreen struct {
	Screen
	Data goxic.PhIdxs
}

var scrnShips shipsScreen

func (s *shipsScreen) loadTmpl(page *WebPage) {
	ts := page.from("ships.html", App.Lang)
	goxic.MustIndexMap(s, ts[""], false, GxName.Convert)
}

func wuiLoadShips(currentShip *ship.Ship) (res []*ship.Ship) {
	shipIds, err := ship.TheShips.List()
	if err != nil {
		panic(err)
	}
	if currentShip != nil {
		res = []*ship.Ship{currentShip}
	}
	for _, sid := range shipIds {
		if currentShip != nil && sid == currentShip.Id {
			continue
		}
		s := ship.TheShips.Load(sid, "")
		res = append(res, s)
	}
	return res
}

const shipsTab = "ships"

func (s *shipsScreen) ServeHTTP(wr http.ResponseWriter, rq *http.Request) {
	theCurrentTab = 0
	if offline.isOffline(shipsTab, wr, rq) {
		return
	}
	var bt goxic.BounT
	var h WuiHdr
	s.init(&bt, &h, shipsTab)
	ships := wuiLoadShips(cmdr.Ship.Ship)
	bt.Bind(s.Data, Json{V: ships})
	readState(noErr(func() {
		goxic.Must(bt.WriteTo(wr))
	}))
}
