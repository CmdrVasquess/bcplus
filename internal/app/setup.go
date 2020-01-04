package app

import (
	"os"
)

func setup(app *BCpApp) (firstTime bool) {
	if _, err := os.Stat(app.dataDir); os.IsNotExist(err) {
		log.Infoa("create `data dir`", app.dataDir)
		err := os.MkdirAll(app.dataDir, 0700)
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
