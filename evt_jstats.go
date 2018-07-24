package main

import (
	"encoding/json"
	"os"

	l "git.fractalqb.de/fractalqb/qblog"
)

func jstatRead(file string) (map[string]interface{}, error) {
	f, err := os.Open(file)
	if err != nil {
		log.Logf(l.Error, "cannot open stat file '%s': %s", file, err)
		return nil, err
	}
	defer f.Close()
	res := make(map[string]interface{})
	dec := json.NewDecoder(f)
	err = dec.Decode(&res)
	if err != nil {
		log.Logf(l.Error, "cannot parse stat file '%s': %s", file, err)
		return nil, err
	}
	return res, nil
}

func jstatMarket(statFile string) {
	stat, err := jstatRead(statFile)
	if err == nil && theEddn != nil {
		go eddnSendCommodities(theEddn, stat)
	}
}

func jstatStatus(statFile string) {
	_, err := jstatRead(statFile)
	if err != nil {
		return
	}
	// TODO update location
	//	if  theEddn != nil {
	//		go eddnSendCommodities(theEddn, stat)
	//	}
}
