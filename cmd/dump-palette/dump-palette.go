package main

// dumps a 768-byte palette file to a PNG

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
)

const STRIDE = 3 // R G B
const PALETTE_SIZE = 256 * STRIDE
const IMG_WIDTH = 16
const IMG_HEIGHT = 16

func createPalettedImage(data [PALETTE_SIZE]byte, scale int) *image.Paletted {
	var palette color.Palette
	for i := 0; i < PALETTE_SIZE; i += 3 {
		c := color.RGBA{data[i], data[i+1], data[i+2], 255}
		palette = append(palette, c)
	}
	img := image.NewPaletted(image.Rect(0, 0, IMG_WIDTH*scale, IMG_HEIGHT*scale), palette)
	for i, _ := range palette {
		x := i % IMG_WIDTH
		y := i / IMG_WIDTH
		for j := x * scale; j < (x+1)*scale; j++ {
			for k := y * scale; k < (y+1)*scale; k++ {
				img.SetColorIndex(j, k, uint8(i))
			}
		}
	}
	return img
}

func init() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s <pal file> <png file>\n", os.Args[0])
		flag.PrintDefaults()
	}
}

func main() {
	var scale int
	var data [PALETTE_SIZE]byte

	flag.IntVar(&scale, "scale", 2, "image scale")
	flag.Parse()

	if flag.NArg() < 2 {
		flag.Usage()
		os.Exit(2)
	}
	palFile, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatalf("Could not open pal file: %v\n", err)
	}
	defer palFile.Close()
	outFile, err := os.Create(flag.Arg(1))
	if err != nil {
		log.Fatalf("Could not open output file: %v\n", err)
	}
	defer outFile.Close()
	numRead, err := palFile.Read(data[:])
	if err != nil {
		log.Fatalf("Could not read pal file: %v\n", err)
	}
	if numRead != PALETTE_SIZE {
		log.Fatalf("Read palette size is not %d bytes\n", PALETTE_SIZE)
	}
	img := createPalettedImage(data, scale)
	if encodeErr := png.Encode(outFile, img); encodeErr != nil {
		log.Fatalf("Could not encode png: %v\n", encodeErr)
	}
}
