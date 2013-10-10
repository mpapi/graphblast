package graphblast

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestTimeSeriesAdd(t *testing.T) {
	ts := NewTimeSeries(2, "")
	when := time.Unix(1400000000, 0)
	ts.Add(when, 1, nil)

	if len(ts.Values) != 1 {
		t.Error("Add did not record value")
	}
	if ts.Values[when.Format(time.RFC3339Nano)] != 1 {
		t.Error("Add recorded wrong value")
	}
	if ts.Min != 1 && ts.Max != 1 {
		t.Error("Add recorded wrong value stat")
	}
}

func TestTimeSeriesAddError(t *testing.T) {
	ts := NewTimeSeries(2, "")
	when := time.Unix(1400000000, 0)
	ts.Add(when, 1, errors.New("fail"))

	if len(ts.Values) != 0 {
		t.Error("Add recorded value for error")
	}
	if ts.Count != 0 || ts.Filtered != 0 || ts.Errors != 1 {
		t.Error("Add recorded wrong count stat for error")
	}
}

func TestTimeSeriesAddFiltered(t *testing.T) {
	ts := NewTimeSeries(2, "")
	ts.Allowed = Range{Countable(0), Countable(100)}
	when := time.Unix(1400000000, 0)
	ts.Add(when, 101, nil)

	if len(ts.Values) != 0 {
		t.Error("Add recorded value for filtered")
	}
	if ts.Count != 0 || ts.Filtered != 1 || ts.Errors != 0 {
		t.Error("Add recorded wrong count stat for filtered")
	}
}

func TestTimeSeriesAddWindowed(t *testing.T) {
	ts := NewTimeSeries(1, "")
	ts.Add(time.Unix(1400000000, 0), 1, nil)
	ts.Add(time.Unix(1400000001, 0), 2, nil)

	if len(ts.Values) != 1 {
		t.Error("Add recorded too many values for windowed")
	}
	if ts.Values[time.Unix(1400000001, 0).Format(time.RFC3339Nano)] != 2 {
		t.Error("Add dropped the wrong value for windowed")
	}
	if ts.Count != 2 || ts.Filtered != 0 || ts.Errors != 0 {
		t.Error("Add recorded wrong count stat for filtered")
	}
}

func TestTimeSeriesChanged(t *testing.T) {
	ts := NewTimeSeries(1, "")
	changed, next := ts.Changed(0)
	if changed {
		t.Error("Changed incorrectly reported change")
	}
	if next != 0 {
		t.Error("Changed incorrectly returned new indicator")
	}

	ts.Add(time.Unix(1400000000, 0), 1, nil)
	changed, next = ts.Changed(0)
	if !changed {
		t.Error("Changed incorrectly reported no change")
	}
	if next <= 0 {
		t.Error("Changed did not return a larger indicator value")
	}
}

func TestTimeSeriesRead(t *testing.T) {
	ts := NewTimeSeries(1, "")
	reader := strings.NewReader("1\n2\n")
	errs := make(chan error, 2)
	ts.Read(reader, errs)
	if ts.Count != 2 {
		t.Error("Read failed to read the input fully")
	}
	if ts.Errors != 0 {
		t.Error("Read failed to read the input without errors")
	}
	if len(ts.Values) != 1 {
		t.Error("Read failed to discard values for windowing")
	}
	for _, val := range ts.Values {
		if val != 2 {
			t.Error("Read failed to discard the correct value when windowing")
		}
	}

	reader = strings.NewReader("5\na\n")
	ts.Read(reader, errs)
	if ts.Count != 3 {
		t.Error("Read failed to read the good part of a bad input")
	}
	if ts.Errors != 1 {
		t.Error("Read failed to signal an error in the input")
	}
	if len(ts.Values) != 1 {
		t.Error("Read failed to discard values for windowing")
	}
	for _, val := range ts.Values {
		if val != 5 {
			t.Error("Read failed to discard the correct value when windowing")
		}
	}
}
