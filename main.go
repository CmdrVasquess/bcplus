package main

import (
	_ "expvar"
	"flag"
	"fmt"
	_ "net/http/pprof"
	"os"

	"github.com/CmdrVasquess/BCplus/core"
)

//go:generate versioner -bno build_no -pkg core -p BCp -t Date ./VERSION ./core/version.go

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintln(os.Stderr, "For more information, see:")
	fmt.Fprintln(os.Stderr, "\thttps://cmdrvasquess.github.io/BCplus/")
	fmt.Fprintln(os.Stderr, "Flags:")
	flag.PrintDefaults()
}

func main() {
	fmt.Println(core.AppDesc)
	flag.StringVar(&core.FlagJDir, "j", core.DefaultJournalDir(), "Game directory with journal files")
	flag.StringVar(&core.FlagDDir, "d", core.DefaultDataDir(), core.AppNameShort+" data directory")
	flag.StringVar(&core.FlagEddn, "eddn", "",
		`Send events to EDDN. Select one of:
- off     : dont send data to EDDN
- anon    : send as 'anonymous'
- scramble: send as a unique, persistent id not derived from commander name
- cmdr    : send with commander name
- test    : send to test schema with scrambled uploader`)
	flag.UintVar(&core.FlagWuiPort, "p", 1337, "port number for the web ui")
	flag.StringVar(&core.FlagMacros, "macros", "", "use macro file")
	flag.StringVar(&core.FlagTheme, "theme", "dark", "select web theme")
	flag.BoolVar(&core.LogV, "v", false, "Log verbose (aka debug level)")
	flag.BoolVar(&core.LogVV, "vv", false, "Log very verbose (aka trace level)")
	flag.Usage = usage
	flag.Parse()
	core.FlagLogLevel()
	core.FlagCheckEddn()
	core.Run()
}
