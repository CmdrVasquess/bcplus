package app

import (
	"regexp"
	"strings"
	"time"
)

//go:generate stringer -type SlotType
type SlotType int

const (
	StUndef SlotType = iota
	StFixed
	StCore
	StOptional
	StHardpoint
	StUtility
	StDeco
)

type JeEngineering struct {
	BlueprintName      string
	Level              int
	ExperimentalEffect string `json:",omitempty"`
}

type JeLdoModule struct {
	Slot        string
	Item        string
	Engineering *JeEngineering `json:",omitempty"`
}

func (m *JeLdoModule) Type() SlotType {
	if res, ok := jeSlotMap[m.Slot]; ok {
		return res
	}
	// TODO do we need submatch
	if match := jsOptRegex.FindStringSubmatch(m.Slot); match != nil {
		return StOptional
	}
	if match := jsHardpRegex.FindStringSubmatch(m.Slot); match != nil {
		if match[1] == "Tiny" {
			return StUtility
		}
		return StHardpoint
	}
	switch {
	case strings.HasPrefix(m.Slot, "Decal"):
		return StDeco
	case strings.HasPrefix(m.Slot, "ShipName"):
		return StDeco
	case strings.HasPrefix(m.Slot, "Bobble"):
		return StDeco
	case strings.HasPrefix(m.Slot, "ShipID"):
		return StDeco
	}
	return StUndef
}

type JeLoadout struct {
	Ts            time.Time `json:"timestamp"`
	CargoCapacity int
	HullHealth    int
	HullValue     int64
	Rebuy         int64
	MaxJumpRange  float64
	Ship          string
	UnladenMass   float64
	ShipID        int
	ModulesValue  int64
	ShipIdent     string
	ShipName      string
	FuelCapacity  struct {
		Reserve float64
		Main    int
	}
	Modules []JeLdoModule
}

var (
	jsOptRegex   = regexp.MustCompile(`^Slot(\d+)_Size(\d+)$`)
	jsHardpRegex = regexp.MustCompile(`^(.+)Hardpoint(\d+)$`)
	jeSlotMap    = map[string]SlotType{
		"ShipCockpit":            StFixed,
		"CargoHatch":             StFixed,
		"PlanetaryApproachSuite": StFixed,
		"Armour":                 StCore,
		"PowerPlant":             StCore,
		"MainEngines":            StCore,
		"FrameShiftDrive":        StCore,
		"LifeSupport":            StCore,
		"PowerDistributor":       StCore,
		"Radar":                  StCore,
		"FuelTank":               StCore,
		"WeaponColour":           StDeco,
		"EngineColour":           StDeco,
		"VesselVoice":            StDeco,
	}
)
