package cmdr

import (
	"encoding/json"
	"math"
	"os"
	"time"

	l "git.fractalqb.de/fractalqb/qblog"
)

var log = l.Std("bc+cdr:")
var LogConfig = l.Package(log)

type Location struct {
	SysId    int64
	LocId    int64
	Docked   bool
	Lat, Lon float32
}

func (loc *Location) Clear() {
	loc.SysId = 0
	loc.LocId = 0
	loc.Docked = false
	loc.ClearGeo()
}

func (loc *Location) ClearGeo() {
	loc.Lat = float32(math.NaN())
	loc.Lon = float32(math.NaN())
}

func (loc *Location) ValidGeo() bool {
	lat, lon := float64(loc.Lat), float64(loc.Lon)
	return !math.IsNaN(lat) && !math.IsInf(lat, 0) &&
		!math.IsNaN(lon) && !math.IsInf(lon, 0)
}

type Jump struct {
	SysId int64
	When  time.Time
}

const JumpHistLen = 100

type JumpHist struct {
	Hist  []Jump
	WrPtr int
}

func (jh *JumpHist) Len() int { return len(jh.Hist) }

func (jh *JumpHist) Add(sysId int64, t time.Time) {
	j := Jump{SysId: sysId, When: t.Round(time.Second)}
	if len(jh.Hist) < JumpHistLen {
		jh.Hist = append(jh.Hist, j)
	} else {
		jh.Hist[jh.WrPtr] = j
		if jh.WrPtr++; jh.WrPtr >= len(jh.Hist) {
			jh.WrPtr = 0
		}
	}
}

type Rank struct {
	Level    int
	Progress int
}

type State struct {
	Name        string
	Scrambled   string
	Creds, Loan int64
	Ranks       struct {
		Combat, Trade, Explore, CQC, Imps, Feds Rank
	}
	Rep struct {
		Imps, Feds, Allis float32
	}
	MinorRep   map[string]float32  `json:",omitempty"`
	Missions   map[uint32]*Mission `json:",omitempty"`
	JStatFlags uint32
	Loc, Home  Location
	InShip     int
	Ships      map[int]*Ship
	Mats       map[Material]*MatState
	RcpDmnd    map[RcpDef][]int
	EddnMode   string
	Edsm       struct {
		Name   string
		ApiKey string // TODO store somwhere secure!
	}
	Jumps JumpHist
}

func NewState(init *State) *State {
	if init == nil {
		init = new(State)
	}
	init.MinorRep = make(map[string]float32)
	init.Missions = make(map[uint32]*Mission)
	init.Ships = make(map[int]*Ship)
	init.Mats = make(map[Material]*MatState)
	init.RcpDmnd = make(map[RcpDef][]int)
	return init
}

func (s *State) Save(filename string) error {
	log.Logf(l.Linfo, "save commander state to '%s'", filename)
	tmpnm := filename + "~"
	f, err := os.Create(tmpnm)
	if err != nil {
		return err
	}
	defer func() {
		if f != nil {
			f.Close()
		}
		f = nil
	}()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "\t")
	err = enc.Encode(s)
	if err != nil {
		return err
	}
	err = f.Close()
	f = nil
	if err != nil {
		return err
	}
	err = os.Rename(tmpnm, filename)
	return err
}

func (s *State) Load(filename string) error {
	log.Logf(l.Linfo, "load commander state from '%s'", filename)
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(s)
	return err
}
