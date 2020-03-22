package main

import (
	"fmt"
	"runtime"
)

// "time"

//  _ "net/http/pprof"

// urls := []string{"http://google.com"}
// 	for _, url := range urls {
// 		resp, err := http.Get(url)
// 		if err != nil {
// 			fmt.Printf("%v\n", err)
// 			os.Exit(1)
// 		}
// 		b, err := ioutil.ReadAll(resp.Body)
// 		resp.Body.Close()
// 		if err != nil {
// 			fmt.Printf("%v\n", err)
// 			os.Exit(1)
// 		}
// 		fmt.Printf("%s", b)

// func fetch(url string, ch chan<- string) {
// 	start := time.Now()
// 	resp, err := http.Get(url)
// 	if err != nil {
// 		ch <- fmt.Srintf(err)
// 		return
// 	}
// 	nbytes, err := io.Copy(ioutil.Discard, resp.Body)
// 	resp.Body.Close()
// 	if err != nil {
// 		ch <- fmt.Srintf("while reading %s: %v\n", url, err)
// 		return
// 	}
// 	secs := time.Since(start).Seconds()
// 	ch <- fmt.Sprintf("%.2fs %7d %s", secs, nbytes, url)
// }

func main() {

	cpus := runtime.NumCPU()
	fmt.Println(cpus)
	// var d clusters.Observations

	// objects := [][3]float64{
	// 	{1, 1, 1},
	// 	{1, 1, 1},
	// 	{1, 1, 1},
	// 	{1, 1, 1},
	// 	{2, 3, 15},
	// 	{1, 1, 1},
	// 	{2, 2, 2},
	// 	{1, 1, 1},
	// 	{2, 2, 2},
	// }
	// for _, pix := range objects {
	// 	d = append(d, clusters.Coordinates{
	// 		float64(pix[0]),
	// 		float64(pix[1]),
	// 		float64(pix[2]),
	// 	})
	// }

	// km := kmeans.New()
	// clusters, _ := km.Partition(d, 6)
	// fmt.Println(clusters)

	// objects := [][3]uint8{
	// 	{15, 34, 250},
	// 	{15, 34, 250},
	// 	{1, 1, 1},
	// 	{2, 2, 2},
	// 	{2, 3, 15},
	// 	{15, 34, 250},
	// 	{2, 2, 2},
	// 	{1, 1, 1},
	// 	{2, 2, 2},
	// }

	objectsPix := []uint8{
		15, 34, 250, 255,
		15, 34, 250, 255,
		1, 1, 1, 255,
		2, 2, 2, 255,
		2, 3, 15, 255,
		15, 34, 250, 255,
		2, 2, 2, 255,
		1, 1, 1, 255,
		2, 2, 2, 255,
	}

	// objects := [][3]float64{
	// 	{1, 1, 1},
	// 	{15, 34, 250},
	// 	{1, 1, 1},
	// 	{2, 2, 2},
	// 	{2, 3, 15},
	// 	{15, 34, 250},
	// 	{2, 2, 2},
	// 	{1, 1, 1},
	// 	{1, 1, 1},
	// }

	KmeansPartition2(objectsPix)
	// fmt.Println(acolor)

	// icolor := RGBToInt(dominantColorsByte["yellow"]) //13356134
	// acolor := IntToRGB(icolor)
	// fmt.Println(acolor)

	// af := "D.jpg"
	// // GetImageFromJpeg(af)
	// image, Dx, Dy, _ := GetImageFromJpeg(af)
	// colorA, colorB, colorC, _ := DominantColors(image, Dx, Dy)
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
