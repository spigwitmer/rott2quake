package wad2

import (
	"bytes"
	"encoding/binary"
	"github.com/nfnt/resize"
	"gitlab.com/camtap/lumps/pkg/imgutil"
	"image"
	"image/draw"
)

type MIPTexture struct {
	Name      [16]byte
	Width     int32
	Height    int32
	Scale1Pos int32 // pos to 1:1 data
	Scale2Pos int32 // pos to data scaled to half
	Scale4Pos int32 // pos to data scaled to 1/4
	Scale8Pos int32 // pos to data scaled to 1/8
}

type QKPicHeader struct {
	Width      uint16
	Height     uint16
	LeftOffset int16
	TopOffset  int16
}

func imageToIndexByteArray(img *image.Paletted) []byte {
	width := img.Bounds().Max.X - img.Bounds().Min.X
	height := img.Bounds().Max.Y - img.Bounds().Min.Y
	data := make([]byte, width*height)

	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {
			data[i*width+j] = img.ColorIndexAt(j, i)
		}
	}

	return data
}

func PalettedImageToQuakePic(img *image.Paletted) ([]byte, error) {
	var picData bytes.Buffer
	var pic QKPicHeader
	pic.Width = uint16(img.Bounds().Max.X - img.Bounds().Min.X)
	pic.Height = uint16(img.Bounds().Max.Y - img.Bounds().Min.Y)
	pic.LeftOffset = 0
	pic.TopOffset = 0
	transparencyValue := uint8(0xff)
	posts := make(map[uint16][]byte)
	for i := uint16(0); i < pic.Height; i++ {
		startingColumn := uint8(0)
		var rowData, postData bytes.Buffer
		for j := uint16(0); j < pic.Width; j++ {
			pixValue := img.ColorIndexAt(int(j), int(i))
			if pixValue == transparencyValue && startingColumn == 0 {
				continue
			} else if pixValue == transparencyValue && startingColumn > 0 {
				// dump new post
				postLength := uint8(j) - uint8(startingColumn)

				// offset
				rowData.WriteByte(startingColumn)
				// length
				rowData.WriteByte(postLength)
				// unused
				rowData.WriteByte(0x00)
				// data
				_, err := rowData.Write(postData.Bytes())
				if err != nil {
					return nil, err
				}
				// unused
				rowData.WriteByte(0x00)

				postData.Reset()
				startingColumn = 0
			} else {
				postData.WriteByte(pixValue)
			}
		}

		if startingColumn > 0 {
			// dump remainder into new post
			postLength := uint8(pic.Width) - uint8(startingColumn)

			// offset
			rowData.WriteByte(startingColumn)
			// length
			rowData.WriteByte(postLength)
			// unused
			rowData.WriteByte(0x00)
			// data
			_, err := rowData.Write(postData.Bytes())
			if err != nil {
				return nil, err
			}
			// unused
			rowData.WriteByte(0x00)
		}

		posts[i] = rowData.Bytes()
		posts[i] = append(posts[i], 0xff)
	}

	err := binary.Write(&picData, binary.LittleEndian, pic)
	if err != nil {
		return nil, err
	}

	// write offsets followed by data
	rowOffset := uint32(binary.Size(pic))
	for i := uint16(0); i < pic.Height; i++ {
		if berr := binary.Write(&picData, binary.LittleEndian, rowOffset); berr != nil {
			return nil, berr
		}
		rowOffset += uint32(len(posts[i]))
	}
	for i := uint16(0); i < pic.Height; i++ {
		_, werr := picData.Write(posts[i])
		if werr != nil {
			return nil, werr
		}
	}

	return picData.Bytes(), nil
}

func scalePalettedImage(img *image.Paletted, invFactor int) *image.Paletted {
	width := img.Bounds().Dx() / invFactor
	height := img.Bounds().Dy() / invFactor
	timg := resize.Resize(uint(width), uint(height), img, resize.Lanczos3)
	// TODO: this is disgusting. We go from Paletted to RGBA back to
	// Paletted. Find a simpler, more efficient way.
	simg := image.NewRGBA(image.Rect(0, 0, width, height))
	pimg := image.NewPaletted(image.Rect(0, 0, width, height), imgutil.RottPalette)
	draw.Draw(simg, simg.Bounds(), timg, timg.Bounds().Min, draw.Src)

	// quantize image to Quake palette on the fly
	for i := 0; i < width; i++ {
		for j := 0; j < height; j++ {
			pimg.SetColorIndex(i, j, uint8(imgutil.QuakePalette.Index(simg.At(i, j))))
		}
	}
	return pimg
}

func PalettedImageToMIPTexture(img *image.Paletted) ([]byte, error) {
	// convert a peletted image to MIP texture data by scaling the image
	// 1/2x, 1/4x, and 1/8x
	var mip MIPTexture
	// name + width + height + 1pos + 2pos + 4pos + 8pos
	headerSize := int32(40)
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()
	resizedHalf := scalePalettedImage(img, 2)
	resizedFourth := scalePalettedImage(img, 4)
	resizedEighth := scalePalettedImage(img, 8)
	img.Palette = imgutil.RottPalette
	resizedHalf.Palette = imgutil.RottPalette
	resizedFourth.Palette = imgutil.RottPalette
	resizedEighth.Palette = imgutil.RottPalette

	imgutil.TranslateRottPalettedImageToQuake(img)

	mip.Width = int32(width)
	mip.Height = int32(height)
	mip.Scale1Pos = headerSize
	mip.Scale2Pos = headerSize + int32(len(img.Pix))
	mip.Scale4Pos = headerSize + int32(len(img.Pix)+len(resizedHalf.Pix))
	mip.Scale8Pos = headerSize + int32(len(img.Pix)+len(resizedHalf.Pix)+len(resizedFourth.Pix))

	b := new(bytes.Buffer)
	if err := binary.Write(b, binary.LittleEndian, &mip); err != nil {
		return nil, err
	}
	b.Write(img.Pix)
	b.Write(resizedHalf.Pix)
	b.Write(resizedFourth.Pix)
	b.Write(resizedEighth.Pix)
	return b.Bytes(), nil
}
