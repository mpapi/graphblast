.PHONY: lint test

graphblast: graphblast.go script.js index.html lint test
	go build
	objcopy graphblast \
		--add-section script.js=script.js \
		--add-section index.html=index.html \
		graphblast.out
	mv graphblast.out graphblast
	strip graphblast

test: graphblast.go
	go test -v

lint: script.js
	jshint script.js
