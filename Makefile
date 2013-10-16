.PHONY: all deps lint test coverage

all: lint test bin/graphblast

bin/graphblast: *.go bind/*.go bundle/*.go graphblast/*.go assets/*
	go build -o bin/graphblast graphblast/graphblast.go
	objcopy bin/graphblast \
		--add-section assets/script.js=assets/script.js \
		--add-section assets/index.html=assets/index.html \
		bin/graphblast.out
	mv bin/graphblast.out bin/graphblast
	strip bin/graphblast

test: *.go bind/*.go bundle/*.go graphblast/*.go
	go test github.com/hut8labs/graphblast

coverage:
	./bin/gocov test -v github.com/hut8labs/graphblast > .coverage.json
	./bin/gocov annotate -ceiling=100 .coverage.json

lint: assets/script.js
	jshint assets/script.js

deps:
	npm update jshint
	go get -u github.com/axw/gocov
