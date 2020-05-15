package main

import (
	"bytes"
	"encoding/binary"
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

type Palette struct {
	R, G, B uint8
}

var (
	PaletteData [256]Palette
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

// convert wall data to PNG before writing
func dumpWallDataToFile(destFhnd io.WriteSeeker, lumpReader io.Reader) (int64, error) {
	// assumes 64x64 (standard mandated by ROTT)
	const width, height = 64, 64

	var rawImgData [width * height]byte
	numRead, err := lumpReader.Read(rawImgData[:])
	if err != nil {
		return 0, err
	}
	if numRead != width*height {
		return 0, errors.New("numRead != width*height???")
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {
			palette := PaletteData[rawImgData[(i*width)+j]]
			img.SetRGBA(i, j, color.RGBA{palette.R, palette.G, palette.B, 255})
		}
	}

	if err := png.Encode(destFhnd, img); err != nil {
		return 0, err
	}

	return destFhnd.Seek(0, io.SeekCurrent)
}

// convert patch data to PNG before writing
func dumpPatchDataToFile(destFhnd io.WriteSeeker, lumpInfo *LumpHeader, lumpReader io.Reader) (int64, error) {
	// https://doomwiki.org/wiki/Picture_format

	// read entire lump to perform random access
	patchBytes := make([]byte, lumpInfo.Size)
	_, err := lumpReader.Read(patchBytes)
	if err != nil {
		return 0, err
	}
	lumpBuffer := bytes.NewReader(patchBytes)

	var patchHeader RottPatchHeader
	if err := binary.Read(lumpBuffer, binary.LittleEndian, &patchHeader); err != nil {
		log.Fatal(err)
	}

	columnOffsets := make([]uint16, patchHeader.Width)
	if err := binary.Read(lumpBuffer, binary.LittleEndian, &columnOffsets); err != nil {
		return 0, err
	}

	img := image.NewRGBA(image.Rect(0, 0, int(patchHeader.Width), int(patchHeader.Height)))
	for idx, cOffset := range columnOffsets {
		_, err := lumpBuffer.Seek(int64(cOffset), io.SeekStart)
		if err != nil {
			return 0, err
		}
		rowstart := byte(0)
		for rowstart != 255 {
			rowstart, err := lumpBuffer.ReadByte()
			if err != nil {
				return 0, err
			}
			if rowstart == 255 {
				break
			}

			var pixelCount uint8
			err = binary.Read(lumpBuffer, binary.LittleEndian, &pixelCount)
			if err != nil {
				return 0, err
			}

			// read dummy byte
			_, err = lumpBuffer.ReadByte()
			if err != nil {
				return 0, err
			}

			for i := uint8(0); i < pixelCount-1; {
				paletteCode, err := lumpBuffer.ReadByte()
				if err != nil {
					return 0, err
				}

				pixel := PaletteData[paletteCode]
				img.SetRGBA(idx, int(i+rowstart), color.RGBA{pixel.R, pixel.G, pixel.B, 255})
			}

			// read another dummy byte
			_, err = lumpBuffer.ReadByte()
			if err != nil {
				return 0, err
			}
		}
	}

	if err = png.Encode(destFhnd, img); err != nil {
		return 0, err
	}

	return destFhnd.Seek(0, io.SeekCurrent)
}

func dumpLumpDataToFile(wadFile *IWAD, lumpInfo *LumpHeader, destFname string,
	dataType string) {
	lumpReader, err := wadFile.LumpData(lumpInfo)
	if err != nil {
		log.Fatalf("Could not get lump data reader for %s: %v\n", destFname, err)
	}
	fmt.Printf("dumping %s\n", destFname)
	destfhnd, err := os.Create(destFname)
	if err != nil {
		log.Fatalf("Could not write to %s: %v\n", destFname, err)
	}
	defer destfhnd.Close()

	switch dataType {
	case "wall":
		_, err = dumpWallDataToFile(destfhnd, lumpReader)
	case "midi":
		_, err = dumpRawLumpDataToFile(destfhnd, lumpReader)
	case "patch":
		_, err = dumpPatchDataToFile(destfhnd, lumpInfo, lumpReader)
	default:
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

		paletteLump, err := wadFile.GetLump("PAL")
		if err != nil {
			log.Fatal(err)
		}
		paletteLumpData, err := wadFile.LumpData(paletteLump)
		if err != nil {
			log.Fatal(err)
		}
		if err = binary.Read(paletteLumpData, binary.LittleEndian, &PaletteData); err != nil {
			log.Fatal(err)
		}

		if err := os.MkdirAll(destDir, 0755); err != nil {
			log.Fatalf("Could not create dest dir: %v\n", err)
		}

		dataType := "raw"
		for i := uint32(0); i < wadFile.Header.NumLumps; i += 1 {
			lumpInfo := wadFile.LumpDirectory[i]
			switch lumpInfo.NameString() {
			case "WALLSTRT":
				dataType = "wall"
			case "WALLSTOP":
				dataType = "raw"
			case "SONGSTRT":
				dataType = "midi"
			case "ANIMSTRT":
				dataType = "wall"
			case "EXITSTRT":
				dataType = "raw"
			case "ABVWSTRT":
				dataType = "raw"
			case "ABVMSTRT":
				dataType = "raw"
			case "HMSKSTRT":
				dataType = "raw"
			case "GUNSTART":
				dataType = "raw"
			case "ELEVSTRT":
				dataType = "raw"
			case "DOORSTRT":
				dataType = "patch"
			case "SIDESTRT":
				dataType = "raw"
			case "MASKSTRT":
				dataType = "raw"
			case "UPDNSTRT":
				dataType = "raw"
			case "SKYSTART":
				dataType = "raw"
			case "ORDRSTRT":
				dataType = "raw"
			case "SHAPSTRT":
				dataType = "raw"
			case "DIGISTRT":
				dataType = "raw"
			case "G_START":
				dataType = "raw"
			case "PCSTART":
				dataType = "raw"
			case "ADSTART":
				dataType = "raw"
			}
			if lumpInfo.Size > 0 {
				destFname := fmt.Sprintf("%s/%s", destDir, lumpInfo.NameString())
				switch dataType {
				case "wall":
					destFname = fmt.Sprintf("%s.png", destFname)
				case "midi":
					destFname = fmt.Sprintf("%s.mid", destFname)
				default:
					destFname = fmt.Sprintf("%s.dat", destFname)
				}
				dumpLumpDataToFile(wadFile, lumpInfo, destFname, dataType)
			}
		}
	}
}
