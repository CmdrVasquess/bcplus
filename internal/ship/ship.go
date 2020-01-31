package ship

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/CmdrVasquess/bcplus/internal/galaxy"

	"git.fractalqb.de/fractalqb/ggja"
	"git.fractalqb.de/fractalqb/qbsllm"
)

var (
	log    = qbsllm.New(qbsllm.Lnormal, "ship", nil, nil)
	LogCfg = qbsllm.Config(log)
)

type Engineerable struct {
	Blueprint int64  `json:",omitempty"`
	EngyLevel int    `json:",omitempty"`
	XEffect   string `json:",omitempty"`
}

type AlloyRating int

const (
	AlloyLight AlloyRating = iota
	AlloyReinforced
	AlloyMilitary
	AlloyMirrored
	AlloyReactive
)

type ModRating int

const (
	RateE ModRating = iota
	RateD
	RateC
	RateB
	RateA
)

type Alloy struct {
	Engineerable
	Rating AlloyRating
}

type Module struct {
	Engineerable
	Size   int
	Rating ModRating
}

func (m *Module) update(mod ggja.Obj) {
	s, c := findSizeClass(mod.MStr("Item"))
	m.Size = s
	m.Rating = c
	m.updEngy(mod)
}

func (m *Module) updEngy(mod ggja.Obj) {
	if engy := mod.Obj("Engineering"); engy == nil {
		m.Blueprint = 0
		m.EngyLevel = 0
		m.XEffect = ""
	} else {
		m.Blueprint = engy.MInt64("BlueprintID")
		m.EngyLevel = engy.MInt("Level")
		m.XEffect = engy.Str("ExperimentalEffect", "")
	}
}

type NamedModule struct {
	Module
	Name string
}

type Mount int

const (
	Fixed Mount = iota
	Gimballed
	Turret
)

func ParseMount(s string) Mount {
	s = strings.ToLower(s)
	switch s {
	case "fixed":
		return Fixed
	case "gimballed":
		return Gimballed
	case "turret":
		return Turret
	}
	return Mount(-1)
}

type Weapon struct {
	Engineerable
	Size  MountSize
	Mount Mount
	Name  string
}

func (m *NamedModule) update(size int, mod ggja.Obj) {
	item := mod.MStr("Item")
	match := rgxIntItem.FindStringSubmatch(item)
	if match == nil {
		m.Name = item
		m.Size = -1
		m.Rating = -1
	} else {
		m.Name = match[1]
		i, _ := strconv.Atoi(match[3])
		m.Size = i
		i, _ = strconv.Atoi(match[5])
		m.Rating = ModRating(i)
	}
	m.updEngy(mod)
}

type Ship struct {
	Type    TypeRef
	Id      int
	Name    string
	Ident   string
	Rebuy   int
	MaxJump float64
	Cargo   int
	Alloy   Alloy
	Core    [FuelTank + 1]Module
	Opt     []NamedModule
	Mount   [HugeWeapon + 1][]Weapon
	Berth   struct {
		Sys  galaxy.SysDesc
		Port string
	}
	StateAt time.Time
}

func (shp *Ship) Update(ldo ggja.Obj) {
	shty := TheTypes.Load(ldo.MStr("Ship"))
	if shty != nil {
		shp.Type.ShipType = shty
	}
	shp.Id = ldo.MInt("ShipID")
	shp.Name = ldo.MStr("ShipName")
	shp.Ident = ldo.MStr("ShipIdent")
	shp.Rebuy = ldo.MInt("Rebuy")
	shp.MaxJump = ldo.MF64("MaxJumpRange")
	shp.Cargo = ldo.MInt("CargoCapacity")
	for _, bm := range ldo.MArr("Modules").Bare {
		mod := ggja.Obj{Bare: bm.(ggja.GenObj), OnError: ldo.OnError}
		slot := mod.MStr("Slot")
		switch slot {
		case "Armour":
			if grade := findGrade(mod.MStr("Item")); grade >= 0 {
				shp.Alloy.Rating = AlloyRating(grade)
			}
			continue
		case "PowerPlant":
			shp.Core[PowerPlant].update(mod)
			continue
		case "MainEngines":
			shp.Core[Thrusters].update(mod)
			continue
		case "FrameShiftDrive":
			shp.Core[FrameShiftDrive].update(mod)
			continue
		case "LifeSupport":
			shp.Core[LifeSupport].update(mod)
			continue
		case "PowerDistributor":
			shp.Core[PowerDistributor].update(mod)
			continue
		case "Radar":
			shp.Core[Sensors].update(mod)
			continue
		case "FuelTank":
			shp.Core[FuelTank].update(mod)
			continue
		}
		if match := rgxHpSlot.FindStringSubmatch(slot); match != nil {
			shp.updateHp(mod, match)
		} else if match = rgxOptSlot.FindStringSubmatch(slot); match != nil {
			idx, _ := strconv.Atoi(match[1])
			sz, _ := strconv.Atoi(match[2])
			if idx >= len(shp.Opt) {
				tmp := make([]NamedModule, idx+1)
				copy(tmp, shp.Opt)
				shp.Opt = tmp
			}
			shp.Opt[idx].update(sz, mod)
		}
	}
}

func hpStripTail(item string) (head, tail string) {
	sep := strings.LastIndexByte(item, '_')
	if sep < 0 {
		return item, ""
	}
	head = item[:sep]
	tail = item[sep+1:]
	return head, tail
}

func parseHpItem(item string) (msz MountSize, name string, mount Mount) {
	if !strings.HasPrefix(item, "hpt_") {
		panic("not a hardpoint item: " + item)
	}
	item = item[4:]
	var tag string
	item, tag = hpStripTail(item)
	msz = ParseMountSize(tag)
	name, tag = hpStripTail(item)
	if tag != "" {
		mount = ParseMount(tag)
	}
	return msz, name, mount
}

func (shp *Ship) updateHp(mod ggja.Obj, slotDesc []string) {
	slotSz := ParseMountSize(slotDesc[1])
	slotIdx, _ := strconv.Atoi(slotDesc[2])
	mlvl := &shp.Mount[slotSz]
	if slotIdx >= len(*mlvl) {
		nlvl := make([]Weapon, slotIdx+1)
		copy(nlvl, *mlvl)
		*mlvl = nlvl
	}
	mountSz, name, mount := parseHpItem(mod.MStr("Item"))
	(*mlvl)[slotIdx].Size = mountSz
	(*mlvl)[slotIdx].Mount = mount
	(*mlvl)[slotIdx].Name = name
}

type ShipRef struct {
	*Ship
}

func (tr ShipRef) MarshalJSON() ([]byte, error) {
	if tr.Ship == nil {
		return []byte("null"), nil
	}
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%d", tr.Id)
	return buf.Bytes(), nil
}

func (tr *ShipRef) UnmarshalJSON(b []byte) error {
	name := string(b)
	if name == "null" || name == "" {
		tr.Ship = nil
	} else {
		id, err := strconv.Atoi(name)
		if err != nil {
			return err
		}
		tr.Ship = TheShips.Load(id, "")
	}
	return nil
}

var (
	rgxClass   = regexp.MustCompile(`[cC]lass(\d+)`)
	rgxGrade   = regexp.MustCompile(`[gG]rade(\d+)`)
	rgxIntItem = regexp.MustCompile(`^int_(.+?)(_size(\d+))?(_class(\d+))?$`)
)

func findClass(s string) ModRating {
	match := rgxClass.FindStringSubmatch(s)
	if match == nil {
		return -1
	}
	res, _ := strconv.Atoi(match[1])
	return ModRating(res - 1)
}

func findSizeClass(s string) (int, ModRating) {
	return findSize(s), findClass(s)
}

func findGrade(s string) ModRating {
	match := rgxGrade.FindStringSubmatch(s)
	if match == nil {
		return -1
	}
	res, _ := strconv.Atoi(match[1])
	return ModRating(res - 1)
}
