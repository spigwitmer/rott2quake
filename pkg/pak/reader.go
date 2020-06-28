package pak

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"gitlab.com/camtap/lumps/pkg/lumps"
	"io"
)

var (
	pakMagic = []byte{'P', 'A', 'C', 'K'}
)

type PAKEntry struct {
	reader *PAKReader
	Header PAKEntryHeader
}

type PAKEntryHeader struct {
	EntryName [56]byte
	Offset    uint32
	FSize     uint32
}

func (p *PAKEntry) Name() string {
	nullpos := bytes.Index(p.Header.EntryName[:], []byte{'\x00'})
	return string(p.Header.EntryName[0:nullpos])
}

func (p *PAKEntry) Size() int { return int(p.Header.FSize) }

func (p *PAKEntry) Print() {
	fmt.Printf("%s (%d bytes)\n", p.Name(), p.Size())
}

func (p *PAKEntry) Open() (io.Reader, error) {
	_, err := p.reader.fhnd.Seek(int64(p.Header.Offset), io.SeekStart)
	if err != nil {
		return nil, err
	}
	return io.LimitReader(p.reader.fhnd, int64(p.Size())), nil
}

func (p *PAKEntry) GuessFileTypeAndSubdir() (string, string) {
	// path is already part of the name
	return "raw", ""
}

type PAKHeader struct {
	TableOffset uint32
	TableSize   uint32
}

type PAKReader struct {
	fhnd      io.ReadSeeker
	Header    PAKHeader
	Directory []PAKEntry
	EntryMap  map[string]*PAKEntry
}

func NewPAKReader(fhnd io.ReadSeeker) (*PAKReader, error) {
	var pr PAKReader
	var magic [4]byte

	// check that file starts with "PACK"
	if _, err := fhnd.Read(magic[:]); err != nil {
		return nil, err
	}
	if !bytes.Equal(magic[:], pakMagic[:]) {
		return nil, errors.New("bad magic")
	}
	if err := binary.Read(fhnd, binary.LittleEndian, &pr.Header); err != nil {
		return nil, err
	}

	// build map of package contents
	if _, err := fhnd.Seek(int64(pr.Header.TableOffset), io.SeekStart); err != nil {
		return nil, err
	}
	entryLen := 64
	numEntries := int(pr.Header.TableSize) / entryLen
	pr.EntryMap = make(map[string]*PAKEntry)
	fmt.Printf("entryLen: %d, numEntries: %d\n", entryLen, numEntries)
	for i := 0; i < numEntries; i++ {
		var pe PAKEntry
		if err := binary.Read(fhnd, binary.LittleEndian, &pe.Header); err != nil {
			return nil, err
		}
		pe.reader = &pr
		pr.Directory = append(pr.Directory, pe)
		// Name will include full path ("foo/bar/baz.bsp")
		pr.EntryMap[pe.Name()] = &pe
	}

	pr.fhnd = fhnd
	return &pr, nil
}

type PAKIterator struct {
	reader *PAKReader
	idx    int
}

func (p *PAKIterator) Next() lumps.ArchiveEntry {
	if p.idx >= len(p.reader.Directory) {
		return nil
	}
	entry := p.reader.Directory[p.idx]
	p.idx++
	return &entry
}

func (p *PAKReader) List() lumps.ArchiveIterator {
	return &PAKIterator{p, 0}
}

func (p *PAKReader) Type() string { return "quake" }

func (p *PAKReader) GetEntry(name string) (lumps.ArchiveEntry, error) {
	if entry, ok := p.EntryMap[name]; ok {
		return entry, nil
	}
	return nil, fmt.Errorf("Entry \"%s\" not found", name)
}
