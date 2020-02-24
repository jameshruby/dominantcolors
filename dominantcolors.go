package main

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"path"
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
		HandleError(err, "")
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

const rgbLen = 3

func ColorToRGBHexString(color [rgbLen]byte) string {
	return fmt.Sprintf("#%X%X%X", color[0], color[1], color[2])
}

func DownloadImage(url string) (string, error) {
	response, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

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

func GetImageFromJpeg(imagefilename string) (*image.RGBA, image.Config, error) {
	image.RegisterFormat("jpg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
	var imageConfig image.Config
	var rgbImage *image.RGBA
	testImage, err := os.Open(imagefilename)
	if err != nil {
		return rgbImage, imageConfig, err
	}
	defer testImage.Close()

	imageConfig, _, err = image.DecodeConfig(testImage)
	if err != nil {
		return rgbImage, imageConfig, fmt.Errorf("Error: Image config failed %v", err)
	}
	testImage.Seek(0, 0)
	imageData, _, err := image.Decode(testImage)
	if err != nil {
		return rgbImage, imageConfig, err
	}

	//make RGBA out of it
	imageBounds := imageData.Bounds()
	rgbImage = image.NewRGBA(image.Rect(0, 0, imageBounds.Dx(), imageBounds.Dy()))
	draw.Draw(rgbImage, rgbImage.Bounds(), imageData, imageBounds.Min, draw.Src)
	return rgbImage, imageConfig, nil
}

func DominantColors(image *image.RGBA, width int, height int) ([rgbLen]byte, [rgbLen]byte, [rgbLen]byte, error) {
	var cA, cB, cC [rgbLen]byte
	if width == 0 || height == 0 {
		return cA, cB, cC, errors.New("image size was 0")
	}
	//build a map of unique colors and its sum, pix is array with colors just stacked behind each other
	imgPix := image.Pix
	uniqueColors := make(map[[rgbLen]byte]int)
	var aCount, bCount, cCount int
	const rgbaLen = 4
	for i := 0; i < len(imgPix); i += rgbaLen {
		var pixel [rgbLen]byte
		copy(pixel[:], imgPix[i:i+rgbLen:i+rgbLen]) //getting RGBA [125][126][243][255] [100][2][56][255]
		colorOccurences := uniqueColors[pixel] + 1
		switch {
		case colorOccurences > aCount:
			aCount = colorOccurences
			cA = pixel
		case colorOccurences > bCount:
			bCount = colorOccurences
			cB = pixel
		case colorOccurences > cCount:
			cCount = colorOccurences
			cC = pixel
		}
		uniqueColors[pixel] = colorOccurences
	}
	//guard for less colorfull images
	if bCount == 0 {
		cB = cA
	}
	if cCount == 0 {
		cC = cA
	}
	return cA, cB, cC, nil
}
