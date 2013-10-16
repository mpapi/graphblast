package bind

import (
	"testing"
)

type TestStruct struct {
	Foo   int    `default:"1000" help:"the initial value of foo"`
	Bar   string `help:"bleh"`
	Quux  []string
	Barf  map[string]string
	Float float64
	Bool  bool
}

func TestBind(t *testing.T) {
	f := &TestStruct{}
	Bind(f, map[string][]string{
		"foo":   []string{"1"},
		"bar":   []string{"2"},
		"quux":  []string{"3", "4", "5"},
		"float": []string{"6"},
		"bool":  []string{"true"},
	})
	if f.Foo != 1 {
		t.Error("Failed to bind int value")
	}
	if f.Bar != "2" {
		t.Error("Failed to bind string value")
	}
	if len(f.Quux) != 3 {
		t.Error("Failed to bind string slice value")
	}
	if f.Float != 6 {
		t.Error("Failed to bind float value")
	}
	if !f.Bool {
		t.Error("Failed to bind bool value")
	}
}

func TestGenerateFlags(t *testing.T) {
	f := &TestStruct{}
	flagSet, ok := GenerateFlags(f, "yep")
	if !ok {
		t.Error("Failed to generate flags")
	}

	flagSet.Parse([]string{"-foo=39", "-bar=blah"})
	if f.Foo != 39 {
		t.Error("Failed to bind int value")
	}
	if f.Bar != "blah" {
		t.Error("Failed to bind string value")
	}
	if len(f.Quux) != 0 {
		t.Error("Failed to bind string slice value")
	}
	if f.Float != 0 {
		t.Error("Failed to bind float value")
	}
	if f.Bool {
		t.Error("Failed to bind bool value")
	}
}
