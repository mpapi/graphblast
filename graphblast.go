package main

import (
	"bufio"
	"bundle"
	"bytes"
	"container/list"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
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

// The type of the items to parse from stdin and count in the histogram.
type Countable float64

// Parses a countable value from a string, and returns a non-nil error if
// parsing fails.
func Parse(str string) (Countable, error) {
	d, err := strconv.ParseFloat(str, 64)
	return Countable(d), err
}

// Returns the bucket (as a string) of which the countable value should
// increment the count, given the bucket size.
func (d Countable) Bucket(size int) string {
	if d < 0 {
		d -= Countable(size)
	}
	return strconv.Itoa(int(d) / size * size)
}

func doRead(input io.Reader, errors chan error, process func(string)) {
	logger("starting to read data")
	reader := bufio.NewReader(input)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			logger("finished reading data due to %v", err)
			errors <- err
			return
		}
		process(line)
	}
}

type Graph interface {
	Changed(int) (bool, int)
	Read(chan error)
}

type LogFile struct {
	Values map[string]string

	Layout string // the layout to use (interpreted by JS)
	Label  string // the label of the display
	Window int    // the number of lines to retain

	Colors   string // the colors to use when displaying the graph
	FontSize string // the CSS font size to use when displaying the graph

	Count    int // the number of values encountered so far
	Filtered int // the number of values filtered out so far
	Errors   int // the number of values skipped due to errors so far
}

func NewLogFile(window int, label string) *LogFile {
	return &LogFile{
		Layout: "logfile",
		Values: make(map[string]string, 1024),
		Label:  label,
		Window: window}
}

func (lf *LogFile) Changed(indicator int) (bool, int) {
	if lf.Count <= indicator {
		return false, indicator
	}
	return true, lf.Count
}

func (lf *LogFile) Add(line string, err error) {
	if err != nil {
		lf.Errors += 1
		return
	}

	lf.Values[fmt.Sprintf("%v", lf.Count)] = line
	lf.Count += 1
	if len(lf.Values) > lf.Window {
		delete(lf.Values, fmt.Sprintf("%v", lf.Count-lf.Window-1))
	}
}

func (lf *LogFile) Read(errs chan error) {
	doRead(os.Stdin, errs, func(line string) {
		lf.Add(strings.TrimSpace(line), nil)
	})
}

// TODO list of x/y pairs (x is string, y is countable -- then only send deltas)
// TODO make use of embedding

type ScatterPlot struct {
	Values map[string]Countable

	Layout string // the layout to use (interpreted by JS)
	Label  string // the label of the histogram
	Width  int    // the maximum graph width in pixels
	Height int    // the maximum graph height in pixels

	Colors   string // the colors to use when displaying the graph
	FontSize string // the CSS font size to use when displaying the graph

	Min Countable // the minimum value encountered so far
	Max Countable // the maximum value encountered so far

	Count    int // the number of values encountered so far
	Filtered int // the number of values filtered out so far
	Errors   int // the number of values skipped due to errors so far
}

func NewScatterPlot(label string) *ScatterPlot {
	return &ScatterPlot{
		Layout: "scatterplot",
		Values: make(map[string]Countable, 1024),
		Label:  label,
		Min:    Countable(math.Inf(1)),
		Max:    Countable(math.Inf(-1))}
}

func (sp *ScatterPlot) Changed(indicator int) (bool, int) {
	if sp.Count <= indicator {
		return false, indicator
	}
	return true, sp.Count
}

// TODO interface Collection with methods for updating min/max/etc.

func (sp *ScatterPlot) Add(x Countable, val Countable, err error) {
	if err != nil {
		sp.Errors += 1
		return
	} else if val < Countable(*min) || val > Countable(*max) {
		sp.Filtered += 1
		return
	}

	if val < sp.Min {
		sp.Min = val
	}
	if val > sp.Max {
		sp.Max = val
	}

	sp.Count += 1
	sp.Values[fmt.Sprintf("%v", x)] = val
}

func (sp *ScatterPlot) Read(errs chan error) {
	doRead(os.Stdin, errs, func(line string) {
		parts := strings.SplitN(strings.TrimSpace(line), " ", 2)
		if len(parts) != 2 {
			sp.Add(0, 0, errors.New("invalid line"))
			return
		}
		parsedX, err := Parse(parts[0])
		if err != nil {
			sp.Add(0, 0, err)
			return
		}
		parsedVal, err := Parse(parts[1])
		sp.Add(parsedX, parsedVal, err)
	})
}

type TimeSeries struct {
	Values map[string]Countable
	times  *list.List

	Layout string // the layout to use (interpreted by JS)
	Label  string // the label of the histogram
	Width  int    // the maximum graph width in pixels
	Height int    // the maximum graph height in pixels
	Window int    // the number of points to retain

	Colors   string // the colors to use when displaying the graph
	FontSize string // the CSS font size to use when displaying the graph

	Min Countable // the minimum value encountered so far
	Max Countable // the maximum value encountered so far

	Count    int // the number of values encountered so far
	Filtered int // the number of values filtered out so far
	Errors   int // the number of values skipped due to errors so far
}

func NewTimeSeries(window int, label string) *TimeSeries {
	return &TimeSeries{
		times:  list.New(),
		Layout: "time-series",
		Values: make(map[string]Countable, 1024),
		Window: window,
		Label:  label,
		Min:    Countable(math.Inf(1)),
		Max:    Countable(math.Inf(-1))}
}

func (ts *TimeSeries) Changed(indicator int) (bool, int) {
	if ts.Count <= indicator {
		return false, indicator
	}
	return true, ts.Count
}

// TODO interface Collection with methods for updating min/max/etc.

func (ts *TimeSeries) Add(when time.Time, val Countable, err error) {
	if err != nil {
		ts.Errors += 1
		return
	} else if val < Countable(*min) || val > Countable(*max) {
		ts.Filtered += 1
		return
	}

	if val < ts.Min {
		ts.Min = val
	}
	if val > ts.Max {
		ts.Max = val
	}

	ts.Count += 1
	key := when.Format(time.RFC3339Nano)
	ts.times.PushBack(key)
	ts.Values[key] = val
	if ts.times.Len() > ts.Window {
		drop := ts.times.Front()
		ts.times.Remove(drop)
		dropped := drop.Value.(string)
		delete(ts.Values, dropped)
	}
}

func (ts *TimeSeries) Read(errors chan error) {
	doRead(os.Stdin, errors, func(line string) {
		parsed, err := Parse(strings.TrimSpace(line))
		ts.Add(time.Now(), parsed, err)
	})
}

// Collects and buckets values. Stats (min, max, total, etc.) are computed as
// countable values come in.
type Histogram struct {
	Values map[string]Countable

	Layout string // the layout to use (interpreted by JS)
	Bucket int    // the histogram bucket size
	Label  string // the label of the histogram
	Wide   bool   // whether to use the alternate wide graph orientation
	Width  int    // the maximum graph width in pixels
	Height int    // the maximum graph height in pixels

	Colors   string // the colors to use when displaying the graph
	FontSize string // the CSS font size to use when displaying the graph

	Min Countable // the minimum value encountered so far
	Max Countable // the maximum value encountered so far
	Sum Countable // the sum of values encountered so far

	Count    int // the number of values encountered so far
	Filtered int // the number of values filtered out so far
	Errors   int // the number of values skipped due to errors so far
}

// Returns a new histogram. The bucket size is used to count values that
// fall within a different size. The `label` and `wide` options control
// the display of the rendered graph.
func NewHistogram(bucket int, label string, wide bool) *Histogram {
	return &Histogram{
		Layout: "histogram",
		Values: make(map[string]Countable, 1024),
		Bucket: bucket,
		Label:  label,
		Wide:   wide,
		Min:    Countable(math.Inf(1)),
		Max:    Countable(math.Inf(-1))}
}

// Returns whether the graph has changed since the `indicator` value was
// returned, and a new indicator if it has..
func (hist *Histogram) Changed(indicator int) (bool, int) {
	if hist.Count <= indicator {
		return false, indicator
	}
	return true, hist.Count
}

// Adds a countable value, modifying the stats and counts accordingly.
func (hist *Histogram) Add(val Countable, err error) {
	if err != nil {
		hist.Errors += 1
		return
	} else if val < Countable(*min) || val > Countable(*max) {
		hist.Filtered += 1
		return
	}

	if val < hist.Min {
		hist.Min = val
	}
	if val > hist.Max {
		hist.Max = val
	}
	hist.Sum += val
	hist.Count += 1
	hist.Values[val.Bucket(hist.Bucket)] += 1
}

// Read and parse countable values from stdin, add them to a histogram and
// update stats.
func (hist *Histogram) Read(errors chan error) {
	logger("starting to read data")
	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			logger("finished reading data due to %v", err)
			errors <- err
			return
		}
		hist.Add(Parse(strings.TrimSpace(line)))
	}
}

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

func logger(format string, v ...interface{}) {
	if !*verbose {
		return
	}
	log.Printf(format+"\n", v...)
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

func logRequest(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger("(%v) starting %v %v", r.RemoteAddr, r.Method, r.URL)
		defer logger("(%v) finished %v %v", r.RemoteAddr, r.Method, r.URL)
		handler(w, r)
	}
}

func buildGraph(arg string) Graph {
	switch arg {
	case "histogram":
		graph := NewHistogram(*bucket, *label, *wide)
		graph.Width = *width
		graph.Height = *height
		graph.Colors = *colors
		graph.FontSize = *fontSize
		return graph
	case "time-series":
		graph := NewTimeSeries(65, *label)
		graph.Width = *width
		graph.Height = *height
		graph.Colors = *colors
		graph.FontSize = *fontSize
		return graph
	case "scatterplot":
		graph := NewScatterPlot(*label)
		graph.Width = *width
		graph.Height = *height
		graph.Colors = *colors
		graph.FontSize = *fontSize
		return graph
	case "logfile":
		graph := NewLogFile(5, *label)
		graph.Colors = *colors
		graph.FontSize = *fontSize
		return graph
		// TODO window
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
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	flag.Parse()

	graph := buildGraph(flag.Arg(0))
	// TODO have graph return a FlagSet

	readerrors := make(chan error)
	watchers := make(ErrorWatchers, 0)
	go watchers.Broadcast(readerrors)

	go graph.Read(readerrors)

	ticker := time.NewTicker(time.Duration(*delay) * time.Second)

	indexfile := bundle.ReadFile("index.html")
	indexpage := template.Must(template.New("index").Parse(string(indexfile)))
	http.HandleFunc("/", logRequest(func(w http.ResponseWriter, r *http.Request) {
		msg, err := json.Marshal(&graph)
		if err != nil {
			fmt.Println("FAIL", err)
			return
		}
		indexpage.Execute(w, string(msg))
	}))

	scriptfile := bytes.NewReader(bundle.ReadFile("script.js"))
	http.HandleFunc("/script.js", logRequest(func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, "script.js", time.Now(), scriptfile)
	}))

	http.HandleFunc("/data", logRequest(func(w http.ResponseWriter, r *http.Request) {
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

	logger("listening on %v", *listen)
	http.ListenAndServe(*listen, nil)
}
