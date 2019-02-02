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
	log "git.fractalqb.de/fractalqb/qbsllm"

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
	lgr.Infoa("save BC+ state to `file`", filename)
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
	lgr.Infoa("load BC+ state from `file`", filename)
	f, err := os.Open(filename)
	if os.IsNotExist(err) {
		s.clear()
		if s.MatCats == nil {
			s.MatCats = make(map[string]string)
		}
		lgr.Warna("BC+ `state` not exists", filename)
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
	lgr.Info(log.Str("running bc+ event loop…"))
	var wuiupd webui.UIUpdate
	for evt := range bcpEventQ {
		lgr.Tracea("bc+ event from `src`", evt.Source)
		wuiupd = 0
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
			lgr.Errora("unknown source `tag` in main `event`", evt.Source, evt)
		}
		var cmd interface{}
		if wuiupd&webui.UIReload == webui.UIReload {
			cmd = &webui.WsCmdLoad{
				WsCommand: webui.WsCommand{Cmd: webui.WsLoadCmd},
			}
		} else if wuiupd != 0 {
			const updLogLvl = log.Ldebug
			updHdr := wuiupd&webui.UIHdr == webui.UIHdr
			upd := webui.NewWsCmdUpdate(updHdr, nil)
			switch {
			case webui.Update(wuiupd, webui.UIShips):
				lgr.Args(updLogLvl, "wui update ships (`header`)", updHdr)
				upd.Tpc = webui.TpcShipsData(theCmdr)
			case webui.Update(wuiupd, webui.UISurface):
				lgr.Args(updLogLvl, "wui update surface (`header`)", updHdr)
				upd.Tpc = webui.TpcSurfaceData(theCmdr)
			default:
				lgr.Wr(updLogLvl, log.Str("wui update header"))
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
		lgr.Infoa("init galaxy DB from `file`", fnm)
		err := res.RunSql(0, fnm)
		if err != nil {
			lgr.Panica("`err`", err)
		}
		newDB = true
	}
	gxyv, err := res.Version()
	if err != nil {
		lgr.Panica("`err`", err)
	}
	if !newDB {
		fnm := resFile(sqlCreate)
		lgr.Infoa("check galaxy DB from `file`", fnm)
		err := res.RunSql(gxyv, fnm)
		if err != nil {
			lgr.Panica("`err`", err)
		}
		v, err := res.Version()
		if err != nil {
			lgr.Panica("`err`", err)
		}
		if v > gxyv {
			lgr.Infoa("updated galaxy DB from `version`", gxyv)
		}
		gxyv = v
	}
	lgr.Infoa("galaxy DB `version`", gxyv)
	return res
}

func cmdrDir(name string) string {
	if len(name) == 0 {
		lgr.Warn(log.Str("empty commander name for directory"))
		name = "_anonymous_"
	} else {
		name = strings.Replace(name, " ", "_", -1)
	}
	res := filepath.Join(FlagDDir, name)
	if _, err := os.Stat(res); os.IsNotExist(err) {
		lgr.Debuga("create commander's `dir`", res)
		err = os.Mkdir(res, 0777)
		if err != nil {
			lgr.Warna("cannot create commander's `dir`", res)
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
			lgr.Errora("while saving commander state: `err`", err)
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
			lgr.Errora("cannot switch to commander `name`: `err`", name, err)
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
			lgr.Warna("EDDN connect failed: `err`", err)
			theEddn = nil
		} else {
			lgr.Infoa("connected to EDDN as `uploader` / `app` / `version`",
				upldr, AppNameShort, vstr)
		}
	} else {
		theCmdr = nil
		theEddn = nil
		lgr.Info(log.Str("disconnected from EDDN"))
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
		lgr.Fatala("illegal EDDN `mode`", eddnMode)
	}
}

func Run() {
	var err error
	lgr.Wr(log.Linfo,
		log.Fmt("goEDDNc v%d.%d.%d%s", eddn.Major, eddn.Minor, eddn.Bugfix, eddn.Quality))
	lgr.Wr(log.Ldebug, log.Fmt("BC+ root: '%s'", bcpRoot))
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
	lgr.Info(log.Str("BC+ interrupted with Ctrl-C, shutting down…"))
	lgr.Info(log.Str("closing galaxy repo"))
	theGalaxy.Close()
	if theCmdr != nil {
		err = theCmdr.Save(cmdrFile(theCmdr.Name, cmdrState))
		if err != nil {
			lgr.Errora("while saving commander state: `err`", err)
		}
	}
	err = bcpState.save(stateFileName())
	if err != nil {
		lgr.Errora("while saving BC+ state: `err`", err)
	}
	nameMaps.Save()
	lgr.Info(log.Str("Fly safe commander! o7"))
}
