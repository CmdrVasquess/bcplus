package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"git.fractalqb.de/fractalqb/pack"
	"github.com/CmdrVasquess/bcplus/internal/common"
)

var (
	runDir     = flag.String("C", "", "change to directory before packing")
	outDir     = flag.String("d", ".", "target directory")
	packType   = flag.String("pack", "", "pack archive from distro dir: zip")
	ddapAction = flag.String("ddap", "", "Action for Dist Directory after packing")
	distNm     string
)

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func purgeFilter(dir string, info os.FileInfo) bool {
	skip := false
	if !(info.IsDir() || info.Mode().IsRegular()) {
		skip = true
	}
	name := info.Name()
	switch {
	case strings.HasSuffix(name, ".go"):
		skip = true
	case strings.HasSuffix(name, "~"):
		skip = true
	case strings.HasPrefix(name, "."):
		skip = true
	}
	if skip {
		log.Printf("SKIP: %s/%s", dir, info.Name())
	}
	return !skip
}

func dist() string {
	distDir := filepath.Join(*outDir, distNm)
	if _, err := os.Stat(distDir); !os.IsNotExist(err) {
		must(os.RemoveAll(distDir))
	}
	must(os.Mkdir(distDir, 0777))
	must(pack.CopyToDir(distDir, pack.OsDepExe,
		"bcplus", "util/screenshot/screenshot"))
	must(pack.CopyToDir(distDir, nil,
		//"LICENSE",
		//"README.md",
		"VERSION",
	))
	must(pack.CopyTree(distDir, "assets", nil, purgeFilter))
	basenm := filepath.Join(distDir, "assets/s/js/vue")
	must(os.Rename(basenm+".min.js", basenm+".js"))
	return distDir
}

func main() {
	flag.Parse()
	var err error
	*outDir, err = filepath.Abs(*outDir)
	if err != nil {
		log.Fatal(err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	if len(*runDir) > 0 {
		log.Printf("switching run directory to '%s'", *runDir)
		err := os.Chdir(*runDir)
		if err != nil {
			log.Fatal(err)
		}
	}
	distNm = fmt.Sprintf("BCplus-%d.%d.%d%s",
		common.BCpMajor,
		common.BCpMinor,
		common.BCpPatch,
		common.BCpQuality)
	distDir := dist()
	switch *packType {
	case "zip":
		must(pack.ZipDist(distDir+".zip", distNm, distDir))
	default:
		if len(*packType) > 0 {
			log.Fatalf("unsupported archive type '%s'", *packType)
		}
	}
	err = os.Chdir(cwd)
	if err != nil {
		log.Fatal(err)
	}
	if len(*packType) > 0 {
		switch {
		case len(*ddapAction) == 0:
			log.Printf("remove dist dir '%s'", distDir)
			must(os.RemoveAll(distDir))
		case *ddapAction != ".":
			log.Printf("rename dist dir '%s' → '%s'", distDir, *ddapAction)
			if _, err := os.Stat(*ddapAction); !os.IsNotExist(err) {
				must(os.RemoveAll(*ddapAction))
			}
			must(os.Rename(distDir, *ddapAction))
		}
	}
}
