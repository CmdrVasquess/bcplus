package core

import (
	log "git.fractalqb.de/fractalqb/qbsllm"
	"github.com/CmdrVasquess/BCplus/cmdr"
	"github.com/CmdrVasquess/BCplus/common"
	"github.com/CmdrVasquess/BCplus/galaxy"
	edsm "github.com/CmdrVasquess/goEDSMc"
)

var edsmSvc = edsm.NewService(edsm.Life)

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
		logEdsm.Panica("`err`", err)
	}
	return ssys.Name
}

func (es *edsmState) SetSysName(v string) {

}

func (es *edsmState) SysCoo() []float64 {
	ssys, err := theGalaxy.GetSystem(es.Loc.SysId)
	if err != nil {
		logEdsm.Panica("`err`", err)
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

var sysResolveQ = make(chan common.SysResolve, 24)

func sysResolver() {
	lgr.Info(log.Str("running EDSM system resolver"))
NEXT_RQ:
	for rq := range sysResolveQ {
		for _, sysNm := range rq.Names {
			ssys, err := theGalaxy.MustSystem(sysNm)
			if err != nil {
				lgr.Errora("`err`", err)
				continue NEXT_RQ
			}
			if galaxy.V3dValid(ssys.Coos) {
				continue
			}
			sys, err := edsmSvc.System(sysNm, edsm.SYSTEM_COOS)
			if err != nil {
				lgr.Errora("`err`", err)
				continue NEXT_RQ
			}
			if sys == nil {
				lgr.Warna("no system `name` from edsm", sysNm)
				continue
			}
			galaxy.V3dSet3(&ssys.Coos, sys.Coords.X, sys.Coords.Y, sys.Coords.Z)
			lgr.Debuga("resolved `system` `coos`", sysNm, ssys.Coos)
			_, err = theGalaxy.PutSystem(ssys)
			if err != nil {
				lgr.Errora("`err`", err)
				continue NEXT_RQ
			}
		}
	}
	lgr.Info(log.Str("EDSM system resolver terminated"))
}
