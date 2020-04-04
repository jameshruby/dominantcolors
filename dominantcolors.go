package main

import (
	"errors"
	"image"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
	"runtime"
	"sync"
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
	//build a map of unique colors and its sum, pix is array with colors just stacked behind each other
	imgPix := image.Pix

	const rgbaLen = 4

	lenImgPix := len(imgPix)
	partitionsLen := int(math.RoundToEven((float64(lenImgPix/rgbaLen) / float64(partitionsCount)))) * rgbaLen

	toIntSlice := func(color []byte) int {
		r, g, b := int(color[0]), int(color[1]), int(color[2])
		rgb := ((r & 0xFF) << 16) | ((g & 0xFF) << 8) | (b & 0xFF)
		return rgb
	}
	countPartition := func(imgPix []uint8) <-chan map[int]int {
		uniqueColorsCh := make(chan map[int]int)
		go func(imgPix []uint8) {
			uniqueColors := make(map[int]int)
			for i := 0; i < len(imgPix); i += rgbaLen {
				pixel := toIntSlice(imgPix[i : i+4 : i+4])
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

	intToRGB := func(rgb int) [3]byte {
		r := (rgb >> 16) & 0xFF
		g := (rgb >> 8) & 0xFF
		b := rgb & 0xFF
		return [3]byte{byte(r), byte(g), byte(b)}
	}

	return intToRGB(rcolors[0]), intToRGB(rcolors[1]), intToRGB(rcolors[2]), nil
}
