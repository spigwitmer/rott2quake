#!/bin/bash

set -eu
mapfname=${1:-"r2q-data/DARKWAR.RTL"}
outdir=${2:-"map-data"}
rm -rf $outdir
./rott2quake -wad-out $PWD/out.wad -rtl ${mapfname} -rtl-map-outdir ${outdir} r2q-data/DARKWAR.WAD
pushd ${outdir}
ln -s ../darkwar imgs
popd
