package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	gxy "github.com/CmdrVasquess/BCplus/galaxy"
	l "github.com/fractalqb/qblog"
)

//go:generate stringer -type=RankType
type RankType int

const (
	RnkCombat RankType = iota
	RnkTrade
	RnkExplore
	RnkImp
	RnkFed
	RnkCqc
	RnkCount
)

type Timestamp time.Time

var sepStatSz int
var sepBodySz int

func init() {
	dummy := make([]byte, 4)
	sepStatSz = utf8.EncodeRune(dummy, gxy.SepStation)
	sepBodySz = utf8.EncodeRune(dummy, gxy.SepBody)
}

func (t *Timestamp) MarshalJSON() (res []byte, err error) {
	str := (*time.Time)(t).Format(time.RFC3339)
	res, err = json.Marshal(str)
	return res, err
}

func (t *Timestamp) UnmarshalJSON(json []byte) (err error) {
	jstr := string(json)
	jstr = jstr[1 : len(jstr)-1]
	if ts, err := time.Parse(time.RFC3339, jstr); err != nil {
		return err
	} else {
		*t = Timestamp(ts)
	}
	return nil
}

type GmState struct {
	T            Timestamp
	IsBeta       bool
	Cmdr         Commander
	Tj2j         time.Duration // min. time: jump to jump
	TrvlPlanShip ShipRef
	JumpHist     []*Jump         `json:",omitempty"`
	MatCatHide   map[string]bool `json:",omitempty"`
	MatFlt       MatFilter
	evtBacklog   []map[string]interface{}
	next1stJump  bool
	creds        *CmdrCreds
}

func (g *GmState) isOffline() bool {
	return len(g.Cmdr.Name) == 0
}

const jumpHistMax = 101

type SysRef struct {
	*gxy.StarSys
}

func (sr SysRef) MarshalJSON() (res []byte, err error) {
	if sr.StarSys == nil {
		res, err = json.Marshal("-")
	} else {
		res, err = json.Marshal(sr.Name())
	}
	return res, err
}

func (sr *SysRef) UnmarshalJSON(json []byte) error {
	jstr := string(json)
	jstr = jstr[1 : len(jstr)-1]
	if jstr == "-" {
		sr.StarSys = nil
	} else {
		ssys := theGalaxy.GetSystem(jstr)
		sr.StarSys = ssys
	}
	return nil
}

type Jump struct {
	First  bool // 1st jump after load → statistics etc.
	Sys    SysRef
	Arrive Timestamp
}

func (stat *GmState) addJump(ssys *gxy.StarSys, t Timestamp) {
	if ssys == nil {
		panic("new jump in history without star system")
	}
	hist := append(stat.JumpHist, nil)
	newJump := &Jump{stat.next1stJump, SysRef{ssys}, t}
	i := len(hist) - 2
	for i >= 0 {
		if !time.Time(t).Before(time.Time(hist[i].Arrive)) {
			break
		}
		hist[i+1] = hist[i]
		i--
	}
	if i >= 0 && !time.Time(t).After(time.Time(hist[i].Arrive)) {
		glog.Log(l.Warn, "duplicate jump in history: ", time.Time(t), ssys.Name)
	}
	hist[i+1] = newJump
	if len(hist) > jumpHistMax {
		hist = hist[1:]
	}
	if cap(hist) > 2*len(hist) {
		tmp := make([]*Jump, len(hist))
		copy(tmp, hist)
		hist = tmp
	}
	stat.JumpHist = hist
}

func NewGmState() *GmState {
	res := GmState{
		Cmdr:        *NewCommander(),
		MatCatHide:  make(map[string]bool),
		next1stJump: true,
	}
	return &res
}

func (s *GmState) clear() {
	s.Cmdr.clear()
	s.JumpHist = nil
	s.evtBacklog = nil
	s.next1stJump = true
	s.creds.Clear()
}

func (s *GmState) save(w io.Writer) error {
	je := json.NewEncoder(w)
	je.SetIndent("", "  ")
	err := je.Encode(s)
	return err
}

func (s *GmState) load(r io.Reader) {
	jd := json.NewDecoder(r)
	if err := jd.Decode(s); err != nil {
		log.Printf("faild to load: %s", err)
		s.clear()
	} else { // TODO just debugging
		for i, j := range s.JumpHist {
			if j.Sys.StarSys == nil {
				glog.Fatalf("loaded jump #%d/%d with nil system", i, len(s.JumpHist))
			}
		}
	}
}

type LocRef struct {
	gxy.Location
}

func (lr LocRef) Nil() bool {
	return lr.Location == nil
}

func (lr LocRef) MarshalJSON() (res []byte, err error) {
	if lr.Location == nil {
		res, err = json.Marshal("-")
	} else {
		res, err = json.Marshal(lr.String())
	}
	return res, err
}

func (lr *LocRef) UnmarshalJSON(json []byte) error {
	jstr := string(json)
	jstr = jstr[1 : len(jstr)-1]
	if jstr == "-" {
		lr.Location = nil
	} else {
		if sep := strings.IndexRune(jstr, gxy.SepStation); sep > 0 {
			sysNm := strings.Trim(jstr[sep+sepStatSz:], " \t")
			ssys := theGalaxy.GetSystem(sysNm)
			stnNm := strings.Trim(jstr[:sep], " \t")
			stn := ssys.GetStation(stnNm)
			lr.Location = stn
		} else if sep := strings.IndexRune(jstr, gxy.SepBody); sep > 0 {
			sysNm := strings.Trim(jstr[:sep], " \t")
			ssys := theGalaxy.GetSystem(sysNm)
			bdyNm := strings.Trim(jstr[sep+sepBodySz:], " \t")
			bdy := ssys.GetBody(bdyNm)
			lr.Location = bdy
		} else {
			jstr = strings.Trim(jstr, " \t")
			ssys := theGalaxy.GetSystem(jstr)
			lr.Location = ssys
		}
		if lr.Location == nil {
			glog.Logf(l.Error, "unmarshal LocRef: cannot resolve '%s'", jstr)
		}
	}
	return nil
}

type ShipRef struct {
	*Ship
}

func (shr ShipRef) MarshalJSON() (res []byte, err error) {
	if shr.Ship == nil {
		res, err = json.Marshal("-")
	} else if len(shr.Name) > 0 {
		res, err = json.Marshal(fmt.Sprintf("%s (%d)", shr.Name, shr.ID))
	} else {
		res, err = json.Marshal(fmt.Sprintf("(%d)", shr.ID))
	}
	return res, err
}

var shipRefRgx = regexp.MustCompile("\\((\\d+)\\)")

func (shr *ShipRef) UnmarshalJSON(json []byte) error {
	jstr := string(json)
	jstr = jstr[1 : len(jstr)-1]
	if jstr == "-" {
		shr.Ship = nil
	} else {
		match := shipRefRgx.FindStringSubmatch(jstr)
		if match == nil {
			glog.Logf(l.Error, "cannot resolve ship-ref: '%s'", jstr)
			shr.Ship = nil
		} else {
			shipId, _ := strconv.Atoi(match[1])
			shr.Ship = theGame.Cmdr.ShipById(shipId)
			if shr.Ship == nil {
				glog.Logf(l.Error, "json unmarshal: cannot resolve ship-id %d", shipId)
			}
		}
	}
	return nil
}

type Destination struct {
	Loc  LocRef   `json:"Location"`
	Tags []string `json:",omitempty"`
	Note string   `json:",omitempty"`
}

func (dst *Destination) HasTag(tag string) bool {
	for _, t := range dst.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

func (dst *Destination) Tag(tags ...string) {
	for _, t := range tags {
		if !dst.HasTag(t) {
			dst.Tags = append(dst.Tags, t)
		}
	}
}

func (dst *Destination) Untag(tags ...string) {
	ntgs := make([]string, 0, len(dst.Tags)-len(tags))
NextHaveTag:
	for _, ht := range dst.Tags {
		for _, rt := range tags {
			if ht == rt {
				continue NextHaveTag
			}
		}
		ntgs = append(ntgs, ht)
	}
	dst.Tags = ntgs
}

type SynthRef string

func synthRef(syn *gxy.Synthesis, level int) SynthRef {
	return SynthRef(fmt.Sprintf("%d:%s", level, syn.Name))
}

func (sr SynthRef) split() (name string, level int) {
	sep := strings.IndexRune(string(sr), ':')
	name = string(sr)[sep+1:]
	lvl, _ := strconv.Atoi(string(sr)[:sep])
	return name, lvl
}

func (sr SynthRef) Get() (synth *gxy.Synthesis, level int) {
	var snm string
	snm, level = sr.split()
	synth = theGalaxy.Synthesis(snm)
	return synth, level
}

type Material struct {
	Have int16
	Need int16
}

type Commander struct {
	Name    string
	Credits int64
	Loan    int64
	Ranks   [RnkCount]uint8
	RnkPrgs [RnkCount]uint8
	Friends []string `json:",omitempty"`
	Ships   []*Ship
	CurShip ShipRef
	Home    LocRef
	Loc     LocRef
	MatsRaw CmdrsMats         `json:",omitempty"`
	MatsMan CmdrsMats         `json:",omitempty"`
	MatsEnc CmdrsMats         `json:",omitempty"`
	Dests   []*Destination    `json:"Destinations,omitempty"`
	Synth   map[SynthRef]uint `json:",omitempty"`
}

func NewCommander() *Commander {
	res := Commander{
		MatsRaw: make(map[string]*Material),
		MatsMan: make(map[string]*Material),
		MatsEnc: make(map[string]*Material),
		Synth:   make(map[SynthRef]uint),
	}
	return &res
}

func (cmdr *Commander) NeedSynth(syn *gxy.Synthesis, lvl uint, count uint) {
	key := synthRef(syn, int(lvl))
	if count == 0 {
		delete(cmdr.Synth, key)
	} else {
		glog.Logf(l.Info, "set syref %s = %d", string(key), count)
		cmdr.Synth[key] = count
	}
}

func (cmdr *Commander) ShipById(shipId int) *Ship {
	for _, ship := range cmdr.Ships {
		if ship.ID == shipId {
			return ship
		}
	}
	return nil
}

func (cmdr *Commander) SellShip(ship *Ship, t Timestamp) {
	if ship != nil {
		if ship == cmdr.CurShip.Ship {
			cmdr.CurShip.Ship = nil
		}
		ship.ID = -ship.ID - 1
		ship.Sold = &t
	}
}

func (cmdr *Commander) SellShipId(shipId int, t Timestamp) {
	cmdr.SellShip(cmdr.ShipById(shipId), t)
}

func (cmdr *Commander) Material(jName string) *Material {
	switch theGalaxy.MatCategory(jName) {
	case gxy.Raw:
		return cmdr.MatsRaw[jName]
	case gxy.Man:
		return cmdr.MatsMan[jName]
	case gxy.Enc:
		return cmdr.MatsEnc[jName]
	default:
		return nil
	}
}

func (cmdr *Commander) FindDest(loc gxy.Location) *Destination {
	for _, dst := range cmdr.Dests {
		if dst.Loc.Location.String() == loc.String() {
			return dst
		}
	}
	return nil
}

func (cmdr *Commander) GetDest(loc gxy.Location) (res *Destination) {
	res = cmdr.FindDest(loc)
	if res == nil {
		res = &Destination{
			Loc: LocRef{loc},
		}
		cmdr.Dests = append(cmdr.Dests, res)
	}
	return res
}

func (cmdr *Commander) RmDest(loc gxy.Location) (res bool) {
	ndst := make([]*Destination, 0, len(cmdr.Dests)-1)
	for _, d := range cmdr.Dests {
		if d.Loc.Location.String() != loc.String() {
			ndst = append(ndst, d)
		} else {
			res = true
		}
	}
	cmdr.Dests = ndst
	return res
}

type CmdrsMats map[string]*Material

func (cm CmdrsMats) SetHave(mat string, n int16) {
	m, ok := cm[mat]
	if !ok {
		m = &Material{Have: n}
		cm[mat] = m
	} else {
		m.Have = n
	}
}

func (cm CmdrsMats) clearHave() {
	for m, hn := range cm {
		if hn.Need == 0 {
			delete(cm, m)
		} else {
			hn.Have = 0
		}
	}
}

func (cmdr *Commander) clear() {
	cmdr.Name = ""
	cmdr.Credits = 0
	cmdr.Loan = 0
	cmdr.Friends = nil
	cmdr.Ships = nil
	cmdr.CurShip.Ship = nil
	cmdr.Home.Location = nil
	cmdr.Loc.Location = nil
	cmdr.MatsRaw = make(CmdrsMats)
	cmdr.MatsMan = make(CmdrsMats)
	cmdr.MatsEnc = make(CmdrsMats)
	cmdr.Dests = nil
}

type JumpStats struct {
	DistMax    float32
	DistSum    float32
	DistCount  int
	BoostSum   float32
	BoostCount int
}

type Ship struct {
	ID     int
	Type   string
	Name   string
	Ident  string
	Bought *Timestamp `json:",omitempty"`
	Sold   *Timestamp `json:",omitempty"`
	Loc    LocRef     `json:"Location,omitempty"`
	Jump   JumpStats
}
