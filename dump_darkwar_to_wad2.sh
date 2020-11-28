#!/bin/bash

set -eu
lumpname=${1:-""}
wadfile=${2:-r2q-data/DARKWAR.WAD}
destdir=${3:-darkwar}
rm -rf darkwar
if [ "x${lumpname}" = "x" ]; then
    ./rott2quake -wad-out out.wad -dump ${wadfile} ${destdir} | tee rott2quake.out
else
    ./rott2quake -wad-out out.wad -lname "${lumpname}" -dump ${wadfile} ${destdir} | tee rott2quake.out
fi
#dlv debug ./cmd/rott2quake -- -dump-data ${wadfile} ${destdir}
