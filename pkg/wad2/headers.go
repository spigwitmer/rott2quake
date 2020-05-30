package wad2

import (
	"bytes"
)

var (
	wad2Magic = [4]byte{'W', 'A', 'D', '2'}
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
	return string(bytes.Trim(l.Name[:], "\x00"))
}

type Lump struct {
	Name string
	Data []byte
}

type WAD2Header struct {
	Magic      [4]byte
	NumEntries int32
	DirOffset  int32
}
