TOP=$(shell pwd)
DESTDIR ?= /usr/local
VERSION = 1.0.$(shell echo `git log --oneline | wc -l`)

RONN := $(shell which ronn >/dev/null 2>&1 && echo "ronn -w --organization=PandaStream" || echo "@echo 'Could not generate manpage because ronn is missing. gem install ronn' || ")
RONNS = $(wildcard man/*.ronn)
ROFFS = $(RONNS:.ronn=)

.PHONY: all man install html gh-pages
all: socketmaster man

socketmaster: *.go
	go fmt
	go build -o socketmaster

%.1: %.1.ronn
	$(RONN) -r $<

man: $(ROFFS)

html:
	$(RONN) -W5 -s toc man/*.ronn

gh-pages: html
	git stash
	git checkout gh-pages
	mv man/socketmaster.1.html index.html
	git add index.html
	git commit -m "build"
	git checkout master
	git stash pop || true

deb:
	rm -rf $(TOP)/fpm
	$(MAKE) install DESTDIR=$(TOP)/fpm/usr/local
	fpm -s dir -t deb -n socketmaster -v $(VERSION) -C $(TOP)/fpm --license MIT --vendor PandaStream --maintainer "<jonas@pandastream.com>" --url http://zimbatm.github.com/socketmaster .

version:
	@echo socketmaster v$(VERSION)

release:
	git tag v$(VERSION)

clean:
	rm -f socketmaster man/*.1
	rm -rf fpm
	rm -f *.deb

install: all
	install -d bin $(DESTDIR)/bin
	install -d man $(DESTDIR)/man/man1
	install -C socketmaster $(DESTDIR)/bin/socketmaster
	cp -R man/*.1 $(DESTDIR)/man/man1
