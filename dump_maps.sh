#!/bin/bash

set -eu
mapfname=${1:-"lumps-data/DARKWAR.RTL"}
outdir=${2:-"map-data"}
rm -rf $outdir
./lumps -wad-out $PWD/out.wad -rtl ${mapfname} -rtl-map-outdir ${outdir} lumps-data/DARKWAR.WAD
pushd ${outdir}
ln -s ../darkwar imgs
popd
