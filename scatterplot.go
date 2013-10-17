package graphblast

import (
	"errors"
	"fmt"
	"io"
	"math"
	"strings"
)

type ScatterPlot struct {
	Values map[string]Countable

	Layout string // the layout to use (interpreted by JS)
	Label  string // the label of the histogram
	Width  int    // the maximum graph width in pixels
	Height int    // the maximum graph height in pixels

	Allowed Range

	Colors   string // the colors to use when displaying the graph
	FontSize string // the CSS font size to use when displaying the graph

	Min Countable // the minimum value encountered so far
	Max Countable // the maximum value encountered so far

	Count    int // the number of values encountered so far
	Filtered int // the number of values filtered out so far
	Errors   int // the number of values skipped due to errors so far
}

func NewScatterPlot() *ScatterPlot {
	return &ScatterPlot{
		Layout:  "scatterplot",
		Values:  make(map[string]Countable, 1024),
		Allowed: Range{Countable(math.Inf(-1)), Countable(math.Inf(1))},
		Min:     Countable(math.Inf(1)),
		Max:     Countable(math.Inf(-1))}
}

func (sp *ScatterPlot) Changed(indicator int) (bool, int) {
	if sp.Count <= indicator {
		return false, indicator
	}
	return true, sp.Count
}

func (sp *ScatterPlot) Add(x Countable, val Countable, err error) {
	if err != nil {
		sp.Errors += 1
		return
	} else if !sp.Allowed.Contains(val) {
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
	// TODO This is an ugly hack
	sp.Values[fmt.Sprintf("%v|%v", x, sp.Count)] = val
}

func (sp *ScatterPlot) Read(reader io.Reader) error {
	return doRead(reader, func(line string) {
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
