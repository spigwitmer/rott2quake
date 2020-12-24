SOURCES := $(wildcard cmd/*/*.go pkg/*/*.go)
GOPRIVATE := '*'
export GOPRIVATE

TARGETS = rott2quake dump-palette

all: rott2quake dump-palette

rott2quake: $(SOURCES)
	go build -o $@ ./cmd/$@
dump-palette: $(SOURCES)
	go build -o $@ ./cmd/$@

.PHONY: show-sources
show-sources:
	ls -1 $(SOURCES)

.PHONY: dump-darkwar-to-wad2
dump-darkwar-to-wad2: rott2quake
	./dump_darkwar_to_wad2.sh

.PHONY: dump-maps-dusk
dump-maps-dusk: rott2quake
	./dump_maps_dusk.sh

.PHONY: dump-maps
dump-maps: rott2quake
	./dump_maps.sh

.PHONY: gofmt
gofmt: $(SOURCES)
	gofmt -w $(SOURCES)
