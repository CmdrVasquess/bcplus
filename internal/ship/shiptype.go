package ship

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"git.fractalqb.de/fractalqb/ggja"
)

type CoreModules int

const (
	PowerPlant CoreModules = iota
	Thrusters
	FrameShiftDrive
	LifeSupport
	PowerDistributor
	Sensors
	FuelTank
)

type SlotRestriction uint

const (
	MilSlot SlotRestriction = (1 << iota)
)

type SlotType struct {
	Size     int
	Restrict uint
}

type MountSize int

const (
	Utility MountSize = iota
	SmallWeapon
	MidWeapon
	LargeWeapon
	HugeWeapon
)

func ParseMountSize(s string) MountSize {
	switch strings.ToLower(s) {
	case "tiny":
		return Utility
	case "small":
		return SmallWeapon
	case "medium":
		return MidWeapon
	case "large":
		return LargeWeapon
	case "huge":
		return HugeWeapon
	}
	return MountSize(-1)
}

type ShipType struct {
	Name   string
	Mounts [HugeWeapon + 1]int
	Core   [FuelTank + 1]int
	Opt    []SlotType
}

type TypeRef struct {
	*ShipType
}

func (tr TypeRef) MarshalJSON() ([]byte, error) {
	if tr.ShipType == nil {
		return []byte("nil"), nil
	}
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "\"%s\"", tr.Name)
	return buf.Bytes(), nil
}

func (tr *TypeRef) UnmarshalJSON(b []byte) error {
	if len(b) < 2 || b[0] != '"' || b[len(b)-1] != '"' {
		return fmt.Errorf("invalid ship-type ref '%s'", string(b))
	}
	name := string(b[1 : len(b)-1])
	if name == "" {
		tr.ShipType = nil
	} else {
		tr.ShipType = TheTypes.Load(name)
	}
	return nil
}

var (
	rgxHpSlot  = regexp.MustCompile(`^(.+)Hardpoint(\d+)$`)
	rgxOptSlot = regexp.MustCompile(`^Slot(\d+)_Size(\d+)$`)
	rgxSize    = regexp.MustCompile(`[sS]ize(\d+)`)
)

func findSize(s string) int {
	match := rgxSize.FindStringSubmatch(s)
	if match == nil {
		return -1
	}
	res, _ := strconv.Atoi(match[1])
	return res
}

func (st *ShipType) Refine(ldo ggja.Obj) (changed bool) {
	if nm := ldo.MStr("Ship"); nm != st.Name {
		st.Name = nm
		changed = true
	}
	for _, bm := range ldo.MArr("Modules").Bare {
		mod := ggja.Obj{Bare: bm.(ggja.GenObj), OnError: ldo.OnError}
		slot := mod.MStr("Slot")
		switch slot {
		case "PowerPlant":
			if sz := findSize(mod.MStr("Item")); sz > st.Core[PowerPlant] {
				st.Core[PowerPlant] = sz
				changed = true
			}
			continue
		case "MainEngines":
			if sz := findSize(mod.MStr("Item")); sz > st.Core[Thrusters] {
				st.Core[Thrusters] = sz
				changed = true
			}
			continue
		case "FrameShiftDrive":
			if sz := findSize(mod.MStr("Item")); sz > st.Core[FrameShiftDrive] {
				st.Core[FrameShiftDrive] = sz
				changed = true
			}
			continue
		case "LifeSupport":
			if sz := findSize(mod.MStr("Item")); sz > st.Core[LifeSupport] {
				st.Core[LifeSupport] = sz
				changed = true
			}
			continue
		case "PowerDistributor":
			if sz := findSize(mod.MStr("Item")); sz > st.Core[PowerDistributor] {
				st.Core[PowerDistributor] = sz
				changed = true
			}
			continue
		case "Radar":
			if sz := findSize(mod.MStr("Item")); sz > st.Core[Sensors] {
				st.Core[Sensors] = sz
				changed = true
			}
			continue
		case "FuelTank":
			if sz := findSize(mod.MStr("Item")); sz > st.Core[FuelTank] {
				st.Core[FuelTank] = sz
				changed = true
			}
			continue
		}
		if match := rgxHpSlot.FindStringSubmatch(slot); match != nil {
			changed = changed || st.rfnHardpoint(mod, match)
		} else if match = rgxOptSlot.FindStringSubmatch(slot); match != nil {
			idx, _ := strconv.Atoi(match[1])
			sz, _ := strconv.Atoi(match[2])
			if idx >= len(st.Opt) {
				tmp := make([]SlotType, idx+1)
				copy(tmp, st.Opt)
				st.Opt = tmp
				changed = true
			}
			if sz > st.Opt[idx].Size {
				st.Opt[idx].Size = sz
				changed = true
			}
		}
	}
	return changed
}

func (st *ShipType) rfnHardpoint(hp ggja.Obj, slot []string) (chg bool) {
	hpno, _ := strconv.Atoi(slot[2])
	switch slot[1] {
	case "Tiny":
		if st.Mounts[Utility] < hpno {
			st.Mounts[Utility] = hpno
			chg = true
		}
	case "Small":
		if st.Mounts[SmallWeapon] < hpno {
			st.Mounts[SmallWeapon] = hpno
			chg = true
		}
	case "Medium":
		if st.Mounts[MidWeapon] < hpno {
			st.Mounts[MidWeapon] = hpno
			chg = true
		}
	case "Large":
		if st.Mounts[LargeWeapon] < hpno {
			st.Mounts[LargeWeapon] = hpno
			chg = true
		}
	case "Huge":
		if st.Mounts[HugeWeapon] < hpno {
			st.Mounts[HugeWeapon] = hpno
			chg = true
		}
	default:
		panic("unknown hardpoint " + slot[0])
	}
	return chg
}
