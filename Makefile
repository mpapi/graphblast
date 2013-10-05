.PHONY: all lint test coverage

all: lint test graphblast

graphblast: graphblast.go src/graphblast/*.go script.js index.html
	go build
	objcopy graphblast \
		--add-section script.js=script.js \
		--add-section index.html=index.html \
		graphblast.out
	mv graphblast.out graphblast
	strip graphblast

test:
	go test -v . graphblast bundle

coverage:
	./bin/gocov test -v graphblast > report.json
	./bin/gocov annotate -ceiling=100 report.json

lint: script.js
	jshint script.js
