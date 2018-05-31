// bcplus project main.go
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	c "github.com/CmdrVasquess/BCplus/cmdr"
	gxy "github.com/CmdrVasquess/BCplus/galaxy"
	edsm "github.com/CmdrVasquess/goEDSMc"
	"github.com/CmdrVasquess/watched"
	"github.com/fractalqb/namemap"
	l "github.com/fractalqb/qblog"
)

func init() {
	assetPathRoot = os.Args[0]
	assetPathRoot = filepath.Dir(filepath.Clean(assetPathRoot))
	docsPath = filepath.Join(assetPathRoot, "docs")
	assetPathRoot = filepath.Join(assetPathRoot, "bcplus.d")
	var err error
	if assetPathRoot, err = filepath.Abs(assetPathRoot); err != nil {
		panic(err)
	}
	glog.Logf(l.Info, "assets: %s", assetPathRoot)
	nmNavItem = namemap.MustLoad(assetPath("nm/navitems.xsx")).
		FromStd().
		To(false, "lang:").
		Verify("nav items", "std → lang:")
	nmRnkCombat = namemap.MustLoad(assetPath("nm/rnk_combat.xsx")).
		FromStd().
		To(false, "lang:").
		Verify("combat ranks", "std → lang:")
	nmRnkTrade = namemap.MustLoad(assetPath("nm/rnk_trade.xsx")).
		FromStd().
		To(false, "lang:").
		Verify("trade ranks", "std → lang:")
	nmRnkExplor = namemap.MustLoad(assetPath("nm/rnk_explore.xsx")).
		FromStd().
		To(false, "lang:").
		Verify("explorer ranks", "std → lang:")
	nmRnkCqc = namemap.MustLoad(assetPath("nm/rnk_cqc.xsx")).
		FromStd().
		To(false, "lang:").
		Verify("CQC ranks", "std → lang:")
	nmRnkFed = namemap.MustLoad(assetPath("nm/rnk_feds.xsx")).
		FromStd().
		To(false, "lang:").
		Verify("federation ranks", "std → lang:")
	nmRnkImp = namemap.MustLoad(assetPath("nm/rnk_imps.xsx")).
		FromStd().
		To(false, "lang:").
		Verify("imperial ranks", "std → lang:")
	nmShipType = namemap.MustLoad(assetPath("nm/shiptype.xsx")).
		FromStd().
		Verify("ship types", "std → lang:")
	nmMatType = namemap.MustLoad(assetPath("nm/resctypes.xsx")).
		FromStd().
		To(false, "lang:").
		Verify("rescource types", "std → lang:")
	nmMats = namemap.MustLoad(assetPath("nm/materials.xsx")).
		FromStd().
		To(false, "lang:").
		Verify("materials", "std → lang:")
	nmMatsXRef = nmMats.Base().FromStd().To(false, "wikia")
	nmMatsId = nmMats.Base().FromStd().To(false, "id")
	nmMatsIdRev = nmMats.Base().From("id", false).To(true)
	nmMatGrade = namemap.MustLoad(assetPath("nm/matgrade.xsx")).FromStd()
	nmMGrdRaw = nmMatGrade.Base().DomainIdx("raw")
	nmBdyCats = namemap.MustLoad(assetPath("nm/body-cats.xsx")).
		FromStd().
		To(false, "lang:").
		Verify("materials", "std → lang:")
	nmSynthLvl = namemap.MustLoad(assetPath("nm/synthlevel.xsx")).
		FromStd().
		To(false, "short:").
		Verify("materials", "std → short:")
}

type bcEvent struct {
	source rune
	data   interface{}
}

var theStateLock = sync.RWMutex{}
var theGalaxy *gxy.Galaxy
var theGame = c.NewGmState()
var credsKey []byte
var eventq = make(chan bcEvent, 128)
var signals = make(chan os.Signal, 1)

var jrnlDir string
var dataDir string
var enableJMacros bool
var verybose bool

var nmNavItem namemap.FromTo
var nmRnkCombat namemap.FromTo
var nmRnkTrade namemap.FromTo
var nmRnkExplor namemap.FromTo
var nmRnkCqc namemap.FromTo
var nmRnkFed namemap.FromTo
var nmRnkImp namemap.FromTo
var nmShipType namemap.From
var nmMatType namemap.FromTo
var nmMats namemap.FromTo
var nmMatsXRef namemap.FromTo
var nmMatsId namemap.FromTo
var nmMatsIdRev namemap.FromTo
var nmBdyCats namemap.FromTo
var nmSynthLvl namemap.FromTo
var nmMatGrade namemap.From
var nmMGrdRaw int

//go:generate ./genversion.sh
func BCpDescribe(wr io.Writer) {
	fmt.Fprintf(wr, "BoardComputer+ v%d.%d.%d%s / %s on %s (%s)",
		BCpMajor,
		BCpMinor,
		BCpBugfix,
		BCpQuality,
		runtime.Version(),
		runtime.GOOS,
		BCpDate)
}

func BCpDescStr() string {
	buf := bytes.NewBuffer(nil)
	BCpDescribe(buf)
	return buf.String()
}

func saveState(beta bool) {
	var w *os.File
	var err error
	defer func() {
		if w != nil {
			w.Close()
		}
	}()
	if theGame.Cmdr.Name == "" {
		glog.Logf(l.Info, "empty state, nothing to save")
	} else {
		var fnm string
		if beta {
			fnm = theGame.Cmdr.Name + "-beta.json"
		} else {
			fnm = theGame.Cmdr.Name + ".json"
		}
		fnm = filepath.Join(dataDir, fnm)
		tnm := fnm + "~"
		w, err = os.Create(tnm)
		if err != nil {
			glog.Logf(l.Error, "cannot save game status to '%s': %s", fnm, err)
		} else {
			glog.Logf(l.Info, "save state to %s", fnm)
			err := theGame.Save(w)
			w.Close()
			w = nil
			if err != nil {
				glog.Log(l.Error, err)
			} else if err = os.Rename(tnm, fnm); err != nil {
				glog.Log(l.Error, err)
			}
			fnm = filepath.Join(dataDir, "latestevent")
			if beta {
				fnm += ".beta"
			}
			w, err = os.Create(fnm)
			if err != nil {
				glog.Logf(l.Error, "cannot write latest event time to %s: %s", fnm, err)
			} else if err = json.NewEncoder(w).Encode(&theGame.T); err != nil {
				glog.Logf(l.Error, "cannot write latest event time to %s: %s", fnm, err)
			}
			w.Close()
			w = nil
		}

	}
	theGalaxy.Close()
	saveMacros(filepath.Join(dataDir, "macros.xsx"))
}

func loadState(cmdrNm string, beta bool) bool {
	var fnm string
	if beta {
		fnm = fmt.Sprintf("%s-beta.json", cmdrNm)
		if _, err := os.Stat(fnm); os.IsNotExist(err) {
			fnm = fmt.Sprintf("%s.json", cmdrNm)
		}
	} else {
		fnm = fmt.Sprintf("%s.json", cmdrNm)
	}
	fnm = filepath.Join(dataDir, fnm)
	glog.Logf(l.Info, "load state from %s", fnm)
	if r, err := os.Open(fnm); os.IsNotExist(err) {
		return false
	} else if err == nil {
		defer r.Close()
		theGame.Load(r)
		if len(credsKey) > 0 {
			loadCreds(cmdrNm)
		}
		return true
	} else {
		panic("load commander: " + err.Error())
	}
}

func loadCreds(cmdrNm string) error {
	if theGame.Creds == nil {
		theGame.Creds = &c.CmdrCreds{}
	} else {
		theGame.Creds.Clear()
	}
	theEdsm.Creds = &theGame.Creds.Edsm
	filenm := filepath.Join(dataDir, cmdrNm+".pgp")
	glog.Logf(l.Info, "load credentials from %s", filenm)
	if _, err := os.Stat(filenm); os.IsNotExist(err) {
		glog.Logf(l.Warn, "commander %s's credentials do not exist", cmdrNm)
		return nil
	}
	f, err := os.Open(filenm)
	if err != nil {
		return err
	}
	defer f.Close()
	err = theGame.Creds.Read(f, credsKey)
	if err != nil {
		glog.Logf(l.Warn, "failed to read credentials for %s: %s", cmdrNm, err)
		theGame.Creds = nil
	}
	return nil
}

var (
	docsPath      string
	assetPathRoot string
)

func assetPath(relPathSlash string) string {
	relPathSlash = filepath.FromSlash(relPathSlash)
	return filepath.Join(assetPathRoot, relPathSlash)
}

const (
	esrcUsr = 'u'
)

func eventLoop() {
	if feedEdsm {
		glog.Log(l.Info, "loading event discard list from EDSM…")
		var dscs []string
		err := theEdsm.Discard(&dscs)
		if err != nil {
			glog.Logf(l.Error, "cannot load discard list from EDSM: %s", err)
			edsmDiscard = nil
		} else {
			edsmDiscard = make(map[string]bool)
			for _, d := range dscs {
				edsmDiscard[d] = true
			}
			glog.Logf(l.Debug, "EDSM discard list has %d entries", len(edsmDiscard))
			if verybose {
				for e, _ := range edsmDiscard {
					glog.Logf(l.Trace, "EDSM discard: %s", e)
				}
			}
			theEdsm.Game = (*EdsmState)(theGame)
			if theGame.Creds != nil {
				theEdsm.Creds = &theGame.Creds.Edsm
			}
		}
	}
	glog.Log(l.Info, "starting main event loop…")
	for e := range eventq {
		switch e.source {
		case watched.EscrJournal:
			func() {
				defer func() {
					if r := recover(); r != nil {
						glog.Logf(l.Error, "journal event error: %s", r)
					}
				}()
				dispatchJournal(&theStateLock, theGame, e.data.([]byte))
			}()
		case esrcUsr:
			func() {
				defer func() {
					if r := recover(); r != nil {
						glog.Logf(l.Error, "user event error: %s", r)
					}
				}()
				dispatchUser(&theStateLock, theGame, e.data.(map[string]interface{}))
			}()
		case watched.EscrStatus:
			func() {
				defer func() {
					if r := recover(); r != nil {
						glog.Logf(l.Error, "user event error: %s", r)
					}
				}()
				dispStfStatus(&theStateLock, theGame, e.data.(string))
			}()
		default:
			glog.Logf(l.Warn, "no handler for event source: %c", e.source)
		}
	}
}

func checkJournals(jrnlDir string) {
	var lets time.Time
	tsfnm := filepath.Join(dataDir, "latestevent")
	rd, err := os.Open(tsfnm)
	if err != nil {
		glog.Log(l.Error, "failed to read latest timestamp: %s", err)
		return
	}
	defer rd.Close()
	err = json.NewDecoder(rd).Decode(&lets)
	if err != nil {
		glog.Log(l.Error, "failed to read latest timestamp: %s", err)
		return
	}
	catchUpWithJournal(lets, jrnlDir)
}

func main() {
	flag.StringVar(&dataDir, "d", defaultDataDir(),
		"directory to store BC+ data")
	flag.StringVar(&jrnlDir, "j", defaultJournalDir(),
		"directory with journal files")
	flag.UintVar(&webGuiPort, "p", 1337,
		"web GUI port")
	verbose := flag.Bool("v", false, "verbose logging")
	flag.BoolVar(&verybose, "vv", false, "very verbose logging")
	flag.BoolVar(&acceptHistory, "hist", false, "accept historic events")
	loadCmdr := flag.String("cmdr", "", "preload commander")
	promptKey := flag.Bool("pmk", false, "prompt for credential master key")
	flag.DurationVar(&macroPause, "mcrp", 60*time.Millisecond,
		"set the delay between macro elements")
	flag.BoolVar(&enableJMacros, "jmacros", true, "enable journal macro engine")
	flag.IntVar(&tspLimit, "tsp-limit", 120, "set the limit for TSP in travel planning")
	flag.BoolVar(&feedEdsm, "edsm", false, "WiP: send events to EDSM (s.a. pmk & credentials)")
	edsmTestSvc := flag.Bool("edsm.test", false, "run against EDSM test server")
	showHelp := flag.Bool("h", false, "show help")
	flag.Parse()
	if *showHelp {
		BCpDescribe(os.Stdout)
		fmt.Println()
		flag.Usage()
		os.Exit(0)
	}
	if verybose {
		glog.SetLevel(l.Trace)
	} else if *verbose {
		glog.SetLevel(l.Debug)
	}
	glog.Logf(l.Info, BCpDescStr())
	glog.Logf(l.Info, "data    : %s\n", dataDir)
	var err error
	if *promptKey {
		credsKey = c.PromptCredsKey("")
	}
	if _, err = os.Stat(dataDir); os.IsNotExist(err) {
		err = os.MkdirAll(dataDir, 0777)
		if err != nil {
			glog.Fatal("cannot create data dir: %s", err.Error())
		}
	}
	glog.Logf(l.Info, "journals: %s\n", jrnlDir)
	loadMacros(filepath.Join(dataDir, "macros.xsx"))
	theGalaxy, err = gxy.OpenGalaxy(
		filepath.Join(dataDir, "systems.json"),
		assetPath("data/"))
	if err != nil {
		glog.Fatal(err)
	}
	c.SetTheGalaxy(theGalaxy)
	if len(*loadCmdr) > 0 {
		loadState(*loadCmdr, false)
	}
	if *edsmTestSvc {
		glog.Log(l.Info, "switch to EDSM test servers")
		theEdsm = edsm.NewService(edsm.Test)
	}
	checkJournals(jrnlDir)
	jdir := watched.JournalDir{
		Dir: jrnlDir,
		PerJLine: func(line []byte) {
			cpy := make([]byte, len(line))
			copy(cpy, line)
			eventq <- bcEvent{watched.EscrJournal, cpy}
		},
		OnStatChg: func(tag rune, statFile string) {
			eventq <- bcEvent{tag, statFile}
		},
		Quit: make(chan bool),
	}
	go jdir.Watch()
	go eventLoop()
	runWebGui()
	signal.Notify(signals, os.Interrupt)
	// up & running – wait for Ctrl-C…
	<-signals
	jdir.Quit <- true
	glog.Log(l.Info, "BC+ interrupted")
	theStateLock.RLock()
	saveState(theGame.IsBeta)
	theStateLock.RUnlock()
	glog.Log(l.Info, "bye…")
}
