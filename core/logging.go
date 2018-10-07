package core

import (
	"io"
	"os"

	"git.fractalqb.de/fractalqb/qblog"
	"github.com/CmdrVasquess/BCplus/cmdr"
	"github.com/CmdrVasquess/BCplus/common"
	"github.com/CmdrVasquess/BCplus/galaxy"
	"github.com/CmdrVasquess/BCplus/webui"
	"github.com/CmdrVasquess/watched"
)

var (
	log         = qblog.Std("bcplus:")
	logEddn     = log.NewSub("bc+edn:")
	logEdsm     = log.NewSub("bc+eds:")
	LogV, LogVV bool
)

func init() {
	common.LogConfig.SetParent(log)
	watched.LogConfig.SetParent(log)
	galaxy.LogConfig.SetParent(log)
	cmdr.LogConfig.SetParent(log)
	webui.LogConfig.SetParent(log)
	logFile, _ := os.Create("BCplus.log")
	logWr := io.MultiWriter(logFile, os.Stderr)
	log.SetOutput(logWr)
}

func FlagLogLevel() {
	if LogVV {
		log.SetLevel(qblog.Ltrace)
	} else if LogV {
		log.SetLevel(qblog.Ldebug)
	}
}
