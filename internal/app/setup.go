package app

import (
	"os"
	"path/filepath"

	"github.com/CmdrVasquess/bcplus/internal/common"
)

const shipsDir = "ships"

func setup(app *BCpApp) (firstTime bool) {
	if _, err := os.Stat(app.dataDir); os.IsNotExist(err) {
		log.Infoa("create `data dir`", app.dataDir)
		err := os.MkdirAll(filepath.Join(app.dataDir, shipsDir), common.DirFileMode)
		if err != nil {
			log.Fatale(err)
		}
		firstTime = true
	}
	if !firstTime {
		if _, err := os.Stat(app.dataBCpApp()); os.IsNotExist(err) {
			firstTime = true
		}
	}
	return firstTime
}
