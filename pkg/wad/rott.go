package wad

// Rise of the Triad specific stuff

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"gitlab.com/camtap/lumps/pkg/imgutil"
	"gitlab.com/camtap/lumps/pkg/lumps"
	"gitlab.com/camtap/lumps/pkg/rtl"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
)

// wall and sky textures
type RottPatchHeader struct {
	OrigSize     int16
	Width        int16
	Height       int16
	LeftOffset   int16
	TopOffset    int16
	Transparency int16
}

// floors and ceilings
type RottLpicHeader struct {
	Width  uint16
	Height uint16
	OrgX   uint16
	OrgY   uint16
}

type RottPicHeader struct {
	Width  uint8
	Height uint8
}

// https://doomwiki.org/wiki/Picture_format
func GetImageFromPatchData(lumpInfo lumps.ArchiveEntry, lumpReader io.Reader, iwad lumps.ArchiveReader) (*image.Paletted, error) {
	// read entire lump to perform random access
	patchBytes := make([]byte, lumpInfo.Size())
	_, err := lumpReader.Read(patchBytes)
	if err != nil {
		return nil, err
	}
	lumpBuffer := bytes.NewReader(patchBytes)

	var patchHeader PatchHeader
	if err := binary.Read(lumpBuffer, binary.LittleEndian, &patchHeader); err != nil {
		return nil, err
	}

	columnOffsets := make([]uint16, patchHeader.Width)
	if err := binary.Read(lumpBuffer, binary.LittleEndian, &columnOffsets); err != nil {
		return nil, err
	}

	pal := imgutil.GetPalette(iwad.Type())
	if pal == nil {
		return nil, fmt.Errorf("Game %s does not have a palette", iwad.Type())
	}
	img := image.NewPaletted(image.Rect(0, 0, int(patchHeader.Width), int(patchHeader.Height)), *pal)
	// default to transparent
	for i := 0; i < 128; i++ {
		for j := 0; j < 128; j++ {
			img.SetColorIndex(i, j, 0)
		}
	}
	for idx, cOffset := range columnOffsets {
		_, err := lumpBuffer.Seek(int64(cOffset), io.SeekStart)
		if err != nil {
			return nil, err
		}
		rowstart := byte(0)
		for rowstart != 255 {
			rowstart, err := lumpBuffer.ReadByte()
			if err != nil {
				return nil, err
			}
			if rowstart == 255 {
				break
			}

			var pixelCount uint8
			err = binary.Read(lumpBuffer, binary.LittleEndian, &pixelCount)
			if err != nil {
				return nil, err
			}

			if pixelCount > 0 {
				for i := uint8(0); i < pixelCount-1; i++ {
					paletteCode, err := lumpBuffer.ReadByte()
					if err != nil {
						return nil, err
					}

					img.SetColorIndex(idx, int(i+rowstart), paletteCode)
				}
			}

			// read dummy byte
			_, err = lumpBuffer.ReadByte()
			if err != nil {
				return nil, err
			}
		}
	}

	return img, nil
}

// convert patch data to PNG before writing
func DumpPatchDataToFile(destFhnd io.WriteSeeker, lumpInfo lumps.ArchiveEntry, lumpReader io.Reader, iwad lumps.ArchiveReader) (int64, error) {
	img, err := GetImageFromPatchData(lumpInfo, lumpReader, iwad)
	if err != nil {
		return 0, err
	}

	if err = png.Encode(destFhnd, img); err != nil {
		return 0, err
	}

	return destFhnd.Seek(0, io.SeekCurrent)
}

func GetImageFromPicData(lumpInfo lumps.ArchiveEntry, lumpReader io.Reader, iwad lumps.ArchiveReader) (*image.Paletted, error) {
	var header RottPicHeader

	if err := binary.Read(lumpReader, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	rawData := make([]uint8, int(header.Width)*int(header.Height)*4)

	if err := binary.Read(lumpReader, binary.LittleEndian, rawData); err != nil {
		return nil, err
	}

	pal := imgutil.GetPalette(iwad.Type())
	if pal == nil {
		return nil, fmt.Errorf("Game %s does not have a palette", iwad.Type())
	}
	img := image.NewPaletted(image.Rect(0, 0, int(header.Width)*4, int(header.Height)), *pal)
	for planenum := 0; planenum < 4; planenum++ {
		for i := 0; i < int(header.Height); i++ {
			for j := 0; j < int(header.Width); j++ {
				rawPos := (i * int(header.Width)) + j
				val := rawData[rawPos+(int(header.Width)*int(header.Height)*planenum)]
				img.SetColorIndex((j*4)+planenum, i, val)
			}
		}
	}

	return img, nil
}

// convert VGA planar data to PNG
func DumpPicDataToFile(destFhnd io.WriteSeeker, lumpInfo lumps.ArchiveEntry, lumpReader io.Reader, iwad lumps.ArchiveReader) (int64, error) {
	img, err := GetImageFromPicData(lumpInfo, lumpReader, iwad)
	if err != nil {
		return 0, err
	}

	if err := png.Encode(destFhnd, img); err != nil {
		return 0, err
	}

	return destFhnd.Seek(0, io.SeekCurrent)
}

func GetImageFromLpicData(lumpInfo lumps.ArchiveEntry, lumpReader io.Reader, iwad lumps.ArchiveReader) (*image.Paletted, error) {
	var header RottLpicHeader

	if err := binary.Read(lumpReader, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	rawData := make([]uint8, 128*128)

	if err := binary.Read(lumpReader, binary.LittleEndian, rawData); err != nil {
		return nil, err
	}

	pal := imgutil.GetPalette(iwad.Type())
	if pal == nil {
		return nil, fmt.Errorf("Game %s does not have a palette", iwad.Type())
	}
	img := image.NewPaletted(image.Rect(0, 0, 128, 128), *pal)
	for i := 0; i < 128; i++ {
		for j := 0; j < 128; j++ {
			img.SetColorIndex(j, i, rawData[(i*128)+j])
		}
	}

	return img, nil
}

// convert floor and ceiling data to PNG
func DumpLpicDataToFile(destFhnd io.WriteSeeker, lumpInfo lumps.ArchiveEntry, lumpReader io.Reader, iwad lumps.ArchiveReader) (int64, error) {
	img, err := GetImageFromLpicData(lumpInfo, lumpReader, iwad)

	if err != nil {
		return 0, err
	}

	if err := png.Encode(destFhnd, img); err != nil {
		return 0, err
	}

	return destFhnd.Seek(0, io.SeekCurrent)
}

// https://doomwiki.org/wiki/Picture_format
func GetImageFromTransPatchData(lumpInfo lumps.ArchiveEntry, lumpReader io.Reader, iwad lumps.ArchiveReader) (*image.Paletted, error) {
	// read entire lump to perform random access
	patchBytes := make([]byte, lumpInfo.Size())
	_, err := lumpReader.Read(patchBytes)
	if err != nil {
		return nil, err
	}
	lumpBuffer := bytes.NewReader(patchBytes)

	var patchHeader RottPatchHeader
	if err := binary.Read(lumpBuffer, binary.LittleEndian, &patchHeader); err != nil {
		log.Fatal(err)
	}

	columnOffsets := make([]uint16, patchHeader.Width)
	if err := binary.Read(lumpBuffer, binary.LittleEndian, &columnOffsets); err != nil {
		return nil, err
	}

	pal := imgutil.GetPalette(iwad.Type())
	if pal == nil {
		return nil, fmt.Errorf("Game %s does not have a palette", iwad.Type())
	}
	img := image.NewPaletted(image.Rect(0, 0, int(patchHeader.Width), int(patchHeader.Height)), *pal)
	// default to transparent
	for i := 0; i < 128; i++ {
		for j := 0; j < 128; j++ {
			img.SetColorIndex(i, j, 0)
		}
	}
	for idx, cOffset := range columnOffsets {
		//log.Printf("\tcOffset(%d): %04x", idx, cOffset)
		_, err := lumpBuffer.Seek(int64(cOffset), io.SeekStart)
		if err != nil {
			return nil, err
		}
		rowstart := byte(0)
		for rowstart != 255 {
			rowstart, err := lumpBuffer.ReadByte()
			if err != nil {
				return nil, err
			}
			//log.Printf("\t\trowstart(%d): %02x", idx, rowstart)
			if rowstart == 255 {
				break
			}

			pixelCount, err := lumpBuffer.ReadByte()
			if err != nil {
				return nil, err
			}
			//log.Printf("\t\tpixelCount(%d): %d", idx, pixelCount)

			if pixelCount == 0 {
				continue
			}

			src, err := lumpBuffer.ReadByte()
			if err != nil {
				return nil, err
			}

			//log.Printf("\t\tsrc(%d): %02x", idx, src)
			if src == 254 {
				// TODO: translucency shiz?
				for i := byte(0); i < pixelCount; i++ {
					img.SetColorIndex(idx, int(i+rowstart), 0)
				}
			} else {
				img.SetColorIndex(idx, int(rowstart), src)
				for i := uint8(1); i < pixelCount; i++ {
					paletteCode, err := lumpBuffer.ReadByte()
					if err != nil {
						return nil, err
					}
					img.SetColorIndex(idx, int(i+rowstart), paletteCode)
				}
			}
		}
	}
	return img, nil
}

// convert translucent patch data to PNG before writing
func DumpTransPatchDataToFile(destFhnd io.WriteSeeker, lumpInfo lumps.ArchiveEntry, lumpReader io.Reader, iwad lumps.ArchiveReader) (int64, error) {
	img, err := GetImageFromTransPatchData(lumpInfo, lumpReader, iwad)
	if err != nil {
		return 0, err
	}

	if err = png.Encode(destFhnd, img); err != nil {
		return 0, err
	}

	return destFhnd.Seek(0, io.SeekCurrent)
}

type RGB struct {
	R, G, B uint8
}
type LBMHeader struct {
	Width   uint16
	Height  uint16
	Palette [256]RGB
}

// convert expression of freedom from euclidian oppression to PNG
func DumpLBMDataToFile(destFhnd io.WriteSeeker, lumpInfo lumps.ArchiveEntry, lumpReader io.Reader, iwad lumps.ArchiveReader) (int64, error) {
	var header LBMHeader
	var palette color.Palette

	if err := binary.Read(lumpReader, binary.LittleEndian, &header); err != nil {
		return 0, nil
	}

	getByte := func() (uint8, error) {
		var mybyte [1]byte
		numread, err := lumpReader.Read(mybyte[:])
		if numread != 1 && err == nil {
			return 0, errors.New("Could not read byte")
		}
		return mybyte[0], err
	}

	for _, pd := range header.Palette {
		palette = append(palette, color.RGBA{pd.R, pd.G, pd.B, 255})
	}

	img := image.NewPaletted(image.Rect(0, 0, int(header.Width), int(header.Height)), palette)
	for i := uint16(0); i < header.Height; i++ {
		for j := uint16(0); j < header.Width; {
			rep, err := getByte()
			if err != nil {
				return 0, err
			}
			if rep > 0x80 {
				rep = (rep ^ 0xff) + 2
				val, err := getByte()
				if err != nil {
					return 0, err
				}
				k := uint16(0)
				for ; k < uint16(rep); k++ {
					img.SetColorIndex(int(j)+int(k), int(i), val)
				}
				j += uint16(rep)
			} else if rep < 0x80 {
				rep++
				for k := uint16(0); k < uint16(rep); k++ {
					val, err := getByte()
					if err != nil {
						return 0, err
					}
					img.SetColorIndex(int(j)+int(k), int(i), val)
				}
				j += uint16(rep)
			} else { // == 0x80
				j--
			}
		}
	}

	if err := png.Encode(destFhnd, img); err != nil {
		return 0, err
	}

	return destFhnd.Seek(0, io.SeekCurrent)
}

var TypeOneOffs = map[string][2]string{
	"SND_ON":   [2]string{"patch", "widgets"},
	"SND_OFF":  [2]string{"patch", "widgets"},
	"BLOCK1":   [2]string{"patch", "block"},
	"BLOCK2":   [2]string{"patch", "block"},
	"BLOCK3":   [2]string{"patch", "block"},
	"CACHEBAR": [2]string{"patch", "misc"},
	"INFO1":    [2]string{"patch", "info"},
	"INFO2":    [2]string{"patch", "info"},
	"INFO3":    [2]string{"patch", "info"},
	"INFO4":    [2]string{"patch", "info"},
	"INFO5":    [2]string{"patch", "info"},
	"INFO6":    [2]string{"patch", "info"},
	"INFO7":    [2]string{"patch", "info"},
	"INFO8":    [2]string{"patch", "info"},
	"INFO9":    [2]string{"patch", "info"},
	"DEADJOE":  [2]string{"patch", "boss-deaths"},
	"DEADROBO": [2]string{"patch", "boss-deaths"},
	"DEADSTEV": [2]string{"patch", "boss-deaths"},
	"DEADTOM":  [2]string{"patch", "boss-deaths"},
	"HSWITCH5": [2]string{"tpatch", "masked"},
	"HSWITCH8": [2]string{"tpatch", "masked"},
	"HSWTICH4": [2]string{"patch", "masked"},
	"HSWITCH6": [2]string{"patch", "masked"},
	"HSWITCH7": [2]string{"patch", "masked"},
	"HSWTCH9":  [2]string{"patch", "masked"},
	"PAL":      [2]string{"raw", "misc"},
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

func ROTTGuessFileTypeAndSubdir(entry *WADEntry) (string, string) {
	entryName := entry.Name()
	var dataType, subdir string

	// masked walls have more specific requirements on what type of
	// format it is
	for _, maskedentry := range rtl.MaskedWalls {
		if entryName == maskedentry.Bottom {
			return "tpatch", "masked"
		} else if entryName == maskedentry.Above || entryName == maskedentry.Middle {
			return "patch", "masked"
		} else if entryName == maskedentry.Side {
			return "wall", "masked"
		}
	}

	for _, maskedentry := range rtl.Platforms {
		if entryName == maskedentry.Bottom {
			return "tpatch", "masked"
		} else if entryName == maskedentry.Above || entryName == maskedentry.Middle {
			return "patch", "masked"
		} else if entryName == maskedentry.Side {
			return "wall", "masked"
		}
	}

	if oneOff, found := TypeOneOffs[entry.Name()]; found {
		return oneOff[0], oneOff[1]
	}

	// TODO: fix this. this is terribly written.
	// map things. skiplist things. do anything besides
	// traversing through the entire directory
	for _, direntry := range entry.Reader.LumpDirectory {
		if direntry.NameString() == entryName {
			return dataType, subdir
		}
		switch direntry.NameString() {
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
			subdir = "masked-unknown"
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
	}

	return dataType, subdir
}
