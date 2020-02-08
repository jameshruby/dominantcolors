package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"sort"
)

func DominantColorsFromURLToCSV(urlListFile string, csvFilename string) {
	linksScanner := OpenTheList(urlListFile)
	//create CSV file
	outputCSV, err := os.Create(csvFilename)
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}
	writerCSV := csv.NewWriter(outputCSV)

	for linksScanner.Scan() {
		url := linksScanner.Text()
		filename := DownloadImage(url)
		image, imageConfig := GetImageFromJpeg(filename)
		colorA, colorB, colorC := DominantColors(image, imageConfig.Width, imageConfig.Height)
		os.Remove(filename)
		err = writerCSV.Write([]string{url, ColorToRGBHexString(colorA), ColorToRGBHexString(colorB), ColorToRGBHexString(colorC)})
		if err != nil {
			fmt.Println(err)
		}
	}
	writerCSV.Flush()
	outputCSV.Close()
}

func OpenTheList(urlListFile string) *bufio.Scanner {
	//open the file
	file, err := os.Open(urlListFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return scanner
}

func ColorToRGBHexString(color color.Color) string {
	r, g, b, _ := color.RGBA()
	ra, ga, ba := uint8(r/0x101), uint8(g/0x101), uint8(b/0x101)
	return fmt.Sprintf("#%X%X%X", ra, ga, ba)
}

func DownloadImage(url string) string {
	response, e := http.Get(url)
	if e != nil {
		log.Fatal(e)
	}
	defer response.Body.Close()
	// TODO use path.Base and delete the image when we are done
	//open a file for writing
	filename := path.Base(url)
	file, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Println("image downloaded")
	return filename
}

func GetImageFromJpeg(imagefilename string) (image.Image, image.Config) {
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
	return imageData, imageConfig
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
