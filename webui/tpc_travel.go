package webui

import (
	"net/http"
	"time"
)

const (
	tkeyTravel = "travel"
)

var gxtTravel gxtTopic

type tpcTvlJumpHist struct {
	Sys string
	T   time.Time
}

type tpcTravelData struct {
	JumpHist []tpcTvlJumpHist
}

func newTravel(reuse *tpcTravelData) *tpcTravelData {
	if reuse == nil {
		reuse = new(tpcTravelData)
	}
	cmdr := theCmdr()
	if len(reuse.JumpHist) < cmdr.Jumps.Len() {
		reuse.JumpHist = make([]tpcTvlJumpHist, 0, cmdr.Jumps.Len())
	} else {
		reuse.JumpHist = reuse.JumpHist[:0]
	}
	for i := 0; i < cmdr.Jumps.Len(); i++ {
		idx := i + cmdr.Jumps.WrPtr
		if idx >= cmdr.Jumps.Len() {
			idx = 0
		}
		tmp := tpcTvlJumpHist{
			T: cmdr.Jumps.Hist[idx].When,
		}
		reuse.JumpHist = append(reuse.JumpHist, tmp)
	}
	return reuse
}

func tpcTravel(w http.ResponseWriter, r *http.Request) {
	var data tpcTravelData
	newTravel(&data)
	bt := gxtTravel.NewBounT(nil)
	bindTpcHeader(bt, &gxtTravel)
	bt.BindGen(gxtTravel.TopicData, jsonContent(&data))
	bt.Emit(w)
	currentTopic = UITravel
}
