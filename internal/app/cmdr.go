package app

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/CmdrVasquess/bcplus/internal/ship"

	"git.fractalqb.de/fractalqb/ggja"
	"github.com/jinzhu/gorm"
)

const (
	cmdrFile = "cmdr.json"
	cmdrDb   = "cmdr.db"
)

type MatState struct {
	Have int `json:"have"`
	Free int `json:"free"`
}

type Commander struct {
	Fid          string
	Name         string
	Loc          Location
	Ship         ship.ShipRef
	OnScreenShot ggja.GenArr
	Mats         map[string]MatState `json:"matNeed"`
	Rcps         map[string]int      `json:"rcpNeed"`
	statFlags    uint32
	surfLoc      SurfPos
	db           *gorm.DB
}

func NewCommander(fid, name string) *Commander {
	return &Commander{
		Fid:  fid,
		Name: name,
		Mats: make(map[string]MatState),
		Rcps: make(map[string]int),
	}
}

func (cdmr *Commander) isVoid() bool { return len(cmdr.Fid) == 0 }

func (cmdr *Commander) close() {
	if cmdr.isVoid() {
		return
	}
	log.Debuga("close `commander` `named`", cmdr.Fid, cmdr.Name)
	ship.TheShips.Save(cmdr.Ship.Ship)
	if cmdr.db != nil {
		cmdr.db.Close()
	}
	cmdrf := filepath.Join(cmdrDir(cmdr.Fid), cmdrFile)
	tmpf := cmdrf + "~"
	wr, err := os.Create(tmpf)
	if err != nil {
		log.Errore(err)
	}
	defer wr.Close()
	enc := json.NewEncoder(wr)
	enc.SetIndent("", "  ")
	err = enc.Encode(cmdr)
	if err != nil {
		log.Errore(err)
	}
	wr.Close()
	*cmdr = *NewCommander("", "")
	err = os.Rename(tmpf, cmdrf)
	if err != nil {
		log.Errore(err)
	}
}

func cmdrDir(fid string) string {
	res := filepath.Join(App.dataDir, fid)
	if _, err := os.Stat(res); os.IsNotExist(err) {
		log.Infoa("new cmdr data `dir`", res)
		err := os.MkdirAll(res, 0777)
		if err != nil {
			log.Panice(err)
		}
	}
	return res
}

func (cmdr *Commander) switchTo(fid, name string) {
	var err error
	cmdr.close()
	if len(fid) == 0 {
		*cmdr = *NewCommander("", "")
		return
	}
	ship.TheShips.SetDir(cmdrDir(fid))
	cmdrf := filepath.Join(cmdrDir(fid), cmdrFile)
	if _, err = os.Stat(cmdrf); os.IsNotExist(err) {
		*cmdr = *NewCommander(fid, name)
		log.Debuga("new `commander` `named`", cmdr.Fid, cmdr.Name)
	} else {
		log.Debuga("load `commander` `named` `from`", fid, name, cmdrf)
		*cmdr = *NewCommander(fid, name)
		newMats, newRcps := cmdr.Mats, cmdr.Rcps
		rd, err := os.Open(cmdrf)
		if err != nil {
			log.Panice(err)
		}
		defer rd.Close()
		dec := json.NewDecoder(rd)
		err = dec.Decode(cmdr)
		if err != nil {
			log.Panice(err)
		}
		cmdr.Fid = fid
		if name != "" {
			cmdr.Name = name
		}
		if cmdr.Mats == nil {
			cmdr.Mats = newMats
		}
		if cmdr.Rcps == nil {
			cmdr.Rcps = newRcps
		}
	}
	cmdr.db = openDB(fid)
}
