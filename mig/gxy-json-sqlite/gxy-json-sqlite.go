package main

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"

	gxy "github.com/CmdrVasquess/BCplus/galaxy"
)

func addMaterial(part *gxy.SysPart, nm string, freq float64) {
	part.AddResource(&gxy.Resource{
		Name:    nm,
		OccFreq: float32(freq),
	})
}

func addSystem(jsys map[string]interface{}, repo *gxy.Repo) {
	coos := jsys["Coos"].([]interface{})
	sys := &gxy.System{
		Name: jsys["Name"].(string),
		Coos: gxy.Vec3D{
			coos[0].(float64),
			coos[1].(float64),
			coos[2].(float64),
		},
	}
	_, err := repo.PutSystem(sys)
	if err != nil {
		log.Print(err)
		return
	}
	if tmp, ok := jsys["Bodies"]; ok {
		bdys := tmp.([]interface{})
		for _, tmp := range bdys {
			var part *gxy.SysPart
			jbdy := tmp.(map[string]interface{})
			switch int(jbdy["Category"].(float64)) {
			case 1:
				part = &gxy.SysPart{
					Type:       gxy.Star,
					Name:       jbdy["Name"].(string),
					FromCenter: int(jbdy["Dist"].(float64)),
				}
			case 2:
				part = &gxy.SysPart{
					Type:       gxy.Planet,
					Name:       jbdy["Name"].(string),
					FromCenter: int(jbdy["Dist"].(float64)),
				}
			}
			if part != nil {
				part.Height = gxy.NaN32
				part.Lat = gxy.NaN32
				part.Lon = gxy.NaN32
				_, err = sys.AddPart(part)
				if err != nil {
					log.Print(err)
					return
				}
			}
			if tmp, ok := jbdy["Materials"]; ok {
				matls := tmp.(map[string]interface{})
				for mat, tmp := range matls {
					addMaterial(part, mat, tmp.(float64))
				}
			}
		}
	}
}

func importGalaxy(jsonNm string, repo *gxy.Repo) {
	jf, err := os.Open(jsonNm)
	if err != nil {
		panic(err)
	}
	defer jf.Close()
	jdec := json.NewDecoder(jf)
	xa := repo.XaBegin()
	defer xa.Rollback()
	count := 0
	for {
		jsys := make(map[string]interface{})
		if err := jdec.Decode(&jsys); err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
		addSystem(jsys, repo)
		if count == batchSize {
			count = 0
			xa.Commit()
			xa = repo.XaBegin()
		} else {
			count++
		}
	}
	xa.Commit()
}

var batchSize int

func main() {
	flag.IntVar(&batchSize, "batch", 100, "batch size for DB transactions")
	creSql := flag.String("create", "", "run create script")
	flag.Parse()
	jsonNm := flag.Args()[0]
	dbNm := flag.Args()[1]
	repo, err := gxy.NewRepo(dbNm)
	if err != nil && !os.IsNotExist(err) {
		panic(err)
	}
	defer repo.Close()
	if len(*creSql) > 0 {
		err := repo.RunSql(*creSql)
		if err != nil {
			panic(err)
		}
	}
	importGalaxy(jsonNm, repo)
}
