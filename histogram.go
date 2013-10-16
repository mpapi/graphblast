package graphblast

import (
	"io"
	"math"
	"strings"
)

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

	Allowed Range

	Colors   string // the colors to use when displaying the graph
	FontSize string // the CSS font size to use when displaying the graph

	Min Countable // the minimum value encountered so far
	Max Countable // the maximum value encountered so far
	Sum Countable // the sum of values encountered so far

	Count    int // the number of values encountered so far
	Filtered int // the number of values filtered out so far
	Errors   int // the number of values skipped due to errors so far
}

// Returns a new histogram.
func NewHistogram() *Histogram {
	return &Histogram{
		Layout:  "histogram",
		Values:  make(map[string]Countable, 1024),
		Wide:    false,
		Allowed: Range{Countable(math.Inf(-1)), Countable(math.Inf(1))},
		Min:     Countable(math.Inf(1)),
		Max:     Countable(math.Inf(-1))}
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
	} else if !hist.Allowed.Contains(val) {
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
func (hist *Histogram) Read(reader io.Reader, errors chan error) {
	doRead(reader, errors, func(line string) {
		hist.Add(Parse(strings.TrimSpace(line)))
	})
}
