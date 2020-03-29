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

const PROC_COUNT = 4
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
func main() {
	runtime.GOMAXPROCS(PROC_COUNT) //SET MAX CPUs

	// af := "./testData/test2.jpg"
	
	af := "./testData/test2.jpg"
	// af = "./testData/bw2.jpg"
	af = "c:/Users/winds/Pictures/Van_Gogh_-_Starry_Night-11000px.jpg"
	// af = "c:/Users/winds/Pictures/the-starry-night-vincent-van-gogh.jpg"

	// // // GetImageFromJpeg(af)
	// image, Dx, Dy, _ := GetRGBAImage(af)
	// start := time.Now()
	// colorA, colorB, colorC, _ := DominantColors(image, Dx, Dy)
	// fmt.Println(time.Since(start))
	// fmt.Printf("%v %v %v", colorA, colorB, colorC)

	// f, err := os.Open(af)
	// if err != nil {
	// 	panic(err)
	// }
	// defer f.Close()
	// // Get the content
	// contentType, err := GetFileContentType(f)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("Content Type: " + contentType)
	// fmt.Println("Extension: " + contentType[len(contentType)-3:])

	//TODO ADD PRIMARY IMP

	// var colorA [3]byte
	// s := ColorToRGBHexString(colorA)
	// fmt.Println(s)

	// filename := "./testData/test2.jpg"
	// image, Dx, Dy, err := GetRGBAImage(filename)
	// HandleError(err, "failed to process image "+filename)

	// DistancePoints([3]byte{0, 0, 0}, [3]byte{1, 2, 3})
	// colorA, colorB, colorC, err := DominantColorsKMeans(image, Dx, Dy)

	// colorA, colorB, colorC, err := DominantColors(image, Dx, Dy)
	// fmt.Printf("%v %v %v \n", colorA, colorB, colorC)
	// HandleError(err, "")

	///LOCAL TEST
	// testURLFilename := "./list.txt"
	// testURLFilename := "./testData/listLocal.txt"

	///NET TEST
	// testURLFilename := "./testData/testUrlList.txt"

	////REAL DATA - SUBSET
	// testURLFilename := "./testData/inputSmall.txt"

	////REAL DATA
	// testURLFilename := "./testData/input.txt"
	// start := time.Now()
	// DominantColorsFromURLToCSV(testURLFilename, "colors.csv")
	// elapsed := time.Since(start)
	// fmt.Printf("Elapsed time: %s", elapsed)
}
