package app

import (
	"encoding/json"
	"flag"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"git.fractalqb.de/fractalqb/namemap"

	"git.fractalqb.de/fractalqb/qbsllm"
	"github.com/CmdrVasquess/bcplus/internal/common"
	"github.com/CmdrVasquess/bcplus/internal/ship"
	"github.com/CmdrVasquess/watched"
)

var (
	version = "BC+ " + common.VersionAll
	App     BCpApp
	cmdr    = NewCommander("", "")
	LogWrs  = []io.Writer{os.Stderr, &webLog}
	toSpeak chan<- VoiceMsg
	log     = qbsllm.New(qbsllm.Lnormal, "BC+", nil, nil)
	LogCfg  = qbsllm.Config(log, elogCfg, ship.LogCfg, watched.LogCfg)
	// TODO what to put into BCpApp struct
	flagJournalDir string
	flagWebPort    int
	flagWebAddr    string
	flagWebPin     string
	flagWebTheme   string
	flagCmdr       string
	stateLock      sync.RWMutex

	nmRawMat, nmManMat, nmEncMat *namemap.NameMap
)

func dispatchVoice(channel string, prio int, text string) {
	if text == "" {
		return
	}
	select {
	case toSpeak <- VoiceMsg{Txt: text, Prio: prio, Chan: channel}:
	default:
		log.Warns("speaker queue full, drop message")
	}
}

func readState(do func() error) error {
	stateLock.RLock()
	defer stateLock.RUnlock()
	return do()
}

func writeState(do func() error) error {
	stateLock.Lock()
	defer stateLock.Unlock()
	return do()
}

func noErr(do func()) func() error { return func() error { do(); return nil } }

type BCpApp struct {
	JournalDir string
	dataDir    string
	assetDir   string
	LastEvent  time.Time
	GoOffline  bool
	ApiKey     string
	WebPort    uint16
	WebAddr    string
	WebPin     string
	WebTheme   string
	Speak      SpeakCfg
	debugMode  bool
	Lang       string
	tmpLd      *TmplLoader
}

func (app *BCpApp) dataBCpApp() string {
	return filepath.Join(app.dataDir, "BCplus.json")
}

func (app *BCpApp) loadNameMap(file string) *namemap.NameMap {
	log.Infoa("load name map `file`", file)
	return namemap.MustLoad(filepath.Join(app.assetDir, "nms", file))
}

func (app *BCpApp) load() {
	fnm := app.dataBCpApp()
	rd, err := os.Open(fnm)
	if err != nil {
		log.Fatala("cannot read `config`", fnm)
	}
	defer rd.Close()
	dec := json.NewDecoder(rd)
	if err = dec.Decode(app); err != nil {
		log.Fatala("config read `error`", err)
	}
	if app.debugMode {
		app.LastEvent = time.Time{}
	}
	if err = app.Speak.init(); err != nil {
		log.Fatala("config speak `error`", err)
	}
	log.Infoa("loaded `config`", fnm)
	log.Debuga("use `last timestamp`", app.LastEvent)
}

func (app *BCpApp) save() {
	fnm := app.dataBCpApp()
	tmp := fnm + "~"
	wr, err := os.Create(tmp)
	if err != nil {
		log.Errora("config save `error`", err)
		return
	}
	defer wr.Close()
	enc := json.NewEncoder(wr)
	enc.SetIndent("", "  ")
	err = enc.Encode(app)
	if err != nil {
		log.Errora("config write `error`", err)
		return
	}
	err = wr.Close()
	if err != nil {
		log.Errora("config close `error`", err)
	}
	err = os.Rename(tmp, fnm)
	if err != nil {
		log.Errora("config write `error`", err)
	}
	log.Infoa("saved `config`", fnm)
}

func (app *BCpApp) init() {
	app.Speak.ChanCfg = make(map[string]*ChanConfig)
	showHideCon()
	log.Infoa("BC+ data in `dir`", app.dataDir)
	if !setup(app) {
		app.load()
	}
	ship.TheTypes = ship.TypeRepo(app.dataDir)
	if len(flagJournalDir) > 0 {
		app.JournalDir = flagJournalDir
	} else if len(app.JournalDir) == 0 {
		app.JournalDir = stdJournalDir()
	}
	nmRawMat = app.loadNameMap("raw-mats.xsx")
	nmManMat = app.loadNameMap("man-mats.xsx")
	nmEncMat = app.loadNameMap("enc-mats.xsx")
	log.Infoa("journals in `dir`", app.JournalDir)
	if flagWebPort != 0 {
		app.WebPort = uint16(flagWebPort)
	} else if app.WebPort == 0 {
		app.WebPort = 1337
	}
	if len(flagWebAddr) > 0 {
		app.WebAddr = flagWebAddr
	}
	if flagWebPin == "off" {
		app.WebPin = ""
	} else if len(flagWebPin) > 0 {
		app.WebPin = flagWebPin
	}
	if len(flagWebTheme) > 0 {
		app.WebTheme = flagWebTheme
	} else if app.WebTheme == "" {
		app.WebTheme = "dark"
	}
	err := mustTLSCert(app.dataDir)
	if err != nil {
		log.Fatale(err)
	}
	app.tmpLd = NewTmplLoader(filepath.Join(app.assetDir, "goxic"))
	if flagCmdr != "" {
		cmdr.switchTo(flagCmdr, "")
	}
}

func stdAssetDir() string {
	dir := filepath.Dir(os.Args[0])
	return filepath.Join(dir, "assets")
}

func (app *BCpApp) Flags() {
	flag.StringVar(&flagJournalDir, "j", "", docJournalDir)
	flag.StringVar(&App.dataDir, "d", stdDataDir(), docDataDir)
	flag.StringVar(&App.assetDir, "assets", stdAssetDir(), docAssetDir)
	flag.IntVar(&flagWebPort, "web-port", 0, docWebPort)
	flag.StringVar(&flagWebAddr, "web-addr", "", docWebAddr)
	flag.StringVar(&flagWebPin, "web-pin", "", docWebPin)
	flag.StringVar(&flagWebTheme, "web-theme", "", docWebTheme)
	flag.StringVar(&App.Lang, "lang", "en", docLang)
	flag.BoolVar(&App.debugMode, "debug", false, docDebug)
	flag.StringVar(&flagCmdr, "load-fid", "", "load commander with FID ad start")
}

func (app *BCpApp) Run(signals <-chan os.Signal) {
	log.Infof("running %s", version)
	defer func() {
		if r := recover(); r == nil {
			log.Infof("%s stopped", version)
		} else {
			log.Errorf("panic in %s: %#v", version, r)
		}
	}()
	App.init()
	go eventLoop()
	quitWatch := watchJournalDir(App.JournalDir)
	toSpeak = App.Speak.run()
	go runWebUI()
	<-signals
	log.Infof("BC+ %s interrupted; shutting down...", common.VersionLong)
	cmdr.close()
	app.save()
	close(EventQ)
	close(webUiUpd)
	if toSpeak != nil {
		close(toSpeak)
	}
	close(quitWatch)
}
