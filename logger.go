package main

import (
	"github.com/op/go-logging"
)

// CRITICAL ERROR WARNING NOTICE INFO DEBUG
const logModule = "bcplus"

var glog = logging.MustGetLogger(logModule)

type teeBackend struct {
	bs []logging.Backend
}

func (tbe teeBackend) Log(lvl logging.Level, i int, rec *logging.Record) (err error) {
	for _, b := range tbe.bs {
		if e := b.Log(lvl, i, rec); e != nil {
			err = e
		}
	}
	return err
}

//func init() {
//	tee := teeBackend{}
//	tee.bs = append(tee.bs, logging.NewLogBackend()
//}
