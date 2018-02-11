package main

import (
	"io"
	"os"

	"github.com/CmdrVasquess/BCplus/galaxy"
	"github.com/fractalqb/qblog"
)

const lNotice int = qblog.Warn / 2

var glog = qblog.Std("bcplus:")
var nmlog = glog.NewSub("bc+nmp:")
var ejlog = glog.NewSub("bc+evj:")
var eulog = glog.NewSub("bc+evu:")

func init() {
	galaxy.LogConfig.SetParent(glog)
	logFile, _ := os.Create("BCplus.log")
	logWr := io.MultiWriter(logFile, os.Stderr)
	glog.SetOutput(logWr)
	nmlog.SetOutput(logWr)
	ejlog.SetOutput(logWr)
	eulog.SetOutput(logWr)
}
