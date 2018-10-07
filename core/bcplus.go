package core

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/gofrs/uuid"

	"git.fractalqb.de/fractalqb/ggja"
	l "git.fractalqb.de/fractalqb/qblog"

	"github.com/CmdrVasquess/BCplus/cmdr"
	"github.com/CmdrVasquess/BCplus/common"
	"github.com/CmdrVasquess/BCplus/galaxy"
	"github.com/CmdrVasquess/BCplus/webui"
	"github.com/CmdrVasquess/goEDDNc"
	"github.com/CmdrVasquess/watched"
)

var AppDesc, AppVersion string

func init() {
	AppVersion = fmt.Sprintf("%d.%d.%d%s",
		BCpMajor,
		BCpMinor,
		BCpBugfix,
		BCpQuality,
	)
	AppDesc = fmt.Sprintf(AppNameLong+" v%s / %s on %s (#%d %s)",
		AppVersion,
		runtime.Version(),
		runtime.GOOS,
		BCpBuildNo,
		BCpDate,
	)
	bcpRoot, _ = filepath.Abs(os.Args[0])
	bcpRoot = filepath.Dir(bcpRoot)
	resDir = filepath.Join(bcpRoot, "res")
	bcpState.Commanders = make(map[string]time.Time)
}

func resFile(resName string) string {
	return filepath.Join(resDir, resName)
}

type state struct {
	Version struct {
		Major, Minor, Bugfix int
		Quality              string
		Build                int
	}
	LastEDEvent time.Time
	Commanders  map[string]time.Time
	MatCats     map[string]string
}

func (s *state) clear() {
	s.LastEDEvent = time.Time{}
}

func stateFileName() string {
	res := filepath.Join(FlagDDir, "bcplus.json")
	return res
}

func (s *state) save(filename string) error {
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

func (s *state) load(filename string) error {
	log.Logf(l.Linfo, "load BC+ state from '%s'", filename)
	f, err := os.Open(filename)
	if os.IsNotExist(err) {
		s.clear()
		if s.MatCats == nil {
			s.MatCats = make(map[string]string)
		}
		log.Logf(l.Lwarn, "BC+ state '%s' not exists", filename)
		return nil
	} else if err != nil {
		return err
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(s)
	if s.MatCats == nil {
		s.MatCats = make(map[string]string)
	}
	return err
}

func eventLoop() {
	log.Log(l.Linfo, "running bc+ event loop…")
	var wuiupd webui.UIUpdate
	for evt := range bcpEventQ {
		log.Logf(l.Ltrace, "bc+ event from '%c'", evt.Source)
		switch evt.Source {
		case watched.EscrJournal:
			wuiupd = journalEvent(evt.Data.([]byte))
		case watched.EscrStatus:
			wuiupd = jstatStatus(evt.Data.(string))
		case watched.EscrMarket:
			jstatMarket(evt.Data.(string))
		case watched.EscrShipyard:
			jstatShipyard(evt.Data.(string))
		case watched.EscrOutfit:
			jstatOutfitting(evt.Data.(string))
		case common.BCpEvtSrcWUI:
			userEvent(evt.Data.(ggja.GenObj))
		default:
			wuiupd = 0
			log.Errorf("unknown source tag '%c' in main event: %v", evt.Source, evt)
		}
		var cmd interface{}
		if wuiupd&webui.UIReload == webui.UIReload {
			cmd = &webui.WsCmdLoad{
				WsCommand: webui.WsCommand{Cmd: webui.WsLoadCmd},
			}
		} else if wuiupd != 0 {
			const updLogLvl = l.Ldebug
			updHdr := wuiupd&webui.UIHdr == webui.UIHdr
			upd := webui.NewWsCmdUpdate(updHdr, nil)
			switch {
			case webui.Update(wuiupd, webui.UIShips):
				log.Logf(updLogLvl, "wui update ships (header: %t)", updHdr)
				upd.Tpc = webui.TpcShipsData(theCmdr)
			case webui.Update(wuiupd, webui.UISurface):
				log.Logf(updLogLvl, "wui update surface (header: %t)", updHdr)
				upd.Tpc = webui.TpcSurfaceData(theCmdr)
			default:
				log.Logf(updLogLvl, "wui update header")
			}
			cmd = upd
		}
		if cmd != nil {
			toWsClient <- cmd
		}
	}
}

func goWatchingJournals(jDir string, bcpEvent chan<- common.BCpEvent) *watched.JournalDir {
	res := &watched.JournalDir{
		Dir: jDir,
		PerJLine: func(line []byte) {
			cpy := make([]byte, len(line))
			copy(cpy, line)
			bcpEvent <- common.BCpEvent{Source: watched.EscrJournal, Data: cpy}
		},
		OnStatChg: func(tag rune, statFile string) {
			bcpEvent <- common.BCpEvent{Source: common.Source(tag), Data: statFile}
		},
		Quit: make(chan bool),
	}
	startWith := spoolJouranls(jDir, bcpState.LastEDEvent)
	go res.Watch(startWith)
	return res
}

func openGalaxy() *galaxy.Repo {
	const sqlCreate = "db/create-galaxy.sqlite.sql"
	res, err := galaxy.NewRepo(filepath.Join(FlagDDir, "galaxy.db"))
	newDB := false
	if os.IsNotExist(err) {
		fnm := resFile(sqlCreate)
		log.Logf(l.Linfo, "init galaxy DB from '%s'", fnm)
		err := res.RunSql(0, fnm)
		if err != nil {
			log.Panic(err)
		}
		newDB = true
	}
	gxyv, err := res.Version()
	if err != nil {
		log.Panic(err)
	}
	if !newDB {
		fnm := resFile(sqlCreate)
		log.Logf(l.Linfo, "check galaxy DB from '%s'", fnm)
		err := res.RunSql(gxyv, fnm)
		if err != nil {
			log.Panic(err)
		}
		v, err := res.Version()
		if err != nil {
			log.Panic(err)
		}
		if v > gxyv {
			log.Infof("updated galaxy DB from version: %d", gxyv)
		}
		gxyv = v
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
	res := filepath.Join(FlagDDir, name)
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
	//webui.L10nReset()
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
		FlagCheckEddn()
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
		theEddn.Header.SwName = AppNameShort
		theEddn.Header.SwVersion = vstr
		if err != nil {
			log.Log(l.Lwarn, "EDDN connect failed:", err)
			theEddn = nil
		} else {
			log.Logf(l.Linfo, "connected to EDDN as %s / %s / %s", upldr, AppNameShort, vstr)
		}
	} else {
		theCmdr = nil
		theEddn = nil
		log.Log(l.Linfo, "disconnected from EDDN")
	}
}

const (
	AppNameLong  = "BoardComputer+"
	AppNameShort = "BC+"
)

var (
	FlagJDir    string
	FlagDDir    string
	FlagEddn    string
	FlagWuiPort uint
	FlagMacros  string
	FlagTheme   string
	bcpRoot     string
	resDir      string
	bcpEventQ   = make(chan common.BCpEvent, 128)
	bcpState    state
	theGalaxy   *galaxy.Repo
	theCmdr     *cmdr.State
	theEddn     *eddn.Upload
	stateLock   sync.RWMutex
	toWsClient  chan<- interface{}
	nameMaps    common.NameMaps
)

func FlagCheckEddn() {
	switch FlagEddn {
	case "":
		if theCmdr != nil && len(theCmdr.EddnMode) > 0 {
			eddnMode = theCmdr.EddnMode
		} else {
			eddnMode = flagEddnDefault
		}
	case "off", "anon", "scramble", "cmdr", "test":
		eddnMode = FlagEddn
		if theCmdr != nil {
			theCmdr.EddnMode = FlagEddn
		}
	default:
		log.Fatalf("illegal EDDN mode: %s", eddnMode)
	}
}

func Run() {
	var err error
	log.Logf(l.Ldebug, "BC+ root: '%s'", bcpRoot)
	nameMaps.Load(resDir, FlagDDir, "English\\\\UK")
	bcpState.load(stateFileName())
	bcpState.Version.Major = BCpMajor
	bcpState.Version.Minor = BCpMinor
	bcpState.Version.Bugfix = BCpBugfix
	bcpState.Version.Quality = BCpQuality
	bcpState.Version.Build = BCpBuildNo
	theGalaxy = openGalaxy()
	if len(FlagMacros) > 0 {
		loadMacros(FlagMacros)
	}
	go sysResolver()
	goWatchingJournals(FlagJDir, bcpEventQ)
	toWsClient = webui.Run(&webui.Init{
		DataDir:     FlagDDir,
		ResourceDir: filepath.Join(bcpRoot, "res"),
		CommonName:  AppNameLong,
		Port:        FlagWuiPort,
		BCpVersion:  AppVersion,
		StateLock:   &stateLock,
		Galaxy:      theGalaxy,
		CmdrGetter:  func() *cmdr.State { return theCmdr },
		BCpQ:        bcpEventQ,
		Names:       &nameMaps,
		SysResolve:  sysResolveQ,
		Theme:       FlagTheme,
	})
	go eventLoop()
	// up & running – wait for Ctrl-C…
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	<-signals
	close(sysResolveQ)
	log.Log(l.Linfo, "BC+ interrupted with Ctrl-C, shutting down…")
	log.Log(l.Linfo, "closing galaxy repo")
	theGalaxy.Close()
	if theCmdr != nil {
		err = theCmdr.Save(cmdrFile(theCmdr.Name, cmdrState))
		if err != nil {
			log.Log(l.Lerror, "error while saving commander state:", err)
		}
	}
	err = bcpState.save(stateFileName())
	if err != nil {
		log.Log(l.Lerror, "error while saving BC+ state:", err)
	}
	nameMaps.Save()
	log.Log(l.Linfo, "Fly safe commander! o7")
}
