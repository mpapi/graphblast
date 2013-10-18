package graphblast

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/hut8labs/graphblast/bind"
	"github.com/hut8labs/graphblast/bundle"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

const DEFAULT_GRAPH_NAME = "_"

func Index() http.HandlerFunc {
	indexfile := bundle.ReadFile("assets/index.html")
	indexpage := template.Must(template.New("index").Parse(string(indexfile)))

	namePattern := regexp.MustCompile("^/(?P<name>\\w+)")

	return LogRequest(func(w http.ResponseWriter, r *http.Request) {
		params := ExtractNamed(r.URL.Path, namePattern)
		if len(params["name"]) == 0 {
			params["name"] = DEFAULT_GRAPH_NAME
		}
		indexpage.Execute(w, params["name"])
	})
	// TODO Consider building the JS into the HTML, and removing Script()
}

func Script() http.HandlerFunc {
	scriptfile := bytes.NewReader(bundle.ReadFile("assets/script.js"))
	return LogRequest(func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, "script.js", time.Now(), scriptfile)
	})
}

// ExtractNamed returns a map from capture group names to their matched values
// when the regexp pattern is matched against text.
func ExtractNamed(text string, pattern *regexp.Regexp) map[string]string {
	names := pattern.SubexpNames()
	matches := pattern.FindStringSubmatch(text)

	result := make(map[string]string, len(names)-1)
	for index, name := range names {
		if index == 0 {
			continue
		}
		if index >= len(matches) {
			break
		}
		result[name] = matches[index]
	}
	return result
}

// ParseGraphURL returns a graph name and a graph object built from the
// path and query parameters of the URL. A regular expression is used to
// extract the graph name and type, and is expected to have two named capture
// groups: "name" for the graph name and "type" for the graph type.
func ParseGraphURL(url *url.URL, pattern *regexp.Regexp) (string, Graph) {
	parts := ExtractNamed(url.Path, pattern)
	graphType, ok := parts["type"]
	if !ok {
		return "", nil
	}

	graph := NewGraphFromType(graphType)
	if graph == nil {
		return "", nil
	}

	boundOk := bind.Bind(graph, bind.Parameters(url.Query()))
	if !boundOk {
		return "", nil
	}

	graphName, ok := parts["name"]
	if !ok {
		return "", nil
	}
	return graphName, graph
}

// Inputs returns a HandlerFunc for responding to requests for streaming
// uploads of graph data. When called, the handler creates a new graph based on
// the URL arguments and reads from the request body until the response is
// finished.
func Inputs(requests chan<- GraphRequest) http.HandlerFunc {
	graphPattern := regexp.MustCompile("^/graph/(?P<type>\\w+)/(?P<name>\\w+)")
	return LogRequest(func(w http.ResponseWriter, r *http.Request) {
		name, graph := ParseGraphURL(r.URL, graphPattern)
		if graph == nil {
			http.Error(w, "Invalid graph type or parameters", 400)
		} else {
			PopulateGraph(name, graph, r.Body, requests)
		}
	})
}

// Events returns a HandlerFunc for responding to requests for updates via an
// HTML EventSource (a.k.a. SSE, server-sent events). When called, the handler
// listens for graph data and pushes it to the client as JSON.
func Events(requests chan<- GraphRequest, publisher Publisher) http.HandlerFunc {
	return LogRequest(func(w http.ResponseWriter, r *http.Request) {
		// Get the necessary parts for being an EventSource, or fail.
		flusher, cn, err := toEventSource(w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		messages := publisher.Subscribe(r.RemoteAddr)
		defer publisher.Unsubscribe(r.RemoteAddr)

		requests <- DumpGraphs(r.RemoteAddr)
		for {
			select {
			case _ = <-cn.CloseNotify():
				return

			case msg := <-messages:
				envelope := msg.Envelope()
				contents, msgErr := msg.Contents()
				Log("Sending: %s %s", envelope, string(contents))
				if msgErr != nil {
					// TODO Log it
					continue
				}
				io.WriteString(w, fmt.Sprintf("event: %s\n", envelope))
				io.WriteString(w, fmt.Sprintf("data: %s\n\n", string(contents)))
				flusher.Flush()
			}
		}
	})
}

// Sets up a ResponseWriter for use as an EventSource.
func toEventSource(w http.ResponseWriter) (http.Flusher, http.CloseNotifier, error) {
	f, canf := w.(http.Flusher)
	cn, cancn := w.(http.CloseNotifier)

	if !canf || !cancn {
		return f, cn, errors.New("connection not suitable for EventSource")
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	return f, cn, nil
}
