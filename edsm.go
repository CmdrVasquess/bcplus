package main

import (
	"github.com/CmdrVasquess/BCplus/cmdr"
	edsm "github.com/CmdrVasquess/goEDSMc"
)

type edsmState cmdr.State

var _ edsm.GameState = (*edsmState)(nil)

func (es *edsmState) CmdrName() string {
	return es.Name
}

func (es *edsmState) SysAddr() int64 {
	// TODO
	return -1
}

func (es *edsmState) SetSysAddr(v int64) {

}

func (es *edsmState) SysName() string {
	ssys, err := theGalaxy.GetSystem(es.Loc.SysId)
	if err != nil {
		logEdsm.Panic(err)
	}
	return ssys.Name
}

func (es *edsmState) SetSysName(v string) {

}

func (es *edsmState) SysCoo() []float64 {
	ssys, err := theGalaxy.GetSystem(es.Loc.SysId)
	if err != nil {
		logEdsm.Panic(err)
	}
	return ssys.Coos[:]
}

func (es *edsmState) SetSysCoo(v []float64) {

}

func (es *edsmState) StationId() int64 {
	// TODO
	return -1
}

func (es *edsmState) SetStationId(v int64) {

}

func (es *edsmState) StationName() string {
	// TODO
	return ""
}

func (es *edsmState) SetStationName(v string) {

}

func (es *edsmState) ShipId() int {
	// TODO
	return -1
}

func (es *edsmState) SetShipId(v int) {

}

func (es *edsmState) Command() edsm.Command {
	// TODO
	return -1
}

func (es *edsmState) SetCommand(v edsm.Command) {

}
