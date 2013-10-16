package graphblast

import (
	"bufio"
	"io"
	"strconv"
)

type Range struct {
	Min Countable
	Max Countable
}

func (r *Range) Contains(c Countable) bool {
	return c >= r.Min && c <= r.Max
}

func (r *Range) MarshalJSON() ([]byte, error) {
	return []byte("null"), nil
}

type Graph interface {
	Changed(int) (bool, int)
	Read(io.Reader, chan error)
	// TODO Make it possible to determine and send deltas
}

// The type of the items to parse from stdin and count in the histogram.
type Countable float64

// Parses a countable value from a string, and returns a non-nil error if
// parsing fails.
func Parse(str string) (Countable, error) {
	d, err := strconv.ParseFloat(str, 64)
	return Countable(d), err
}

// Returns the bucket (as a string) of which the countable value should
// increment the count, given the bucket size.
func (d Countable) Bucket(size int) string {
	if size <= 0 {
		size = 1
	}
	if d < 0 {
		d -= Countable(size)
	}
	return strconv.Itoa(int(d) / size * size)
}

func doRead(input io.Reader, errors chan error, process func(string)) {
	Log("starting to read data")
	reader := bufio.NewReader(input)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			Log("finished reading data due to %v", err)
			errors <- err
			return
		}
		process(line)
	}
}
