package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"runtime"
	"strconv"
	"time"
)

const RGB_LENGTH = 3

var BUFFER_SIZE = runtime.NumCPU()

func DominantColorsFromURLToCSV(urlListFile string, csvFilename string) error {
	openTheList := func(urlListFile string) (*bufio.Scanner, *os.File, error) {
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
	linksScanner, fileHandle, err := openTheList(urlListFile)
	if err != nil {
		return err
	}
	chImgInfo := downloadAllImages(linksScanner)
	st, filenames := dominantColorsFromRGBAImage(chImgInfo)
	deleteImages(filenames)
	err = saveEverythingToCSV(st, csvFilename)
	if err != nil {
		return err
	}
	err = fileHandle.Close()
	if err != nil {
		return err
	}
	return nil
}

//? We could probably use only one structure, but since we drain the first channel anyway, we would still have 2channels
type imageInfo struct {
	filename string
	link     string
	err      error
}
type processedImage struct {
	csvInfo [4]string
	err     error
}

func downloadAllImages(linksScanner *bufio.Scanner) <-chan imageInfo {
	chImgInfo := make(chan imageInfo, BUFFER_SIZE)

	downloadImage := func(url string) (string, error) {
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
	go func() {
		for linksScanner.Scan() {
			url := linksScanner.Text()
			filename, err := downloadImage(url)
			if err != nil {
				err = fmt.Errorf("failed to download the image: %v", err)
			}
			chImgInfo <- imageInfo{filename, url, err}
		}
		close(chImgInfo)
	}()
	return chImgInfo
}
func dominantColorsFromRGBAImage(chImgInfo <-chan imageInfo) (<-chan processedImage, <-chan string) {
	toHexString := func(color [RGB_LENGTH]byte) string {
		return fmt.Sprintf("#%X%X%X", color[0], color[1], color[2])
	}
	out := make(chan processedImage, BUFFER_SIZE)
	outnames := make(chan string, BUFFER_SIZE)
	go func() {
		for imgInfo := range chImgInfo {
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
			outnames <- imgInfo.filename
			o.csvInfo = [4]string{imgInfo.link, toHexString(colorA), toHexString(colorB), toHexString(colorC)}
			out <- o
		}
		close(out)
		close(outnames)
	}()
	return out, outnames
}
func deleteImages(filenames <-chan string) {
	go func() {
		for filename := range filenames {
			os.Remove(filename) //ignore errs
		}
	}()
}

//?This is fanIn function which may be bit confusing
func saveEverythingToCSV(st <-chan processedImage, csvFilename string) error {
	outputCSV, err := os.Create(csvFilename)
	if err != nil {
		return fmt.Errorf("failed creating  CSV file: %v", err)
	}
	writerCSV := csv.NewWriter(outputCSV)
	for pi := range st {
		//!From that func perspective, its error, but we won't to handle it gracefully here
		if pi.err != nil {
			err = writerCSV.Write([]string{pi.err.Error()})
			fmt.Println("-- adding to CSV[", pi.csvInfo[0], "]")
		} else {
			err = writerCSV.Write(pi.csvInfo[:])
		}
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

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU()) //*MAKING SURE WE SET MAX CPUs

	//urlFilename = "./testData/inputSmall.txt"s
	urlFilename := "./testData/input.txt"
	csvFilename := "dominantColors.csv"

	if len(os.Args) > 1 {
		urlFilename = os.Args[1]
	}
	if len(os.Args) > 2 {
		csvFilename = os.Args[2]
	}

	start := time.Now()
	err := DominantColorsFromURLToCSV(urlFilename, csvFilename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	elapsed := time.Since(start)

	csvStat, err := os.Stat(csvFilename)
	if err != nil || csvStat.Size() < 1 {
		fmt.Fprintf(os.Stderr, "OUTPUT SIZE CHECK FAILED! %v", err)
		os.Exit(1)
	}
	fmt.Printf("Elapsed time: %s", elapsed)
}
