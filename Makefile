.PHONY: all deps lint test coverage

all: lint test bin/graphblast

bin/graphblast: *.go bundle/*.go graphblast/*.go script.js index.html
	go build -o bin/graphblast graphblast/graphblast.go
	objcopy bin/graphblast \
		--add-section script.js=script.js \
		--add-section index.html=index.html \
		bin/graphblast.out
	mv bin/graphblast.out bin/graphblast
	strip bin/graphblast

test: *.go bundle/*.go graphblast/*.go
	go test github.com/hut8labs/graphblast

coverage:
	./bin/gocov test -v github.com/hut8labs/graphblast > .coverage.json
	./bin/gocov annotate -ceiling=100 .coverage.json

lint: script.js
	jshint script.js

deps:
	npm update jshint
	go get -u github.com/axw/gocov
