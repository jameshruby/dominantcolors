package main

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"sort"
)

func DominantColorsFromJpeg(imagefilename string) (color.Color, color.Color, color.Color) {
	image.RegisterFormat("jpg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
	testImage, err := os.Open(imagefilename)
	if err != nil {
		fmt.Println("Error: File could not be opened")
		os.Exit(1)
	}
	defer testImage.Close()

	imageConfig, _, err := image.DecodeConfig(testImage)
	if err != nil {
		fmt.Println("Error: Image config failed")
		os.Exit(1)
	}
	testImage.Seek(0, 0)
	imageData, _, err := image.Decode(testImage)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return DominantColors(imageData, imageConfig.Width, imageConfig.Height)
}

func DominantColors(image image.Image, width int, height int) (color.Color, color.Color, color.Color) {
	if width == 0 || height == 0 {
		fmt.Println("Warning: Image size was 0")
		return nil, nil, nil
	}
	//build a map of unique colors and its sum
	uniqueColors := make(map[color.Color]int)
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			color := color.RGBAModel.Convert(image.At(x, y))
			uniqueColors[color] = uniqueColors[color] + 1
		}
	}
	//make it into slice as map is not sortable
	//TOOD go sort with another array ? make two slices and just search array of slices
	type ColorCounter struct {
		Key   color.Color
		Value int
	}
	var colorCounterList []ColorCounter
	for color, colorCount := range uniqueColors { //TODO - or just compare the count and save it
		colorCounterList = append(colorCounterList, ColorCounter{color, colorCount})
	}
	sort.Slice(colorCounterList, func(i, j int) bool {
		return colorCounterList[i].Value > colorCounterList[j].Value
	})

	//guard for less colorfull images
	listLen := len(colorCounterList)
	switch {
	case listLen < 2:
		return colorCounterList[0].Key, colorCounterList[0].Key, colorCounterList[0].Key
	case listLen < 3:
		return colorCounterList[0].Key, colorCounterList[1].Key, colorCounterList[1].Key
	default:
		return colorCounterList[0].Key, colorCounterList[1].Key, colorCounterList[2].Key
	}
}
