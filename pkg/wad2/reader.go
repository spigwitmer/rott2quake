package wad2

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type WAD2Reader struct {
	fhnd      io.ReadSeeker
	Header    WAD2Header
	Directory []LumpHeader
}

func NewWAD2Reader(fhnd io.ReadSeeker) (*WAD2Reader, error) {
	var w WAD2Reader

	if err := binary.Read(fhnd, binary.LittleEndian, &w.Header); err != nil {
		return nil, err
	}

	if !bytes.Equal(w.Header.Magic[:], wad2Magic[:]) {
		return nil, fmt.Errorf("Bad wad2 magic: %s", string(w.Header.Magic[:]))
	}

	if w.Header.NumEntries < 0 {
		return nil, fmt.Errorf("Invalid number of entries (%d)", w.Header.NumEntries)
	}

	if _, err := fhnd.Seek(int64(w.Header.DirOffset), io.SeekStart); err != nil {
		return nil, fmt.Errorf("Could not seek to directory: %v\n", err)
	}

	for i := int32(0); i < w.Header.NumEntries; i++ {
		var lh LumpHeader
		if err := binary.Read(fhnd, binary.LittleEndian, &lh); err != nil {
			return nil, err
		}
		w.Directory = append(w.Directory, lh)
	}

	return &w, nil
}

func (w *WAD2Reader) PrintLumps() {
	for _, lumpinfo := range w.Directory {
		fmt.Printf("%s\t\ttype %x\t\t%d bytes\n", lumpinfo.NameString(), lumpinfo.Type, lumpinfo.Size)
	}
}
