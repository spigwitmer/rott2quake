package wad

// WAD utility functions

import (
	"errors"
	"image"
	"image/color"
	"image/png"
	"io"
)

// convert palette image data to PNG before writing
func DumpFlatDataToFile(destFhnd io.WriteSeeker, lumpReader io.Reader, iwad *IWAD, width int, height int) (int64, error) {
	rawImgData := make([]byte, width*height)
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
			palette := iwad.BasePaletteData[rawImgData[(i*width)+j]]
			img.SetRGBA(i, j, color.RGBA{palette.R, palette.G, palette.B, 255})
		}
	}

	if err := png.Encode(destFhnd, img); err != nil {
		return 0, err
	}

	return destFhnd.Seek(0, io.SeekCurrent)
}
