SOURCES := $(wildcard cmd/*/*.go pkg/*/*.go)
GOPRIVATE := '*'
export GOPRIVATE

TARGETS = lumps dump-palette

all: lumps dump-palette

lumps: $(SOURCES)
	go build -o $@ ./cmd/$@
dump-palette: $(SOURCES)
	go build -o $@ ./cmd/$@

.PHONY: show-sources
show-sources:
	ls -1 $(SOURCES)

.PHONY: dump-darkwar-to-wad2
dump-darkwar-to-wad2: lumps
	./dump_darkwar_to_wad2.sh

.PHONY: dump-maps-dusk
dump-maps-dusk: lumps
	./dump_maps_dusk.sh

.PHONY: gofmt
gofmt: $(SOURCES)
	gofmt -w $(SOURCES)
