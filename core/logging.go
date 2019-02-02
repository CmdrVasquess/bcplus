package core

import (
	"io"
	"os"

	"git.fractalqb.de/fractalqb/c4hgol"
	log "git.fractalqb.de/fractalqb/qbsllm"
	"github.com/CmdrVasquess/BCplus/cmdr"
	"github.com/CmdrVasquess/BCplus/common"
	"github.com/CmdrVasquess/BCplus/galaxy"
	"github.com/CmdrVasquess/BCplus/webui"
	"github.com/CmdrVasquess/watched"
)

var (
	lgr         = log.New(log.Lnormal, "bcplus", nil, nil)
	logEddn     = log.New(log.Lnormal, "bc+edn", nil, nil)
	logEdsm     = log.New(log.Lnormal, "bc+eds", nil, nil)
	LogV, LogVV bool
	logCfg      = log.Config(lgr,
		common.LogConfig,
		watched.LogConfig,
		galaxy.LogConfig,
		cmdr.LogConfig,
		webui.LogConfig,
	)
)

func init() {
	logFile, _ := os.Create("BCplus.log")
	logWr := io.MultiWriter(logFile, os.Stderr)
	lgr.SetOutput(logWr)
}

func FlagLogLevel() {
	if LogVV {
		logCfg.SetLevel(c4hgol.Trace)
	} else if LogV {
		logCfg.SetLevel(c4hgol.Debug)
	}
}
