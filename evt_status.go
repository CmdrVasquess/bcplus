package main

import (
	"encoding/json"
	"os"
	"sync"

	c "github.com/CmdrVasquess/BCplus/cmdr"
	l "git.fractalqb.de/fractalqb/qblog"
)

func skipStatEvent(evtFile string) bool {
	stat, err := os.Stat(evtFile)
	return err != nil || stat.Size() == 0
}

func readStatFile(nm string) map[string]interface{} {
	eslog.Logf(l.Debug, "read status file: '%s'", nm)
	f, err := os.Open(nm)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	jdec := json.NewDecoder(f)
	res := make(map[string]interface{})
	err = jdec.Decode(&res)
	if err != nil {
		panic(err)
	}
	return res
}

func dispStfStatus(lock *sync.RWMutex, state *c.GmState, evtFile string) {
	if skipStatEvent(evtFile) {
		eslog.Log(l.Debug, "skip status file:", evtFile)
		return
	}
	msg := map[string]interface{}{
		"cmd":  "statfile",
		"stat": "status",
		"file": readStatFile(evtFile),
	}
	wscSendTo <- msg
}
