package main

import (
	"io"
	"os"

	"github.com/CmdrVasquess/BCplus/galaxy"
	"github.com/CmdrVasquess/watched"
	"github.com/fractalqb/qblog"
)

const lNotice int = qblog.Warn / 2

var glog = qblog.Std("bc+   :")
var nmlog = glog.NewSub("bc+nmp:")
var ejlog = glog.NewSub("bc+evj:")
var eslog = glog.NewSub("bc+evs:")
var eulog = glog.NewSub("bc+evu:")

func init() {
	watched.LogConfig.SetParent(glog)
	galaxy.LogConfig.SetParent(glog)
	logFile, _ := os.Create("BCplus.log")
	logWr := io.MultiWriter(logFile, os.Stderr)
	glog.SetOutput(logWr)
}
