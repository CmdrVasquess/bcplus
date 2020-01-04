package app

import (
	"encoding/json"
	"os"
	"path/filepath"

	"git.fractalqb.de/fractalqb/ggja"
	"github.com/jinzhu/gorm"
)

const (
	cmdrFile = "cmdr.json"
	cmdrDb   = "cmdr.db"
)

type Ship struct {
	Id   string
	Name string
}

type MatState struct {
	Have int `json:"have"`
	Free int `json:"free"`
}

type Head struct {
	Fid  string
	Name string
	Loc  Location
	Ship Ship
}

type Commander struct {
	Head         Head
	OnScreenShot ggja.GenArr
	Mats         map[string]MatState `json:"matNeed"`
	Rcps         map[string]int      `json:"rcpNeed"`
	statFlags    uint32
	surfLoc      SurfPos
	db           *gorm.DB
}

func NewCommander(fid, name string) *Commander {
	return &Commander{
		Head: Head{Fid: fid, Name: name},
		Mats: make(map[string]MatState),
		Rcps: make(map[string]int),
	}
}

func (cdmr *Commander) isVoid() bool { return len(cmdr.Head.Fid) == 0 }

func (cmdr *Commander) close() {
	if cmdr.isVoid() {
		return
	}
	log.Debuga("close `commander` `named`", cmdr.Head.Fid, cmdr.Head.Name)
	if cmdr.db != nil {
		cmdr.db.Close()
	}
	cmdrf := filepath.Join(cmdrDir(cmdr.Head.Fid), cmdrFile)
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
	cmdrf := filepath.Join(cmdrDir(fid), cmdrFile)
	if _, err = os.Stat(cmdrf); os.IsNotExist(err) {
		*cmdr = *NewCommander(fid, name)
		log.Debuga("new `commander` `named`", cmdr.Head.Fid, cmdr.Head.Name)
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
		cmdr.Head.Fid = fid
		if name != "" {
			cmdr.Head.Name = name
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
