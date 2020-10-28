#!/bin/bash

set -eu
mapfname=${1:-"lumps-data/DARKWAR.RTL"}
outdir=${2:-"map-data"}
rm -rf $outdir
# 1.5x the map size or else dusk dude has a hard time getting through
# 1-unit-high throughways such as arch entrances
./lumps -dusk -rtl-map-scale 2 -wad-out $PWD/out.wad -rtl ${mapfname} -rtl-map-outdir ${outdir} lumps-data/DARKWAR.WAD
pushd ${outdir}
# symlink imgs so the HTML map works
ln -s ../darkwar imgs
popd
