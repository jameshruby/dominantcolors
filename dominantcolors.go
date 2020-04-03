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
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"sync"
	"time"
)

const BUFFER_SIZE = PROC_COUNT
const RGB_LEN = 3

func DominantColorsFromURLToCSV(urlListFile string, csvFilename string) {
	linksScanner, fileHandle, err := OpenTheList(urlListFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	chImgInfo := DownloadAllImages(linksScanner)
	st := DominantColorsFromRGBAImage(chImgInfo)
	err = saveEverythingToCSV(st, csvFilename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	fileHandle.Close()
}

type imageInfo struct {
	filename string
	link     string
	err      error
}

type processedImage struct {
	csvInfo [4]string
	err     error
}

func DownloadAllImages(linksScanner *bufio.Scanner) <-chan imageInfo {
	chImgInfo := make(chan imageInfo, BUFFER_SIZE)
	go func() {
		for linksScanner.Scan() {
			url := linksScanner.Text()
			filename, err := DownloadImage(url)
			if err != nil {
				err = fmt.Errorf("failed to download the image: %v", err)
			}
			chImgInfo <- imageInfo{filename, url, err}
		}
		close(chImgInfo)
	}()
	return chImgInfo
}

func DominantColorsFromRGBAImage(chImgInfo <-chan imageInfo) <-chan processedImage {
	out := make(chan processedImage, BUFFER_SIZE)
	go func() {
		for imgInfo := range chImgInfo { //TODO better goroutines handling
			o := processedImage{}
			if imgInfo.err != nil {
				o.err = imgInfo.err
				out <- o
				return
			}
			fmt.Println("-- opening the image " + imgInfo.filename)
			image, Dx, Dy, err := GetRGBAImage(imgInfo.filename)
			if err != nil {
				o.err = fmt.Errorf("failed to open the image[%s]: %v", imgInfo.filename, err)
				out <- o
				return
			}
			fmt.Println("-- opening the image DONE")
			fmt.Println("-- dominant colors " + imgInfo.filename)
			colorA, colorB, colorC, err := DominantColors(image, Dx, Dy)
			if err != nil {
				o.err = fmt.Errorf("failed to get dominantColors[%s]: %v", imgInfo.filename, err)
				out <- o
				return
			}
			// os.Remove(imgInfo.filename) //delete image file
			o.csvInfo = [4]string{imgInfo.link, ColorToRGBHexString(colorA), ColorToRGBHexString(colorB), ColorToRGBHexString(colorC)}
			out <- o
		}
		close(out)
	}()
	return out
}

func saveEverythingToCSV(st <-chan processedImage, csvFilename string) error {
	//create CSV file,
	outputCSV, err := os.Create(csvFilename)
	if err != nil {
		return fmt.Errorf("failed creating  CSV file: %v", err)
	}
	writerCSV := csv.NewWriter(outputCSV)
	for pi := range st {
		if pi.err != nil {
			return pi.err
		}
		err = writerCSV.Write(pi.csvInfo[:])
		if err != nil {
			return fmt.Errorf("CSV writer failed: %v", err)
		}
		fmt.Println("-- adding to CSV[", pi.csvInfo[0], "]")
	}

	writerCSV.Flush()
	err = outputCSV.Close()
	if err != nil {
		return err
	}
	return nil
}

func OpenTheList(urlListFile string) (*bufio.Scanner, *os.File, error) {
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

func ColorToRGBHexString(color [RGB_LEN]byte) string {
	return fmt.Sprintf("#%X%X%X", color[0], color[1], color[2])
}

func RGBToIntSlice(color []byte) int {
	r, g, b := int(color[0]), int(color[1]), int(color[2])
	rgb := ((r & 0xFF) << 16) | ((g & 0xFF) << 8) | (b & 0xFF)
	return rgb
}

func IntToRGB(rgb int) [3]byte {
	r := (rgb >> 16) & 0xFF
	g := (rgb >> 8) & 0xFF
	b := rgb & 0xFF
	return [3]byte{byte(r), byte(g), byte(b)}
}

func DownloadImage(url string) (string, error) {
	fmt.Println("-- downloading " + url)
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
	fmt.Println("-- downloading FINISHED")
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

func DominantColors(image *image.RGBA, width int, height int) ([RGB_LEN]byte, [RGB_LEN]byte, [RGB_LEN]byte, error) {
	if width == 0 || height == 0 {
		var ccA, ccB, ccC [RGB_LEN]byte
		return ccA, ccB, ccC, errors.New("image size was 0")
	}
	//build a map of unique colors and its sum, pix is array with colors just stacked behind each other
	imgPix := image.Pix

	const rgbaLen = 4

	lenImgPix := len(imgPix)
	partitionsCount := PROC_COUNT
	partitionsLen := int(math.RoundToEven((float64(lenImgPix/rgbaLen) / float64(partitionsCount)))) * rgbaLen

	countPartition := func(imgPix []uint8) <-chan map[int]int {
		uniqueColorsCh := make(chan map[int]int)
		go func(imgPix []uint8) {
			uniqueColors := make(map[int]int)
			for i := 0; i < len(imgPix); i += rgbaLen {
				pixel := RGBToIntSlice(imgPix[i : i+4 : i+4])
				uniqueColors[pixel] = uniqueColors[pixel] + 1
			}
			uniqueColorsCh <- uniqueColors
		}(imgPix)
		return uniqueColorsCh
	}
	mergeResults := func(partitionData [][]byte) <-chan map[int]int { //fanIn function
		out := make(chan map[int]int)
		var wg sync.WaitGroup
		wg.Add(len(partitionData))

		for _, p := range partitionData {
			go func(c <-chan map[int]int) {
				for v := range c {
					out <- v
				}
				wg.Done()
			}(countPartition(p))
		}
		go func() {
			wg.Wait()
			close(out)
		}()
		return out
	}

	partitionsData := make([][]byte, partitionsCount)
	begining := 0
	end := partitionsLen
	for i := 0; i < (partitionsCount); i++ {
		partitionsData[i] = imgPix[begining:end]
		begining = end
		end = end + partitionsLen
		if end > lenImgPix {
			end = lenImgPix
		}
	}
	//run the defs
	chn := mergeResults(partitionsData)
	res := make(map[int]int)

	var lastCount, lastColor int
	for i := 0; i < partitionsCount; i++ {
		for lastColor, lastCount = range <-chn {
			res[lastColor] = res[lastColor] + lastCount
		}
	}

	clen := len(res)
	rcolors := make([]int, 3)
	if clen >= 3 {
		// starta := time.Now()
		var cA, cB, cC int
		var aCount, bCount, cCount int
		rcounts := []int{cA, cB, cC}
		rcolors = []int{aCount, bCount, cCount}

		for pixel, colorOccurences := range res {
			for i := 0; i < 3; i++ {
				if colorOccurences > rcounts[i] {
					nextI := i + 1
					if nextI < 3 {
						nextNextI := nextI + 1
						if nextNextI < 3 {
							rcounts[nextNextI] = rcounts[nextI]
							rcolors[nextNextI] = rcolors[nextI]
						}
						rcounts[nextI] = rcounts[i]
						rcolors[nextI] = rcolors[i]
					}
					rcounts[i] = colorOccurences
					rcolors[i] = pixel
					break
				}
			}
		}
		// fmt.Println("A ", time.Since(starta))
	}
	if clen == 2 {
		rcolors[0] = lastColor
		for k, v := range res {
			if v > lastCount {
				rcolors[0], rcolors[1] = k, rcolors[0]
			}
			if v < lastCount {
				rcolors[1] = k
			}
		}
		rcolors[2] = rcolors[1]
	}
	if clen == 1 {
		for i := clen - 1; i < 3; i++ {
			rcolors[i] = lastColor
		}
	}
	return IntToRGB(rcolors[0]), IntToRGB(rcolors[1]), IntToRGB(rcolors[2]), nil
}
