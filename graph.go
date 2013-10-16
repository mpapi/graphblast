package graphblast

import (
	"bufio"
	"io"
	"strconv"
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
	Read(io.Reader, chan error)
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

// Graphs stores a set of graphs by name, and tracks their changes.
type Graphs struct {
	graphs      map[string]Graph
	lastChanged map[string]int
}

// NewGraphs returns an empty graph collection.
func NewGraphs() *Graphs {
	return &Graphs{
		graphs:      make(map[string]Graph, 1),
		lastChanged: make(map[string]int, 1)}
}

// Add stores a graph by name.
func (graphs *Graphs) Add(name string, graph Graph) {
	graphs.graphs[name] = graph
}

// All returns all graphs (by name) regardless of whether they've changed.
func (g *Graphs) All() map[string]Graph {
	return g.graphs
}

// Changed returns the graphs (by name) that have changed since the last call
// to the Changed method.
func (g *Graphs) Changed() (result map[string]Graph) {
	result = make(map[string]Graph)
	for name, graph := range g.graphs {
		changed, lastChanged := graph.Changed(g.lastChanged[name])
		if !changed {
			continue
		}
		g.lastChanged[name] = lastChanged
		result[name] = graph
	}
	return
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

func doRead(input io.Reader, errors chan error, process func(string)) {
	Log("starting to read data")
	reader := bufio.NewReader(input)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			Log("finished reading data due to %v", err)
			errors <- err
			return
		}
		process(line)
	}
}
