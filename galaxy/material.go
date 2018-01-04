package galaxy

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"

	"git.fractalqb.de/xsx"
	xsxtab "git.fractalqb.de/xsx/table"
)

type Material struct {
	JName   string // Name from CMDR's Journal
	Kind    uint8
	Commons int8
}

type MatDemand map[string]int

type Synthesis struct {
	Name      string
	Improves  string
	LvlDemand []MatDemand
}

func loadMaterials(dataDir string) (res map[string]Material, err error) {
	res = make(map[string]Material)
	inf, err := os.Open(filepath.Join(dataDir, "materials.xsx"))
	if err != nil {
		return nil, err
	}
	defer inf.Close()
	xrd := xsx.NewPullParser(bufio.NewReader(inf))
	tDef, err := xsxtab.ReadDef(xrd)
	if err != nil {
		return nil, err
	}
	colKind := tDef.ColIndex("kind")
	if colKind < 0 {
		glog.Fatal("galaxy loading material data: no column 'kind'")
	}
	colJournal := tDef.ColIndex("journal")
	if colJournal < 0 {
		glog.Fatal("galaxy loading material data: no column 'journal'")
	}
	colCmns := tDef.ColIndex("commonness")
	if colCmns < 0 {
		glog.Fatal("galaxy loading material data: no column 'commonness'")
	}
	// TODO evaluate definition(?)
	for row, err := tDef.NextRow(xrd, nil); row != nil; row, err = tDef.NextRow(xrd, row) {
		if err != nil {
			return nil, err
		}
		r, _ := utf8.DecodeRune([]byte(row[colKind]))
		kind := strings.IndexRune("rmd", r)
		if kind < 0 {
			glog.Fatalf("galaxy loading material data: unknown kind '%s'", row[colKind])
		}
		cmns := -1
		if row[colCmns] != "_" {
			cmns, err = strconv.Atoi(row[colCmns])
			if err != nil {
				glog.Fatalf("galaxy loading material data: comminness '%s'", row[colCmns])
			}
		}
		mat := Material{
			JName:   row[colJournal],
			Kind:    uint8(kind),
			Commons: int8(cmns)}
		res[row[colJournal]] = mat
	}
	return res, nil
}
