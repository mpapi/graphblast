package graphblast

import (
	"fmt"
	"regexp"
)

func ExampleExtractNamed() {
	pattern := regexp.MustCompile("^(?P<str>\\w+)\\s*(?P<num>\\d+)$")
	params := ExtractNamed("mystring 9999", pattern)
	fmt.Println(params["str"])
	fmt.Println(params["num"])
	// Output:
	// mystring
	// 9999
}
