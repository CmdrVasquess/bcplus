package ship

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

var (
	TheTypes TypeRepo
	TheShips = ShipRepo{types: &TheTypes}
)

type TypeRepo string

func (tr TypeRepo) file(name string) string {
	return filepath.Join(string(tr), name+".json")
}

func (tr TypeRepo) Load(name string) *ShipType {
	fnm := tr.file(name)
	log.Infoa("load `ship type` from `file`", name, fnm)
	rd, err := os.Open(fnm)
	if os.IsNotExist(err) {
		return &ShipType{Name: name}
	}
	defer rd.Close()
	dec := json.NewDecoder(rd)
	res := new(ShipType)
	err = dec.Decode(res)
	if err != nil {
		panic(err) // TODO
	}
	return res
}

func (tr TypeRepo) Save(shty *ShipType) error {
	fnm := tr.file(shty.Name)
	tfn := fnm + "~"
	log.Infoa("write `ship type` to `file`", shty.Name, fnm)
	wr, err := os.Create(tfn)
	if err != nil {
		return err
	}
	defer wr.Close()
	enc := json.NewEncoder(wr)
	if err = enc.Encode(shty); err == nil {
		wr.Close()
		os.Rename(tfn, fnm)
	}
	return err
}

type ShipRepo struct {
	types *TypeRepo
	dir   string
}

func (sr *ShipRepo) SetDir(dir string) {
	sr.dir = dir
}

func (sr *ShipRepo) List() (ids []int, err error) {
	fs, err := ioutil.ReadDir(sr.dir)
	if err != nil {
		return nil, err
	}
	rgxShip, err := regexp.Compile(`^ship-(\d+).json$`)
	for _, f := range fs {
		if match := rgxShip.FindStringSubmatch(f.Name()); match != nil {
			id, _ := strconv.Atoi(match[1])
			ids = append(ids, id)
		}
	}
	return ids, nil
}

func (sr ShipRepo) file(id int) string {
	return filepath.Join(sr.dir, fmt.Sprintf("ship-%03d.json", id))
}

func (sr ShipRepo) Load(id int, model string) *Ship {
	fnm := sr.file(id)
	log.Infoa("load `ship` `model` from `file`", id, model, fnm)
	rd, err := os.Open(fnm)
	if os.IsNotExist(err) {
		res := &Ship{Id: id}
		if model != "" {
			res.Type.ShipType = sr.types.Load(model)
		}
		return res
	}
	defer rd.Close()
	dec := json.NewDecoder(rd)
	res := new(Ship)
	err = dec.Decode(res)
	if err != nil {
		panic(err) // TODO
	}
	return res
}

func (sr ShipRepo) Save(s *Ship) error {
	if s == nil {
		return nil
	}
	fnm := sr.file(s.Id)
	tfn := fnm + "~"
	log.Infoa("save `ship` to `file`", s.Id, fnm)
	wr, err := os.Create(tfn)
	if err != nil {
		return err
	}
	defer wr.Close()
	enc := json.NewEncoder(wr)
	enc.SetIndent("", "\t")
	s.StateAt = time.Now()
	if err = enc.Encode(s); err == nil {
		wr.Close()
		os.Rename(tfn, fnm)
	}
	return err
}
