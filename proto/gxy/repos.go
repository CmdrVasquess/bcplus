package main

import (
	"fmt"
	"reflect"
	"sync"
)

const (
	tagMask  = 0xfffff
	tagShift = 64 - 20
)

func TypeIdx2Tag(i uint64) (uint64, error) {
	if i&^tagMask != 0 {
		return 0, fmt.Errorf("invalid type index: %d", i)
	}
	return i << tagShift, nil
}

func TypeTag2Idx(t uint64) uint64 { return t >> tagShift }

type OId uint64

func MakeOId(ttag uint64, seq uint64) (OId, error) {
	if seq&(tagMask<<tagShift) != 0 {
		return 0, fmt.Errorf("OId sequence overflow: %d", seq)
	}
	return OId(ttag | seq), nil
}

type EtyLink struct {
	repo *EtyRepo
	oid  OId
	obj  interface{}
}

type EtyRepo struct {
	db    *DbApp
	ttag  uint64
	ops   DbOps
	cache map[OId]*EtyLink
}

func NewEtyRepo(db *DbApp, typeIdx uint64, ops DbOps) (res *EtyRepo, err error) {
	reposLock.Lock()
	defer reposLock.Unlock()
	if res = repos[typeIdx]; res != nil {
		return nil, fmt.Errorf("type index reused: %d", typeIdx)
	}
	tag, err := TypeIdx2Tag(typeIdx)
	if err != nil {
		return nil, err
	}
	res = &EtyRepo{
		db:    db,
		ttag:  tag,
		ops:   ops,
		cache: make(map[OId]*EtyLink),
	}
	repos[typeIdx] = res
	return res, nil
}

func GetEtyRepo(typeIdx uint64) (res *EtyRepo, err error) {
	reposLock.Lock()
	defer reposLock.Unlock()
	if res = repos[typeIdx]; res == nil {
		return nil, fmt.Errorf("unknown type index: %d", typeIdx)
	}
	return res, nil
}

var etyType = reflect.TypeOf(Entity{})

func isEntity(obj interface{}) *Entity {
	objVal := reflect.ValueOf(obj)
	if objVal.Kind() == reflect.Ptr {
		objVal = objVal.Elem()
	}
	for i := 0; i < objVal.NumField(); i++ {
		f := objVal.Field(i)
		if f.Type() == etyType {
			return f.Addr().Interface().(*Entity)
		}
	}
	return nil
}

func (er *EtyRepo) Persist(e interface{}) error {
	ety := isEntity(e)
	if ety == nil {
		return fmt.Errorf("not an entity: %v", e)
	}
	tx, own := er.db.Tx()
	defer own.Rollback(er.db)
	oid, err := er.ops.Create(tx, er.ttag, e)
	if err != nil {
		return err
	}
	eln := &EtyLink{
		repo: er,
		oid:  oid,
		obj:  e,
	}
	ety.etyLink = eln
	er.cache[oid] = eln
	own.Commit(er.db)
	return nil
}

type Entity struct {
	etyLink *EtyLink
}

type EtyRef struct {
	p interface{}
}

var (
	repos     = make(map[uint64]*EtyRepo)
	reposLock sync.Mutex
)
