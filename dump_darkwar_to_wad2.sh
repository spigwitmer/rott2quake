#!/bin/bash

set -eu
lumpname=${1:-""}
wadfile=${2:-lumps-data/DARKWAR.WAD}
destdir=${3:-darkwar}
rm -rf darkwar
if [ "x${lumpname}" = "x" ]; then
    ./lumps -wad-out out.wad -dump ${wadfile} ${destdir}
else
    ./lumps -wad-out out.wad -lname "${lumpname}" -dump ${wadfile} ${destdir}
fi
#dlv debug ./cmd/lumps -- -dump-data ${wadfile} ${destdir}
