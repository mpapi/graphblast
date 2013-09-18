package main

import (
	"bundle"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"graphblast"
	"html/template"
	"io"
	"math"
	"net/http"
	"os"
	"time"
)

// Command-line flags.
var listen = flag.String("listen", ":8080", "address:port to listen on")
var verbose = flag.Bool("verbose", false, "be more verbose")
var label = flag.String("label", "", "graph label")
var min = flag.Float64("min", math.Inf(-1), "minimum accepted value")
var max = flag.Float64("max", math.Inf(1), "maximum accepted value")
var bucket = flag.Int("bucket", 1, "histogram bucket size")
var delay = flag.Int("delay", 5, "delay between updates, in seconds")
var wide = flag.Bool("wide", false, "use wide orientation")
var width = flag.Int("width", 500, "width of the graph, in pixels")
var height = flag.Int("height", 500, "height of the graph, in pixels")
var colors = flag.String("colors", "", "comma-separated: bg, fg, bar color")
var fontSize = flag.String("font-size", "", "font size (CSS)")
var window = flag.Int("window", 1000, "data window size")

// TODO list of x/y pairs (x is string, y is countable -- then only send deltas)
// TODO make use of embedding

// EventSource requests that want to listen for errors (including EOF) can
// register themselves here.
type ErrorWatchers map[string]chan error

// Adds a watcher by name to the map, and returns a new channel they can use
// for listening.
func (ew *ErrorWatchers) Watch(name string) chan error {
	errChan := make(chan error)
	(*ew)[name] = errChan
	return errChan
}

// Removes the watcher by name from the map.
func (ew *ErrorWatchers) Unwatch(name string) {
	errChan, ok := (*ew)[name]
	if ok {
		close(errChan)
		delete(*ew, name)
	}
}

// Takes a channel of errors, and broadcasts any errors that are received on
// that channel to all registered channels.
func (ew *ErrorWatchers) Broadcast(errors chan error) {
	for err := range errors {
		for _, errChan := range *ew {
			errChan <- err
		}
	}
}

// Sets up a ResponseWriter for use as an EventSource.
func EventSource(w http.ResponseWriter) (http.Flusher, http.CloseNotifier, error) {
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

// Writes an object as JSON to a writer. If there are errors serializing, write
// an error JSON object instead.
func sendJSON(writer io.Writer, obj interface{}) {
	msg, err := json.Marshal(obj)
	if err != nil {
		// TODO use ES event
		fmt.Fprint(writer, "data: {\"type\": \"JSON error\"}\n\n")
		return
	}
	fmt.Fprintf(writer, "data: %s\n\n", msg)
}

func buildGraph(arg string) graphblast.Graph {
	allowed := graphblast.Range{
		Min: graphblast.Countable(*min),
		Max: graphblast.Countable(*max)}

	switch arg {
	case "histogram":
		graph := graphblast.NewHistogram(*bucket, *label, *wide)
		graph.Width = *width
		graph.Height = *height
		graph.Colors = *colors
		graph.FontSize = *fontSize
		graph.Allowed = allowed
		return graph
	case "timeseries":
		graph := graphblast.NewTimeSeries(*window, *label)
		graph.Width = *width
		graph.Height = *height
		graph.Colors = *colors
		graph.FontSize = *fontSize
		graph.Allowed = allowed
		return graph
	case "scatterplot":
		graph := graphblast.NewScatterPlot(*label)
		graph.Width = *width
		graph.Height = *height
		graph.Colors = *colors
		graph.FontSize = *fontSize
		graph.Allowed = allowed
		return graph
	case "logfile":
		graph := graphblast.NewLogFile(*window, *label)
		graph.Colors = *colors
		graph.FontSize = *fontSize
		return graph
		// TODO collapse lines
		// TODO send diffs only
		// TODO exit on EOF
		// TODO capture timestamps
	}
	panic("no graph for type")
}

// TODO scatterplot color/size options
// TODO cartogram

func main() {
	flag.Parse()
	graphblast.SetVerboseLogging(*verbose)

	graph := buildGraph(flag.Arg(0))
	// TODO have graph return a FlagSet

	readerrors := make(chan error)
	watchers := make(ErrorWatchers, 0)
	go watchers.Broadcast(readerrors)

	go graph.Read(os.Stdin, readerrors)

	ticker := time.NewTicker(time.Duration(*delay) * time.Second)

	indexfile := bundle.ReadFile("index.html")
	indexpage := template.Must(template.New("index").Parse(string(indexfile)))
	http.HandleFunc("/", graphblast.LogRequest(func(w http.ResponseWriter, r *http.Request) {
		msg, err := json.Marshal(&graph)
		if err != nil {
			fmt.Println("FAIL", err)
			return
		}
		indexpage.Execute(w, string(msg))
	}))

	scriptfile := bytes.NewReader(bundle.ReadFile("script.js"))
	http.HandleFunc("/script.js", graphblast.LogRequest(func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, "script.js", time.Now(), scriptfile)
	}))

	http.HandleFunc("/data", graphblast.LogRequest(func(w http.ResponseWriter, r *http.Request) {
		// Register a channel for passing errors to the HTTP client.
		errors := watchers.Watch(r.RemoteAddr)
		defer watchers.Unwatch(r.RemoteAddr)

		// Get the necessary parts for being an EventSource, or fail.
		flusher, cn, err := EventSource(w)
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
				sendJSON(w, graph)
				flusher.Flush()
			}
		}
	}))

	graphblast.Log("listening on %v", *listen)
	http.ListenAndServe(*listen, nil)
}
