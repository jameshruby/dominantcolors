package main

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"path"
	"sort"
)

func DominantColorsFromURLToCSV(urlListFile string, csvFilename string) {
	linksScanner, fileHandle, err := OpenTheList(urlListFile)
	HandleError(err, "couldn't open the links list")
	defer fileHandle.Close()
	//create CSV file
	outputCSV, err := os.Create(csvFilename)
	HandleError(err, "failed creating CSV file")
	writerCSV := csv.NewWriter(outputCSV)

	for linksScanner.Scan() {
		url := linksScanner.Text()
		filename, err := DownloadImage(url)
		HandleError(err, "failed to download the file")

		image, imageConfig, err := GetImageFromJpeg(filename)
		HandleError(err, "failed to process image")
		colorA, colorB, colorC, err := DominantColors(image, imageConfig.Width, imageConfig.Height)
		HandleError(err, "")

		os.Remove(filename)
		err = writerCSV.Write([]string{url, ColorToRGBHexString(colorA), ColorToRGBHexString(colorB), ColorToRGBHexString(colorC)})
		HandleError(err, "CSV writer failed")
	}
	writerCSV.Flush()
	outputCSV.Close()
	HandleError(err, "failed to close CSV file")
}

func HandleError(err error, extendedMessage string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v: %v\n", extendedMessage, err)
		os.Exit(1)
	}
}

func OpenTheList(urlListFile string) (*bufio.Scanner, io.Closer, error) {
	file, err := os.Open(urlListFile)
	if err != nil {
		return nil, nil, err
	}
	scanner := bufio.NewScanner(file)
	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}
	//we need to return file handle, since we need to close it afterwards
	return scanner, file, nil
}

func ColorToRGBHexString(color color.Color) string {
	r, g, b, _ := color.RGBA()
	ra, ga, ba := uint8(r/0x101), uint8(g/0x101), uint8(b/0x101)
	return fmt.Sprintf("#%X%X%X", ra, ga, ba)
}

func DownloadImage(url string) (string, error) {
	response, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	// TODO use path.Base and delete the image when we are done
	//open a file for writing
	filename := path.Base(url)
	file, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		return "", err
	}
	// fmt.Println("image downloaded")
	return filename, nil
}

func GetImageFromJpeg(imagefilename string) (image.Image, image.Config, error) {
	image.RegisterFormat("jpg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
	var imageConfig image.Config
	var imageData image.Image
	testImage, err := os.Open(imagefilename)
	if err != nil {
		return imageData, imageConfig, err
	}
	defer testImage.Close()

	imageConfig, _, err = image.DecodeConfig(testImage)
	if err != nil {
		return imageData, imageConfig, fmt.Errorf("Error: Image config failed %v", err)
	}
	testImage.Seek(0, 0)
	imageData, _, err = image.Decode(testImage)
	if err != nil {
		return imageData, imageConfig, err
	}
	return imageData, imageConfig, nil
}

func DominantColors(image image.Image, width int, height int) (color.Color, color.Color, color.Color, error) {
	if width == 0 || height == 0 {
		var nilColor color.Color
		return nilColor, nilColor, nilColor, errors.New("image size was 0")
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
		return colorCounterList[0].Key, colorCounterList[0].Key, colorCounterList[0].Key, nil
	case listLen < 3:
		return colorCounterList[0].Key, colorCounterList[1].Key, colorCounterList[1].Key, nil
	default:
		return colorCounterList[0].Key, colorCounterList[1].Key, colorCounterList[2].Key, nil
	}
}
