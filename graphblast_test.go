package main

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

func Test_CountableParse(t *testing.T) {
	c, err := Parse("0")
	if c != 0 && err != nil {
		t.Error("parse did not parse an integer")
	}

	c, err = Parse("-100e5")
	if c != -100e5 && err != nil {
		t.Error("parse did not parse an negative float")
	}

	c, err = Parse("3.1415926535")
	if c != 3.1415926535 && err != nil {
		t.Error("parse did not parse an negative float")
	}

	c, err = Parse("3.1415a")
	if c != 0 || err == nil {
		t.Error("parse did parsed an invalid number")
	}

	c, err = Parse("foo")
	if c != 0 || err == nil {
		t.Error("parse did parsed an invalid number")
	}

	c, err = Parse("")
	if c != 0 || err == nil {
		t.Error("parse did parsed an invalid number")
	}
}

func Test_CountableBucket(t *testing.T) {
	if Countable(4).Bucket(1) != "4" {
		t.Error("bucket failed on int for bucket of size 1")
	}

	if Countable(4.9).Bucket(1) != "4" {
		t.Error("bucket failed on float for bucket of size 1")
	}

	if Countable(0.1).Bucket(1) != "0" {
		t.Error("bucket failed on small float for bucket of size 1")
	}

	if Countable(-0.1).Bucket(1) != "-1" {
		t.Error("bucket failed on negative float for bucket of size 1")
	}

	if Countable(4).Bucket(5) != "0" {
		t.Error("bucket failed on int for bucket of size 5")
	}

	if Countable(4.9).Bucket(5) != "0" {
		t.Error("bucket failed on float for bucket of size 5")
	}

	if Countable(0.1).Bucket(5) != "0" {
		t.Error("bucket failed on small float for bucket of size 5")
	}

	if Countable(-0.1).Bucket(5) != "-5" {
		t.Error("bucket failed on negative float for bucket of size 5")
	}
}
