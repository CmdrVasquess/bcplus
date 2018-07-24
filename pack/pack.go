package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"git.fractalqb.de/fractalqb/pack"
)

//go:generate versioner -p BCp -t Date ../VERSION ./vbcp.go

var runDir = flag.String("C", "", "change to directory before packing")
var outDir = flag.String("d", ".", "target directory")
var packType = flag.String("pack", "", "pack archive from distro dir: zip")
var distNm string

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func purgeFilter(dir string, info os.FileInfo) bool {
	name := info.Name()
	switch {
	case strings.HasSuffix(name, ".go"):
		return false
	case strings.HasSuffix(name, "~"):
		return false
	case strings.HasPrefix(name, "."):
		return false
	}
	return true
}

func dist() string {
	distDir := filepath.Join(*outDir, distNm)
	if _, err := os.Stat(distDir); !os.IsNotExist(err) {
		must(os.RemoveAll(distDir))
	}
	must(os.Mkdir(distDir, 0777))
	must(pack.CopyToDir(distDir,
		"LICENSE",
		"README.md",
		"VERSION",
		"BCplus",
		"jreplay/jreplay",
	))
	must(pack.CopyTree(distDir, "res", purgeFilter))
	must(pack.CopyTree(distDir, "mig", purgeFilter))
	return distDir
}

func main() {
	flag.Parse()
	var err error
	*outDir, err = filepath.Abs(*outDir)
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
	distNm = fmt.Sprintf("BCplus-%d.%d.%d%s", BCpMajor, BCpMinor, BCpBugfix, BCpQuality)
	distDir := dist()
	switch *packType {
	case "zip":
		must(pack.ZipDist(distDir+".zip", distNm, distDir))
	default:
		if len(*packType) > 0 {
			log.Fatalf("unsupported archive type '%s'", *packType)
		}
	}
}
