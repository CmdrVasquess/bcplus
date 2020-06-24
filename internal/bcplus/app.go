package bcplus

import (
	"flag"
	"io"
	"os"
	"path/filepath"

	"github.com/CmdrVasquess/bcplus/internal/common"
	"github.com/CmdrVasquess/goedx"
	"github.com/CmdrVasquess/goedx/apps/bboltgalaxy"
	"github.com/CmdrVasquess/goedx/apps/l10n"

	"git.fractalqb.de/fractalqb/c4hgol"
	"git.fractalqb.de/fractalqb/qbsllm"
)

var (
	App    bcpApp
	LogWrs = []io.Writer{os.Stderr, &webLog}
	LogCfg = c4hgol.Config(qbsllm.NewConfig(log),
		goedx.LogCfg,
		l10n.LogCfg,
	)
)

var (
	log     = qbsllm.New(qbsllm.Lnormal, "BC+", nil, nil)
	edState = goedx.NewEDState()
)

type bcpApp struct {
	goedx.Extension
	WebPort  int
	dataDir  string
	assetDir string
	webAddr  string
	webPin   string
	webTheme string
	appL10n  *l10n.Locales
}

func (bcp *bcpApp) Init() {
	var err error
	edState.Load(bcp.stateFile())
	bcp.CmdrFile = func(cmdr *goedx.Commander) string {
		return filepath.Join(bcp.dataDir, cmdr.FID, "commander.json")
	}
	dir := filepath.Join(bcp.dataDir, "galaxy.bbolt")
	bcp.Galaxy, err = bboltgalaxy.Open(dir)
	if err != nil {
		log.Fatale(err)
	}
	dir = filepath.Join(bcp.dataDir, "l10n")
	bcp.appL10n = l10n.New(dir, edState)
	bcp.AddApp("l10n", bcp.appL10n)
}

func (bcp *bcpApp) Shutdown() {
	bcp.Extension.Stop()
	bcp.appL10n.Close()
	err := bcp.Galaxy.(*bboltgalaxy.Galaxy).Close()
	if err != nil {
		log.Errore(err)
	}
	edState.Save(bcp.stateFile())
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
	flag.IntVar(&App.WebPort, "web-port", 0, docWebPort)
	flag.StringVar(&App.webAddr, "web-addr", "", docWebAddr)
	flag.StringVar(&App.webPin, "web-pin", "", docWebPin)
	flag.StringVar(&App.webTheme, "web-theme", "", docWebTheme)
}

func (app *bcpApp) Run(signals <-chan os.Signal) {
	app.Init()
	log.Infoa("data `dir`", app.dataDir)
	log.Debuga("assert `dir`", app.assetDir)
	go app.Extension.MustRun(true)
	<-signals
	log.Infof("BC+ %s interrupted; shutting down...", common.VersionLong)
	app.Shutdown()
}
