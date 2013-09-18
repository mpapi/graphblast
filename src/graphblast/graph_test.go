package graphblast

import (
	"testing"
)

func TestRangeContains(t *testing.T) {
	r := Range{Min: -1, Max: 1}
	if !r.Contains(0) {
		t.Error("contains failed for a number in range")
	}
	if !r.Contains(-1) {
		t.Error("contains failed to be inclusive for min")
	}
	if !r.Contains(1) {
		t.Error("contains failed to be inclusive for max")
	}
	if r.Contains(-1.1) {
		t.Error("contains failed for a number less than min")
	}
	if r.Contains(1.1) {
		t.Error("contains failed for a number greater than max")
	}
}

func TestCountableParse(t *testing.T) {
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

func TestCountableBucket(t *testing.T) {
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
