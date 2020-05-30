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

func (w *WADWriter) AddLump(name string, data []byte) error {
	newlump := Lump{Name: name, Data: data}
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

	var total int32 = header.DirOffset
	var entrysize int32 = int32(binary.Size(LumpHeader{}))
	for _, lump := range w.Directory {
		if len(lump.Name) > 16 {
			log.Fatalf("lump name %s longer than 16 chars", lump.Name)
		}
		var namebytes [16]byte
		namelen := len(lump.Name)

		// pad name with null bytes
		copy(namebytes[:], []byte(lump.Name))
		if len(lump.Name) < 16 {
			copy(namebytes[namelen:], bytes.Repeat([]byte{'\x00'}, len(lump.Name)-16))
		}

		direntry := LumpHeader{
			FilePos: int32(len(w.Directory))*entrysize + total,
			Size:    int32(len(lump.Data)),
			// TODO: compression?
			MemSize:     int32(len(lump.Data)),
			Type:        0x45,
			Compression: 0,
			Dummy:       0x0000,
			Name:        namebytes,
		}
		if err := binary.Write(dest, binary.LittleEndian, &direntry); err != nil {
			return int64(total), err
		}
		total += int32(len(lump.Data))
	}
	for _, lump := range w.Directory {
		if err := binary.Write(dest, binary.LittleEndian, lump.Data); err != nil {
			return int64(total), err
		}
	}

	return int64(total), nil
}
