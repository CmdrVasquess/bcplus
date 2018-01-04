package galaxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	str "strings"

	"github.com/op/go-logging"
)

// CRITICAL ERROR WARNING NOTICE INFO DEBUG
const logModule = "bc+:gxy"

var glog = logging.MustGetLogger(logModule)

type Galaxy struct {
	glxyfile  string
	sysByName map[string]*StarSys
	Materials map[string]Material
}

func OpenGalaxy(filename string, refData string) (res *Galaxy, err error) {
	res = &Galaxy{
		glxyfile:  filename,
		sysByName: make(map[string]*StarSys)}
	if rd, err := os.Open(res.glxyfile); err == nil {
		defer rd.Close()
		glog.Infof("load galaxy from %s", filename)
		jdec := json.NewDecoder(rd)
		for {
			ssys := new(StarSys)
			if err := jdec.Decode(ssys); err == io.EOF {
				break
			} else if err != nil {
				panic(err)
			}
			res.sysByName[ssys.Name()] = ssys
		}
	}
	weaveGalaxy(res)
	glog.Infof("%d star-systems loaded", len(res.sysByName))
	err = loadRefData(res, refData)
	return res, err
}

func weaveGalaxy(g *Galaxy) {
	for _, ssys := range g.sysByName {
		for _, body := range ssys.Bodies {
			body.ssys = ssys
		}
		for _, stat := range ssys.Stations {
			stat.ssys = ssys
		}
	}
}

func loadRefData(g *Galaxy, dir string) (err error) {
	g.Materials, err = loadMaterials(dir)
	return err
}

func (g *Galaxy) Close() {
	fnm := g.glxyfile + "~"
	w, _ := os.Create(fnm) // error handling; tmp-file + rename
	defer w.Close()
	glog.Infof("save systems to %s", g.glxyfile)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	for _, sys := range g.sysByName {
		enc.Encode(sys)
	}
	w.Close()
	w = nil
	os.Rename(fnm, g.glxyfile) // TODO error handling
}

func (g *Galaxy) FindSystem(name string) *StarSys {
	res, ok := g.sysByName[name]
	if ok {
		return res
	} else {
		return nil
	}
}

func (g *Galaxy) GetSystem(name string) (res *StarSys) {
	res = g.FindSystem(name)
	if res == nil {
		res = &StarSys{}
		res.SetName(name)
		res.Coos.Set1(math.NaN())
		g.sysByName[name] = res
	}
	return res
}

type Location interface {
	String() string
	GCoos() *Vec3D
	System() *StarSys
}

func Dist(from, to Location) float64 {
	res := V3Dist(from.GCoos(), to.GCoos())
	return res
}

//const koo2intScale float64 = 1000.0

//func sysKooToInt(k float64) int32 {
//	return int32(koo2intScale * k)
//}

//func sysKooFromInt(k int32) float64 {
//	return float64(k) / koo2intScale
//}

type StarSys struct {
	Nm       string `json:"Name"`
	Coos     Vec3D
	Bodies   []*SysBody `json:",omitempty"`
	Stations []*Station `json:",omitempty"`
}

func (s *StarSys) String() string {
	return s.Nm
}

func (s *StarSys) System() *StarSys {
	return s
}

func (s *StarSys) GCoos() *Vec3D {
	return &s.Coos
}

func (s *StarSys) Name() string {
	return s.Nm
}

func (s *StarSys) SetName(name string) {
	s.Nm = str.ToUpper(name)
}

func (s *StarSys) FindBody(name string) *SysBody {
	for _, b := range s.Bodies {
		if b.Name == name {
			return b
		}
	}
	return nil
}

func (s *StarSys) GetBody(name string) (res *SysBody) {
	res = s.FindBody(name)
	if res == nil {
		res = &SysBody{
			ssys: s,
			Name: name}
		s.Bodies = append(s.Bodies, res)
	}
	return res
}

func (s *StarSys) FindStation(name string) *Station {
	for _, p := range s.Stations {
		if p.Name == name {
			return p
		}
	}
	return nil
}

func (s *StarSys) GetStation(name string) *Station {
	res := s.FindStation(name)
	if res == nil {
		res = &Station{
			ssys: s,
			Name: name}
		s.Stations = append(s.Stations, res)
	}
	return res
}

type BodyPos struct {
	body   interface{}
	ground bool
	p1, p2 float32
}

func (bp *BodyPos) Body(ssys *StarSys) *SysBody {
	if bdy, ok := bp.body.(*SysBody); ok {
		return bdy
	} else {
		bdyNm, _ := bp.body.(string)
		res := ssys.GetBody(bdyNm)
		return res
	}
}

func (bp *BodyPos) BodyName() string {
	if bdy, ok := bp.body.(*SysBody); ok {
		return bdy.Name
	} else {
		return bp.body.(string)
	}
}

func (bp *BodyPos) Oribit() float32 {
	if bp.ground {
		panic("body-pos: taking orbit from ground position")
	}
	return bp.p1
}

func (bp *BodyPos) SetOrbit(o float32) {
	bp.ground = false
	bp.p1 = o
}

func (bp *BodyPos) GroundPos() (latd float32, lngt float32) {
	if !bp.ground {
		panic("body-pos: taking latitude from non-ground position")
	}
	return bp.p1, bp.p2
}

func (bp *BodyPos) SetGroundPos(latd float32, lngt float32) {
	bp.ground = true
	bp.p1 = latd
	bp.p2 = lngt
}

//go:generate stringer -type=BodyCat
type BodyCat uint8

const (
	Unknown BodyCat = iota
	Star
	Planet
)

type SysBody struct {
	ssys     *StarSys
	Name     string
	Cat      BodyCat `json:"Category"`
	Dist     float32
	stations []*StationPos
	Landable bool
	Mats     map[string]float32 `json:"Materials,omitempty"`
}

const SepBody = '•'

func (b *SysBody) String() string {
	return fmt.Sprintf("%s %c %s", b.ssys.Name(), SepBody, b.Name)
}

func (b *SysBody) GCoos() *Vec3D {
	return &b.ssys.Coos
}

func (b *SysBody) System() *StarSys {
	return b.ssys
}

type StationPos struct {
	BodyPos
	station *Station
}

func (sp *StationPos) MarshalJSON() (res []byte, err error) {
	tmp := make(map[string]interface{})
	tmp["body"] = sp.BodyName()
	if sp.ground {
		lat, lon := sp.GroundPos()
		tmp["latitude"] = lat
		tmp["longitude"] = lon
	} else {
		tmp["orbit"] = sp.Oribit()
	}
	buf := bytes.NewBuffer(nil)
	jenc := json.NewEncoder(buf)
	if err = jenc.Encode(tmp); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (sp *StationPos) UnmarshalJSON(data []byte) error {
	rd := bytes.NewReader(data)
	jdec := json.NewDecoder(rd)
	tmp := make(map[string]interface{})
	if err := jdec.Decode(&tmp); err != nil {
		return err
	}
	sp.body = tmp["body"].(string)
	if orb, ok := tmp["orbit"]; ok {
		sp.SetOrbit(float32(orb.(float64)))
	} else {
		sp.SetGroundPos(float32(tmp["latitude"].(float64)),
			float32(tmp["longitude"].(float64)))
	}
	return nil
}

type Station struct {
	ssys *StarSys
	Name string
	Type string      `json:",omitempty"`
	Pos  *StationPos `json:"Position,omitempty"`
}

func (s *Station) GCoos() *Vec3D {
	return &s.ssys.Coos
}

const SepStation = '/'

func (s *Station) String() string {
	return fmt.Sprintf("%s %c %s", s.Name, SepStation, s.ssys.Name())
}

func (s *Station) System() *StarSys {
	return s.ssys
}

func (s *Station) SetBody(b *SysBody) {
	if s.ssys != b.ssys {
		panic("linking station and body from different systems")
	}
	s.Pos = &StationPos{
		BodyPos: BodyPos{body: b},
		station: s}
	b.stations = append(b.stations, s.Pos)
}
