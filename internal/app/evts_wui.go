package app

import (
	"encoding/json"

	"git.fractalqb.de/fractalqb/c4hgol"

	"git.fractalqb.de/fractalqb/qbsllm"
)

var (
	wuilog    = qbsllm.New(qbsllm.Lnormal, "webui", nil, nil)
	wuilogCfg = c4hgol.Config(qbsllm.NewConfig(jelog))
)

type wuiEvent struct {
	Scr  string
	Cmd  string
	Args interface{}
}

func webuiEvent(raw []byte) {
	wuilog.Debuga("web-ui event: %s", raw)
	var evt wuiEvent
	if err := json.Unmarshal(raw, &evt); err != nil {
		wuilog.Errore(err)
	}
}
