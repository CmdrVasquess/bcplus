package app

import (
	"encoding/json"

	"git.fractalqb.de/fractalqb/qbsllm"
)

var (
	wuilog    = qbsllm.New(qbsllm.Lnormal, "webui", nil, nil)
	wuilogCfg = qbsllm.Config(jelog)
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
