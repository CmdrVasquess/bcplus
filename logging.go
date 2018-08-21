package main

import (
	"io"
	"os"

	"git.fractalqb.de/fractalqb/qblog"
	"github.com/CmdrVasquess/BCplus/cmdr"
	"github.com/CmdrVasquess/BCplus/galaxy"
	"github.com/CmdrVasquess/BCplus/webui"
	"github.com/CmdrVasquess/watched"
)

var (
	log         = qblog.Std("bcplus:")
	logEddn     = log.NewSub("bc+edn:")
	logEdsm     = log.NewSub("bc+eds:")
	logV, logVV bool
)

func init() {
	watched.LogConfig.SetParent(log)
	galaxy.LogConfig.SetParent(log)
	cmdr.LogConfig.SetParent(log)
	webui.LogConfig.SetParent(log)
	logFile, _ := os.Create("BCplus.log")
	logWr := io.MultiWriter(logFile, os.Stderr)
	log.SetOutput(logWr)
}

func flagLogLevel() {
	if logVV {
		log.SetLevel(qblog.Trace)
	} else if logV {
		log.SetLevel(qblog.Debug)
	}
}
