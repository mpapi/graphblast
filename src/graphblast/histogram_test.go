package graphblast

import (
	"errors"
	"testing"
)

func Test_HistogramAdd(t *testing.T) {
	hist := NewHistogram(1, "", false)
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

func Test_HistogramAddError(t *testing.T) {
	hist := NewHistogram(1, "", false)
	hist.Add(1, errors.New("fail"))

	if len(hist.Values) != 0 {
		t.Error("histogram recorded count for error")
	}
	if hist.Count != 0 || hist.Filtered != 0 || hist.Errors != 1 {
		t.Error("histogram recorded wrong count stat")
	}
}

func Test_HistogramAddFiltered(t *testing.T) {
	hist := NewHistogram(1, "", false)
	hist.Allowed = Range{Countable(-1), Countable(0)}
	hist.Add(1, nil)

	if len(hist.Values) != 0 {
		t.Error("histogram recorded count for filtered")
	}
	if hist.Count != 0 || hist.Filtered != 1 || hist.Errors != 0 {
		t.Error("histogram recorded wrong count stat")
	}
}

func Test_HistogramChanged(t *testing.T) {
	hist := NewHistogram(1, "", false)
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
