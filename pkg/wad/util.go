package wad

// WAD utility functions

import (
	"errors"
	"fmt"
	"gitlab.com/camtap/rott2quake/pkg/imgutil"
	"gitlab.com/camtap/rott2quake/pkg/lumps"
	"image"
	"image/color"
	"image/png"
	"io"
)

func GetImageFromFlatData(lumpReader io.Reader, iwad lumps.ArchiveReader, width int, height int, transparent bool) (*image.RGBA, error) {
	rawImgData := make([]byte, width*height)
	numRead, err := lumpReader.Read(rawImgData[:])
	if err != nil {
		return nil, err
	}
	if numRead != width*height {
		return nil, errors.New("numRead != width*height???")
	}

	pal := *imgutil.GetPalette(iwad.Type())
	if pal == nil {
		return nil, fmt.Errorf("Game %s does not have a palette", iwad.Type())
	}
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	if !transparent {
		for i := 0; i < height; i++ {
			for j := 0; j < width; j++ {
				img.SetRGBA(i, j, color.RGBA{0, 0, 0, 0xff})
			}
		}
	}

	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {
			r, g, b, _ := pal[rawImgData[(i*width)+j]].RGBA()
			img.SetRGBA(i, j, color.RGBA{uint8(r), uint8(g), uint8(b), 0xff})
		}
	}

	return img, nil
}

// convert palette image data to PNG before writing
func DumpFlatDataToFile(destFhnd io.WriteSeeker, lumpReader io.Reader, iwad lumps.ArchiveReader, width int, height int, transparent bool) (int64, error) {

	img, err := GetImageFromFlatData(lumpReader, iwad, width, height, transparent)
	if err != nil {
		return 0, err
	}
	if err := png.Encode(destFhnd, img); err != nil {
		return 0, err
	}

	return destFhnd.Seek(0, io.SeekCurrent)
}
