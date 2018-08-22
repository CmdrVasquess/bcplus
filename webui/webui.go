package webui

import (
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"sync"

	l "git.fractalqb.de/fractalqb/qblog"
	"github.com/CmdrVasquess/BCplus/cmdr"
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

var (
	theCmdr      func() *cmdr.State
	theStateLock *sync.RWMutex
)

type Init struct {
	DataDir     string
	ResourceDir string
	CommonName  string
	Lang        string
	Port        uint
	BCpVersion  string
	CmdrGetter  func() *cmdr.State
	StateLock   *sync.RWMutex
}

func (i *Init) configure() {
	theCmdr = i.CmdrGetter
	theStateLock = i.StateLock
}

func offlineFilter(
	w http.ResponseWriter,
	r *http.Request,
	h func(http.ResponseWriter, *http.Request),
) {
	if theCmdr() == nil {
		w.Write(pgOffline)
	} else {
		theStateLock.RLock()
		defer theStateLock.RUnlock()
		h(w, r)
	}
}

func Run(init *Init) {
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
	http.HandleFunc("/", func(wr http.ResponseWriter, rq *http.Request) {
		wr.Header().Set("Content-Type", "text/html; charset=utf-8")
		wr.Write(pgOffline) // TODO error
	})

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
