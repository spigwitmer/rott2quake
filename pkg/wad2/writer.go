package wad2

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"
)

type WADWriter struct {
	Directory []Lump
}

func NewWADWriter() (*WADWriter, error) {
	var writer WADWriter

	return &writer, nil
}

func (w *WADWriter) AddLump(name string, data []byte, ltype int8) error {
	newlump := Lump{Name: name, Data: data, Type: ltype}
	w.Directory = append(w.Directory, newlump)

	return nil
}

func (w *WADWriter) Write(dest io.WriteSeeker) (int64, error) {
	var header WAD2Header

	copy(header.Magic[:], wad2Magic[:])
	header.NumEntries = int32(len(w.Directory))
	header.DirOffset = int32(binary.Size(header))

	if err := binary.Write(dest, binary.LittleEndian, &header); err != nil {
		return 0, err
	}

	entrysize := binary.Size(LumpHeader{})

	// total written
	total := int64(header.DirOffset)
	// calculated position of lump data in file
	var offset int32 = header.DirOffset + int32(entrysize*len(w.Directory))

	for _, lump := range w.Directory {
		if len(lump.Name) > 16 {
			log.Fatalf("lump name %s longer than 16 chars", lump.Name)
		}
		var namebytes [16]byte
		namelen := len(lump.Name)

		// pad name with null bytes
		copy(namebytes[:], []byte(lump.Name))
		if len(lump.Name) < 16 {
			copy(namebytes[namelen:], bytes.Repeat([]byte{'\x00'}, 16-len(lump.Name)))
		}

		direntry := LumpHeader{
			FilePos: offset,
			Size:    int32(len(lump.Data)),
			// TODO: compression?
			MemSize:     int32(len(lump.Data)),
			Type:        lump.Type,
			Compression: 0,
			Dummy:       0x0000,
			Name:        namebytes,
		}
		if err := binary.Write(dest, binary.LittleEndian, &direntry); err != nil {
			return int64(total), err
		}
		offset += int32(len(lump.Data))
		total += int64(entrysize)
	}
	for _, lump := range w.Directory {
		if err := binary.Write(dest, binary.LittleEndian, lump.Data); err != nil {
			return total, err
		}
		total += int64(len(lump.Data))
	}

	return total, nil
}
