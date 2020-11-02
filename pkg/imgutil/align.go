package imgutil

import (
	"image"
)

// resize image to assure the dimensions are a factor of `alignment`
// drawing is preserved in the lower right of the image (dimensions are
// "stretched" on the left and top bounds)
func AlignImageDimensions(img *image.RGBA, alignment int) *image.RGBA {
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

	dimg := image.NewRGBA(image.Rect(0, 0, adjustedWidth, adjustedHeight))
	for i := drawStartX; i < adjustedWidth; i++ {
		for j := drawStartY; j < adjustedHeight; j++ {
			dimg.Set(i, j, img.At(i-drawStartX, j-drawStartY))
		}
	}
	return dimg
}
