package main

import (
	"errors"
	"image"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"runtime"
)

var partitionsCount = runtime.NumCPU()

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

func DominantColors(image *image.RGBA, width int, height int) ([RGB_LENGTH]byte, [RGB_LENGTH]byte, [RGB_LENGTH]byte, error) {
	if width == 0 || height == 0 {
		var ccA, ccB, ccC [RGB_LENGTH]byte
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

	return [RGB_LENGTH]byte{byte(ccA[0]), byte(ccA[1]), byte(ccA[2])}, [RGB_LENGTH]byte{byte(ccB[0]), byte(ccB[1]), byte(ccB[2])}, [RGB_LENGTH]byte{byte(ccC[0]), byte(ccC[1]), byte(ccC[2])}, nil
}
