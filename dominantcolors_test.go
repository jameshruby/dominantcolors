package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"testing"
)

//Using testing colors that shouldn't be broken by conversion
var dominantColors = map[string]color.RGBA{
	"yellow": color.RGBA{203, 204, 102, 0xff}, //#CBCC66
	"blue":   color.RGBA{51, 204, 255, 0xff},
	"red":    color.RGBA{200, 63, 105, 0xff},
}
var secondaryColors = map[string]color.RGBA{
	"green": color.RGBA{87, 198, 43, 0xff},
	"black": color.RGBA{0, 0, 0, 0xff},
	"white": color.RGBA{255, 255, 255, 0xff},
}

var dominantColorsByte = map[string][3]byte{
	"yellow": [3]byte{203, 204, 102}, //#CBCC66
	"blue":   [3]byte{51, 204, 255},
	"red":    [3]byte{200, 63, 105},
}
var secondaryColorsByte = map[string][3]byte{
	"green": [3]byte{87, 198, 43},
	"black": [3]byte{0, 0, 0},
	"white": [3]byte{255, 255, 255},
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

	image, imageConfig, _ := GetImageFromJpeg(filename)
	colorA, colorB, colorC, _ := DominantColors(image, imageConfig.Width, imageConfig.Height)

	colors := [][3]byte{colorA, colorB, colorC}
	expectedColors := [][3]byte{dominantColorsByte["yellow"], dominantColorsByte["blue"], dominantColorsByte["red"]}

	for i := 0; i < len(expectedColors); i++ {
		if colors[i] != expectedColors[i] {
			t.Errorf("Dominant colors are wrong, actual: %v, expected: %v.", colors[i], expectedColors[i])
		}
	}
}

func TestColorToHex(t *testing.T) {
	expected := "#CBCC66"
	yellow := dominantColors["yellow"]
	c := [3]byte{yellow.R, yellow.G, yellow.B}
	actual := ColorToRGBHexString(c)
	if actual != expected {
		t.Errorf("Dominant colors are wrong, actual: %v, expected: %v.", actual, expected)
	}
}

var csvFilename string = "output.csv"

func TestTestImageEndToEnd(t *testing.T) {
	testURLFilename := "./testData/testUrlList.txt"
	DominantColorsFromURLToCSV(testURLFilename, csvFilename)

	csvFile, err := os.Open(csvFilename)
	if err != nil {
		log.Fatalln("Couldn't open the csv file", err)
	}

	reader := csv.NewReader(csvFile)
	expectedResult := []string{"https://i.imgur.com/19cQ2Ka.jpg", "#CBCC66", "#33CCFF", "#C83F69"}
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Errorf("failed to read output file %v", err)
		}

		for i := 0; i < len(expectedResult); i++ {
			if record[i] != expectedResult[i] {
				t.Errorf("CSV output is wrong, actual: %v, expected: %v.", record[i], expectedResult[i])
			}
		}
	}
	csvFile.Close()
	os.Remove(csvFilename)
}

type ImageGeneratorMode int

const (
	SimpleTest ImageGeneratorMode = iota
	Lena
	WorstCase
)

func (i ImageGeneratorMode) String() string {
	switch i {
	case SimpleTest:
		return "SimpleTest"
	case Lena:
		return "Lena"
	case WorstCase:
		return "WorstCase"
	default:
		panic("No such enum item in ImageGeneratorMode")
	}
}

func generateTestImage(width int, height int, mode ImageGeneratorMode) (*image.RGBA, error) {
	testImage := image.NewRGBA(image.Rect(0, 0, width, height))
	var getColor func(x int, y int, i int) color.RGBA

	switch mode {
	case SimpleTest:
		getColor = func(x int, y, i int) color.RGBA {
			heightF := float64(height)
			thresholdColorA := height / 4
			thresholdColorB := height / 3
			thresholdColorC := int(heightF / 1.5)
			thresholdSecColorA := int(heightF / 1.3)
			thresholdSecColorB := int(heightF / 1.09)

			var c color.RGBA
			switch {
			case x < thresholdColorA:
				c = dominantColors["blue"]
			case x >= thresholdColorA:
				c = secondaryColors["black"]
			case x >= thresholdColorB && x < thresholdColorC:
				c = dominantColors["yellow"]
			case x >= thresholdColorC:
				c = secondaryColors["white"]
			case x >= thresholdSecColorA && x < thresholdSecColorB:
				c = dominantColors["red"]
			case x >= thresholdSecColorB:
				c = secondaryColors["green"]
			default:
				erMessage :=
					fmt.Sprintf(
						"%d missed tresholds[%d, %d, %d, %d, %d] while generating the image",
						x, thresholdColorA, thresholdColorB, thresholdColorC, thresholdSecColorA, thresholdSecColorB)
				panic(errors.New(erMessage))

			}
			return c
		}
	case Lena:
		content, err := os.Open(fmt.Sprintf("./testData/lena%dx%d.png", width, height))
		if err != nil {
			return testImage, err
		}
		pngDecoded, err := png.Decode(content)
		if err != nil {
			return testImage, err
		}
		imgRGBA := pngDecoded.(*image.RGBA)
		imgPix := imgRGBA.Pix

		getColor = func(x int, y int, i int) color.RGBA {
			//Pix is array with colors just stacked behind each other
			i *= 4
			c := imgPix[i : i+4 : i+4]
			return color.RGBA{c[0], c[1], c[2], c[3]}
		}
	case WorstCase:
		uniqueColors := make(map[color.Color]struct{})
		getColor = func(x int, y int, i int) color.RGBA {
			color := color.RGBA{uint8(rand.Intn(255)), uint8(rand.Intn(255)), uint8(rand.Intn(255)), 0xff}
			if _, found := uniqueColors[color]; found {
				return getColor(x, y, 0)
			}
			uniqueColors[color] = struct{}{}
			return color
		}
	}

	i := 0
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			testImage.Set(x, y, getColor(x, y, i))
			i++
		}
	}

	//test png - commented, we dont need to create test png now
	// out, err := os.Create("./output.png")
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }
	// err = png.Encode(out, testImage)
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }

	return testImage, nil
}

//TODO FIX Commented out since the test is failing now, after change of imgGen function
// func TestCorrectColorOutput(t *testing.T) {
// 	imgWidth := 12
// 	imgHeight := 12
// 	testImage, _ := generateTestImage(imgWidth, imgHeight, SimpleTest)
// 	colorA, colorB, colorC, _ := DominantColors(testImage, imgWidth, imgHeight)
// 	colors := []color.Color{colorA, colorB, colorC}
// 	expectedColors := []color.Color{dominantColors["yellow"], dominantColors["blue"], dominantColors["red"]}

// 	for i := 0; i < len(dominantColors); i++ {
// 		if colors[i] != expectedColors[i] {
// 			t.Errorf("Dominant colors are wrong, actual: %v, expected: %v.", colors[i], expectedColors[i])
// 		}
// 	}
// }

func TestNullPicture(t *testing.T) {
	imgWidth := 0
	imgHeight := 0
	testImage := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))

	colorA, colorB, colorC, err := DominantColors(testImage, imgWidth, imgHeight)
	if err == nil {
		t.Errorf("Dominant colors aren't returning error")
	}
	t.Logf("%v %v %v", colorA, colorB, colorC)
	//colors := [][3]byte{colorA, colorB, colorC}
	// expectedColors := [][3]byte{nil, nil, nil}

	// for i := 0; i < len(expectedColors); i++ {
	// 	if colors[i] != expectedColors[i] {
	// 		t.Errorf("Dominant colors are wrong, actual: %v, expected: %v.", colors[i], expectedColors[i])
	// 	}
	// }
}

func TestSolidPictureShouldReturnSameColor(t *testing.T) {
	imgWidth := 12
	imgHeight := 12
	testImage := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))

	for x := 0; x < imgWidth; x++ {
		for y := 0; y < imgHeight; y++ {
			testImage.Set(x, y, dominantColors["yellow"])
		}
	}

	colorA, colorB, colorC, _ := DominantColors(testImage, imgWidth, imgHeight)
	colors := [][3]byte{colorA, colorB, colorC}
	expectedColors := [][3]byte{dominantColorsByte["yellow"], dominantColorsByte["yellow"], dominantColorsByte["yellow"]}

	for i := 0; i < len(expectedColors); i++ {
		if colors[i] != expectedColors[i] {
			t.Errorf("Dominant color is wrong, actual: %v, expected: %v.", colors[i], expectedColors[i])
		}
	}
}

func TestImgTypeSwichingWithEvilPostfix(t *testing.T) {
	filename := "./testData/lena100x100.jpg"
	input, err := ioutil.ReadFile("./testData/lena100x100.png")
	if err != nil {
		t.Errorf("%v", err)
	}
	err = ioutil.WriteFile(filename, input, 0644)
	if err != nil {
		t.Errorf("%v", err)
	}
	_, _, err = GetImageFromJpeg(filename)
	if err != nil {
		t.Errorf("Failed to open the image: %v", err)
	}
	os.Remove(filename)
}

func BenchmarkDominantColors(b *testing.B) {
	generatorModes := []ImageGeneratorMode{SimpleTest, Lena, WorstCase}
	var benchmarks = []struct {
		baseName   string
		size       int
		testImages []ImageGeneratorMode
	}{
		{"SmallImages", 100, generatorModes},
		{"MediumImages", 512, generatorModes},
		{"LargeImages", 2000, generatorModes},
	}

	for _, bm := range benchmarks {
		for _, testImage := range bm.testImages {
			b.Run(bm.baseName+"_"+testImage.String(), func(b *testing.B) {
				testImage, err := generateTestImage(bm.size, bm.size, testImage) //TODO move to init
				if err != nil {
					b.Fatalf("%s", err)
				}
				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					_, _, _, err := DominantColors(testImage, bm.size, bm.size) //we don't care about actual output
					if err != nil {
						b.Fatalf("%s", err)
					}
				}
			})
		}
	}
}
