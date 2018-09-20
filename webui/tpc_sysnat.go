package webui

import (
	"net/http"
)

const (
	tkeySysNat = "sysnat"
)

var gxtSysNat gxtTopic

type tpcSysNatData struct {
	MatNms map[string]string
}

func newSysNat(reuse *tpcSysNatData) *tpcSysNatData {
	if reuse == nil {
		reuse = new(tpcSysNatData)
	}
	reuse.MatNms = make(map[string]string)
	nmap.Material.Base().ForEach(nmap.Material.Base().StdDomain,
		func(ed string) {
			loc, _ := nmap.Material.Map(ed)
			reuse.MatNms[ed] = loc
		},
	)
	// TODO
	return reuse
}

func tpcSysNat(w http.ResponseWriter, r *http.Request) {
	var data tpcSysNatData
	newSysNat(&data)
	bt := gxtSysNat.NewBounT(nil)
	bindTpcHeader(bt, &gxtSysNat)
	bt.BindP(gxtSysNat.TopicData, "null")
	bt.Emit(w)
}
