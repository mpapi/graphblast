package main

import (
	"testing"
)

func Test_StatsAdd(t *testing.T) {
	stats := Stats{0, 0, 0, 0, 0, 0, 1}
	stats.Add(1)

	expected := Stats{0, 1, 1, 1, 0, 0, 1}
	if stats != expected {
		t.Error("stats did not record count")
	}

	stats.Add(-0.5)
	expected = Stats{-0.5, 1, 0.5, 2, 0, 0, 1}
	if stats != expected {
		t.Error("stats did not record negative count")
	}
}

func Test_StatsAddFiltered(t *testing.T) {
	stats := Stats{0, 0, 0, 0, 0, 0, 1}
	stats.AddFiltered()

	expected := Stats{0, 0, 0, 0, 1, 0, 1}
	if stats != expected {
		t.Error("stats did not record filtered")
	}
}

func Test_StatsAddError(t *testing.T) {
	stats := Stats{0, 0, 0, 0, 0, 0, 1}
	stats.AddError()

	expected := Stats{0, 0, 0, 0, 0, 1, 1}
	if stats != expected {
		t.Error("stats did not record error")
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
