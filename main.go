package main

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/CmdrVasquess/BCplus/cmdr"
	"github.com/CmdrVasquess/BCplus/galaxy"
	eddn "github.com/CmdrVasquess/goEDDNc"
	"github.com/CmdrVasquess/watched"
	l "git.fractalqb.de/fractalqb/qblog"
)

//go:generate versioner -bno build_no -p BCp -t Date ./VERSION ./version.go
var appDesc string

func init() {
	appDesc = fmt.Sprintf(appNameLong+" v%d.%d.%d%s / %s on %s (#%d %s)",
		BCpMajor,
		BCpMinor,
		BCpBugfix,
		BCpQuality,
		runtime.Version(),
		runtime.GOOS,
		BCpBuildNo,
		BCpDate)
	bcpRoot, _ = filepath.Abs(os.Args[0])
	bcpRoot = filepath.Dir(bcpRoot)
}

func resFile(resName string) string {
	return filepath.Join(bcpRoot, "res", resName)
}

type State struct {
	LastEDEvent time.Time
}

func (s *State) Clear() {
	s.LastEDEvent = time.Time{}
}

func stateFileName() string {
	res := filepath.Join(flagDDir, "bcplus.json")
	return res
}

func (s *State) Save(filename string) error {
	log.Logf(l.Info, "save BC+ state to '%s'", filename)
	tmpnm := filename + "~"
	f, err := os.Create(tmpnm)
	if err != nil {
		return err
	}
	defer func() {
		if f != nil {
			f.Close()
		}
		f = nil
	}()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "\t")
	err = enc.Encode(s)
	if err != nil {
		return err
	}
	err = f.Close()
	f = nil
	if err != nil {
		return err
	}
	err = os.Rename(tmpnm, filename)
	return err
}

func (s *State) Load(filename string) error {
	log.Logf(l.Info, "load BC+ state from '%s'", filename)
	f, err := os.Open(filename)
	if os.IsNotExist(err) {
		s.Clear()
		log.Logf(l.Warn, "BC+ state '%s' not exists", filename)
		return nil
	} else if err != nil {
		return err
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(s)
	return err
}

type evtSource rune

type bcpEvent struct {
	source evtSource
	data   interface{}
}

func eventLoop() {
	log.Log(l.Info, "running bc+ event loop…")
	for evt := range bcpEventQ {
		log.Logf(l.Trace, "bc+ event from '%c': %v", evt.source, evt.data)
		switch evt.source {
		case watched.EscrJournal:
			journalEvent(evt.data.([]byte))
		case watched.EscrStatus:
			jstatStatus(evt.data.(string))
		case watched.EscrMarket:
			jstatMarket(evt.data.(string))
		}
	}
}

func goWatchingJournals(jDir string, bcpEvents chan<- bcpEvent) *watched.JournalDir {
	res := &watched.JournalDir{
		Dir: jDir,
		PerJLine: func(line []byte) {
			cpy := make([]byte, len(line))
			copy(cpy, line)
			bcpEvents <- bcpEvent{watched.EscrJournal, cpy}
		},
		OnStatChg: func(tag rune, statFile string) {
			bcpEvents <- bcpEvent{evtSource(tag), statFile}
		},
		Quit: make(chan bool),
	}
	startWith := spoolJouranls(jDir, bcpState.LastEDEvent)
	go res.Watch(startWith)
	return res
}

func openGalaxy() *galaxy.Repo {
	res, err := galaxy.NewRepo(filepath.Join(flagDDir, "galaxy.db"))
	if os.IsNotExist(err) {
		fnm := resFile("galaxy/create-sqlite.sql")
		log.Logf(l.Info, "init galaxy DB from '%s'", fnm)
		err := res.RunSql(fnm)
		if err != nil {
			log.Panic(err)
		}
	}
	gxyv, err := res.Version()
	if err != nil {
		log.Panic(err)
	}
	log.Logf(l.Info, "galaxy DB version: %d", gxyv)
	return res
}

func cmdrDir(name string) string {
	if len(name) == 0 {
		log.Log(l.Warn, "empty commander name for directory")
		name = "_anonymous_"
	} else {
		name = strings.Replace(name, " ", "_", -1)
	}
	res := filepath.Join(flagDDir, name)
	if _, err := os.Stat(res); os.IsNotExist(err) {
		log.Logf(l.Debug, "create commander dir '%s'", res)
		err = os.Mkdir(res, 0777)
		if err != nil {
			log.Logf(l.Warn, "cannot create commander dir '%s'", res)
		}
	}
	return res
}

func cmdrStateFile(name string) string {
	res := cmdrDir(name)
	res = filepath.Join(res, "state.json")
	return res
}

func switchToCommander(name string) {
	if theCmdr != nil {
		if theCmdr.Name == name {
			return
		}
		err := theCmdr.Save(cmdrStateFile(theCmdr.Name))
		if err != nil {
			log.Log(l.Error, "error while saving commander state:", err)
		}
	}
	theGalaxy.ClearCache()
	if len(name) > 0 {
		theCmdr = cmdr.NewState(theCmdr)
		err := theCmdr.Load(cmdrStateFile(name))
		if os.IsNotExist(err) {
			theCmdr.Name = name
			err = nil
		} else if err != nil {
			log.Logf(l.Error, "cannot switch to commander '%s': %s", name, err)
			theCmdr = nil
			return
		}
		flagCheckEddn()
		var upldr string
		switch eddnMode {
		case "cmdr":
			upldr = theCmdr.Name
		case "scramble", "test":
			md5sum := md5.Sum([]byte(theCmdr.Name))
			upldr = base64.StdEncoding.
				WithPadding(base64.NoPadding).
				EncodeToString(md5sum[:])
			upldr = "Src" + upldr
		case "anon":
			upldr = "anonymous"
		default:
			theEddn = nil
			return
		}
		vstr := fmt.Sprintf("%d.%d.%d%s", BCpMajor, BCpMinor, BCpBugfix, BCpQuality)
		theEddn = &eddn.Upload{Vaildate: true, TestUrl: eddnMode == "test"}
		theEddn.Http.Timeout = 5 * time.Second
		theEddn.Header.Uploader = upldr
		theEddn.Header.SwName = appNameShort
		theEddn.Header.SwVersion = vstr
		if err != nil {
			log.Log(l.Warn, "EDDN connect failed:", err)
			theEddn = nil
		} else {
			log.Logf(l.Info, "connected to EDDN as %s / %s / %s", upldr, appNameShort, vstr)
		}
	} else {
		theCmdr = nil
		theEddn = nil
		log.Log(l.Info, "disconnected from EDDN")
	}
}

const (
	appNameLong     = "BoardComputer+"
	appNameShort    = "BC+"
	flagEddnDefault = "off"
	flagEddnOff     = "off"
)

var (
	flagJDir  string
	flagDDir  string
	flagEddn  string
	bcpRoot   string
	bcpEventQ = make(chan bcpEvent, 128)
	bcpState  State
	theGalaxy *galaxy.Repo
	theCmdr   *cmdr.State
	theEddn   *eddn.Upload
	eddnMode  string
	stateLock sync.RWMutex
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintln(os.Stderr, "For more information, see:")
	fmt.Fprintln(os.Stderr, "\thttps://cmdrvasquess.github.io/BCplus/")
	fmt.Fprintln(os.Stderr, "Flags:")
	flag.PrintDefaults()
}

func flagCheckEddn() {
	switch flagEddn {
	case "":
		if theCmdr != nil && len(theCmdr.EddnMode) > 0 {
			eddnMode = theCmdr.EddnMode
		} else {
			eddnMode = flagEddnDefault
		}
	case "off", "anon", "scramble", "cmdr", "test":
		eddnMode = flagEddn
		if theCmdr != nil {
			theCmdr.EddnMode = flagEddn
		}
	default:
		log.Fatalf("illegal EDDN mode: %s", eddnMode)
	}
}

func main() {
	fmt.Println(appDesc)
	flag.StringVar(&flagJDir, "j", defaultJournalDir(), "Game directory with journal files")
	flag.StringVar(&flagDDir, "d", defaultDataDir(), appNameShort+" data directory")
	flag.StringVar(&flagEddn, "eddn", "", "Send to EDDN {off, anon, scramble, cmdr, test}")
	flag.BoolVar(&logV, "v", false, "Log verbose (aka debug level)")
	flag.BoolVar(&logVV, "vv", false, "Log very verbose (aka trace level)")
	flag.Usage = usage
	flag.Parse()
	flagLogLevel()
	flagCheckEddn()
	var err error
	log.Logf(l.Debug, "BC+ root: '%s'", bcpRoot)
	bcpState.Load(stateFileName())
	theGalaxy = openGalaxy()
	goWatchingJournals(flagJDir, bcpEventQ)
	go eventLoop()

	// up & running – wait for Ctrl-C…
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	<-signals
	log.Log(l.Info, "BC+ interrupted with Ctrl-C, shutting down…")
	log.Log(l.Info, "closing galaxy repo")
	theGalaxy.Close()
	if theCmdr != nil {
		err = theCmdr.Save(cmdrStateFile(theCmdr.Name))
		if err != nil {
			log.Log(l.Error, "error while saving commander state:", err)
		}
	}
	err = bcpState.Save(stateFileName())
	if err != nil {
		log.Log(l.Error, "error while saving BC+ state:", err)
	}
	log.Log(l.Info, "Fly safe commander! o7")
}
