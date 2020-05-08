package main

import (
    "bytes"
    "encoding/binary"
    "fmt"
)

var (
    rtlMagic = [4]byte{'R', 'T', 'L', '\x00'}
)

type RTLVersion struct {
    Signature [4]byte
    Version uint32
}

type RTLHeader struct {
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
    RTLVersion
    RTLHeader
    fhnd io.ReadSeeker
}

func NewRTL(r io.ReadSeeker) (*RTL, error) {
    var r RTL
    r.fhnd = r

    if err := binary.Read(r, binary.LittleEndian, &i); err != nil {
        return nil, err
    }

    if !bytes.Equal(i.Signature[:], rtlMagic[:]) {
        return nil, fmt.Errorf("not an RTL file")
    }
}

func (r *RTL) MapName() (string) {
    return string(bytes.Trin(l.Name[:], "\x00"))
}
