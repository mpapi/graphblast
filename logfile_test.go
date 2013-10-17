package graphblast

import (
	"errors"
	"strings"
	"testing"
)

func TestLogFileAdd(t *testing.T) {
	lf := NewLogFile()
	lf.Add("line1", nil)
	if lf.Count != 1 {
		t.Errorf("Add didn't increment the count")
	}
	if lf.Errors != 0 {
		t.Errorf("Add incremented the error count")
	}
	if len(lf.Values) != 1 {
		t.Errorf("Add didn't store the data")
	}
	if lf.Values["0"] != "line1" {
		t.Errorf("Add didn't correctly store the right data")
	}
}

func TestLogFileAddError(t *testing.T) {
	lf := NewLogFile()
	lf.Add("line1", errors.New("fail"))
	if lf.Errors != 1 {
		t.Errorf("Add with error didn't increment the error count")
	}
	if lf.Count != 0 {
		t.Errorf("Add with error incremented the count")
	}
	if len(lf.Values) != 0 {
		t.Errorf("Add with error stored some data")
	}
}

func TestLogFileAddWindow(t *testing.T) {
	lf := NewLogFile()
	lf.Window = 1
	lf.Add("line1", nil)
	lf.Add("line2", nil)
	if lf.Errors != 0 {
		t.Errorf("Add incremented the error count")
	}
	if lf.Count != 2 {
		t.Errorf("Add didn't increment the count")
	}
	if len(lf.Values) != 1 {
		t.Errorf("Add didn't window the data")
	}
	if lf.Values["1"] != "line2" {
		t.Errorf("Add didn't correctly drop data when windowing")
	}
}

func TestLogFileChanged(t *testing.T) {
	lf := NewLogFile()
	changed, next := lf.Changed(0)
	if changed {
		t.Error("log should be unchanged")
	}
	if next != 0 {
		t.Error("Changed should return indicator when nothing has changed")
	}

	lf.Add("line1", nil)
	changed, next = lf.Changed(0)
	if !changed {
		t.Error("log should be changed")
	}
	if next <= 0 {
		t.Error("Changed should return a new indicator after changes")
	}
}

func TestLogFileRead(t *testing.T) {
	reader := strings.NewReader("line1\nline2\n")

	lf := NewLogFile()
	lf.Read(reader)
	if lf.Count != 2 {
		t.Error("Read failed to read the input fully")
	}
	if lf.Errors != 0 {
		t.Error("Read failed to read the input without errors")
	}
	if lf.Values["1"] != "line2" {
		t.Error("Read failed to read the correct values")
	}
}
