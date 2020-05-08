package main

import (
    "bytes"
    "encoding/binary"
    "fmt"
    "io"
)

var (
    rtlMagic = [4]byte{'R', 'T', 'L', '\x00'}
)

type RTLHeader struct {
    Signature [4]byte
    Version uint32
    Used uint32
    CRC [4]byte
    RLEWTag [4]byte
    MapSpecials uint32
    WallPlaneOffset uint32
    SpritePlaneOffset uint32
    InfoPlaneOffset uint32
    WallPlaneLength uint32
    SpritePlaneLength uint32
    InfoPlaneLength uint32
    Name [24]byte
}

type RTL struct {
    fhnd io.ReadSeeker
    Header RTLHeader
}

func NewRTL(rfile io.ReadSeeker) (*RTL, error) {
    var r RTL
    r.fhnd = rfile

    if err := binary.Read(rfile, binary.LittleEndian, &r.Header); err != nil {
        return nil, err
    }

    if !bytes.Equal(r.Header.Signature[:], rtlMagic[:]) {
        return nil, fmt.Errorf("not an RTL file")
    }

    return &r, nil
}

func (r *RTL) MapName() (string) {
    return string(bytes.Trim(r.Header.Name[:], "\x00"))
}
