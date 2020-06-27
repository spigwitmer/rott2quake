package wad

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"gitlab.com/camtap/lumps/pkg/lumps"
	"image/color"
	"io"
)

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

type WADEntry struct {
	LumpName   string
	Reader     *WADReader
	LumpHeader *LumpHeader
}

func NewWADEntry(name string, header *LumpHeader, reader *WADReader) *WADEntry {
	var w WADEntry
	w.LumpName = name
	w.Reader = reader
	w.LumpHeader = header
	return &w
}

func (w *WADEntry) Name() string {
	return w.LumpName
}

func (w *WADEntry) Size() int {
	return int(w.LumpHeader.Size)
}

func (w *WADEntry) Open() (io.Reader, error) {
	header, err := w.Reader.GetLump(w.LumpName)
	if err != nil {
		return nil, err
	}
	reader, err := w.Reader.LumpData(header)
	if err != nil {
		return nil, err
	}
	return reader, err
}

func (w *WADEntry) Print() {
	fmt.Printf("%s (%d bytes)\n", w.Name(), w.LumpHeader.Size)
}

type WADIterator struct {
	Reader *WADReader
	idx    uint32
}

func (w *WADIterator) Next() lumps.ArchiveEntry {
	if int(w.idx) >= len(w.Reader.LumpDirectory) {
		return nil
	}
	lheader := w.Reader.LumpDirectory[int(w.idx)]
	w.idx++
	return NewWADEntry(lheader.NameString(), lheader, w.Reader)
}

func (i *WADReader) List() lumps.ArchiveIterator {
	var iter WADIterator
	iter.Reader = i
	iter.idx = 0
	return &iter
}

func (i *WADReader) GetLump(name string) (*LumpHeader, error) {
	for _, ld := range i.LumpDirectory {
		if ld.NameString() == name {
			return ld, nil
		}
	}
	return nil, fmt.Errorf("lump %s not found", name)
}

func (i *WADReader) GetEntry(name string) (lumps.ArchiveEntry, error) {
	lheader, err := i.GetLump(name)
	if err != nil {
		return nil, err
	}

	return NewWADEntry(name, lheader, i), nil
}

func (i *WADReader) Type() string { return "rott" }

func (i *WADReader) LumpData(l *LumpHeader) (io.Reader, error) {
	if _, err := i.fhnd.Seek(int64(l.FilePos), io.SeekStart); err != nil {
		return nil, err
	}
	// TODO: virtual lump entries with 0 size?
	return io.LimitReader(i.fhnd, int64(l.Size)), nil
}
