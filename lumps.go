package main

import (
    "flag"
    "fmt"
    "io"
    "log"
    "os"
)

func init() {
    flag.Usage = func() {
        fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s <.WAD file> [dest dir]\n", os.Args[0])
        flag.PrintDefaults()
    }
}

func dumpLumpDataToFile(wadFile *IWAD, lumpInfo *LumpHeader, destFname string) {
    lumpReader, err := wadFile.LumpData(lumpInfo)
    if err != nil {
        log.Fatalf("Could not get lump data reader for %s: %v\n", destFname, err)
    }
    destfhnd, err := os.Create(destFname)
    if err != nil {
        log.Fatalf("Could not write to %s: %v\n", destFname, err)
    }
    defer destfhnd.Close()
    numWritten, err := io.Copy(destfhnd, lumpReader)
    if err != nil {
        log.Fatalf("Could not copy to %s: %v\n", destFname, err)
    }
    if numWritten != int64(lumpInfo.Size) {
        log.Fatalf("numWritten != lumpInfo.Size ??!\n")
    }
}

func main() {
    var dumpLumpData, printLumps bool

    flag.BoolVar(&dumpLumpData, "dump-data", false, "Dump Lump Data")
    flag.BoolVar(&printLumps, "print-lumps", false, "Print Lump Directory")
    flag.Parse()
    if flag.NArg() < 1 {
        flag.Usage()
        os.Exit(2)
    }

    fhnd, err := os.Open(flag.Arg(0))
    if err != nil {
        log.Fatalf("Could not open file: %v\n", err)
    }
    wadFile, err := NewIWAD(fhnd)
    if err != nil {
        log.Fatalf("Could not open IWAD: %v\n", err)
    }

    fmt.Printf("WAD file has %d lumps\n", wadFile.Header.NumLumps)

    if printLumps {
        wadFile.PrintLumps()
    }
    if dumpLumpData {
        if flag.NArg() < 2 {
            flag.Usage()
            os.Exit(2)
        }
        destDir := flag.Arg(1)
        for i := uint32(0); i < wadFile.Header.NumLumps; i += 1 {
            lumpInfo := wadFile.LumpDirectory[i]
            destFname := fmt.Sprintf("%s/%s.dat", destDir, lumpInfo.NameString())
            fmt.Printf("dumping %s\n", destFname)
            dumpLumpDataToFile(wadFile, lumpInfo, destFname)
        }
    }
}
