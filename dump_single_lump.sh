#!/bin/bash

set -eu
DEBUG="${DEBUG:-""}"
lumpname=${1:-PAUSED}
lumptype=${2:-""}
wadfile=${3:-lumps-data/DARKWAR.WAD}
destdir=${4:-darkwar}
rm -rf ${destdir}

if [ "x$DEBUG" != "x" ]; then
    dlv debug ./cmd/lumps -- -dump-raw \
        -lname "${lumpname}" \
        -ltype "${lumptype}" \
        -dump-data ${wadfile} \
        ${destdir}
else
    ./lumps -dump-raw -lname "${lumpname}" -ltype "${lumptype}" -dump-data ${wadfile} ${destdir}
fi
