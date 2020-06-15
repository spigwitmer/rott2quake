package wad

import (
	"bytes"
)

var (
	iwadMagic = [4]byte{'I', 'W', 'A', 'D'}
	pwadMagic = [4]byte{'P', 'W', 'A', 'D'}
)

type WADHeader struct {
	Magic           [4]byte
	NumLumps        uint32 // NOTE: vanilla doom reads these as signed ints
	DirectoryOffset uint32
}

type LumpHeader struct {
	FilePos uint32
	Size    uint32
	Name    [8]byte
}

func (l *LumpHeader) NameString() string {
	return string(bytes.Trim(l.Name[:], "\x00"))
}

// sprites (guns, actors, etc.)
type PatchHeader struct {
	OrigSize   int16
	Width      int16
	Height     int16
	LeftOffset int16
	TopOffset  int16
}
