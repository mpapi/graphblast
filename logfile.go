package graphblast

import (
	"fmt"
	"io"
	"strings"
)

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

func (lf *LogFile) Read(reader io.Reader, errs chan error) {
	doRead(reader, errs, func(line string) {
		// TODO Capture timestamps for log lines
		lf.Add(strings.TrimSpace(line), nil)
	})
}
