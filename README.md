# Graphblast

`graphblast` reads numerical data from stdin and pushes it to your web browser
as a shareable graph that updates in realtime. It's the missing link between
your trusty Unix utilities -- `tail`, `grep`, `sed`, etc. -- and a
full-featured graphing package.

## Quick start

To procure, build, and run Graphblast, you'll need:

* [Go][] >= 1.1
* Git

If you're modifying the frontend, you'll also want Node.js for linting the
frontend JavaScript.

```shell
go get github.com/hut8labs/graphblast
cd $GOPATH/src/github.com/hut8labs/graphblast
make bin/graphblast
```

Now try it out:

```shell
while true; do echo $RANDOM && sleep 1s; done | ./bin/graphblast -verbose timeseries
```

and point your browser at [http://localhost:8080](http://localhost:8080).

[go]: http://golang.org/doc/install
