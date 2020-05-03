package edpc

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

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
	dir string
}

func (s *Stub) Init() error {
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
