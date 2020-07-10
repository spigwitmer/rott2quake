Converts [Rise of the Triad](https://www.gog.com/game/rise_of_the_triad__dark_war) maps and textures to Quake BSP and WAD files.

## Building

### Ingredients

* Go 1.14+
* Make

### Building the CLI tool

```bash
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

This will dump the following map data into a new folder: an HTML file containing the map grid, 3 files showing the wall/sprite/info plane values, and a .map file of the converted level that can be generated with TrenchBroom or ericw-tools.

NOTE: you need to generate a Quake .wad file from the invocation above and pass the path to it as `-wad-out`

```bash
./lumps -wad-out quake-rott.wad -rtl DARKWAR.RTL -rtl-map-outdir <dest dir>
```

If you're generating maps to play in Dusk, use at least 1.5x scale:
```bash
./lumps -wad-out quake-rott.wad -rtl DARKWAR.RTL -rtl-map-scale 1.5 -rtl-map-outdir <dest dir>
```

### Listing textures in a .wad file

ROTT:
```bash
./lumps -list DARKWAR.WAD
```

Quake:
```bash
./lumps -list -quake QUAKE101.WAD


## Supported items

- [x] World structure
- [x] Masked walls
- [x] Platforms
- [ ] Trampolines (TODO)
- [ ] Weapon placement (TODO)
- [ ] Enemy placement (TODO)
- [ ] Switchplates (TODO)
- [ ] Moving Walls (TODO)
- [ ] Obstacles e.g. flamethrowers, crushers (maybe)


## Known issues

- Tops of platform are (intentionally) not rendered
- Scale cannot go past 3x without bad things happening. Quake won't
  render the floor or ceiling.
