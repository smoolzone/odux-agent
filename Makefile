VERSION ?= 0.1.1
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -X github.com/smoolzone/odux-agent/cmd.Version=$(VERSION) -X github.com/smoolzone/odux-agent/cmd.Commit=$(COMMIT) -X github.com/smoolzone/odux-agent/cmd.BuildDate=$(DATE)

# Where `make install` puts the binary (override: make install PREFIX=~/bin)
PREFIX ?= /usr/local

build:
	go build -ldflags "$(LDFLAGS)" -o odux .

install: build
	install -d $(PREFIX)/bin
	install -m 0755 odux $(PREFIX)/bin/odux
	@echo "installed: $(PREFIX)/bin/odux"

uninstall:
	rm -f $(PREFIX)/bin/odux
	@echo "removed: $(PREFIX)/bin/odux"

test:
	go test ./...
