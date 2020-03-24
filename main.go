package main

import (
	"fmt"
	"runtime"
	"time"
)

func main() {

	runtime.GOMAXPROCS(1) //SET MAX CPUs
	
	af := "./testData/test2.jpg"
	// af = "./testData/bw2.jpg"
	af = "c:/Users/winds/Pictures/Van_Gogh_-_Starry_Night-11000px.jpg"
	// af = "c:/Users/winds/Pictures/the-starry-night-vincent-van-gogh.jpg"

	// // GetImageFromJpeg(af)
	image, Dx, Dy, _ := GetRGBAImage(af)
	start := time.Now()
	colorA, colorB, colorC, _ := DominantColors(image, Dx, Dy)
	fmt.Println(time.Since(start))
	fmt.Printf("%v %v %v", colorA, colorB, colorC)

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
