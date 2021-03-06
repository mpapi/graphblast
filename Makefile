.PHONY: all deps lint test coverage

OUTPUT = bin/graphblast
GO_SRC = *.go bundle/*.go bind/*.go graphblast/*.go
GO_PKG = github.com/hut8labs/graphblast \
		 github.com/hut8labs/graphblast/bind \
		 github.com/hut8labs/graphblast/bundle

all: lint test $(OUTPUT)

$(OUTPUT): $(GO_SRC) assets/*
	go build -o $@ graphblast/graphblast.go
	objcopy $@ $@.out \
		--add-section assets/script.js=assets/script.js \
		--add-section assets/index.html=assets/index.html
	mv $@.out $@
	strip $@

test: $(GO_SRC)
	go test -timeout=1s $(GO_PKG)

coverage:
	./bin/gocov test -v -timeout=1s $(GO_PKG) > .coverage.json
	./bin/gocov annotate -ceiling=100 .coverage.json

lint: $(GO_SRC) assets/script.js
	jshint assets/script.js
	gofmt -d $(GO_SRC) | tee /dev/stderr | wc -l | grep -q ^0$

deps:
	npm install jshint
	go get -u github.com/axw/gocov
