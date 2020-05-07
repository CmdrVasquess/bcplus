package edpc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/CmdrVasquess/bcplus/itf"

	"git.fractalqb.de/fractalqb/qbsllm"
	"github.com/CmdrVasquess/bcplus/internal/common"
)

const (
	idxFile = "index.json"
)

var (
	log    = qbsllm.New(qbsllm.Lnormal, "edpc", nil, nil)
	LogCfg = qbsllm.Config(log)
)

type Stub struct {
	dir       string
	LocUpdate chan itf.Location
}

func (s *Stub) Init() error {
	s.LocUpdate = make(chan itf.Location, 8)
	go s.locUpdater()
	return nil
}

func (s *Stub) SetCmdr(fid string, cmdrDir string) error {
	if stat, err := os.Stat(cmdrDir); os.IsNotExist(err) {
		log.Infoa("init `edpc dir`", cmdrDir)
		os.MkdirAll(cmdrDir, common.DirFileMode)
		wr, err := os.Create(filepath.Join(cmdrDir, idxFile))
		if err != nil {
			return err
		}
		defer wr.Close()
		enc := json.NewEncoder(wr)
		if err = enc.Encode(Data{}); err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else if !stat.IsDir() {
		return fmt.Errorf("not a directory: %s", cmdrDir)
	}
	s.dir = cmdrDir
	return nil
}

func (s *Stub) ListStories() (res []Story, err error) {
	rd, err := os.Open(filepath.Join(s.dir, idxFile))
	if err != nil {
		return nil, err
	}
	defer rd.Close()
	dec := json.NewDecoder(rd)
	var d Data
	if err = dec.Decode(&d); err != nil {
		return nil, err
	}
	return d.Stories, nil
}

func (s *Stub) locUpdater() {
	log.Infos("running location updater")
	defer log.Infos("location updater terminated")
	httpClt := http.Client{Timeout: 5 * time.Second}
	for loc := range s.LocUpdate {
		locpath := []string{url.PathEscape(loc.SysName)}
		log.Tracea("update `location`", loc)
		if loc.RefType == itf.NoRefType {
			locpath = append(locpath, url.PathEscape(loc.Mode.String()))
		}
		pstr := path.Join(locpath...)
		resp, err := httpClt.Head(pstr)
		if err == nil {
			log.Infoa("discovered hint `at`", pstr)
			resp.Body.Close()
		} else {
			log.Debuga("no hint `at`: `req-err`", pstr, err)
		}
	}
}
