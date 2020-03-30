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

func DominantColorsFromURLToCSV(urlListFile string, csvFilename string) {
	chImgInfo := DownloadAllImages(urlListFile)
	st, chErr := DominantColorsFromRGBAImage(chImgInfo)
	err := saveEverythingToCSV(st, chErr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}

//TODO not sure which approach will work better with goroutines/ structures vs channel/slice merge
func saveEverythingToCSV(st <-chan [4]string, errors <-chan error) error {
	//create CSV file,
	csvFilename := "huhu.csv"
	outputCSV, err := os.Create(csvFilename)
	if err != nil {
		return fmt.Errorf("failed creating  CSV file: %v", err)
	}
	writerCSV := csv.NewWriter(outputCSV)
	//TODO can this run concurrently too ?
	// go func() {
	// for line := range st {
	for {

		select {
		case err := <-errors:
			return fmt.Errorf("CSV writer failed: %v", err)
		case line := <-st:
			err = writerCSV.Write(line[:])
			if err != nil {
				return err
			}
			fmt.Println("-- adding to CSV[", line[0], "]")
		}

		//TODO FIX if this errs the channel stays undrained
		// if err != nil {
		// 	return err
		// }
	}

	writerCSV.Flush()
	err = outputCSV.Close()
	if err != nil {
		return err
	}
	// }()
	return nil
}

func DominantColorsFromRGBAImage(chImgInfo <-chan imageInfo) (<-chan [4]string, <-chan error) {
	out := make(chan [4]string, BUFFER_SIZE)
	errors := make(chan error, 0)
	go func() {
		for imgInfo := range chImgInfo { //TODO better goroutines handling
			if imgInfo.err != nil {
				errors <- fmt.Errorf("failed to download given image: %v", imgInfo.err)
				return
			}
			fmt.Println("-- opening the image " + imgInfo.filename)
			image, Dx, Dy, err := GetRGBAImage(imgInfo.filename)
			if err != nil {
				errors <- fmt.Errorf("failed to open the image[%s]: %v", imgInfo.filename, err)
				return
			}
			fmt.Println("-- opening the image DONE")
			fmt.Println("-- dominant colors " + imgInfo.filename)
			colorA, colorB, colorC, err := DominantColors(image, Dx, Dy)
			if err != nil {
				errors <- fmt.Errorf("failed to get dominantColors[%s]: %v", imgInfo.filename, err)
				return
			}
			out <- [4]string{imgInfo.link, ColorToRGBHexString(colorA), ColorToRGBHexString(colorB), ColorToRGBHexString(colorC)}
		}
		close(out)
	}()
	return out, nil
	//remove temp file err = os.Remove(filename)
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
	link     string
	err      error
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

func DownloadAllImages(linksFile string) <-chan imageInfo {
	linksScanner, fileHandle, err := OpenTheList(linksFile)
	chImgInfo := make(chan imageInfo, BUFFER_SIZE)
	if err != nil {
		chImgInfo <- imageInfo{"", "", nil}
		return chImgInfo
	}

	//defer fileHandle.Close()
	go func() {
		for linksScanner.Scan() {
			url := linksScanner.Text()
			filename, err := DownloadImage(url)
			chImgInfo <- imageInfo{filename, url, fmt.Errorf("failed to download the file[%s]: %v", filename, err)}
		}
		close(chImgInfo)
		fileHandle.Close()
	}()
	return chImgInfo
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

func DominantColors(image *image.RGBA, width int, height int) ([rgbLen]byte, [rgbLen]byte, [rgbLen]byte, error) {
	if width == 0 || height == 0 {
		var ccA, ccB, ccC [rgbLen]byte
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
