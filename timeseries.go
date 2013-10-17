package graphblast

import (
	"container/list"
	"io"
	"math"
	"strings"
	"time"
)

type TimeSeries struct {
	Values map[string]Countable
	times  *list.List

	Layout string // the layout to use (interpreted by JS)
	Label  string // the label of the histogram
	Width  int    // the maximum graph width in pixels
	Height int    // the maximum graph height in pixels
	Window int    // the number of points to retain

	Allowed Range

	Colors   string // the colors to use when displaying the graph
	FontSize string // the CSS font size to use when displaying the graph

	Min Countable // the minimum value encountered so far
	Max Countable // the maximum value encountered so far

	Count    int // the number of values encountered so far
	Filtered int // the number of values filtered out so far
	Errors   int // the number of values skipped due to errors so far
}

func NewTimeSeries() *TimeSeries {
	return &TimeSeries{
		times:   list.New(),
		Layout:  "time-series",
		Window:  100,
		Values:  make(map[string]Countable, 1024),
		Allowed: Range{Countable(math.Inf(-1)), Countable(math.Inf(1))},
		Min:     Countable(math.Inf(1)),
		Max:     Countable(math.Inf(-1))}
}

func (ts *TimeSeries) Changed(indicator int) (bool, int) {
	if ts.Count <= indicator {
		return false, indicator
	}
	return true, ts.Count
}

func (ts *TimeSeries) Add(when time.Time, val Countable, err error) {
	if err != nil {
		ts.Errors += 1
		return
	} else if !ts.Allowed.Contains(val) {
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

func (ts *TimeSeries) Read(reader io.Reader) error {
	return doRead(reader, func(line string) {
		parsed, err := Parse(strings.TrimSpace(line))
		ts.Add(time.Now(), parsed, err)
	})
}
