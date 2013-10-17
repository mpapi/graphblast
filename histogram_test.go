package graphblast

import (
	"errors"
	"strings"
	"testing"
)

func TestHistogramAdd(t *testing.T) {
	hist := NewHistogram()
	hist.Add(1, nil)

	if len(hist.Values) != 1 {
		t.Error("histogram did not record count")
	}
	if hist.Values["1"] != 1 {
		t.Error("histogram recorded wrong count")
	}
	if hist.Min != 1 || hist.Max != 1 || hist.Sum != 1 {
		t.Error("histogram recorded wrong value stat")
	}
	if hist.Count != 1 || hist.Filtered != 0 || hist.Errors != 0 {
		t.Error("histogram recorded wrong count stat")
	}

	hist.Add(-0.5, nil)

	if len(hist.Values) != 2 {
		t.Error("histogram did not record count")
	}
	if hist.Values["-1"] != 1 {
		t.Error("histogram recorded wrong count")
	}
	if hist.Min != -0.5 || hist.Max != 1 || hist.Sum != 0.5 {
		t.Error("histogram recorded wrong value stat")
	}
	if hist.Count != 2 || hist.Filtered != 0 || hist.Errors != 0 {
		t.Error("histogram recorded wrong count stat")
	}
}

func TestHistogramAddError(t *testing.T) {
	hist := NewHistogram()
	hist.Add(1, errors.New("fail"))

	if len(hist.Values) != 0 {
		t.Error("histogram recorded count for error")
	}
	if hist.Count != 0 || hist.Filtered != 0 || hist.Errors != 1 {
		t.Error("histogram recorded wrong count stat")
	}
}

func TestHistogramAddFiltered(t *testing.T) {
	hist := NewHistogram()
	hist.Allowed = Range{Countable(-1), Countable(0)}
	hist.Add(1, nil)

	if len(hist.Values) != 0 {
		t.Error("histogram recorded count for filtered")
	}
	if hist.Count != 0 || hist.Filtered != 1 || hist.Errors != 0 {
		t.Error("histogram recorded wrong count stat")
	}
}

func TestHistogramChanged(t *testing.T) {
	hist := NewHistogram()
	changed, next := hist.Changed(0)
	if changed {
		t.Error("histogram should be unchanged")
	}
	if next != 0 {
		t.Error("Changed should return indicator when nothing has changed")
	}

	hist.Add(1, nil)
	changed, next = hist.Changed(0)
	if !changed {
		t.Error("histogram should be changed")
	}
	if next <= 0 {
		t.Error("Changed should return a new indicator after changes")
	}
}

func TestHistogramRead(t *testing.T) {
	hist := NewHistogram()
	reader := strings.NewReader("5\n5\n")
	hist.Read(reader)
	if hist.Count != 2 {
		t.Error("Read failed to read the input fully")
	}
	if hist.Errors != 0 {
		t.Error("Read failed to read the input without errors")
	}
	if hist.Values["5"] != 2 {
		t.Error("Read failed to read the correct values")
	}

	reader = strings.NewReader("5\na\n")
	hist.Read(reader)
	if hist.Count != 3 {
		t.Error("Read failed to read the good part of a bad input")
	}
	if hist.Errors != 1 {
		t.Error("Read failed to signal an error in the input")
	}
	if hist.Values["5"] != 3 {
		t.Error("Read failed to read the correct values")
	}
}
