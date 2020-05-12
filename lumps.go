package main

import (
    "errors"
    "flag"
    "fmt"
    "image"
    "image/color"
    "image/png"
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

func dumpRawLumpDataToFile(destFhnd io.WriteSeeker, lumpReader io.Reader) (int64, error) {
    return io.Copy(destFhnd, lumpReader)
}

// convert to PNG before writing
func dumpWallDataToFile(destFhnd io.WriteSeeker, lumpReader io.Reader) (int64, error) {
    // assumes 64x64
    // TODO: color
    const width, height = 64, 64

    var rawImgData [width*height]byte
    numRead, err := lumpReader.Read(rawImgData[:])
    if err != nil {
        return 0, err
    }
    if numRead != width*height {
        return 0, errors.New("numRead != width*height???")
    }

    img := image.NewGray(image.Rect(0, 0, width, height))
    for i := 0; i < height; i++ {
        for j := 0; j < width; j++ {
            img.SetGray(i, j, color.Gray{uint8(rawImgData[(i*width)+j])})
        }
    }

    if err := png.Encode(destFhnd, img); err != nil {
        return 0, err
    }

    return destFhnd.Seek(0, io.SeekCurrent)
}

func dumpLumpDataToFile(wadFile *IWAD, lumpInfo *LumpHeader, destFname string,
                        wallData bool) {
    lumpReader, err := wadFile.LumpData(lumpInfo)
    if err != nil {
        log.Fatalf("Could not get lump data reader for %s: %v\n", destFname, err)
    }
    if wallData {
        destFname = fmt.Sprintf("%s.png", destFname)
    } else {
        destFname = fmt.Sprintf("%s.dat", destFname)
    }
    fmt.Printf("dumping %s\n", destFname)
    destfhnd, err := os.Create(destFname)
    if err != nil {
        log.Fatalf("Could not write to %s: %v\n", destFname, err)
    }
    defer destfhnd.Close()

    if wallData {
        _, err = dumpWallDataToFile(destfhnd, lumpReader)
    } else {
        _, err = dumpRawLumpDataToFile(destfhnd, lumpReader)
    }
    if err != nil {
        log.Fatalf("Could not copy to %s: %v\n", destFname, err)
    }
}

func main() {
    var dumpLumpData, printLumps bool
    var rtlFile, rtlMapOutdir string
    var rtl *RTL
    var printRTLInfo bool

    flag.StringVar(&rtlFile, "rtl", "", "RTL file")
    flag.BoolVar(&printRTLInfo, "print-rtl-info", false, "Print RTL metadata")
    flag.StringVar(&rtlMapOutdir, "rtl-map-outdir", "", "Write RTL ASCII map out to this folder")
    flag.BoolVar(&dumpLumpData, "dump-data", false, "Dump Lump Data out to dest dir")
    flag.BoolVar(&printLumps, "print-lumps", false, "Print Lump Directory")
    flag.Parse()
    if flag.NArg() < 1 {
        flag.Usage()
        os.Exit(2)
    }

    if rtlFile != "" {
        rtlFhnd, err := os.Open(rtlFile)
        if err != nil {
            log.Fatalf("Could not open RTL file %s: %v\n", rtlFile, err)
        }
        defer rtlFhnd.Close()

        rtl, err = NewRTL(rtlFhnd)
        if err != nil {
            log.Fatalf("Could not parse RTL file: %v\n", err)
        }
    }

    if rtl != nil && printRTLInfo {
        rtl.PrintMetadata()
    }

    if rtlMapOutdir != "" {
        if err := os.MkdirAll(rtlMapOutdir, 0755); err != nil {
            log.Fatalf("Could not create outdir: %v\n", err)
        }
        for idx, md := range rtl.MapData {
            if md.Header.Used == 0 {
                continue
            }
            rtlMapFile := fmt.Sprintf("%s/map%03d.txt", rtlMapOutdir, idx+1)
            wallFhnd, err := os.Create(rtlMapFile)
            if err != nil {
                log.Fatalf("Could not open %s for writing: %v\n", rtlMapFile, err)
            }
            defer wallFhnd.Close()
            err = md.DumpWallToFile(wallFhnd)
            if err != nil {
                log.Fatalf("Could not write map to %s: %v\n", rtlMapFile, err)
            }
        }
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

        if err := os.MkdirAll(destDir, 0755); err != nil {
            log.Fatalf("Could not create dest dir: %v\n", err)
        }

        wallData := false
        for i := uint32(0); i < wadFile.Header.NumLumps; i += 1 {
            lumpInfo := wadFile.LumpDirectory[i]
            switch lumpInfo.NameString() {
            case "WALLSTRT":
                wallData = true
            case "WALLSTOP":
                wallData = false
            }
            if lumpInfo.Size > 0 {
                destFname := fmt.Sprintf("%s/%s", destDir, lumpInfo.NameString())
                dumpLumpDataToFile(wadFile, lumpInfo, destFname, wallData)
            }
        }
    }
}
