package wad

// Rise of the Triad specific utility functions

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
)

// returns lump containing palette data
func GetPaletteData(iwad *IWAD) ([]Palette, error) {
	paletteData := make([]Palette, 256)

	paletteLump, err := iwad.GetLump("PAL")
	if err != nil {
		return nil, err
	}
	paletteLumpData, err := iwad.LumpData(paletteLump)
	if err != nil {
		return nil, err
	}
	if err = binary.Read(paletteLumpData, binary.LittleEndian, &paletteData); err != nil {
		return nil, err
	}

	return paletteData, nil
}

// convert patch data to PNG before writing
func DumpPatchDataToFile(destFhnd io.WriteSeeker, lumpInfo *LumpHeader, lumpReader io.Reader, iwad *IWAD) (int64, error) {
	// https://doomwiki.org/wiki/Picture_format

	paletteData, err := GetPaletteData(iwad)
	if err != nil {
		return 0, err
	}

	// read entire lump to perform random access
	patchBytes := make([]byte, lumpInfo.Size)
	_, err = lumpReader.Read(patchBytes)
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

			for i := uint8(0); i < pixelCount-1; i++ {
				paletteCode, err := lumpBuffer.ReadByte()
				if err != nil {
					return 0, err
				}

				pixel := paletteData[paletteCode]
				img.SetRGBA(idx, int(i+rowstart), color.RGBA{pixel.R, pixel.G, pixel.B, 255})
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

// convert floor and ceiling data to PNG
func DumpLpicDataToFile(destFhnd io.WriteSeeker, lumpInfo *LumpHeader, lumpReader io.Reader, iwad *IWAD) (int64, error) {
	var header RottLpicHeader

	paletteData, err := GetPaletteData(iwad)
	if err != nil {
		return 0, err
	}

	if err := binary.Read(lumpReader, binary.LittleEndian, &header); err != nil {
		return 0, err
	}

	rawData := make([]uint8, 128*128)

	if err := binary.Read(lumpReader, binary.LittleEndian, rawData); err != nil {
		return 0, err
	}

	img := image.NewRGBA(image.Rect(0, 0, 128, 128))
	for i := 0; i < 128; i++ {
		for j := 0; j < 128; j++ {
			pixel := paletteData[rawData[(i*128)+j]]
			img.SetRGBA(j, i, color.RGBA{pixel.R, pixel.G, pixel.B, 255})
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

	paletteData, err := GetPaletteData(iwad)
	if err != nil {
		return 0, err
	}

	// read entire lump to perform random access
	patchBytes := make([]byte, lumpInfo.Size)
	_, err = lumpReader.Read(patchBytes)
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

			for i := uint8(0); i < pixelCount-1; i++ {
				paletteCode, err := lumpBuffer.ReadByte()
				if err != nil {
					return 0, err
				}

				pixel := paletteData[paletteCode]
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
