export GOPATH=$(PWD)

all: socketmaster

ex: examples
examples: childserver

socketmaster: src/bin/socketmaster/*.go src/socketmaster/*.go
	go fmt socketmaster
	go fmt bin/socketmaster
	go build bin/socketmaster

childserver: examples/childserver.go src/socketmaster/*.go
	go fmt socketmaster
	go fmt examples/childserver.go
	go build examples/childserver.go
