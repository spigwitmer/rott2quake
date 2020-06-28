package lumps

import (
	"io"
)

type ArchiveEntry interface {
	// returns a Reader that represents the raw (and, if applicable,
	// decompressed) contents of the entry
	Open() (io.Reader, error)
	// returns the name/path
	Name() string
	// returns the uncompressed file size
	Size() int
	// prints a description of the entry to os.Stdout
	Print()
	// proposes a file type and destination folder for dumping
	GuessFileTypeAndSubdir() (string, string)
}

type ArchiveIterator interface {
	// returns the next entry in the archive's directory
	Next() ArchiveEntry
}

type ArchiveReader interface {
	// return an iterator to all entries in the archive
	List() ArchiveIterator
	// returns game type
	Type() string
	// returns an entry of the given name or path
	GetEntry(name string) (ArchiveEntry, error)
}
