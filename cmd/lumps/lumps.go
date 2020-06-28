package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"gitlab.com/camtap/lumps/pkg/lumps"
	"gitlab.com/camtap/lumps/pkg/pak"
	rtlfile "gitlab.com/camtap/lumps/pkg/rtl"
	"gitlab.com/camtap/lumps/pkg/wad"
	"gitlab.com/camtap/lumps/pkg/wad2"
)

var TypeOneOffs = map[string][2]string{
	"SND_ON":   [2]string{"patch", "widgets"},
	"SND_OFF":  [2]string{"patch", "widgets"},
	"DEADJOE":  [2]string{"patch", "boss-deaths"},
	"DEADROBO": [2]string{"patch", "boss-deaths"},
	"DEADSTEV": [2]string{"patch", "boss-deaths"},
	"DEADTOM":  [2]string{"patch", "boss-deaths"},
	"LICENSE":  [2]string{"raw", "misc"},
	"IMFREE":   [2]string{"lbm", "misc"},
	"BOOTBLOD": [2]string{"lbm", "misc"},
	"BOOTNORM": [2]string{"lbm", "misc"},
	"SVENDOR":  [2]string{"lbm", "misc"},
	"DEADBOSS": [2]string{"lbm", "misc"},
	"MMBK":     [2]string{"pic", "misc"},
	"PAUSED":   [2]string{"pic", "misc"},
	"WAIT":     [2]string{"pic", "misc"},
	"TNUMB":    [2]string{"pic", "misc"},
	"BATTP":    [2]string{"pic", "misc"},
	"DOOR2":    [2]string{"wall", "doors"},
	"EDOOR":    [2]string{"wall", "doors"},
	"RAMDOOR1": [2]string{"wall", "doors"},
	"SDOOR4":   [2]string{"wall", "doors"},
	"SNADOOR":  [2]string{"wall", "doors"},
	"SNDOOR":   [2]string{"wall", "doors"},
	"SNKDOOR":  [2]string{"wall", "doors"},
	"TNADOOR":  [2]string{"wall", "doors"},
	"TNDOOR":   [2]string{"wall", "doors"},
	"TNKDOOR":  [2]string{"wall", "doors"},
	"TRIDOOR1": [2]string{"wall", "doors"},
	"SIDE8":    [2]string{"wall", "side"},
	"SIDE21":   [2]string{"wall", "side"},
	"LOCK1":    [2]string{"wall", "side"},
	"LOCK2":    [2]string{"wall", "side"},
	"LOCK3":    [2]string{"wall", "side"},
	"LOCK4":    [2]string{"wall", "side"},
	"SIDE13":   [2]string{"wall", "side"},
	"SIDE16":   [2]string{"wall", "side"},
	"SIDE17":   [2]string{"wall", "side"},
}

func init() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s <.WAD file> [dest dir]\n", os.Args[0])
		flag.PrintDefaults()
	}
}

func dumpRawLumpDataToFile(destFhnd io.WriteSeeker, lumpReader io.Reader) (int64, error) {
	return io.Copy(destFhnd, lumpReader)
}

func dumpLumpDataToFile(archive lumps.ArchiveReader, entry lumps.ArchiveEntry, destFname string,
	dataType string, wad2Writer *wad2.WADWriter) {
	lumpReader, err := entry.Open()
	if err != nil {
		log.Fatalf("Could not get lump data reader for %s: %v\n", destFname, err)
	}
	fmt.Printf("dumping %s as %s\n", destFname, dataType)
	err = os.MkdirAll(filepath.Dir(destFname), 0755)
	if err != nil {
		log.Fatalf("Could not create folder for %s: %v\n", destFname, err)
	}
	destfhnd, err := os.Create(destFname)
	if err != nil {
		log.Fatalf("Could not write to %s: %v\n", destFname, err)
	}
	defer destfhnd.Close()

	switch dataType {
	case "wall":
		// assumes 64x64 (standard mandated by ROTT)
		_, err = wad.DumpFlatDataToFile(destfhnd, lumpReader, archive, 64, 64)
	case "sky":
		// assumes 256x200 (standard mandated by ROTT)
		_, err = wad.DumpFlatDataToFile(destfhnd, lumpReader, archive, 256, 200)
	case "midi":
		_, err = dumpRawLumpDataToFile(destfhnd, lumpReader)
	case "patch":
		_, err = wad.DumpPatchDataToFile(destfhnd, entry, lumpReader, archive)
	case "tpatch":
		_, err = wad.DumpTransPatchDataToFile(destfhnd, entry, lumpReader, archive)
	case "lpic":
		_, err = wad.DumpLpicDataToFile(destfhnd, entry, lumpReader, archive)
	case "pic":
		_, err = wad.DumpPicDataToFile(destfhnd, entry, lumpReader, archive)
	case "lbm":
		_, err = wad.DumpLBMDataToFile(destfhnd, entry, lumpReader, archive)
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
			lumpReader, err := entry.Open()
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

	if wad2Writer != nil {
		// dump select lumps into outgoing quake wad
		// specifically, we want:
		// * palette
		// * sky/floor textures
		// just dump the palette data as is, convert sky/floor data
		// into MIP textures
		if entry.Name() == "PAL" {
			var paletteData [768]byte
			rawLumpReader, err := entry.Open()
			if err != nil {
				log.Fatalf("Could not get %s lump data: %v\n", entry.Name(), err)
			}
			_, err = rawLumpReader.Read(paletteData[:])
			if err != nil {
				log.Fatalf("Could not read palette data: %v\n", err)
			}
			wad2Writer.AddLump("PALETTE", paletteData[:], wad2.LT_RAW)
		} else if dataType == "sky" {
			rawLumpReader, err := entry.Open()
			if err != nil {
				log.Fatalf("Could not get %s lump data: %v\n", entry.Name(), err)
			}
			img, err := wad.GetImageFromFlatData(rawLumpReader, archive, 256, 200)
			if err != nil {
				log.Fatalf("Could not get flat data image: %v\n", err)
			}
			mipdata, err := wad2.PalettedImageToMIPTexture(img)
			if err != nil {
				log.Fatalf("Could not get MIP texture from flat: %v\n", err)
			}
			wad2Writer.AddLump(entry.Name(), mipdata, wad2.LT_MIPTEX)
		} else if dataType == "lpic" {
			rawLumpReader, err := entry.Open()
			if err != nil {
				log.Fatalf("Could not get %s lump data: %v\n", entry.Name(), err)
			}
			img, err := wad.GetImageFromLpicData(entry, rawLumpReader, archive)
			if err != nil {
				log.Fatalf("Could not get lpic data image: %v\n", err)
			}
			mipdata, err := wad2.PalettedImageToMIPTexture(img)
			if err != nil {
				log.Fatalf("Could not get MIP texture from flat: %v\n", err)
			}
			wad2Writer.AddLump(entry.Name(), mipdata, wad2.LT_MIPTEX)
		} else if dataType == "pic" {
			rawLumpReader, err := entry.Open()
			if err != nil {
				log.Fatalf("Could not get %s lump data: %v\n", entry.Name(), err)
			}
			img, err := wad.GetImageFromPicData(entry, rawLumpReader, archive)
			if err != nil {
				log.Fatalf("Could not get pic data image: %v\n", err)
			}
			mipdata, err := wad2.PalettedImageToMIPTexture(img)
			if err != nil {
				log.Fatalf("Could not get MIP texture from flat: %v\n", err)
			}
			wad2Writer.AddLump(entry.Name(), mipdata, wad2.LT_MIPTEX)
		} else if dataType == "wall" {
			if animWall, frameNum := rtlfile.GetAnimatedWallInfo(entry.Name()); animWall != nil {
				// dump animated wall
				wad2LumpName := fmt.Sprintf("+%d%s", frameNum-1, animWall.StartingLump)
				rawLumpReader, err := entry.Open()
				if err != nil {
					log.Fatalf("Could not get %s lump data: %v\n", entry.Name(), err)
				}
				img, err := wad.GetImageFromFlatData(rawLumpReader, archive, 64, 64)
				if err != nil {
					log.Fatalf("Could not get wall data image: %v\n", err)
				}
				mipdata, err := wad2.PalettedImageToMIPTexture(img)
				if err != nil {
					log.Fatalf("Could not get MIP texture from flat: %v\n", err)
				}
				wad2Writer.AddLump(strings.ToLower(wad2LumpName), mipdata, wad2.LT_MIPTEX)

			}
		}
	}
}

func main() {
	var dumpLumpData, printLumps, dumpRaw bool
	var rtlFile, rtlMapOutdir, lumpName, lumpType string
	var wadOut string
	var isQuakeWad, isPak bool
	var convertToDusk bool
	var rtl *rtlfile.RTL
	var printRTLInfo bool
	var wadExtractor lumps.ArchiveReader

	flag.StringVar(&rtlFile, "rtl", "", "RTL file")
	flag.BoolVar(&isPak, "pak", false, "Input file is Quake .pak file")
	flag.StringVar(&lumpName, "lname", "", "Dump data only for this lump")
	flag.StringVar(&lumpType, "ltype", "", "force specific lump type (only relevant when -lname is specified)")
	flag.BoolVar(&printRTLInfo, "print-rtl-info", false, "Print RTL metadata (requires -rtl)")
	flag.StringVar(&wadOut, "wad-out", "", "output ripped image assets to Quake wad2 file (requires -dump-lump-data)")
	flag.BoolVar(&isQuakeWad, "quake", false, "wad specified is from Quake, not ROTT")
	flag.BoolVar(&convertToDusk, "dusk", false, "convert assets to Dusk rather than Quake")
	flag.StringVar(&rtlMapOutdir, "rtl-map-outdir", "", "Write RTL ASCII map out to this folder")
	flag.BoolVar(&dumpLumpData, "dump", false, "Dump Lump Data out to dest dir")
	flag.BoolVar(&dumpRaw, "dump-raw", false, "Dump raw lump data alongside rendered")
	flag.BoolVar(&printLumps, "list", false, "Print Lump Directory")
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
		if rtl == nil {
			log.Fatalf("Must provide RTL file when dumping map data")
		}
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
	if isQuakeWad {
		wadExtractor, err = wad2.NewWAD2Reader(fhnd)
		if err != nil {
			log.Fatalf("Could not open Quake wad for reading: %v\n", err)
		}

		quakeWad := wadExtractor.(*wad2.WAD2Reader)
		fmt.Printf("WAD2 file has %d lumps\n", len(quakeWad.Directory))
	} else if isPak {
		wadExtractor, err = pak.NewPAKReader(fhnd)
		if err != nil {
			log.Fatalf("Could not open Quake PAK for reading: %v\n", err)
		}

		quakePak := wadExtractor.(*pak.PAKReader)
		fmt.Printf("PAK file has %d entries\n", len(quakePak.Directory))
	} else {
		// default to ROTT wad
		wadExtractor, err = wad.NewIWAD(fhnd)
		if err != nil {
			log.Fatalf("Could not open IWAD: %v\n", err)
		}

		rottWad := wadExtractor.(*wad.WADReader)
		fmt.Printf("WAD file has %d lumps\n", len(rottWad.LumpDirectory))
	}

	if printLumps {
		iter := wadExtractor.List()
		for entry := iter.Next(); entry != nil; entry = iter.Next() {
			entry.Print()
		}
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

		var wadOutFile *os.File
		var wad2Out *wad2.WADWriter
		if wadOut != "" {
			if err := os.MkdirAll(path.Dir(wadOut), 0755); err != nil {
				log.Fatalf("Could not create wad out dir: %v\n", err)
			}
			if wadOutFile, err = os.Create(wadOut); err != nil {
				log.Fatalf("Could not open wad file %s: %v\n", wadOut, err)
			}

			if wad2Out, err = wad2.NewWADWriter(); err != nil {
				log.Fatalf("Could not create WAD2 writer: %v\n", err)
			}

			defer wadOutFile.Close()
		}

		subdir := ""
		dataType := "raw"
		wadIterator := wadExtractor.List()
		for lumpInfo := wadIterator.Next(); lumpInfo != nil; lumpInfo = wadIterator.Next() {
			if lumpName != "" && lumpInfo.Name() != lumpName {
				continue
			}
			switch lumpInfo.Name() {
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
				subdir = "sounds-digital"
			case "G_START":
				dataType = "raw"
				subdir = "sounds"
			case "PCSTART":
				dataType = "raw"
				subdir = "sounds-pcspkr"
			case "ADSTART":
				dataType = "raw"
				subdir = "sounds-adlib"
			case "WALLSTOP", "EXITSTOP", "ELEVSTOP", "DOORSTOP", "SIDESTOP", "MASKSTOP",
				"UPDNSTOP", "SKYSTOP", "ORDRSTOP", "SHAPSTOP", "DIGISTOP", "PCSTOP", "ADSTOP":
				dataType = "raw"
				subdir = ""
			case "PAL":
				dataType = "raw"
				subdir = "misc"
			}

			lastDataType, lastSubdir := dataType, subdir
			oneOffInfo, isOneOff := TypeOneOffs[lumpInfo.Name()]
			if isOneOff {
				dataType = oneOffInfo[0]
				subdir = oneOffInfo[1]
			}
			if lumpInfo.Size() > 0 {
				var destFname string
				if subdir == "" {
					destFname = fmt.Sprintf("%s/%s", destDir, lumpInfo.Name())
				} else {
					if err := os.MkdirAll(destDir+"/"+subdir, 0755); err != nil {
						log.Fatal(err)
					}
					destFname = fmt.Sprintf("%s/%s/%s", destDir, subdir, lumpInfo.Name())
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
				case "lbm":
					destFname = fmt.Sprintf("%s.png", destFname)
				case "midi":
					destFname = fmt.Sprintf("%s.mid", destFname)
				default:
					destFname = fmt.Sprintf("%s.dat", destFname)
				}
				dumpLumpDataToFile(wadExtractor, lumpInfo, destFname, dataType, wad2Out)
				if dumpRaw {
					dumpLumpDataToFile(wadExtractor, lumpInfo, destFname+".raw", "raw", nil)
				}
			}
			if isOneOff {
				dataType = lastDataType
				subdir = lastSubdir
			}
		}

		if wadOutFile != nil {
			wad2written, err := wad2Out.Write(wadOutFile)
			if err != nil {
				log.Fatalf("Could not write out wad file: %v\n", err)
			} else {
				fmt.Printf("Wad file %s written (%d bytes)\n", wadOut, wad2written)
			}
		}
	}
}
