package webui

import (
	"net/http"
)

const (
	tkeySysPop = "syspop"
)

var gxtSysPop gxtTopic

func tpcSysPop(w http.ResponseWriter, r *http.Request) {
	bt := gxtSysPop.NewBounT(nil)
	bindTpcHeader(bt, &gxtSysPop)
	bt.BindP(gxtSysPop.TopicData, "null")
	bt.Emit(w)
	CurrentTopic = UISysPop
}
