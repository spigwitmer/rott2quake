package lumps

import (
	"io"
)

type ArchiveEntry interface {
	Open() (io.Reader, error)
	Name() string
	Size() int
	Print()
}

type ArchiveIterator interface {
	Next() ArchiveEntry
}

type ArchiveReader interface {
	List() ArchiveIterator
	Type() string
	GetEntry(name string) (ArchiveEntry, error)
}
