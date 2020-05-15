package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
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

type RottPatchHeader struct {
	OrigSize     uint16
	Width        uint16
	Height       uint16
	LeftOffset   uint16
	TopOffset    uint16
	Transparency uint16
}

func (l *LumpHeader) NameString() string {
	return string(bytes.Trim(l.Name[:], "\x00"))
}

type IWAD struct {
	fhnd          io.ReadSeeker
	Header        WADHeader
	LumpDirectory []*LumpHeader
}

func readLumpHeadersFromIWAD(r io.ReadSeeker, wad *IWAD) error {
	if _, err := r.Seek(int64(wad.Header.DirectoryOffset), io.SeekStart); err != nil {
		return err
	}
	var i uint32
	for i = 0; i < wad.Header.NumLumps; i += 1 {
		var newLump LumpHeader
		if err := binary.Read(r, binary.LittleEndian, &newLump); err != nil {
			return err
		}
		wad.LumpDirectory = append(wad.LumpDirectory, &newLump)
	}
	return nil
}

func NewIWAD(r io.ReadSeeker) (*IWAD, error) {
	var i IWAD
	i.fhnd = r

	if err := binary.Read(r, binary.LittleEndian, &i.Header); err != nil {
		return nil, err
	}

	if !bytes.Equal(i.Header.Magic[:], iwadMagic[:]) {
		return nil, fmt.Errorf("not an IWAD file")
	}

	if err := readLumpHeadersFromIWAD(r, &i); err != nil {
		return nil, err
	}

	return &i, nil
}

func (i *IWAD) PrintLumps() {
	var nl uint32
	for nl = 0; nl < i.Header.NumLumps; nl += 1 {
		lumpHeader := i.LumpDirectory[nl]
		fmt.Printf("%s (%d bytes at 0x%x)\n", lumpHeader.NameString(), lumpHeader.Size, lumpHeader.FilePos)
	}
}

func (i *IWAD) GetLump(name string) (*LumpHeader, error) {
	for _, ld := range i.LumpDirectory {
		if ld.NameString() == name {
			return ld, nil
		}
	}
	return nil, fmt.Errorf("lump %s not found", name)
}

func (i *IWAD) LumpData(l *LumpHeader) (io.Reader, error) {
	if _, err := i.fhnd.Seek(int64(l.FilePos), io.SeekStart); err != nil {
		return nil, err
	}
	// TODO: virtual lump entries with 0 size?
	return io.LimitReader(i.fhnd, int64(l.Size)), nil
}
