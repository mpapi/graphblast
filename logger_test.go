package graphblast

import (
	"bytes"
	"log"
	"testing"
)

func TestLogNoVerbose(t *testing.T) {
	output := new(bytes.Buffer)
	SetLogger(log.New(output, "", 0))
	SetVerboseLogging(false)
	Log("foo")
	if output.String() != "" {
		t.Errorf("Log wrote data even though !verbose")
	}
}

func TestLogVerbose(t *testing.T) {
	output := new(bytes.Buffer)
	SetLogger(log.New(output, "", 0))
	SetVerboseLogging(true)
	Log("foo")
	logged := output.String()
	if logged == "" {
		t.Errorf("Log wrote no data even though verbose")
	}
	if logged != "foo\n" {
		t.Errorf("Log wrote wrong data (%#v)", logged)
	}
}
