package main

import (
	"image"
	"image/color"
	"testing"
)

//Using testing colors that shouldn't be broken by conversion
var dominantColors = map[string]color.RGBA{
	"yellow": color.RGBA{203, 204, 102, 0xff},
	"blue":   color.RGBA{51, 204, 255, 0xff},
	"red":    color.RGBA{200, 63, 105, 0xff},
}
var secondaryColors = map[string]color.RGBA{
	"green": color.RGBA{87, 198, 43, 0xff},
}

/*
	this test uses already made image to avoid differences between
	RGB and YCbCr

	also see https://stackoverflow.com/questions/47550838/unexpected-inaccurate-image-color-conversions-in-go

	Original code:
	imgWidth := 12
	imgHeight := 12
	testImage := generateTestImage(imgWidth, imgHeight)
	file, _ := os.Create(filename)
	jpeg.Encode(file, testImage, nil)

	And in dominantcolors
	b := imageData.Bounds()
	rgbImage := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(rgbImage, rgbImage.Bounds(), imageData, b.Min, draw.Src)
*/

func TestJpegColorOutput(t *testing.T) {

	filename := "./testData/test.jpg"
	colorA, colorB, colorC := DominantColorsFromJpeg(filename)
	colors := []color.Color{colorA, colorB, colorC}
	expectedColors := []color.Color{dominantColors["yellow"], dominantColors["blue"], dominantColors["red"]}

	r, g, b, _ := colors[0].RGBA()
	//ra, ga, ba := color.YCbCrToRGB(uint8(r/0x101), uint8(g/0x101), uint8(b/0x101))
	t.Logf("%d, %d, %d", r, g, b)
	t.Logf("%d, %d, %d", r/0x101, g/0x101, b/0x101)

	for i := 0; i < len(expectedColors); i++ {
		if colors[i] != expectedColors[i] {
			t.Errorf("Dominant colors are wrong, actual: %v, expected: %v.", colors[i], expectedColors[i])
		}
	}
}

func generateTestImage(width int, height int) image.Image {
	testImage := image.NewRGBA(image.Rect(0, 0, width, height))
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			switch {
			case x < 3:
				testImage.Set(x, y, dominantColors["blue"])
			case x == 3:
				testImage.Set(x, y, color.Black)
			case x >= 4 && x < 8:
				testImage.Set(x, y, dominantColors["yellow"])
			case x == 8:
				testImage.Set(x, y, color.White)
			case x >= 9 && x < 11:
				testImage.Set(x, y, dominantColors["red"])
			case x == 11:
				testImage.Set(x, y, secondaryColors["green"])
			}
		}
	}
	return testImage
}

func TestCorrectColorOutput(t *testing.T) {
	imgWidth := 12
	imgHeight := 12
	testImage := generateTestImage(imgWidth, imgHeight)
	colorA, colorB, colorC := DominantColors(testImage, imgWidth, imgHeight)
	colors := []color.Color{colorA, colorB, colorC}
	expectedColors := []color.Color{dominantColors["yellow"], dominantColors["blue"], dominantColors["red"]}

	for i := 0; i < len(dominantColors); i++ {
		if colors[i] != expectedColors[i] {
			t.Errorf("Dominant colors are wrong, actual: %v, expected: %v.", colors[i], expectedColors[i])
		}
	}
}

func TestNullPicture(t *testing.T) {
	imgWidth := 0
	imgHeight := 0
	testImage := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))

	colorA, colorB, colorC := DominantColors(testImage, imgWidth, imgHeight)
	colors := []color.Color{colorA, colorB, colorC}
	expectedColors := []color.Color{nil, nil, nil}

	for i := 0; i < len(expectedColors); i++ {
		if colors[i] != expectedColors[i] {
			t.Errorf("Dominant colors are wrong, actual: %v, expected: %v.", colors[i], expectedColors[i])
		}
	}
}

func TestSolidPictureShouldReturnSameColor(t *testing.T) {
	imgWidth := 12
	imgHeight := 12
	testImage := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))

	for x := 0; x < imgWidth; x++ {
		for y := 0; y < imgHeight; y++ {
			testImage.Set(x, y, dominantColors["orange"])
		}
	}

	colorA, colorB, colorC := DominantColors(testImage, imgWidth, imgHeight)
	colors := []color.Color{colorA, colorB, colorC}
	expectedColors := []color.Color{dominantColors["orange"], dominantColors["orange"], dominantColors["orange"]}

	for i := 0; i < len(expectedColors); i++ {
		if colors[i] != expectedColors[i] {
			t.Errorf("Dominant color is wrong, actual: %v, expected: %v.", colors[i], expectedColors[i])
		}
	}
}
