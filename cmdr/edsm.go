package cmdr

import (
	"github.com/CmdrVasquess/BCplus/galaxy"
	edsm "github.com/CmdrVasquess/goEDSM"
)

func (gs *GmState) CmdrName() string {
	if edsmCmdr := gs.Creds.Edsm.EdsmCmdr; len(edsmCmdr) == 0 {
		return gs.Cmdr.Name
	} else {
		return edsmCmdr
	}
}

func (gs *GmState) SysAddr() int64 {
	// TODO
	return edsm.Unknown
}

func (gs *GmState) SysName() string {
	return gs.Cmdr.Loc.System().Name()
}

func (gs *GmState) SysCoo() []float64 {
	coos := &(gs.Cmdr.Loc.System().Coos)
	if galaxy.V3dValid(coos) {
		return coos[:]
	} else {
		return nil
	}

}

func (gs *GmState) StationId() int64 {
	// TODO
	return edsm.Unknown
}

func (gs *GmState) StationName() string {
	if stn, ok := gs.Cmdr.Loc.Ref.(*galaxy.Station); ok {
		return stn.Name
	} else {
		return ""
	}
}

func (gs *GmState) ShipId() int {
	if shp := gs.Cmdr.CurShip.Ship; shp == nil {
		return edsm.Unknown
	} else {
		return shp.ID
	}
}

func (gs *GmState) Command() edsm.Command {
	// TODO
	return edsm.Unknown
}
