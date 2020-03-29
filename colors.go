package main

import "fmt"

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
func ColorToRGBHexString(color [rgbLen]byte) string {
	return fmt.Sprintf("#%X%X%X", color[0], color[1], color[2])
}
