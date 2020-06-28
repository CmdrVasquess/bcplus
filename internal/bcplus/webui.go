package bcplus

import (
	"crypto/subtle"
	"fmt"
	"math/rand"
	"net/http"
	"path/filepath"
	"time"
)

func webPIN(h http.HandlerFunc) http.HandlerFunc {
	if len(App.webPin) == 0 {
		return h
	}
	return func(wr http.ResponseWriter, rq *http.Request) {
		_, pass, ok := rq.BasicAuth()
		if ok && subtle.ConstantTimeCompare([]byte(pass), []byte(App.webPin)) != 1 {
			ok = false
			// ConstTimeComp still varies with length
			time.Sleep(time.Duration(300+rand.Intn(300)) * time.Millisecond)
			log.Warna("wrong web-pin from `remote`", rq.RemoteAddr)
		}
		if !ok {
			wr.Header().Set("WWW-Authenticate", `Basic realm="Password: BC+ Web PIN"`)
			http.Error(wr, "Unauthorized", http.StatusUnauthorized)
			return
		}
		h(wr, rq)
	}
}

func webRoutes() {
	htStatic := http.FileServer(http.Dir(filepath.Join(App.assetDir, "s")))
	http.HandleFunc("/s/", webPIN(http.StripPrefix("/s", htStatic).ServeHTTP))
	http.HandleFunc("/ws/log", webPIN(logWs))
}

func runWebUI() {
	// webLoadTmpls()
	webRoutes()
	keyf := filepath.Join(App.dataDir, keyFile)
	crtf := filepath.Join(App.dataDir, certFile)
	lstn := fmt.Sprintf("%s:%d", App.webAddr, App.WebPort)
	if App.webTLS {
		log.Infoa("run web ui on https `addr`", lstn)
		log.Fatale(http.ListenAndServeTLS(lstn, crtf, keyf, nil))
	} else {
		log.Infoa("run web ui on http `addr`", lstn)
		log.Fatale(http.ListenAndServe(lstn, nil))
	}
}
