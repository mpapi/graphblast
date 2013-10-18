package graphblast

import (
	"bufio"
	"io"
	"strconv"
	"time"
)

type Range struct {
	Min Countable
	Max Countable
}

func (r *Range) Contains(c Countable) bool {
	return c >= r.Min && c <= r.Max
}

func (r *Range) MarshalJSON() ([]byte, error) {
	return []byte("null"), nil
}

type Graph interface {
	Changed(int) (bool, int)
	Read(io.Reader) error
	// TODO Make it possible to determine and send deltas
}

// NewGraphFromType returns an unconfigured graph object for the type of graph
// corresponding to graphType, or nil if there is no type of graph with that
// name.
func NewGraphFromType(graphType string) Graph {
	switch graphType {
	case "logfile":
		return NewLogFile()
	case "timeseries":
		return NewTimeSeries()
	case "scatterplot":
		return NewScatterPlot()
	case "histogram":
		return NewHistogram()
	default:
		return nil
	}
}

type Graphs struct {
	named   map[string]Graph
	changed map[string]int
}

// GraphRequest sequences modifications to an internal collection of Graphs.
type GraphRequest func(*Graphs, Subscribers)

// CreateGraph adds a new Graph to a collection.
func CreateGraph(name string, graph Graph) GraphRequest {
	return func(graphs *Graphs, subs Subscribers) {
		graphs.named[name] = graph
		body := map[string]string{"name": name}
		subs.Send(NewJSONMessage("__created", body))
	}
}

// CompleteGraph notifies subscribers that a Graph is no longer being updated,
// possibly due to an error.
func CompleteGraph(name string, err error) GraphRequest {
	return func(graphs *Graphs, subs Subscribers) {
		if err != nil {
			graphs.changed[name] = 0
			body := map[string]string{"name": name, "reason": err.Error()}
			subs.Send(NewJSONMessage("__completed", body))
		}
	}
}

// DumpGraphs sends all Graphs in a collection to a single subscriber.
func DumpGraphs(subscriber string) GraphRequest {
	return func(graphs *Graphs, subs Subscribers) {
		for name, graph := range graphs.named {
			CreateGraph(name, graph)(graphs, subs)
			subs.Send(NewJSONMessageTo([]string{subscriber}, name, graph))
		}
	}
}

// NotifyChanges sends all Graphs that have changed (since the last call to
// NotifyChanges) to all subscribers.
func NotifyChanges() GraphRequest {
	return func(graphs *Graphs, subs Subscribers) {
		for name, graph := range graphs.named {
			changed, indicator := graph.Changed(graphs.changed[name])
			graphs.changed[name] = indicator
			if !changed {
				continue
			}
			subs.Send(NewJSONMessage(name, graph))
		}
	}
}

// ProcessGraphRequests maintains an internal collection of Graphs, listens for
// GraphRequests, and applies them to the collection.
func ProcessGraphRequests(requests <-chan GraphRequest, subs Subscribers) {
	graphs := &Graphs{make(map[string]Graph), make(map[string]int)}
	for requestFunc := range requests {
		requestFunc(graphs, subs)
	}
}

// PeriodicallyNotifyChanges sends a NotifyChanges request on a channel every n
// seconds.
func PeriodicallyNotifyChanges(requests chan<- GraphRequest, seconds int) {
	updateFreq := time.Duration(seconds) * time.Second
	for _ = range time.Tick(updateFreq) {
		requests <- NotifyChanges()
	}
}

// PopulateGraph is a convenience function for updating a Graph from an input
// stream, sending creation and completion requests at the appropriate times.
func PopulateGraph(name string, graph Graph, input io.Reader, requests chan<- GraphRequest) {
	requests <- CreateGraph(name, graph)
	var err error
	defer func() {
		requests <- CompleteGraph(name, err)
	}()
	err = graph.Read(input)
}

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
	if size <= 0 {
		size = 1
	}
	if d < 0 {
		d -= Countable(size)
	}
	return strconv.Itoa(int(d) / size * size)
}

func doRead(input io.Reader, process func(string)) error {
	Log("starting to read data")
	reader := bufio.NewReader(input)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			Log("finished reading data due to %v", err)
			return err
		}
		process(line)
	}
}
