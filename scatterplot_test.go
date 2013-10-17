package graphblast

import (
	"errors"
	"strings"
	"testing"
)

func TestScatterPlotAdd(t *testing.T) {
	sp := NewScatterPlot()
	sp.Add(1, 1, nil)

	if len(sp.Values) != 1 {
		t.Error("Add did not record pair")
	}
	if sp.Values["1|1"] != 1 {
		t.Error("Add recorded wrong pair")
	}

	sp.Add(1, 2, nil)
	if len(sp.Values) != 2 {
		t.Error("Add did not record second pair")
	}
	if sp.Values["1|2"] != 2 {
		t.Error("Add recorded wrong second pair")
	}
}

func TestScatterPlotAddError(t *testing.T) {
	sp := NewScatterPlot()
	sp.Add(1, 1, errors.New("fail"))

	if len(sp.Values) != 0 {
		t.Error("Add recorded pair for error")
	}
	if sp.Count != 0 || sp.Filtered != 0 || sp.Errors != 1 {
		t.Error("Add recorded wrong count stat")
	}
}

func TestScatterPlotAddFiltered(t *testing.T) {
	sp := NewScatterPlot()
	sp.Allowed = Range{Countable(-1), Countable(1)}
	sp.Add(0, -100, nil)

	if len(sp.Values) != 0 {
		t.Error("Add didn't filter pair")
	}
	if sp.Count != 0 || sp.Filtered != 1 || sp.Errors != 0 {
		t.Error("Add recorded wrong count stat")
	}
}

func TestScatterPlotAddNoFilter(t *testing.T) {
	sp := NewScatterPlot()
	sp.Allowed = Range{Countable(-1), Countable(1)}
	sp.Add(-100, 0, nil)

	if len(sp.Values) != 1 {
		t.Error("Add did not record pair")
	}
	if sp.Values["-100|1"] != 0 {
		t.Error("Add recorded wrong pair")
	}
}

func TestScatterPlotChanged(t *testing.T) {
	sp := NewScatterPlot()
	changed, next := sp.Changed(0)
	if changed {
		t.Error("Scatterplot should be changed")
	}
	if next != 0 {
		t.Error("Changed should return indicator when nothing has changed")
	}

	sp.Add(1, 1, nil)
	changed, next = sp.Changed(0)
	if !changed {
		t.Error("Scatterplot should be changed")
	}
	if next <= 0 {
		t.Error("Changed should return a new indicator after changes")
	}
}

func TestScatterPlotRead(t *testing.T) {
	sp := NewScatterPlot()
	reader := strings.NewReader("10 100\n20 200\n")
	sp.Read(reader)

	if sp.Count != 2 {
		t.Error("Read failed to read the input fully")
	}
	if sp.Errors != 0 {
		t.Error("Read failed to read the input without errors")
	}
	if sp.Values["10|1"] != 100 || sp.Values["20|2"] != 200 {
		t.Error("Read failed to read the correct values")
	}

	reader = strings.NewReader("30 300\na\n")
	sp.Read(reader)
	if sp.Count != 3 {
		t.Error("Read failed to read the good part of a bad input")
	}
	if sp.Errors != 1 {
		t.Error("Read failed to signal an error in the input")
	}
	if sp.Values["30|3"] != 300 {
		t.Error("Read failed to read the correct values")
	}

	reader = strings.NewReader("a 100\n40 b\n")
	sp.Read(reader)
	if sp.Errors != 3 {
		t.Error("Read failed to signal an error for bad values")
	}
}
