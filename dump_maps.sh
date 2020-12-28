#!/bin/bash

set -xeu
mapfname=${1:-"r2q-data/DARKWAR.RTL"}
outdir=${2:-"map-data"}
shift; shift
rm -rf $outdir
./rott2quake -wad-out $PWD/out.wad -rtl ${mapfname} -rtl-map-scale 1.5 -rtl-map-outdir ${outdir} "$@" r2q-data/DARKWAR.WAD
pushd ${outdir}
ln -s ../darkwar imgs
popd
