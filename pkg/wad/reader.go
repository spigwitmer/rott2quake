package wad

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image/color"
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

// sprites (guns, actors, etc.)
type PatchHeader struct {
	OrigSize   int16
	Width      int16
	Height     int16
	LeftOffset int16
	TopOffset  int16
}

// wall and sky textures
type RottPatchHeader struct {
	OrigSize     int16
	Width        int16
	Height       int16
	LeftOffset   int16
	TopOffset    int16
	Transparency int16
}

// floors and ceilings
type RottLpicHeader struct {
	Width  uint16
	Height uint16
	OrgX   uint16
	OrgY   uint16
}

type RottPicHeader struct {
	Width  uint8
	Height uint8
}

type Palette struct {
	R, G, B uint8
}

func (l *LumpHeader) NameString() string {
	return string(bytes.Trim(l.Name[:], "\x00"))
}

type WADReader struct {
	fhnd            io.ReadSeeker
	Header          WADHeader
	BasePaletteData color.Palette
	LumpDirectory   []*LumpHeader
}

func getBasePaletteData(iwad *WADReader) error {
	paletteData := make([]Palette, 256)

	paletteLump, err := iwad.GetLump("PAL")
	if err != nil {
		return err
	}
	paletteLumpData, err := iwad.LumpData(paletteLump)
	if err != nil {
		return err
	}
	if err = binary.Read(paletteLumpData, binary.LittleEndian, &paletteData); err != nil {
		return err
	}

	for _, wp := range paletteData {
		iwad.BasePaletteData = append(iwad.BasePaletteData, color.RGBA{wp.R, wp.G, wp.B, 255})
	}

	return nil
}

func readLumpHeadersFromWAD(r io.ReadSeeker, wad *WADReader) error {
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

func NewIWAD(r io.ReadSeeker) (*WADReader, error) {
	var i WADReader
	i.fhnd = r

	if err := binary.Read(r, binary.LittleEndian, &i.Header); err != nil {
		return nil, err
	}

	if !bytes.Equal(i.Header.Magic[:], iwadMagic[:]) {
		return nil, fmt.Errorf("not an IWAD file")
	}

	if err := readLumpHeadersFromWAD(r, &i); err != nil {
		return nil, err
	}

	if err := getBasePaletteData(&i); err != nil {
		return nil, err
	}

	return &i, nil
}

func (i *WADReader) PrintLumps() {
	var nl uint32
	for nl = 0; nl < i.Header.NumLumps; nl += 1 {
		lumpHeader := i.LumpDirectory[nl]
		fmt.Printf("%d: %s (%d bytes at 0x%x)\n", nl, lumpHeader.NameString(), lumpHeader.Size, lumpHeader.FilePos)
	}
}

func (i *WADReader) GetLump(name string) (*LumpHeader, error) {
	for _, ld := range i.LumpDirectory {
		if ld.NameString() == name {
			return ld, nil
		}
	}
	return nil, fmt.Errorf("lump %s not found", name)
}

func (i *WADReader) LumpData(l *LumpHeader) (io.Reader, error) {
	if _, err := i.fhnd.Seek(int64(l.FilePos), io.SeekStart); err != nil {
		return nil, err
	}
	// TODO: virtual lump entries with 0 size?
	return io.LimitReader(i.fhnd, int64(l.Size)), nil
}
