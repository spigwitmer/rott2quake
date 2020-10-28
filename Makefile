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

.PHONY: dump-maps-dusk
dump-maps-dusk: lumps
	./dump_maps_dusk.sh
