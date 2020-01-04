package app

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"git.fractalqb.de/fractalqb/goxic"
)

type shipsScreen struct {
	Screen
	Data goxic.PhIdxs
}

var scrnShips shipsScreen

func (s *shipsScreen) loadTmpl(page *WebPage) {
	ts := page.from("ships.html", App.Lang)
	goxic.MustIndexMap(s, ts[""], false, gxName.Convert)
}

type wuiShipsData []*wuiShip

func wuiLoadShip(file string, ship *JeLoadout) {
	rd, err := os.Open(file)
	if err != nil {
		log.Fatale(err)
	}
	defer rd.Close()
	dec := json.NewDecoder(rd)
	err = dec.Decode(&ship)
	if err != nil {
		log.Fatale(err)
	}
}

func (shd *wuiShipsData) readFrom(dir string) {
	dirls, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Panice(err)
	}
	var jship JeLoadout
	for _, dire := range dirls {
		wuiLoadShip(filepath.Join(dir, dire.Name()), &jship)
		ship := &wuiShip{
			ShipID:    jship.ShipID,
			AsOf:      DateTime(jship.Ts),
			Ship:      jship.Ship,
			ShipName:  jship.ShipName,
			ShipIdent: strings.ToUpper(jship.ShipIdent),
			MaxJump:   float32(jship.MaxJumpRange),
		}
		for _, jmdl := range jship.Modules {
			mdl := &wuiShipModule{
				Slot: jmdl.Slot,
				Item: jmdl.Item,
			}
			switch jmdl.Type() {
			case StCore:
				ship.Core = append(ship.Core, mdl)
			case StOptional:
				ship.Opt = append(ship.Opt, mdl)
			case StHardpoint:
				ship.Weap = append(ship.Weap, mdl)
			case StUtility:
				ship.Util = append(ship.Util, mdl)
			}
		}
		*shd = append(*shd, ship)
	}
}

type wuiShip struct {
	ShipID    int
	AsOf      DateTime
	Ship      string
	ShipName  string
	ShipIdent string
	MaxJump   float32
	Core      []*wuiShipModule
	Opt       []*wuiShipModule
	Weap      []*wuiShipModule
	Util      []*wuiShipModule
}

type wuiShipModule struct {
	Slot string
	Item string
}

var shipFileRegex = regexp.MustCompile(`^ship-(\d+).json$`)

func (s *shipsScreen) ServeHTTP(wr http.ResponseWriter, rq *http.Request) {
	if cmdr.isVoid() {
		http.NotFound(wr, rq)
		return
	}
	var ships wuiShipsData
	ships.readFrom(filepath.Join(cmdrDir(cmdr.Head.Fid), "ships"))
	bt := s.NewBounT(nil)
	bt.BindP(s.Theme, App.WebTheme)
	bt.BindGen(s.InitHdr, jsonContent(&cmdr.Head))
	bt.BindGen(s.Data, jsonContent(&ships))
	readState(noErr(func() {
		bt.Emit(wr)
	}))
}
