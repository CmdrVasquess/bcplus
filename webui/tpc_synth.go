package webui

import (
	"net/http"
)

const (
	tkeySynth = "synth"
)

var gxtSynth gxtTopic

type tpcSynthData struct {
	MatNms map[string]string
	Stock  map[string]int
	Demand map[string][]int
}

func newSynth(reuse *tpcSynthData) *tpcSynthData {
	if reuse == nil {
		reuse = new(tpcSynthData)
	}
	reuse.MatNms = make(map[string]string)
	reuse.Stock = make(map[string]int)
	reuse.Demand = make(map[string][]int)
	nmap.Material.Base().ForEach(nmap.Material.Base().StdDomain,
		func(ed string) {
			loc, _ := nmap.Material.Map(ed)
			reuse.MatNms[ed] = loc
		},
	)
	cmdr := theCmdr()
	for mat, state := range cmdr.Mats {
		if state.Have > 0 {
			_, key := mat.Parse()
			reuse.Stock[key] = state.Have
		}
	}
	for rcp, needPerGrade := range cmdr.RcpDmnd {
		_, key := rcp.Parse()
		reuse.Demand[key] = needPerGrade
	}
	return reuse
}

func tpcSynth(w http.ResponseWriter, r *http.Request) {
	var data tpcSynthData
	newSynth(&data)
	bt := gxtSynth.NewBounT(nil)
	bindTpcHeader(bt, &gxtSynth)
	bt.BindGen(gxtSynth.TopicData, jsonContent(data))
	bt.Emit(w)
}
