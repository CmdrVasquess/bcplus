package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"git.fractalqb.de/fractalqb/goxic"
	"git.fractalqb.de/fractalqb/goxic/html"
	"git.fractalqb.de/fractalqb/goxic/js"
	"github.com/CmdrVasquess/bcplus/internal/app"
)

var (
	assetDir string
	parser   = html.NewParser()

	tScreen app.Screen
)

func init() {
	parser.CxfMap = map[string]goxic.CntXformer{
		"xml":   html.EscWrap,
		"jsStr": js.EscWrap,
	}
}

func loadFrames() {
	tmap := make(map[string]*goxic.Template)
	err := parser.ParseFile(filepath.Join(assetDir, "goxic/screen.html"), "", tmap)
	if err != nil {
		log.Fatal(err)
	}
	goxic.MustIndexMap(&tScreen, tmap[""], false, app.GxName.Convert)
}

func main() {
	flag.StringVar(&assetDir, "a", ".", "assert dir")
	flag.Parse()
	loadFrames()
	tScreen.NewInitBounT(goxic.Empty, nil).Emit(os.Stdout)
}
