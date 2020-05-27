package wad

// Rise of the Triad specific utility functions

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
)

// convert patch data to PNG before writing
func DumpPatchDataToFile(destFhnd io.WriteSeeker, lumpInfo *LumpHeader, lumpReader io.Reader, iwad *IWAD) (int64, error) {
	// https://doomwiki.org/wiki/Picture_format

	// read entire lump to perform random access
	patchBytes := make([]byte, lumpInfo.Size)
	_, err := lumpReader.Read(patchBytes)
	if err != nil {
		return 0, err
	}
	lumpBuffer := bytes.NewReader(patchBytes)

	var patchHeader PatchHeader
	if err := binary.Read(lumpBuffer, binary.LittleEndian, &patchHeader); err != nil {
		log.Fatal(err)
	}

	columnOffsets := make([]uint16, patchHeader.Width)
	if err := binary.Read(lumpBuffer, binary.LittleEndian, &columnOffsets); err != nil {
		return 0, err
	}

	img := image.NewPaletted(image.Rect(0, 0, int(patchHeader.Width), int(patchHeader.Height)), iwad.BasePaletteData)
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

			for i := uint8(0); i < pixelCount-1; i++ {
				paletteCode, err := lumpBuffer.ReadByte()
				if err != nil {
					return 0, err
				}

				img.SetColorIndex(idx, int(i+rowstart), paletteCode)
			}

			// read dummy byte
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

// convert VGA planar data to PNG
func DumpPicDataToFile(destFhnd io.WriteSeeker, lumpInfo *LumpHeader, lumpReader io.Reader, iwad *IWAD) (int64, error) {
	var header RottPicHeader

	if err := binary.Read(lumpReader, binary.LittleEndian, &header); err != nil {
		return 0, err
	}

	rawData := make([]uint8, int(header.Width)*int(header.Height)*4)

	if err := binary.Read(lumpReader, binary.LittleEndian, rawData); err != nil {
		return 0, err
	}

	img := image.NewPaletted(image.Rect(0, 0, int(header.Width)*4, int(header.Height)), iwad.BasePaletteData)
	for planenum := 0; planenum < 4; planenum++ {
		for i := 0; i < int(header.Height); i++ {
			for j := 0; j < int(header.Width); j++ {
				rawPos := (i * int(header.Width)) + j
				val := rawData[rawPos+(int(header.Width)*int(header.Height)*planenum)]
				img.SetColorIndex((j*4)+planenum, i, val)
			}
		}
	}

	if err := png.Encode(destFhnd, img); err != nil {
		return 0, err
	}

	return destFhnd.Seek(0, io.SeekCurrent)
}

// convert floor and ceiling data to PNG
func DumpLpicDataToFile(destFhnd io.WriteSeeker, lumpInfo *LumpHeader, lumpReader io.Reader, iwad *IWAD) (int64, error) {
	var header RottLpicHeader

	if err := binary.Read(lumpReader, binary.LittleEndian, &header); err != nil {
		return 0, err
	}

	rawData := make([]uint8, 128*128)

	if err := binary.Read(lumpReader, binary.LittleEndian, rawData); err != nil {
		return 0, err
	}

	img := image.NewPaletted(image.Rect(0, 0, 128, 128), iwad.BasePaletteData)
	for i := 0; i < 128; i++ {
		for j := 0; j < 128; j++ {
			img.SetColorIndex(j, i, rawData[(i*128)+j])
		}
	}

	if err := png.Encode(destFhnd, img); err != nil {
		return 0, err
	}

	return destFhnd.Seek(0, io.SeekCurrent)
}

// convert translucent patch data to PNG before writing
func DumpTransPatchDataToFile(destFhnd io.WriteSeeker, lumpInfo *LumpHeader, lumpReader io.Reader, iwad *IWAD) (int64, error) {
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

	fmt.Printf("transpatch_t header: %+v\n", patchHeader)

	img := image.NewPaletted(image.Rect(0, 0, int(patchHeader.Width), int(patchHeader.Height)), iwad.BasePaletteData)
	for idx, cOffset := range columnOffsets {
		fmt.Printf("columnOffset: %d\n", cOffset)
		_, err := lumpBuffer.Seek(int64(cOffset), io.SeekStart)
		if err != nil {
			return 0, err
		}
		rowstart, err := lumpBuffer.ReadByte()
		if err != nil {
			return 0, err
		}
		fmt.Printf("\t\trowstart = %d\n", rowstart) // XXX
		for rowstart != 255 {
			pixelCount, err := lumpBuffer.ReadByte()
			if err != nil {
				return 0, err
			}

			fmt.Printf("\t\tpixelcount = %d\n", rowstart) // XXX
			src, err := lumpBuffer.ReadByte()
			if err != nil {
				return 0, err
			}
			fmt.Printf("\t\tsrc = %d\n", rowstart) // XXX

			if src == 254 {
				// TODO: translucency shiz?
			} else {
				for i := uint8(0); i < pixelCount-1; i++ {
					paletteCode, err := lumpBuffer.ReadByte()
					if err != nil {
						return 0, err
					}
					img.SetColorIndex(idx, int(i+rowstart), paletteCode)
				}
			}
			rowstart, err = lumpBuffer.ReadByte()
			if err != nil {
				return 0, err
			}
			fmt.Printf("\t\trowstart = %d\n", rowstart) // XXX
		}
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
func DumpLBMDataToFile(destFhnd io.WriteSeeker, lumpInfo *LumpHeader, lumpReader io.Reader, iwad *IWAD) (int64, error) {
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
