package graphblast

import (
	"log"
	"net/http"
)

var verbose bool = false
var logfn func(string, ...interface{}) = log.Printf

func SetLogger(logger *log.Logger) {
	logfn = logger.Printf
}

func SetVerboseLogging(v bool) {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	verbose = v
}

func Log(format string, v ...interface{}) {
	if !verbose {
		return
	}
	logfn(format+"\n", v...)
}

func LogRequest(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		Log("(%v) starting %v %v", r.RemoteAddr, r.Method, r.URL)
		defer Log("(%v) finished %v %v", r.RemoteAddr, r.Method, r.URL)
		handler(w, r)
	}
}
