package wad

// WAD utility functions

import (
	"errors"
	"image"
	"image/png"
	"io"
)

func GetImageFromFlatData(lumpReader io.Reader, iwad *WADReader, width int, height int) (*image.Paletted, error) {
	rawImgData := make([]byte, width*height)
	numRead, err := lumpReader.Read(rawImgData[:])
	if err != nil {
		return nil, err
	}
	if numRead != width*height {
		return nil, errors.New("numRead != width*height???")
	}

	img := image.NewPaletted(image.Rect(0, 0, width, height), iwad.BasePaletteData)
	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {
			img.SetColorIndex(i, j, rawImgData[(i*width)+j])
		}
	}

	return img, nil
}

// convert palette image data to PNG before writing
func DumpFlatDataToFile(destFhnd io.WriteSeeker, lumpReader io.Reader, iwad *WADReader, width int, height int) (int64, error) {

	img, err := GetImageFromFlatData(lumpReader, iwad, width, height)
	if err != nil {
		return 0, err
	}
	if err := png.Encode(destFhnd, img); err != nil {
		return 0, err
	}

	return destFhnd.Seek(0, io.SeekCurrent)
}
