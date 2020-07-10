package imgutil

import (
	"image"
)

// resize image to assure the dimensions are a factor of `alignment`
// drawing is preserved in the lower right of the image (dimensions are
// "stretched" on the left and top bounds)
func AlignImageDimensions(img *image.Paletted, alignment int) *image.Paletted {
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()
	if width%alignment == 0 && height%alignment == 0 {
		return img
	}

	var drawStartX, drawStartY int
	adjustedWidth := width
	adjustedHeight := height
	if width%alignment > 0 {
		adjustedWidth += alignment - (width % alignment)
		drawStartX = adjustedWidth - width
	}
	if height%alignment > 0 {
		adjustedHeight += alignment - (height % alignment)
		drawStartY = adjustedHeight - height
	}

	dimg := image.NewPaletted(image.Rect(0, 0, adjustedWidth, adjustedHeight), img.Palette)
	for i := drawStartX; i < adjustedWidth; i++ {
		for j := drawStartY; j < adjustedHeight; j++ {
			dimg.SetColorIndex(i, j, img.ColorIndexAt(i-drawStartX, j-drawStartY))
		}
	}
	return dimg
}
