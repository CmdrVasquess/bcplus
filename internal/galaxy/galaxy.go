package galaxy

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"strings"

	"go.etcd.io/bbolt"
)

type Galaxy struct {
	db *bbolt.DB
}

var (
	bucketSystems   = []byte("syss")
	bucketName2EdId = []byte("n2edid")
)

func OpenGalaxy(file string) (g *Galaxy, err error) {
	g = &Galaxy{}
	g.db, err = bbolt.Open(file, 0666, nil)
	if err != nil {
		return nil, err
	}
	err = g.db.Update(func(tx *bbolt.Tx) error {
		if _, err = tx.CreateBucketIfNotExists(bucketSystems); err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists(bucketName2EdId)
		return err
	})
	return g, err
}

func (g *Galaxy) Close() error {
	if g.db != nil {
		return g.db.Close()
	}
	return nil
}

func (g *Galaxy) PutSystem(system *System) (err error) {
	var edid [binary.MaxVarintLen64]byte
	idlen := binary.PutUvarint(edid[:], system.Addr)
	nm := []byte(strings.ToUpper(system.Name))
	var sys bytes.Buffer
	enc := gob.NewEncoder(&sys)
	err = enc.Encode(system)
	if err != nil {
		return err
	}
	return g.db.Update(func(tx *bbolt.Tx) (err error) {
		syss := tx.Bucket(bucketSystems)
		if err = syss.Put(edid[:idlen], sys.Bytes()); err != nil {
			return err
		}
		n2id := tx.Bucket(bucketName2EdId)
		return n2id.Put(nm, edid[:idlen])
	})
}

func (g *Galaxy) GetSystem(edid uint64) (sys *System, err error) {
	var id [binary.MaxVarintLen64]byte
	idlen := binary.PutUvarint(id[:], edid)
	err = g.db.View(func(tx *bbolt.Tx) error {
		syss := tx.Bucket(bucketSystems)
		raw := syss.Get(id[:idlen])
		if raw == nil {
			return nil
		}
		dec := gob.NewDecoder(bytes.NewReader(raw))
		sys = new(System)
		return dec.Decode(sys)
	})
	return sys, err
}

func init() {
	gob.Register(SysPair{})
	gob.Register(SysBody{})
	gob.Register(SysRing{})
	gob.Register(SysSattelite{})
}

type SysCoos [3]float32

type SysDesc struct {
	Addr uint64
	Name string
	Coos SysCoos
}

type System struct {
	SysDesc
	Center SystemObject
}

type SystemObject interface {
	GravParent() SystemObject
	Name() string
	SetName(n string)
	Visit(parent1st bool, fn func(so SystemObject) (done bool)) bool
}

type GravObject interface {
	GravChilds() []SystemObject
}

type sysObj struct {
	parent SystemObject
	ObjNm  string
}

func (so *sysObj) GravParent() SystemObject { return so.parent }

func (so *sysObj) Name() string { return so.ObjNm }

func (so *sysObj) SetName(n string) { so.ObjNm = n }

type sysGrav struct {
	sysObj
	childs []SystemObject
}

func (sg *sysGrav) vCh(fn func(SystemObject) bool) bool {
	for _, c := range sg.childs {
		if fn(c) {
			return true
		}
	}
	return false
}

func (so *sysGrav) GravChilds() []SystemObject { return so.childs }

type SysPair struct {
	sysGrav
	A, B SystemObject
}

func (so *SysPair) Visit(p1st bool, fn func(so SystemObject) (done bool)) bool {
	return firstOrLast(so, p1st, fn, func() bool {
		if !p1st {
			if so.vCh(fn) {
				return true
			}
		}
		if fn(so.A) {
			return true
		}
		if fn(so.B) {
			return true
		}
		if p1st {
			if so.vCh(fn) {
				return true
			}
		}
		return false
	})
}

//go:generate stringer -type SysBodyType
type SysBodyType int

const (
	UnknownBody SysBodyType = iota
	Star
	Planet
)

type SysBody struct {
	sysGrav
	Type SysBodyType
}

func (so *SysBody) Visit(p1st bool, fn func(so SystemObject) (done bool)) bool {
	return firstOrLast(so, p1st, fn, func() bool { return so.vCh(fn) })
}

//go:generate stringer -type SysRingType
type SysRingType int

const (
	UnknownRing SysRingType = iota
	Belt
	PlanetaryRing
)

type SysRing struct {
	sysObj
	Type SysRingType
}

func (so *SysRing) Visit(_ bool, fn func(so SystemObject) (done bool)) bool {
	return fn(so)
}

//go:generate stringer -type SysSattType
type SysSattType int

const (
	UnknownSatt SysSattType = iota
	Station
	Outpost
	Installation
)

type SysSattelite struct {
	sysObj
	Type SysSattType
}

func (so *SysSattelite) Visit(_ bool, fn func(so SystemObject) (done bool)) bool {
	return fn(so)
}

func firstOrLast(parent SystemObject, p1st bool, fn func(SystemObject) bool, ch func() bool) bool {
	if p1st {
		if fn(parent) {
			return true
		}
		return ch()
	} else if ch() {
		return true
	} else {
		return fn(parent)
	}
}
