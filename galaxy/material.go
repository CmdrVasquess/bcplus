package galaxy

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"

	l "github.com/fractalqb/qblog"
	"github.com/fractalqb/xsx"
	"github.com/fractalqb/xsx/gem"
	xsxtab "github.com/fractalqb/xsx/table"
)

//go:generate stringer -type MatCategory
type MatCategory uint8

const (
	Raw MatCategory = iota
	Man
	Enc
	MCatUndef MatCategory = 255
)

type Material struct {
	JName    string // Name from CMDR's Journal
	Category MatCategory
	Commons  int8
}

type SynthLevel struct {
	Bonus  string
	Demand map[string]uint
}

type Synthesis struct {
	Name     string
	Improves string
	Levels   []SynthLevel
}

func loadMaterials(dataDir string) (res map[string]Material, err error) {
	res = make(map[string]Material)
	matFile := filepath.Join(dataDir, "materials.xsx")
	log.Logf(l.Info, "loading materials from: %s", matFile)
	inf, err := os.Open(matFile)
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
		log.Fatal("galaxy loading material data: no column 'kind'")
	}
	colJournal := tDef.ColIndex("journal")
	if colJournal < 0 {
		log.Fatal("galaxy loading material data: no column 'journal'")
	}
	colCmns := tDef.ColIndex("commonness")
	if colCmns < 0 {
		log.Fatal("galaxy loading material data: no column 'commonness'")
	}
	// TODO evaluate definition(?)
	for row, err := tDef.NextRow(xrd, nil); row != nil; row, err = tDef.NextRow(xrd, row) {
		if err != nil {
			return nil, err
		}
		str := row[colKind].(*gem.Atom).Str
		r, _ := utf8.DecodeRune([]byte(str))
		kind := strings.IndexRune("rmd", r)
		if kind < 0 {
			log.Fatalf("galaxy loading material data: unknown kind '%s'", str)
		}
		cmns := -1
		str = row[colCmns].(*gem.Atom).Str
		if str != "_" {
			cmns, err = strconv.Atoi(str)
			if err != nil {
				log.Fatalf("galaxy loading material data: comminness '%s'", str)
			}
		}
		str = row[colJournal].(*gem.Atom).Str
		mat := Material{
			JName:    str,
			Category: MatCategory(kind),
			Commons:  int8(cmns)}
		res[str] = mat
	}
	log.Logf(l.Info, "loaded %d materials", len(res))
	return res, nil
}

func loadSynth(dataDir string, res *([]Synthesis)) (err error) {
	synFile := filepath.Join(dataDir, "synth.json")
	log.Logf(l.Info, "loading synthesis from: %s", synFile)
	inf, err := os.Open(synFile)
	if err != nil {
		return err
	}
	defer inf.Close()
	jdec := json.NewDecoder(inf)
	err = jdec.Decode(res)
	if err == nil {
		log.Logf(l.Info, "loaded %d synthesis recipes", len(*res))
	}
	return err
}
