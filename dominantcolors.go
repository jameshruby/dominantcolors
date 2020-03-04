package main

/*
EXECUTION TIMES DOESNT MATTER AS MUCH AS OUR ABILLITY TO USE 
RESOURCES 
IT CAN TAKE SECONDS. THERES NO POINT IN SPENDIND AGES ON OBSSESING WITH SPEED
HOW TO MAKE PIPELINE MORE EFICIENT ?
WE STILL HAVE CPU SPIKES....SO THE GOAL SHOULD REALLY BE JUST TO GET RID OF THEM
*/

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

const BUFFER_SIZE = 50

func DominantColorsFromURLToCSV(urlListFile string, csvFilename string) {
	chImgInfo := DownloadAllImages(urlListFile)
	////get all colors
	st := DominantColorsFromRGBAImage(chImgInfo)
	saveEverythingToCSV(st)
}

//TODO not sure which approach will work better with goroutines/ structures vs channel/slice merge
func saveEverythingToCSV(st <-chan [4]string)  {
	//create CSV file,
	csvFilename := "huhu.csv"
	outputCSV, err := os.Create(csvFilename)
	HandleError(err, "failed creating CSV file")
	writerCSV := csv.NewWriter(outputCSV)
	//TODO can this run concurrently too ?
	// go func() {
		for line := range st {
			err = writerCSV.Write(line[:])
			// fmt.Printf("%v %v %v %v \n", line[0], line[1], line[2], line[3])
			// HandleError(err, "CSV writer failed")
		}
		writerCSV.Flush()
		outputCSV.Close()
	// }()
	// HandleError(err, "failed to close CSV file")
}

func DominantColorsFromRGBAImage(chImgInfo <-chan imageInfo) <-chan [4]string { 
	out := make(chan [4]string, 10)  //BUFFER_SIZE
	go func ()  {
		for imgInfo := range chImgInfo {
			// fmt.Println("-- opening the image " + imgInfo.filename)
			image, Dx, Dy, err := GetRGBAImage(imgInfo.filename)
			// fmt.Println("-- opening the image DONE")
			HandleError(err, "failed to process image "+imgInfo.filename)
			// fmt.Println("-- dominant colors " + imgInfo.filename)
			colorA, colorB, colorC, err := DominantColors(image, Dx, Dy)
			// fmt.Println("-- dominant colors DONE")
			HandleError(err, "")
			out <- [4]string{imgInfo.link, ColorToRGBHexString(colorA), ColorToRGBHexString(colorB), ColorToRGBHexString(colorC)}
		}
		close(out)
	}()
	return out

	//remove temp file
	// err = os.Remove(filename)
	// HandleError(err, "")
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

type imageInfo struct {
	filename string
	link string
}
func DownloadAllImagesStub(linksFile string) (<-chan imageInfo) {
	chImgInfo := make(chan imageInfo)
	linksScanner, fileHandle, _ := OpenTheList(linksFile)
	// HandleError(err, "couldn't open the links list")
	go func() {
		for linksScanner.Scan() {
			line := linksScanner.Text()
			chImgInfo <- imageInfo {line,  "LINK_"+line}
		}
		close(chImgInfo)
		fileHandle.Close()
	}()
	return chImgInfo
}
func DownloadAllImages(linksFile string) (<-chan imageInfo) {
	linksScanner, fileHandle, _ := OpenTheList(linksFile)
	chImgInfo := make(chan imageInfo, BUFFER_SIZE)
	// HandleError(err, "couldn't open the links list")
	// defer fileHandle.Close()
	// HandleError(err, "failed creating CSV file")

	go func() {	
		for linksScanner.Scan() {
			url := linksScanner.Text()
			filename, err := DownloadImage(url)
			HandleError(err, "failed to download tjhe file")
			chImgInfo <- imageInfo {filename, url}
		}
		close(chImgInfo)
		fileHandle.Close()
	}()
	return chImgInfo
}
func DownloadImage(url string) (string, error) {
	// fmt.Println("-- downloading " + url)
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
	// fmt.Println("-- downloading FINISHED")
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
