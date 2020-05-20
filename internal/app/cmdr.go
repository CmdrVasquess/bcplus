package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"

	"git.fractalqb.de/fractalqb/ggja"
	"github.com/CmdrVasquess/bcplus/internal/common"
	"github.com/CmdrVasquess/bcplus/internal/galaxy"
	"github.com/CmdrVasquess/bcplus/internal/ship"
	"github.com/CmdrVasquess/bcplus/itf"
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

type FsdJump struct {
	galaxy.SysDesc
	Time  time.Time
	First bool
}

const JumpMax = 51

type Bookmark struct {
	galaxy.SysDesc
	Tags []string `json:",omitempty"`
}

type CmdrLoc struct {
	Sys galaxy.SysDesc
	itf.LocInSys
}

type Commander struct {
	Fid          string
	Name         string
	Loc          CmdrLoc
	Ship         ship.ShipRef
	OnScreenShot ggja.GenArr
	Mats         map[string]MatState `json:"MatNeed"`
	Rcps         map[string]int      `json:"RcpNeed"`
	Bookmarks    []Bookmark
	DestBM       int
	SurfDest     []float64
	JumpHist     []FsdJump
	JumpW        int
	edpcStory    string
	statFlags    uint32
	firstJump    bool
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

func (cmdr *Commander) AddJump(t time.Time, addr uint64, sys string, coos galaxy.SysCoos) {
	if len(cmdr.JumpHist) < JumpMax {
		cmdr.JumpHist = append(cmdr.JumpHist, FsdJump{
			Time:  t,
			First: cmdr.firstJump,
			SysDesc: galaxy.SysDesc{
				Addr: addr,
				Name: sys,
				Coos: coos,
			},
		})
	} else {
		jump := &cmdr.JumpHist[cmdr.JumpW]
		jump.Time = t
		jump.First = cmdr.firstJump
		jump.Addr = addr
		jump.Name = sys
		jump.Coos = coos
		cmdr.JumpW++
	}
	cmdr.firstJump = false
}

func (cmdr *Commander) LastJump() *FsdJump {
	if len(cmdr.JumpHist) == 0 {
		return nil
	}
	if len(cmdr.JumpHist) < JumpMax {
		return &cmdr.JumpHist[len(cmdr.JumpHist)-1]
	}
	idx := cmdr.JumpW - 1
	if idx < 0 {
		idx = len(cmdr.JumpHist) - 1
	}
	return &cmdr.JumpHist[idx]
}

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
		err := os.MkdirAll(res, common.DirFileMode)
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
		cmdr.sanitizeJumpHist()
	}
	cmdr.db = openDB(fid)
	cmdr.edpcStory, err = edpcStub.SetCmdr(fid, filepath.Join(cmdrDir(fid), "edpc"))
	if err != nil {
		log.Panice(err)
	}
}

func (cmdr *Commander) sanitizeJumpHist() {
	if len(cmdr.JumpHist) < 2 {
		return
	}
	sort.Slice(cmdr.JumpHist, func(i, j int) bool {
		ji, jj := &cmdr.JumpHist[i], &cmdr.JumpHist[j]
		return ji.Time.Before(jj.Time)
	})
	last := 0
	for i := range cmdr.JumpHist {
		if i == last {
			continue
		}
		lj, cj := &cmdr.JumpHist[last], &cmdr.JumpHist[i]
		if lj.Addr != cj.Addr {
			last++
			if last != i {
				cmdr.JumpHist[last] = cmdr.JumpHist[i]
			}
		}
	}
	cmdr.JumpHist = cmdr.JumpHist[:last+1]
	if len(cmdr.JumpHist) > JumpMax {
		cmdr.JumpHist = cmdr.JumpHist[:JumpMax]
	}
	cmdr.JumpW = 0
}

func (cmdr *Commander) setLocMode(m itf.TravelMode) (chg Change) {
	if cmdr.Loc.Mode != m {
		cmdr.Loc.Mode = m
		chg = ChgLoc
	}
	return chg
}

func (cmdr *Commander) setLocSys(addr uint64, name string, coos []float32) (chg Change) {
	if cmdr.Loc.Sys.Addr != addr {
		cmdr.Loc.Sys.Addr = addr
		chg = ChgLoc
	}
	if cmdr.Loc.Sys.Name != name {
		cmdr.Loc.Sys.Name = name
		chg = ChgLoc
	}
	for i := range coos {
		if cmdr.Loc.Sys.Coos[i] != coos[i] {
			cmdr.Loc.Sys.Coos[i] = coos[i]
			chg = ChgLoc
		}
	}
	return chg
}

func (cmdr *Commander) setLocRef(t itf.LocRefType, nm string) (chg Change) {
	if cmdr.Loc.RefType != t {
		cmdr.Loc.RefType = t
		chg = ChgLoc
	}
	if cmdr.Loc.Ref != nm {
		cmdr.Loc.Ref = nm
		chg = ChgLoc
	}
	return chg
}

func (cmdr *Commander) setLocCoos(cs ...float64) (chg Change) {
	if len(cs) != len(cmdr.Loc.Coos) {
		chg = ChgCoos
	} else {
		for i, c := range cs {
			if c != cmdr.Loc.Coos[i] {
				chg = ChgCoos
			}
		}
	}
	cmdr.Loc.Coos = cs
	return chg
}

func (cmdr *Commander) setLocAlt(a float64) (cgh Change) {
	if len(cmdr.Loc.Coos) >= 3 {
		if cmdr.Loc.Coos[2] == a {
			return 0
		}
		cmdr.Loc.Coos[2] = a
	} else {
		tmp := []float64{0, 0, a}
		copy(tmp, cmdr.Loc.Coos)
		cmdr.Loc.Coos = tmp
	}
	return ChgLoc
}

func (cmdr *Commander) setLocLatLon(lat, lon float64) (cgh Change) {
	if len(cmdr.Loc.Coos) < 2 {
		cmdr.Loc.Coos = []float64{lat, lon}
		return ChgLoc
	}
	if cmdr.Loc.Coos[0] != lat {
		cmdr.Loc.Coos[0] = lat
		cmdr.Loc.Coos[1] = lon
		return ChgLoc
	}
	if cmdr.Loc.Coos[1] != lon {
		cmdr.Loc.Coos[1] = lon
		return ChgLoc
	}
	return 0
}
