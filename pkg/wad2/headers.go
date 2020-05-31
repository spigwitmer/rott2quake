package wad2

import (
	"bytes"
)

var (
	wad2Magic = [4]byte{'W', 'A', 'D', '2'}
)

var (
	LT_RAW     int8 = 0x40
	LT_PICTURE int8 = 0x42
	LT_MIPTEX  int8 = 0x44
)

type LumpHeader struct {
	FilePos     int32
	Size        int32
	MemSize     int32
	Type        int8
	Compression int8
	Dummy       int16
	Name        [16]byte
}

func (l *LumpHeader) NameString() string {
	// Lumps names are supposed to be null-padded, though
	// garbage data was found in the padding for the names
	// in QUAKE101.wad
	nullpos := bytes.Index(l.Name[:], []byte{'\x00'})
	return string(l.Name[0:nullpos])
}

type Lump struct {
	Name string
	Type int8
	Data []byte
}

type WAD2Header struct {
	Magic      [4]byte
	NumEntries int32
	DirOffset  int32
}
