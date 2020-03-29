package main

import (
	"errors"
	"fmt"
	"image"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
	"sync"
)

const rgbLen = 3

type imageInfo struct {
	filename string
	link     string
	err      error
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
