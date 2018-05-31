package main

import (
	"github.com/CmdrVasquess/BCplus/cmdr"
	"github.com/CmdrVasquess/BCplus/galaxy"
	edsm "github.com/CmdrVasquess/goEDSMc"
)

var theEdsm = edsm.NewService(edsm.Life)
var edsmDiscard map[string]bool
var feedEdsm = false

type EdsmState cmdr.GmState

func (gs *EdsmState) CmdrName() string {
	if edsmCmdr := gs.Creds.Edsm.EdsmCmdr; len(edsmCmdr) == 0 {
		return gs.Cmdr.Name
	} else {
		return edsmCmdr
	}
}

func (gs *EdsmState) SysAddr() int64 {
	// TODO
	return edsm.Unknown
}

func (gs *EdsmState) SysName() string {
	return gs.Cmdr.Loc.System().Name()
}

func (gs *EdsmState) SysCoo() []float64 {
	coos := &(gs.Cmdr.Loc.System().Coos)
	if galaxy.V3dValid(coos) {
		return coos[:]
	} else {
		return nil
	}

}

func (gs *EdsmState) StationId() int64 {
	// TODO
	return edsm.Unknown
}

func (gs *EdsmState) StationName() string {
	if stn, ok := gs.Cmdr.Loc.Ref.(*galaxy.Station); ok {
		return stn.Name
	} else {
		return ""
	}
}

func (gs *EdsmState) ShipId() int {
	if shp := gs.Cmdr.CurShip.Ship; shp == nil {
		return edsm.Unknown
	} else {
		return shp.ID
	}
}

func (gs *EdsmState) Command() edsm.Command {
	// TODO
	return edsm.Unknown
}
