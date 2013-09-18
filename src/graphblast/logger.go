package graphblast

import (
	"log"
	"net/http"
)

var verbose bool = false

func SetVerboseLogging(v bool) {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	verbose = v
}

func Log(format string, v ...interface{}) {
	if !verbose {
		return
	}
	log.Printf(format+"\n", v...)
}

func LogRequest(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		Log("(%v) starting %v %v", r.RemoteAddr, r.Method, r.URL)
		defer Log("(%v) finished %v %v", r.RemoteAddr, r.Method, r.URL)
		handler(w, r)
	}
}
