package graphblast

import (
	"bundle"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"time"
)

func Index(graph Graph) http.HandlerFunc {
	indexfile := bundle.ReadFile("index.html")
	indexpage := template.Must(template.New("index").Parse(string(indexfile)))

	return LogRequest(func(w http.ResponseWriter, r *http.Request) {
		msg, err := json.Marshal(&graph)
		if err != nil {
			Log("error: %s", err)
			return
		}
		// TODO Use the initial JSON in the HTML template, or remove it here
		indexpage.Execute(w, string(msg))
	})
	// TODO Consider building the JS into the HTML, and removing Script()
}

func Script() http.HandlerFunc {
	scriptfile := bytes.NewReader(bundle.ReadFile("script.js"))
	return LogRequest(func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, "script.js", time.Now(), scriptfile)
	})
}

func Events(updateFreq time.Duration, watchers ErrorWatchers, graph Graph) http.HandlerFunc {
	return LogRequest(func(w http.ResponseWriter, r *http.Request) {
		ticker := time.NewTicker(updateFreq)
		defer ticker.Stop()

		// Register a channel for passing errors to the HTTP client.
		errors := watchers.Watch(r.RemoteAddr)
		defer watchers.Unwatch(r.RemoteAddr)

		// Get the necessary parts for being an EventSource, or fail.
		flusher, cn, err := toEventSource(w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		sendJSON(w, graph)
		flusher.Flush()

		changed, lastCount := false, 0
		for {
			select {
			case _ = <-cn.CloseNotify():
				return

			case err = <-errors:
				sendJSON(w, map[string]error{"error": err})
				flusher.Flush()
				return

			case _ = <-ticker.C:
				changed, lastCount = graph.Changed(lastCount)
				if !changed {
					continue
				}
				// TODO Only send graph deltas
				sendJSON(w, graph)
				flusher.Flush()
			}
		}
	})
}

// Writes an object as JSON to a writer. If there are errors serializing, write
// an error JSON object instead.
func sendJSON(writer io.Writer, obj interface{}) {
	msg, err := json.Marshal(obj)
	if err != nil {
		// TODO Use an EventSource event here to distinguish data from errors
		fmt.Fprint(writer, "data: {\"type\": \"JSON error\"}\n\n")
		return
	}
	fmt.Fprintf(writer, "data: %s\n\n", msg)
}

// Sets up a ResponseWriter for use as an EventSource.
func toEventSource(w http.ResponseWriter) (http.Flusher, http.CloseNotifier, error) {
	f, canf := w.(http.Flusher)
	cn, cancn := w.(http.CloseNotifier)

	if !canf || !cancn {
		return f, cn, errors.New("connection not suitable for EventSource")
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	return f, cn, nil
}
