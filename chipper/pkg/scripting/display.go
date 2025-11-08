package scripting

// taken from https://github.com/fforchino/vector-go-sdk

import "image"

func ConvertPixesTo16BitRGB(r uint32, g uint32, b uint32, a uint32, opacityPercentage uint16) uint16 {
	R, G, B := uint16(r/257), uint16(g/8193), uint16(b/257)

	R = R * opacityPercentage / 100
	G = G * opacityPercentage / 100
	B = B * opacityPercentage / 100

	//The format appears to be: 000bbbbbrrrrrggg

	var Br uint16 = (uint16(B & 0xF8)) << 5 // 5 bits for blue  [8..12]
	var Rr uint16 = (uint16(R & 0xF8))      // 5 bits for red   [3..7]
	var Gr uint16 = (uint16(G))             // 3 bits for green [0..2]

	out := uint16(Br | Rr | Gr)
	//println(fmt.Sprintf("%d,%d,%d -> R: %016b G: %016b B: %016b = %016b", R, G, B, Rr, Gr, Br, out))
	return out
}

func ConvertPixelsToRawBitmap(image image.Image, opacityPercentage int) []uint16 {
	imgHeight, imgWidth := image.Bounds().Max.Y, image.Bounds().Max.X
	bitmap := make([]uint16, imgWidth*imgHeight)

	for y := 0; y < imgHeight; y++ {
		for x := 0; x < imgWidth; x++ {
			r, g, b, a := image.At(x, y).RGBA()
			bitmap[(y)*imgWidth+(x)] = ConvertPixesTo16BitRGB(r, g, b, a, uint16(opacityPercentage))
		}
	}
	return bitmap
}
