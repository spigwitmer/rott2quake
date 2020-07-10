Converts [Rise of the Triad](https://www.gog.com/game/rise_of_the_triad__dark_war) maps and textures to Quake BSP and WAD files.

## Building

```bash
go get -v
make
```

## Invocation

```bash
./lumps -help
```

### Dumping textures to a folder and .wad file

This will dump out (most) textures in ROTT's .wad file to a destination folder as well as a .wad file usable in Quake:

```bash
./lumps -wad-out quake-rott.wad -dump DARKWAR.WAD <dest dir>
```

### Dumping maps to a folder

This will dump the following map data into a new folder: an HTML file containing the map grid, 3 files showing the wall/sprite/info plane values, and a .bsp file of the converted map.

NOTE: you need to generate a Quake .wad file from the invocation above and pass the path to it as `-wad-out`

```bash
./lumps -wad-out quake-rott.wad -rtl DARKWAR.RTL -rtl-map-outdir <dest dir>
```

### Listing textures in a .wad file

ROTT:
```bash
./lumps -list DARKWAR.WAD
```

Quake:
```bash
./lumps -list -quake QUAKE101.WAD
