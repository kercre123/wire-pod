package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
)

const SCREEN_WIDTH, SCREEN_HEIGHT = 184, 96 // 240,240 // 180,240

const OUTPUT_FILE_PATH = "boot_anim.raw"

// static image as single frame is quick to disappear compared to a gifs
// extra frames gives them more stage time, anything below 30 may not be observed by humans
const STATIC_IMAGE_FRAMES = 30

func Path(rel string) string {
	if filepath.IsAbs(rel) {
		return rel
	}
	abs, _ := filepath.Abs(rel)
	return abs
}

func fileExists(file string) (bool, error) {
	_, err := os.Stat(file)
	if err != nil {
		return false, err
	}
	return true, err
}

func getGifDimensions(gif *gif.GIF) (x, y int) {
	var lowestX int
	var lowestY int
	var highestX int
	var highestY int

	for _, img := range gif.Image {
		if img.Rect.Min.X < lowestX {
			lowestX = img.Rect.Min.X
		}
		if img.Rect.Min.Y < lowestY {
			lowestY = img.Rect.Min.Y
		}
		if img.Rect.Max.X > highestX {
			highestX = img.Rect.Max.X
		}
		if img.Rect.Max.Y > highestY {
			highestY = img.Rect.Max.Y
		}
	}

	return highestX - lowestX, highestY - lowestY
}

func convertPixesTo16BitRGB(r uint32, g uint32, b uint32, a uint32) uint16 {
	R, G, B := int(r/257), int(g/257), int(b/257)

	return uint16((int(R>>3) << 11) |
		(int(G>>2) << 5) |
		(int(B>>3) << 0))
}

func isFileValid(filePath string) (file *os.File, err error) {
	_, err = fileExists(filePath)
	if err != nil {
		return nil, err
	}

	file, err = os.Open(filePath)

	if err != nil {
		return nil, err
	}

	return file, nil
}

func getGifFrames(file *os.File) ([]*image.Paletted, error) {
	gifImages, err := gif.DecodeAll(file)
	if err != nil {
		return nil, err
	}
	imgWidth, imgHeight := getGifDimensions(gifImages)

	if imgHeight != SCREEN_HEIGHT || imgWidth != SCREEN_WIDTH {
		return nil, errors.New(fmt.Sprintf("width %dpx height %dpx file is required.", SCREEN_WIDTH, SCREEN_HEIGHT))

	}

	return gifImages.Image, nil
}

func prepareOutputFile() (*os.File, error) {
	outputFilePath := os.Getenv("ANIM_OUTPUT_FILE_PATH") + OUTPUT_FILE_PATH
	outputFileExists, _ := fileExists(outputFilePath)

	if outputFileExists {
		os.Remove(outputFilePath)
	}

	outPutFile, err := os.OpenFile(outputFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}
	return outPutFile, nil
}

func convertPixelsToRawBitmap(image *image.RGBA) []uint16 {
	imgHeight, imgWidth := image.Bounds().Max.Y, image.Bounds().Max.X
	bitmap := make([]uint16, SCREEN_WIDTH*SCREEN_HEIGHT)

	for y := 0; y < imgHeight; y++ {
		for x := 0; x < imgWidth; x++ {
			bitmap[(y)*SCREEN_WIDTH+(x)] = convertPixesTo16BitRGB(image.At(x, y).RGBA())
		}
	}
	return bitmap
}

func writeFramesToOutputFile(file *os.File, frames []*image.Paletted) {
	overpaintImage := image.NewRGBA(image.Rect(0, 0, frames[0].Bounds().Max.X, frames[0].Bounds().Max.Y))
	draw.Draw(overpaintImage, overpaintImage.Bounds(), frames[0], image.ZP, draw.Src)

	for _, srcImg := range frames {
		draw.Draw(overpaintImage, overpaintImage.Bounds(), srcImg, image.ZP, draw.Over)
		bitmap := convertPixelsToRawBitmap(overpaintImage)

		for _, ui := range bitmap {
			binary.Write(file, binary.LittleEndian, ui)
		}

	}
}

func getGifImages(gifPath string) ([]*image.Paletted, error) {
	inputFile, err := isFileValid(gifPath)
	if err != nil {
		return nil, err
	}

	defer inputFile.Close()

	gifImages, err := getGifFrames(inputFile)
	if err != nil {
		return nil, err
	}
	return gifImages, nil
}

func getPNGRaw(pngPath string) ([]uint16, error) {
	pngFile, err := os.Open(pngPath)
	if err != nil {
		return nil, err
	}
	defer pngFile.Close()

	pngImage, err := png.Decode(pngFile)
	if err != nil {
		return nil, err
	}

	imageHeight, imageWidth := pngImage.Bounds().Max.Y, pngImage.Bounds().Max.X

	if imageHeight != SCREEN_HEIGHT || imageWidth != SCREEN_WIDTH {
		return nil, errors.New(fmt.Sprintf("width %dpx height %dpx file is required.", SCREEN_WIDTH, SCREEN_HEIGHT))
	}

	overpaintImage := image.NewRGBA(image.Rect(0, 0, imageWidth, imageHeight))
	draw.Draw(overpaintImage, overpaintImage.Bounds(), pngImage, image.ZP, draw.Src)

	return convertPixelsToRawBitmap(overpaintImage), nil
}

func getJPGRaw(jpgPath string) ([]uint16, error) {
	jpgFile, err := os.Open(jpgPath)
	if err != nil {
		return nil, err
	}
	defer jpgFile.Close()

	jpgImage, err := jpeg.Decode(jpgFile)
	if err != nil {
		return nil, err
	}

	imageHeight, imageWidth := jpgImage.Bounds().Max.Y, jpgImage.Bounds().Max.X
	if imageHeight != SCREEN_HEIGHT || imageWidth != SCREEN_WIDTH {
		return nil, errors.New(fmt.Sprintf("width %dpx height %dpx file is required.", SCREEN_WIDTH, SCREEN_HEIGHT))
	}

	overpaintImage := image.NewRGBA(image.Rect(0, 0, imageWidth, imageHeight))
	draw.Draw(overpaintImage, overpaintImage.Bounds(), jpgImage, image.ZP, draw.Src)

	return convertPixelsToRawBitmap(overpaintImage), nil
}

func main() {
	imagePaths := os.Args[1:]

	outPutFile, err := prepareOutputFile()
	if err != nil {
		fmt.Println(err)
		return
	}

	var processErr string

	for _, imagePath := range imagePaths {
		imageExtension := strings.ToLower(filepath.Ext(Path(imagePath)))
		imagePathAbs := Path(imagePath)
		if imageExtension == ".gif" {
			if gifImages, err := getGifImages(imagePathAbs); err != nil {
				processErr = fmt.Sprintf("%s: %s ", err.Error(), imagePath)
				break
			} else {
				writeFramesToOutputFile(outPutFile, gifImages)
			}

		} else if imageExtension == ".png" {
			if pngImage, err := getPNGRaw(imagePathAbs); err != nil {
				processErr = fmt.Sprintf("%s: %s ", err.Error(), imagePath)
				break
			} else {
				for i := 0; i < STATIC_IMAGE_FRAMES; i++ {
					binary.Write(outPutFile, binary.LittleEndian, pngImage)
				}
			}
		} else if imageExtension == ".jpg" || imageExtension == ".jpeg" {
			if pngImage, err := getJPGRaw(imagePathAbs); err != nil {
				processErr = fmt.Sprintf("%s: %s ", err.Error(), imagePath)
				break
			} else {
				for i := 0; i < STATIC_IMAGE_FRAMES; i++ {
					binary.Write(outPutFile, binary.LittleEndian, pngImage)
				}
			}
		} else {
			processErr = fmt.Sprintf("Unsupported file format: %s ", imagePath)
			break
		}
	}

	defer outPutFile.Close()

	if processErr != "" {
		fmt.Println(fmt.Sprintf("Error- %s", processErr))
		return
	}

	fmt.Println(fmt.Sprintf("Done! frames merged to create screen animation file: %s", OUTPUT_FILE_PATH))
}
