package webui

import (
	"encoding/json"
	"io"
	"net/http"

	gxc "git.fractalqb.de/fractalqb/goxic"
)

const (
	tkeySysPop = "syspop"
)

var gxtSysPop struct {
	*gxc.Template
	HeaderData []int
}

func tpcSysPop(w http.ResponseWriter, r *http.Request) {
	var hdr Header
	newHeader(&hdr)
	bt := gxtSysPop.NewBounT(nil)
	bt.BindGen(gxtSysPop.HeaderData, func(wr io.Writer) int {
		enc := json.NewEncoder(wr)
		err := enc.Encode(hdr)
		if err != nil {
			panic(err)
		}
		return 1
	})
	bt.Emit(w)
}
