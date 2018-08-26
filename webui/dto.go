package webui

import (
	"github.com/CmdrVasquess/BCplus/cmdr"
)

type HdrSysLoc struct {
	Name string
	Coos [3]float64
}

type Header struct {
	Cmdr   string
	Ship   *cmdr.Ship
	System HdrSysLoc
	Home   *HdrSysLoc
}

const WsUpdCmd = "update"

type WsCmdUpdate struct {
	WsCommand
	Hdr Header
	Tpc interface{} `json:",omitempty"`
}

func NewWsCmdUpdate(reuse *WsCmdUpdate) *WsCmdUpdate {
	if reuse == nil {
		reuse = new(WsCmdUpdate)
	}
	reuse.Cmd = WsUpdCmd
	newHeader(&reuse.Hdr)
	return reuse
}

func newHeader(reuse *Header) *Header {
	if reuse == nil {
		reuse = new(Header)
	}
	if cmdr := theCmdr(); cmdr == nil {
		reuse.Cmdr = "–"
		// TODO
	} else {
		reuse.Cmdr = cmdr.Name
		if cmdr.InShip >= 0 {
			reuse.Ship = cmdr.Ships[cmdr.InShip]
		}
		ssys, _ := theGalaxy.GetSystem(cmdr.Loc.SysId)
		if ssys != nil {
			reuse.System.Name = ssys.Name
			reuse.System.Coos = ssys.Coos
		}
		if cmdr.Home.SysId > 0 {
			ssys, _ = theGalaxy.GetSystem(cmdr.Home.SysId)
			if ssys != nil {
				reuse.Home = &HdrSysLoc{
					Name: ssys.Name,
					Coos: ssys.Coos,
				}
			}
		}
	}
	return reuse
}
