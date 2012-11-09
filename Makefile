DESTDIR=/usr/local

all: socketmaster

socketmaster: *.go
	go fmt
	go build

install: socketmaster
	install -C socketmaster $(DESTDIR)/bin/socketmaster
