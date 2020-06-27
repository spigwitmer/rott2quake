package wad2

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"gitlab.com/camtap/lumps/pkg/lumps"
	"io"
)

type WAD2Entry struct {
	Header   *LumpHeader
	Reader   *WAD2Reader
	LumpName string
}

func NewWAD2Entry(name string, header *LumpHeader, reader *WAD2Reader) *WAD2Entry {
	var w WAD2Entry
	w.LumpName = name
	w.Header = header
	w.Reader = reader
	return &w
}

func (w *WAD2Entry) Open() (io.Reader, error) {
	_, err := w.Reader.fhnd.Seek(int64(w.Header.FilePos), io.SeekStart)
	if err != nil {
		return nil, err
	}
	// TODO: what about compression?
	return io.LimitReader(w.Reader.fhnd, int64(w.Header.MemSize)), nil
}

func (w *WAD2Entry) Name() string { return w.LumpName }

func (w *WAD2Entry) Size() int { return int(w.Header.MemSize) }

func (w *WAD2Entry) Print() {
	fmt.Printf("%s\t\ttype %x\t\t%d bytes\n", w.Name(), w.Header.Type, int(w.Header.Size))
}

type WAD2Reader struct {
	fhnd            io.ReadSeeker
	Header          WAD2Header
	Directory       []LumpHeader
	DirectoryByName map[string]int
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
		// NOTE: this breaks if there were 2 lumps of the same name
		// though that's a big if
		w.DirectoryByName[lh.NameString()] = int(i)
	}

	return &w, nil
}

func (w *WAD2Reader) Type() string { return "quake" }

func (w *WAD2Reader) GetEntry(name string) (lumps.ArchiveEntry, error) {
	if idx, found := w.DirectoryByName[name]; found {
		direntry := w.Directory[idx]
		return NewWAD2Entry(name, &direntry, w), nil
	}
	return nil, fmt.Errorf("Entry %s not found", name)
}

func (w *WAD2Reader) List() lumps.ArchiveIterator {
	return &WAD2Iterator{w, 0}
}

type WAD2Iterator struct {
	Reader *WAD2Reader
	idx    int
}

func (w *WAD2Iterator) Next() lumps.ArchiveEntry {
	if int(w.idx) >= len(w.Reader.Directory) {
		return nil
	}
	lheader := w.Reader.Directory[int(w.idx)]
	w.idx++
	return NewWAD2Entry(lheader.NameString(), &lheader, w.Reader)
}
