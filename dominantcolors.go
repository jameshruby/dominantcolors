package main

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"image"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"
)


func DominantColorsFromURLToCSV(urlListFile string, csvFilename string) {
	filenames, links := DownloadAllImages(urlListFile)
	//get all colors
	var allColors [][3]string
	
	for _, filename := range filenames {
		image, Dx, Dy, err := GetRGBAImage(filename)
		HandleError(err, "failed to process image "+filename)
		colorA, colorB, colorC, err := DominantColors(image, Dx, Dy)
		HandleError(err, "")
		//remove temp file
		// err = os.Remove(filename)
		// HandleError(err, "")
		str := [3]string{ColorToRGBHexString(colorA), ColorToRGBHexString(colorB), ColorToRGBHexString(colorC)}
		allColors = append(allColors, str)
	}

	//create CSV file
	outputCSV, err := os.Create(csvFilename)
	HandleError(err, "failed creating CSV file")
	writerCSV := csv.NewWriter(outputCSV)
	//TODO not sure which approach will work better with goroutines/ structures vs channel/slice merge
	for i := 0; i < len(links); i++ {
		err = writerCSV.Write([]string{links[i], allColors[i][0], allColors[i][1], allColors[i][2] })
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


func DownloadAllImagesStub(linksFile string) ([]string, []string) {
	filenames := []string{}
	links := []string{}
	linksScanner, fileHandle, err := OpenTheList(linksFile)
	HandleError(err, "couldn't open the links list")
	for linksScanner.Scan() {
		filenames = append(filenames, linksScanner.Text())
		links = append(links, linksScanner.Text())
	}
	fileHandle.Close()
	return filenames, links
}
func DownloadAllImages(linksFile string) ([]string, []string) {
	linksScanner, fileHandle, err := OpenTheList(linksFile)
	links := []string{}
	HandleError(err, "couldn't open the links list")
	defer fileHandle.Close()
	HandleError(err, "failed creating CSV file")

	var filenames []string
	for linksScanner.Scan() {
		url := linksScanner.Text()
		filename, err := DownloadImage(url)
		HandleError(err, "failed to download tjhe file")
		filenames = append(filenames, filename)
		links = append(links, url)
	}
	return filenames, links
}
func DownloadImage(url string) (string, error) {
	response, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	timePrefix := strconv.FormatInt(int64(time.Now().UnixNano()), 10)
	filename := timePrefix + "_" + path.Base(url)
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

func GetRGBAImage(imagefilename string) (img *image.RGBA, Dx int, Dy int, err error) {
	var rgbImage *image.RGBA
	testImage, err := os.Open(imagefilename)
	if err != nil {
		return rgbImage, 0, 0, err
	}
	defer testImage.Close()
	if err != nil {
		return rgbImage, 0, 0, err
	}
	_, err = testImage.Seek(0, 0) //TODO is this really needed ?
	if err != nil {
		return rgbImage, 0, 0, err
	}
	imageData, _, err := image.Decode(testImage)
	if err != nil {
		return rgbImage, 0, 0, err
	}
	//make RGBA out of it
	imageBounds := imageData.Bounds()
	rgbImage = image.NewRGBA(image.Rect(0, 0, imageBounds.Dx(), imageBounds.Dy()))
	draw.Draw(rgbImage, rgbImage.Bounds(), imageData, imageBounds.Min, draw.Src)
	return rgbImage, imageBounds.Dx(), imageBounds.Dy(), nil
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
