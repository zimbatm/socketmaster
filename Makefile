DESTDIR ?= /usr/local

RONN := $(shell which ronn >/dev/null 2>&1 && echo "ronn -w --organization=PandaStream" || echo "@echo 'Could not generate manpage because ronn is missing. gem install ronn' || ")
RONNS = $(wildcard man/*.ronn)
ROFFS = $(RONNS:.ronn=)

.PHONY: all man install
all: socketmaster man

socketmaster: *.go
	go fmt
	go build -o socketmaster

%.1: %.1.ronn
	$(RONN) -r $<

man: $(ROFFS)

clean:
	rm -f socketmaster
	rm man/*.1

install: all
	install -d bin $(DESTDIR)/bin
	install -d man $(DESTDIR)/man/man1
	install -C socketmaster $(DESTDIR)/bin/socketmaster
	cp -R man/*.1 $(DESTDIR)/share/man/man1
