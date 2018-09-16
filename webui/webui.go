package webui

import (
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"sync"

	gxc "git.fractalqb.de/fractalqb/goxic"
	l "git.fractalqb.de/fractalqb/qblog"
	"github.com/CmdrVasquess/BCplus/cmdr"
	"github.com/CmdrVasquess/BCplus/common"
	"github.com/CmdrVasquess/BCplus/galaxy"
)

const (
	certFile = "webui.cert"
	keyFile  = "webui.key"
)

var (
	log       = l.Std("bc+wui:")
	LogConfig = l.Package(log)
	pgOffline []byte
)

type gxtTopic struct {
	*gxc.Template
	HeaderData []int
	TopicData  []int
}

var (
	theGalaxy    *galaxy.Repo
	theCmdr      func() *cmdr.State
	theStateLock *sync.RWMutex
	theBCpQ      chan<- common.BCpEvent
	nmap         *common.NameMaps
)

type Init struct {
	DataDir     string
	ResourceDir string
	CommonName  string
	Lang        string
	Port        uint
	BCpVersion  string
	Galaxy      *galaxy.Repo
	CmdrGetter  func() *cmdr.State
	StateLock   *sync.RWMutex
	BCpQ        chan<- common.BCpEvent
	Names       *common.NameMaps
}

func (i *Init) configure() {
	theGalaxy = i.Galaxy
	theCmdr = i.CmdrGetter
	theStateLock = i.StateLock
	theBCpQ = i.BCpQ
	nmap = i.Names
}

func offlineFilter(h func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if theCmdr() == nil {
			w.Write(pgOffline)
		} else {
			theStateLock.RLock()
			defer theStateLock.RUnlock()
			h(w, r)
		}
	}
}

type topic struct {
	key  string
	path string
	hdlr func(http.ResponseWriter, *http.Request)
	nav  string
}

var topics = []*topic{
	&topic{key: tkeySysPop, path: "/syspop", hdlr: tpcSysPop},
	&topic{key: tkeySysNat, path: "/sysnat", hdlr: tpcSysNat},
	&topic{key: tkeySynth, path: "/synth", hdlr: tpcSynth},
}

func getTopic(key string) *topic {
	for _, t := range topics {
		if t.key == key {
			return t
		}
	}
	panic(fmt.Errorf("requested unknown topic '%s'", key))
}

func Run(init *Init) chan<- interface{} {
	log.Log(l.Linfo, "Initialize Web UI…")
	init.configure()
	loadTemplates(init.ResourceDir, init.Lang, init.BCpVersion)
	err := mustTLSCert(init.DataDir, init.CommonName)
	if err != nil {
		log.Panic(err)
	}
	htStatic := http.FileServer(http.Dir(filepath.Join(init.ResourceDir, "s")))
	http.Handle("/s/", http.StripPrefix("/s", htStatic))
	http.HandleFunc("/ws", serveWs)
	//	http.HandleFunc("/", func(wr http.ResponseWriter, rq *http.Request) {
	//		wr.Header().Set("Content-Type", "text/html; charset=utf-8")
	//		wr.Write(pgOffline) // TODO error
	//	})
	http.HandleFunc("/", offlineFilter(tpcSysPop))
	for _, tdef := range topics {
		http.HandleFunc(tdef.path, offlineFilter(tdef.hdlr))
	}
	go wscHub()
	go http.ListenAndServeTLS(fmt.Sprintf(":%d", init.Port),
		filepath.Join(init.DataDir, certFile),
		filepath.Join(init.DataDir, keyFile),
		nil)
	addls, err := ownAddrs()
	if err != nil {
		log.Panic(err)
	} else {
		log.Log(l.Linfo, "Local Web UI address:")
		log.Logf(l.Linfo, "\thttps://localhost:%d/", init.Port)
		log.Log(l.Linfo, "This host's addresses to connect to Web UI from remote:")
		for _, addr := range addls {
			log.Logf(l.Linfo, "\thttps://%s:%d/", addr, init.Port)
		}
	}
	return wscSendTo
}

func ownAddrs() (res []string, err error) {
	ifaddrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	for _, addr := range ifaddrs {
		if nip, ok := addr.(*net.IPNet); ok {
			if nip.IP.IsLoopback() {
				continue
			}
			if ip := nip.IP.To4(); ip != nil {
				res = append(res, nip.IP.String())
			} else if ip := nip.IP.To16(); ip != nil {
				res = append(res, fmt.Sprintf("[%s]", nip.IP.String()))
			}
		}
	}
	return res, err
}
