package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var label = flag.String("label", "", "graph label")
var min = flag.Float64("min", math.Inf(-1), "minimum accepted value")
var max = flag.Float64("max", math.Inf(1), "maximum accepted value")
var bucket = flag.Int("bucket", 1, "histogram bucket size")
var delay = flag.Int("delay", 5, "delay between updates, in seconds")
var wide = flag.Bool("wide", false, "use wide orientation")

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
	if d < 0 {
		d -= Countable(size)
	}
	return strconv.Itoa(int(d) / size * size)
}

// Stats are computed as countable values come in, and are sent out with
// histogram data in Messages.
type Stats struct {
	Min      Countable // the minimum value encountered so far
	Max      Countable // the maximum value encountered so far
	Total    Countable // the sum of values encountered so far
	Values   int       // the number of values encountered so far
	Filtered int       // the number of values filtered out so far
	Bucket   int       // the histogram bucket size
	Label    string    // the optional graph label
}

// Adds a countable value to Stats, modifying the stats accordingly.
func (s *Stats) Add(c Countable) {
	if c < s.Min {
		s.Min = c
	}
	if c > s.Max {
		s.Max = c
	}
	s.Total += c
	s.Values += 1
}

func (s *Stats) AddFiltered() {
	s.Filtered += 1
}

// Wraps a histogram and stats, for JSON encoding.
type Message struct {
	Stats     *Stats
	Histogram *map[string]int
	Wide      bool
}

// Read and parse countable values from stdin, add them to a histogram and
// update stats.
func Read(hist *map[string]int, stats *Stats) {
	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			// TODO signal done
			break
		} else if err != nil {
			// TODO signal error
			break
		}

		key, err := Parse(strings.TrimSpace(line))
		if err != nil {
			// TODO signal failed
			continue
		} else if key < Countable(*min) || key > Countable(*max) {
			stats.AddFiltered()
			continue
		}
		stats.Add(key)
		(*hist)[key.Bucket(*bucket)] += 1
	}
}

func main() {
	flag.Parse()

	hist := make(map[string]int)
	stats := &Stats{0, 0, 0, 0, 0, *bucket, *label}
	ticker := time.NewTicker(time.Duration(*delay) * time.Second)

	go Read(&hist, stats)

	indexpage := template.Must(template.ParseFiles("index.html"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		msg, err := json.Marshal(Message{stats, &hist, *wide})
		if err != nil {
			fmt.Println("FAIL", err)
			return
		}
		indexpage.Execute(w, string(msg))
	})

	http.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {
		f, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "no flusher", http.StatusInternalServerError)
			return
		}

		cn, ok := w.(http.CloseNotifier)
		if !ok {
			http.Error(w, "no close notifier", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		writeHist := func() bool {
			msg, err := json.Marshal(Message{stats, &hist, *wide})
			if err != nil {
				fmt.Println("FAIL", err)
				fmt.Fprint(w, "data: {\"type\": \"error\"}\n\n")
				return false
			}
			fmt.Fprintf(w, "data: %s\n\n", msg)
			f.Flush()
			return true
		}

		if !writeHist() {
			return
		}

		lastvalues := 0
		for {
			select {
			case _ = <-cn.CloseNotify():
				return

			case _ = <-ticker.C:
				if stats.Values <= lastvalues {
					continue
				}
				lastvalues = stats.Values

				if !writeHist() {
					break
				}
			}
		}
	})
	http.ListenAndServe(":8080", nil)
}
