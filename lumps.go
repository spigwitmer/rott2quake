package main

import (
    "flag"
    "fmt"
    "log"
    "os"
)

func init() {
    flag.Usage = func() {
        fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s <.WAD file>\n", os.Args[0])
        flag.PrintDefaults()
    }
}

func main() {
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

    wadFile.PrintLumps()
}
