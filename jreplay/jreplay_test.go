package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"
)

const tst_dirnm = "jreplay-test.d"
const tstdir_perm = 0777

var srcdir string = filepath.Join(tst_dirnm, "src")
var dstdir string = filepath.Join(tst_dirnm, "dst")

func newSrcFile(nm string, lineNo int) (newFileName string) {
	newFileName = filepath.Join(srcdir, nm)
	f, _ := os.Create(newFileName)
	defer f.Close()
	for i := 1; i <= lineNo; i++ {
		fmt.Fprintf(f, "line %d\n", i)
	}
	return
}

func compare(fnm string) error {
	ofNm := filepath.Join(srcdir, fnm)
	cfNm := filepath.Join(dstdir, fnm)
	ofSt, _ := os.Stat(ofNm)
	cfSt, _ := os.Stat(cfNm)
	if ofSt.Size() != cfSt.Size() {
		errors.New(fmt.Sprintf("sized differ for %s: orig=%d / copy=%d",
			ofSt.Size(),
			cfSt.Size()))
	}
	of, _ := os.Open(ofNm)
	defer of.Close()
	cf, _ := os.Open(cfNm)
	defer cf.Close()
	const BUF_SZ = 1024
	oBuf, cBuf := make([]byte, BUF_SZ), make([]byte, BUF_SZ)
	pos := 0
	for ordc, err := of.Read(oBuf); err != io.EOF; ordc, err = of.Read(oBuf) {
		crdc, _ := cf.Read(cBuf)
		if ordc != crdc {
			return errors.New(
				fmt.Sprintf("byte read: orig=%d / copy=%d", ordc, crdc))
		}
		for i := 0; i < ordc; i++ {
			if oBuf[i] != cBuf[i] {
				return errors.New(
					fmt.Sprint("content differs @byte %d", pos))
			}
			pos++
		}
	}
	return nil
}

func TestReplay(t *testing.T) {
	if _, err := os.Stat(tst_dirnm); !os.IsNotExist(err) {
		log.Printf("test dir '%s' exists => purge!", tst_dirnm)
		os.RemoveAll(tst_dirnm)
	}
	os.MkdirAll(srcdir, tstdir_perm)
	os.Mkdir(dstdir, tstdir_perm)
	fnm := newSrcFile("Journal.log", 10)
	replay(fnm, dstdir)
	err := compare("Journal.log")
	if err != nil {
		t.Error(err)
	}
}
