SOURCES := $(wildcard cmd/*/*.go pkg/*/*.go)
GOPRIVATE := '*'
export GOPRIVATE

ADDITIONAL_WADS ?= $(shell pwd)/r2q-data/quake101.wad
FGD_FILE ?= $(shell pwd)/r2q-data/dusk4.fgd

additionalWadsParam := $(foreach wadPath,$(ADDITIONAL_WADS),-add-wad $(wadPath))
fgdParam := -fgd $(FGD_FILE)


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
	./dump_maps_dusk.sh r2q-data/DARKWAR.RTL map-data $(additionalWadsParam) $(fgdParam)

.PHONY: dump-maps
dump-maps: rott2quake
	./dump_maps.sh r2q-data/DARKWAR.RTL map-data $(additionalWadsParam)

.PHONY: gofmt
gofmt: $(SOURCES)
	gofmt -w $(SOURCES)
