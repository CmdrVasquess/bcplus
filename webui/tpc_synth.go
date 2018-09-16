package webui

import (
	"encoding/json"
	"io"
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
	var hdr Header
	newHeader(&hdr)
	var data tpcSynthData
	newSynth(&data)
	bt := gxtSynth.NewBounT(nil)
	bt.BindGen(gxtSynth.HeaderData, func(wr io.Writer) int {
		enc := json.NewEncoder(wr)
		enc.SetIndent("", "\t")
		err := enc.Encode(hdr)
		if err != nil {
			panic(err)
		}
		return 1 // TODO howto determine the correct length
	})
	bt.BindGen(gxtSynth.TopicData, func(wr io.Writer) int {
		enc := json.NewEncoder(wr)
		enc.SetIndent("", "\t")
		err := enc.Encode(data)
		if err != nil {
			panic(err)
		}
		return 1 // TODO howto determine the correct length
	})
	bt.Emit(w)
}
