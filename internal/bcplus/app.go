package bcplus

import (
	"flag"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"git.fractalqb.de/fractalqb/c4hgol"
	"git.fractalqb.de/fractalqb/qbsllm"
	"github.com/CmdrVasquess/bcplus/internal/common"
	"github.com/CmdrVasquess/bcplus/internal/wapp"
	"github.com/CmdrVasquess/goedx"
	"github.com/CmdrVasquess/goedx/apps/bboltgalaxy"
	"github.com/CmdrVasquess/goedx/apps/l10n"
)

var (
	App    bcpApp
	LogWrs = []io.Writer{os.Stderr, &webLog}
	LogCfg = c4hgol.Config(qbsllm.NewConfig(log),
		goedx.LogCfg,
		l10n.LogCfg,
		wapp.LogCfg,
	)
)

var (
	log     = qbsllm.New(qbsllm.Lnormal, "BC+", nil, nil)
	edState = goedx.NewEDState()
)

const (
	l10nDir = "l10n"
)

type bcpApp struct {
	goedx.Extension
	WebPort    int
	dataDir    string
	assetDir   string
	webAddr    string
	webPin     string
	webTheme   string
	webTLS     bool
	debugModes string
	appL10n    *l10n.Locales
}

func ensureDatadir(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		log.Infoa("create `data dir`", dir)
		if err = os.MkdirAll(dir, common.DirFileMode); err != nil {
			log.Fatale(err)
		}
	}
	l10n := filepath.Join(dir, l10nDir)
	if _, err := os.Stat(l10n); os.IsNotExist(err) {
		log.Infoa("create `localization dir`", l10n)
		if err = os.Mkdir(l10n, common.DirFileMode); err != nil {
			log.Fatale(err)
		}
	}
}

func (bcp *bcpApp) Init() {
	ensureDatadir(bcp.dataDir)
	if bcp.webTLS {
		mustTLSCert(bcp.dataDir)
	}
	var err error
	if err = edState.Load(bcp.stateFile()); os.IsNotExist(err) {
		log.Infoa("state `file` not exists", bcp.stateFile())
	}
	bcp.EDState = edState
	if strings.IndexRune(App.debugModes, 'h') < 0 {
		bcp.JournalAfter = edState.LastJournalEvent
	}
	bcp.CmdrFile = func(cmdr *goedx.Commander) string {
		return filepath.Join(bcp.dataDir, cmdr.FID, "commander.json")
	}
	dir := filepath.Join(bcp.dataDir, "galaxy.bbolt")
	bcp.Galaxy, err = bboltgalaxy.Open(dir)
	if err != nil {
		log.Fatale(err)
	}
	dir = filepath.Join(bcp.dataDir, l10nDir)
	bcp.appL10n = l10n.New(dir, edState)
	bcp.AddApp("l10n", bcp.appL10n)
	initWebUI()
}

func (bcp *bcpApp) SaveState() {
	var cmdrFile string
	if bcp.EDState.Cmdr != nil && bcp.EDState.Cmdr.FID != "" {
		cmdrFile = bcp.CmdrFile(bcp.EDState.Cmdr)
	}
	err := edState.Save(bcp.stateFile(), cmdrFile)
	if err != nil {
		log.Errore(err)
	}
}

func (bcp *bcpApp) Shutdown() {
	bcp.Extension.Stop()
	bcp.appL10n.Close()
	err := bcp.Galaxy.(*bboltgalaxy.Galaxy).Close()
	if err != nil {
		log.Errore(err)
	}
	bcp.SaveState()
}

func (bcp *bcpApp) stateFile() string {
	return filepath.Join(bcp.dataDir, "bcplus.json")
}

func stdAssetDir() string {
	dir := filepath.Dir(os.Args[0])
	return filepath.Join(dir, "assets")
}

func (bcp *bcpApp) Flags() {
	jDir, err := goedx.FindJournals()
	if err != nil {
		jDir = ""
	}
	flag.StringVar(&App.JournalDir, "j", jDir, docJournalDir)
	flag.StringVar(&App.dataDir, "d", stdDataDir(), docDataDir)
	flag.StringVar(&App.assetDir, "assets", stdAssetDir(), docAssetDir)
	flag.IntVar(&App.WebPort, "web-port", 1337, docWebPort)
	flag.StringVar(&App.webAddr, "web-addr", "", docWebAddr)
	flag.StringVar(&App.webPin, "web-pin", "", docWebPin)
	flag.StringVar(&App.webTheme, "web-theme", "", docWebTheme)
	flag.BoolVar(&App.webTLS, "web-tls", true, docWebTLS)
	flag.StringVar(&App.debugModes, "debug", "", docDebug)
}

func (app *bcpApp) Run(signals <-chan os.Signal) {
	log.Infof("Running BC+ v%d.%d.%d-%s+%d (%s)",
		common.BCpMajor, common.BCpMinor, common.BCpPatch,
		common.BCpQuality, common.BCpBuildNo,
		runtime.Version())
	app.Init()
	log.Infoa("data `dir`", app.dataDir)
	log.Debuga("assert `dir`", app.assetDir)
	go app.Extension.MustRun(true)
	go runWebUI()
	<-signals
	log.Infof("BC+ %s interrupted; shutting down...", common.VersionLong)
	app.Shutdown()
}
