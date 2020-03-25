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
// const HEX16 := 0xFF
func RGBToInt(color [3]byte) int {
	r, g, b := int(color[0]), int(color[1]), int(color[2])
	// rgb := r
	// rgb = (rgb << 8) + g
	// rgb = (rgb << 8) + b

	rgb := ( (r & 0xFF) << 16) | ( (g & 0xFF ) << 8 ) | (b & 0xFF)
	return rgb
}
func IntToRGB(rgb int) [3]byte {
	r := (rgb >> 16) & 0xFF;
	g := (rgb >> 8) & 0xFF;
	b := rgb & 0xFF;
	return [3]byte{byte(r), byte(g), byte(b)}
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
	if width == 0 || height == 0 {
		var ccA, ccB, ccC [rgbLen]byte
		return ccA, ccB, ccC, errors.New("image size was 0")
	}

	clusters := KmeansPartition(image.Pix)
	// fmt.Println("", clusters)
	// const clusterCount = 16
	// clusters := KmeansPartition(image.Pix, clusterCount)
	// sort.Slice(clusters, func(i, j int) bool {
	// 	return len(clusters[i].Points) > len(clusters[j].Points)
	// })

	// if bCount == 0 {
	// 	cB = cA
	// }
	// if cCount == 0 {
	// 	cC = cA
	// }
	// var ccA, ccB, ccC [rgbLen]byte
	// return ccA, ccB, ccC, nil
	// return clusters[0].Points[0], clusters[1].Points[0], clusters[2].Points[0], nil

	var ccA, ccB, ccC = clusters[0], clusters[1], clusters[2]

	return [rgbLen]byte{byte(ccA[0]), byte(ccA[1]), byte(ccA[2])}, [rgbLen]byte{byte(ccB[0]), byte(ccB[1]), byte(ccB[2])}, [rgbLen]byte{byte(ccC[0]), byte(ccC[1]), byte(ccC[2])}, nil
}
