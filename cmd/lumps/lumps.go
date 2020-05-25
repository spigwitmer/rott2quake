package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	rtlfile "gitlab.com/camtap/lumps/pkg/rtl"
	"gitlab.com/camtap/lumps/pkg/wad"
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

func dumpLumpDataToFile(wadFile *wad.IWAD, lumpInfo *wad.LumpHeader, destFname string,
	dataType string) {
	lumpReader, err := wadFile.LumpData(lumpInfo)
	if err != nil {
		log.Fatalf("Could not get lump data reader for %s: %v\n", destFname, err)
	}
	fmt.Printf("dumping %s as %s\n", destFname, dataType)
	destfhnd, err := os.Create(destFname)
	if err != nil {
		log.Fatalf("Could not write to %s: %v\n", destFname, err)
	}
	defer destfhnd.Close()

	switch dataType {
	case "wall":
		// assumes 64x64 (standard mandated by ROTT)
		_, err = wad.DumpFlatDataToFile(destfhnd, lumpReader, wadFile, 64, 64)
	case "sky":
		// assumes 256x200 (standard mandated by ROTT)
		_, err = wad.DumpFlatDataToFile(destfhnd, lumpReader, wadFile, 256, 200)
	case "midi":
		_, err = dumpRawLumpDataToFile(destfhnd, lumpReader)
	case "patch":
		_, err = wad.DumpPatchDataToFile(destfhnd, lumpInfo, lumpReader, wadFile)
	case "tpatch":
		_, err = wad.DumpTransPatchDataToFile(destfhnd, lumpInfo, lumpReader, wadFile)
	case "lpic":
		_, err = wad.DumpLpicDataToFile(destfhnd, lumpInfo, lumpReader, wadFile)
	case "pic":
		_, err = wad.DumpPicDataToFile(destfhnd, lumpInfo, lumpReader, wadFile)
	default:
		_, err = dumpRawLumpDataToFile(destfhnd, lumpReader)
	}

	if err != nil {
		if dataType != "raw" {
			// just dump the raw data instead then
			destfhnd.Close()
			_ = os.Remove(destFname)
			newFname := fmt.Sprintf("%s.dat", destFname)
			log.Printf("Could not copy to %s (%v), writing raw to %s instead", destFname, err, newFname)
			lumpReader, err = wadFile.LumpData(lumpInfo)
			if err != nil {
				log.Fatal(err)
			}
			newfhnd, err := os.Create(newFname)
			if err != nil {
				log.Fatalf("Could not write to %s: %v", newFname, err)
			}
			defer newfhnd.Close()
			_, err = dumpRawLumpDataToFile(newfhnd, lumpReader)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			log.Fatalf("Could not copy to %s: %v\n", destFname, err)
		}
	}
}

func main() {
	var dumpLumpData, printLumps bool
	var rtlFile, rtlMapOutdir, lumpName, lumpType string
	var rtl *rtlfile.RTL
	var printRTLInfo bool

	flag.StringVar(&rtlFile, "rtl", "", "RTL file")
	flag.StringVar(&lumpName, "lname", "", "Dump data only for this lump")
	flag.StringVar(&lumpType, "ltype", "", "force specific lump type (only relevant when -lname is specified)")
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

		rtl, err = rtlfile.NewRTL(rtlFhnd)
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
			rtlRawWallFile := fmt.Sprintf("%s/map%03d-walls.bin", rtlMapOutdir, idx+1)
			rtlRawSpriteFile := fmt.Sprintf("%s/map%03d-sprites.bin", rtlMapOutdir, idx+1)
			rtlRawInfoFile := fmt.Sprintf("%s/map%03d-info.bin", rtlMapOutdir, idx+1)

			wallFhnd, err := os.Create(rtlMapFile)
			if err != nil {
				log.Fatalf("Could not open %s for writing: %v\n", rtlMapFile, err)
			}
			defer wallFhnd.Close()
			err = md.DumpWallToFile(wallFhnd)
			if err != nil {
				log.Fatalf("Could not write map to %s: %v\n", rtlMapFile, err)
			}

			rawWallFhnd, err := os.Create(rtlRawWallFile)
			if err != nil {
				log.Fatalf("Could not open %s for writing: %v\n", rtlRawWallFile, err)
			}
			defer rawWallFhnd.Close()
			if err = binary.Write(rawWallFhnd, binary.LittleEndian, md.WallPlane); err != nil {
				log.Fatalf("Could not write raw map to %s: %v\n", rtlRawWallFile, err)
			}

			rawSpriteFhnd, err := os.Create(rtlRawSpriteFile)
			if err != nil {
				log.Fatalf("Could not open %s for writing: %v\n", rtlRawSpriteFile, err)
			}
			defer rawSpriteFhnd.Close()
			if err = binary.Write(rawSpriteFhnd, binary.LittleEndian, md.SpritePlane); err != nil {
				log.Fatalf("Could not write raw map to %s: %v\n", rtlRawSpriteFile, err)
			}

			rawInfoFhnd, err := os.Create(rtlRawInfoFile)
			if err != nil {
				log.Fatalf("Could not open %s for writing: %v\n", rtlRawInfoFile, err)
			}
			defer rawInfoFhnd.Close()
			if err = binary.Write(rawInfoFhnd, binary.LittleEndian, md.InfoPlane); err != nil {
				log.Fatalf("Could not write raw map to %s: %v\n", rtlRawInfoFile, err)
			}
		}
	}

	fhnd, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatalf("Could not open file: %v\n", err)
	}
	wadFile, err := wad.NewIWAD(fhnd)
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

		if wadFile.BasePaletteData == nil {
			log.Fatalf("Cannot dump IWAD: no pallete data\n")
		}
		destDir := flag.Arg(1)

		if err := os.MkdirAll(destDir, 0755); err != nil {
			log.Fatalf("Could not create dest dir: %v\n", err)
		}

		subdir := ""
		dataType := "raw"
		for i := uint32(0); i < wadFile.Header.NumLumps; i += 1 {
			lumpInfo := wadFile.LumpDirectory[i]
			if lumpName != "" && lumpInfo.NameString() != lumpName {
				continue
			}
			switch lumpInfo.NameString() {
			case "WALLSTRT":
				dataType = "wall"
				subdir = "wall"
			case "SONGSTRT":
				dataType = "midi"
				subdir = "music"
			case "ANIMSTRT":
				dataType = "wall"
				subdir = "anim"
			case "EXITSTRT":
				dataType = "raw"
				subdir = ""
			case "ABVWSTRT":
				dataType = "raw"
				subdir = ""
			case "ABVMSTRT":
				dataType = "raw"
				subdir = ""
			case "HMSKSTRT":
				dataType = "raw"
				subdir = ""
			case "GUNSTART":
				dataType = "patch"
				subdir = "guns"
			case "ELEVSTRT":
				dataType = "wall"
				subdir = "elev"
			case "DOORSTRT":
				dataType = "patch"
				subdir = "doors"
			case "SIDESTRT":
				dataType = "patch"
				subdir = "side"
			case "MASKSTRT":
				dataType = "raw"
				subdir = ""
			case "UPDNSTRT":
				dataType = "lpic"
				subdir = "floors-ceilings"
			case "SKYSTART":
				dataType = "sky"
				subdir = "skies"
			case "ORDRSTRT":
				dataType = "raw"
				subdir = ""
			case "SHAPSTRT":
				dataType = "patch"
				subdir = "shapes"
			case "DIGISTRT":
				dataType = "raw"
				subdir = "sounds"
			case "G_START":
				dataType = "raw"
				subdir = "sounds"
			case "PCSTART":
				dataType = "raw"
				subdir = ""
			case "ADSTART":
				dataType = "raw"
				subdir = ""
			case "WALLSTOP", "EXITSTOP", "ELEVSTOP", "DOORSTOP", "SIDESTOP", "MASKSTOP",
				"UPDNSTOP", "SKYSTOP", "ORDRSTOP", "SHAPSTOP", "DIGISTOP", "PCSTOP", "ADSTOP":
				dataType = "raw"
				subdir = ""
			case "PAL":
				dataType = "raw"
				subdir = "misc"
			}
			if lumpInfo.Size > 0 {
				var destFname string
				if subdir == "" {
					destFname = fmt.Sprintf("%s/%s", destDir, lumpInfo.NameString())
				} else {
					if err := os.MkdirAll(destDir+"/"+subdir, 0755); err != nil {
						log.Fatal(err)
					}
					destFname = fmt.Sprintf("%s/%s/%s", destDir, subdir, lumpInfo.NameString())
				}
				if lumpName != "" && lumpType != "" {
					dataType = lumpType
				}
				switch dataType {
				case "patch":
					destFname = fmt.Sprintf("%s.png", destFname)
				case "tpatch":
					destFname = fmt.Sprintf("%s.png", destFname)
				case "pic":
					destFname = fmt.Sprintf("%s.png", destFname)
				case "lpic":
					destFname = fmt.Sprintf("%s.png", destFname)
				case "wall":
					destFname = fmt.Sprintf("%s.png", destFname)
				case "sky":
					destFname = fmt.Sprintf("%s.png", destFname)
				case "midi":
					destFname = fmt.Sprintf("%s.mid", destFname)
				default:
					destFname = fmt.Sprintf("%s.dat", destFname)
				}
				dumpLumpDataToFile(wadFile, lumpInfo, destFname, dataType)
				dumpLumpDataToFile(wadFile, lumpInfo, destFname+".raw", "raw")
			}
		}
	}
}
