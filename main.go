package main

import (
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

	l "git.fractalqb.de/fractalqb/qblog"
	"github.com/CmdrVasquess/BCplus/cmdr"
	"github.com/CmdrVasquess/BCplus/galaxy"
	"github.com/CmdrVasquess/BCplus/webui"
	eddn "github.com/CmdrVasquess/goEDDNc"
	"github.com/CmdrVasquess/watched"
	"github.com/gofrs/uuid"
)

//go:generate versioner -bno build_no -p BCp -t Date ./VERSION ./version.go
var appDesc, appVersion string

func init() {
	appVersion = fmt.Sprintf("%d.%d.%d%s",
		BCpMajor,
		BCpMinor,
		BCpBugfix,
		BCpQuality,
	)
	appDesc = fmt.Sprintf(appNameLong+" v%s / %s on %s (#%d %s)",
		appVersion,
		runtime.Version(),
		runtime.GOOS,
		BCpBuildNo,
		BCpDate,
	)
	bcpRoot, _ = filepath.Abs(os.Args[0])
	bcpRoot = filepath.Dir(bcpRoot)
	bcpState.Commanders = make(map[string]time.Time)
}

func resFile(resName string) string {
	return filepath.Join(bcpRoot, "res", resName)
}

type State struct {
	Version struct {
		Major, Minor, Bugfix int
		Quality              string
		Build                int
	}
	LastEDEvent time.Time
	Commanders  map[string]time.Time
}

func (s *State) Clear() {
	s.LastEDEvent = time.Time{}
}

func stateFileName() string {
	res := filepath.Join(flagDDir, "bcplus.json")
	return res
}

func (s *State) Save(filename string) error {
	log.Logf(l.Linfo, "save BC+ state to '%s'", filename)
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
	log.Logf(l.Linfo, "load BC+ state from '%s'", filename)
	f, err := os.Open(filename)
	if os.IsNotExist(err) {
		s.Clear()
		log.Logf(l.Lwarn, "BC+ state '%s' not exists", filename)
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
	log.Log(l.Linfo, "running bc+ event loop…")
	for evt := range bcpEventQ {
		log.Logf(l.Ltrace, "bc+ event from '%c': %v", evt.source, evt.data)
		switch evt.source {
		case watched.EscrJournal:
			journalEvent(evt.data.([]byte))
		case watched.EscrStatus:
			jstatStatus(evt.data.(string))
		case watched.EscrMarket:
			jstatMarket(evt.data.(string))
		case watched.EscrShipyard:
			jstatShipyard(evt.data.(string))
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
		log.Logf(l.Linfo, "init galaxy DB from '%s'", fnm)
		err := res.RunSql(fnm)
		if err != nil {
			log.Panic(err)
		}
	}
	gxyv, err := res.Version()
	if err != nil {
		log.Panic(err)
	}
	log.Logf(l.Linfo, "galaxy DB version: %d", gxyv)
	return res
}

func cmdrDir(name string) string {
	if len(name) == 0 {
		log.Log(l.Lwarn, "empty commander name for directory")
		name = "_anonymous_"
	} else {
		name = strings.Replace(name, " ", "_", -1)
	}
	res := filepath.Join(flagDDir, name)
	if _, err := os.Stat(res); os.IsNotExist(err) {
		log.Logf(l.Ldebug, "create commander dir '%s'", res)
		err = os.Mkdir(res, 0777)
		if err != nil {
			log.Logf(l.Lwarn, "cannot create commander dir '%s'", res)
		}
	}
	return res
}

const (
	cmdrState = "state.json"
)

func cmdrFile(cmdr, file string) string {
	res := cmdrDir(cmdr)
	res = filepath.Join(res, file)
	dir := filepath.Dir(res)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0777)
	}
	return res
}

func switchToCommander(name string) {
	if theCmdr != nil {
		if theCmdr.Name == name {
			return
		}
		err := theCmdr.Save(cmdrFile(theCmdr.Name, cmdrState))
		if err != nil {
			log.Log(l.Lerror, "error while saving commander state:", err)
		}
		bcpState.Commanders[theCmdr.Name] = bcpState.LastEDEvent
	}
	theGalaxy.ClearCache()
	if len(name) > 0 {
		theCmdr = cmdr.NewState(theCmdr)
		err := theCmdr.Load(cmdrFile(name, cmdrState))
		if os.IsNotExist(err) {
			theCmdr.Name = name
			err = nil
		} else if err != nil {
			log.Logf(l.Lerror, "cannot switch to commander '%s': %s", name, err)
			theCmdr = nil
			return
		}
		if len(theCmdr.Scrambled) == 0 {
			anon, _ := uuid.NewV4() // TODO what to do on error?
			theCmdr.Scrambled = base64.RawURLEncoding.EncodeToString(anon.Bytes())
		}
		flagCheckEddn()
		bcpState.Commanders[theCmdr.Name] = bcpState.LastEDEvent
		var upldr string
		switch eddnMode {
		case "cmdr":
			upldr = theCmdr.Name
		case "scramble", "test":
			upldr = theCmdr.Scrambled
		case "anon":
			upldr = "anonymous"
		default:
			theEddn = nil
			return
		}
		vstr := fmt.Sprintf("%d.%d.%d%s", BCpMajor, BCpMinor, BCpBugfix, BCpQuality)
		theEddn = &eddn.Upload{Vaildate: true, TestUrl: eddnMode == "test"}
		theEddn.Http.Timeout = eddnTimeout
		theEddn.Header.Uploader = upldr
		theEddn.Header.SwName = appNameShort
		theEddn.Header.SwVersion = vstr
		if err != nil {
			log.Log(l.Lwarn, "EDDN connect failed:", err)
			theEddn = nil
		} else {
			log.Logf(l.Linfo, "connected to EDDN as %s / %s / %s", upldr, appNameShort, vstr)
		}
	} else {
		theCmdr = nil
		theEddn = nil
		log.Log(l.Linfo, "disconnected from EDDN")
	}
}

const (
	appNameLong     = "BoardComputer+"
	appNameShort    = "BC+"
	flagEddnDefault = "off"
	flagEddnOff     = "off"
	eddnTimeout     = 8 * time.Second
)

var (
	flagJDir    string
	flagDDir    string
	flagEddn    string
	flagWuiPort uint
	flagMacros  string
	bcpRoot     string
	bcpEventQ   = make(chan bcpEvent, 128)
	bcpState    State
	theGalaxy   *galaxy.Repo
	theCmdr     *cmdr.State
	theEddn     *eddn.Upload
	eddnMode    string
	stateLock   sync.RWMutex
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
	flag.StringVar(&flagEddn, "eddn", "",
		`Send events to EDDN. Select one of:
- off     : dont send data to EDDN
- anon    : send as 'anonymous'
- scramble: send as a unique, persistent id not derived from commander name
- cmdr    : send with commander name
- test    : send to test schema with scrambled uploader`)
	flag.UintVar(&flagWuiPort, "p", 1337, "port number for the web ui")
	flag.StringVar(&flagMacros, "macros", "", "use macro file")
	flag.BoolVar(&logV, "v", false, "Log verbose (aka debug level)")
	flag.BoolVar(&logVV, "vv", false, "Log very verbose (aka trace level)")
	flag.Usage = usage
	flag.Parse()
	flagLogLevel()
	flagCheckEddn()
	var err error
	log.Logf(l.Ldebug, "BC+ root: '%s'", bcpRoot)
	bcpState.Load(stateFileName())
	bcpState.Version.Major = BCpMajor
	bcpState.Version.Minor = BCpMinor
	bcpState.Version.Bugfix = BCpBugfix
	bcpState.Version.Quality = BCpQuality
	bcpState.Version.Build = BCpBuildNo
	theGalaxy = openGalaxy()
	if len(flagMacros) > 0 {
		loadMacros(flagMacros)
	}
	goWatchingJournals(flagJDir, bcpEventQ)
	go eventLoop()
	webui.Run(&webui.Init{
		DataDir:     flagDDir,
		ResourceDir: filepath.Join(bcpRoot, "res"),
		CommonName:  appNameLong,
		Port:        flagWuiPort,
		BCpVersion:  appVersion,
		StateLock:   &stateLock,
		CmdrGetter:  func() *cmdr.State { return theCmdr },
	})
	// up & running – wait for Ctrl-C…
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	<-signals
	log.Log(l.Linfo, "BC+ interrupted with Ctrl-C, shutting down…")
	log.Log(l.Linfo, "closing galaxy repo")
	theGalaxy.Close()
	if theCmdr != nil {
		err = theCmdr.Save(cmdrFile(theCmdr.Name, cmdrState))
		if err != nil {
			log.Log(l.Lerror, "error while saving commander state:", err)
		}
	}
	err = bcpState.Save(stateFileName())
	if err != nil {
		log.Log(l.Lerror, "error while saving BC+ state:", err)
	}
	log.Log(l.Linfo, "Fly safe commander! o7")
}
