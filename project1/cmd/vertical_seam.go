package cmd

import (
	"computer_vision/lib"
	"fmt"
	"github.com/nfnt/resize"
	"github.com/pkg/errors"
	"image"
	"image/color"
	"image/jpeg"
	"math/rand"
	"os"
)

const pixelSpace = 10

func processVerticalIncrease(img image.Image, noPixelsToIncrease int, mode string) (image.Image, error) {
	imgGray := meta.GetGrayImage(img)
	magnitude := meta.SobelFilter(imgGray)

	if err := meta.PrintMagnitude(magnitude, "test_magnitude.jpeg"); err != nil {
		return nil, errors.Wrapf(err, "could not print magnitude image")
	}

	auxImg := img

	vertical := make([][]int, noPixelsToIncrease)

	fmt.Printf("pixels to increase %v, img size %v\n", noPixelsToIncrease, img.Bounds().Dx())

	for i := 0; i < noPixelsToIncrease; i++ {
		vertical[i] = findOneVertical(magnitude, mode)
		auxImg, magnitude = deleteVertical(vertical[i], auxImg, magnitude)
	}

	// Binary indexed trees for better complexity when finding the number of pixel after inserting stuff.
	aib := make([][]int, len(magnitude[0]) + 1)
	for x := 0; x < len(magnitude[0]); x++ {
		aib[x] = make([]int, len(magnitude) + 1 + noPixelsToIncrease)
	}

	for i := 0; i < noPixelsToIncrease; i++ {
		if len(vertical[i]) != len(magnitude[0]) {
			return nil, errors.New("vertical and magnitude has not the same value")
		}
		for line := range vertical[i] {
			vertical[i][line] += askAib(aib[line], vertical[i][line])
		}
		img = increaseOneVertical(img, vertical[i])

		for line := range vertical[i] {
			updateAib(aib[line], vertical[i][line], 1)
		}
	}
	return img, nil
}

func updateAib(aib []int, poz int, val int) {
	poz ++
	for poz > 0 {
		aib[poz] += val
		poz -= poz & (-poz)
	}
}

func askAib(aib []int, poz int) int {
	ret := 0
	// Aib is using interval [1, n]
	poz ++
	for poz < len(aib) {
		ret += aib[poz]
		poz += poz & (-poz)
	}
	return ret
}

func increaseOneVertical(srcImg image.Image, vertical []int) image.Image {
	srcDim := srcImg.Bounds()
	dstImage := image.NewRGBA(image.Rect(0, 0, srcDim.Dx() + 1, srcDim.Dy()))

	for y := 0; y < srcDim.Dy(); y++ {
		for x := 0; x < srcDim.Dx(); x ++ {
			if x < vertical[y] {
				dstImage.Set(x, y, srcImg.At(x, y))
			} else {
				dstImage.Set(x+1, y, srcImg.At(x, y))
			}
		}
		if vertical[y] == 0 {
			dstImage.Set(vertical[y], y, srcImg.At(vertical[y], y))
			continue
		}

		rS, gS, bS, aS := srcImg.At(vertical[y] - 1, y).RGBA()
		rD, gD, bD, aD := srcImg.At(vertical[y], y).RGBA()

		pixel := color.RGBA{
			R: uint8((rS + rD) >> 9),
			G: uint8((gS + gD) >> 9),
			B: uint8((bS + bD) >> 9),
			A: uint8((aS + aD) >> 9),
		}

		//fmt.Printf("rs=%v rd=%v %v\n", rS, rD, (rS + rD) / 2)
		//fmt.Printf("pixel=\n%v\n old=\n%v\n old2=\n%v\n\n", pixel, srcImg.At(vertical[y], y), srcImg.At(vertical[y], y))

		dstImage.Set(vertical[y], y, pixel)

	}

	return dstImage
}

func proceedVerticalErase(img image.Image, noPixelsToErase int, mode string) (image.Image, error) {
	imgGray := meta.GetGrayImage(img)
	magnitude := meta.SobelFilter(imgGray)

	if err := meta.PrintMagnitude(magnitude, "test_magnitude.jpeg"); err != nil {
		return nil, errors.Wrapf(err, "could not print magnitude image")
	}

	for i := 0; i < noPixelsToErase; i++ {
		vertical := findOneVertical(magnitude, mode)
		img, magnitude = deleteVertical(vertical, img, magnitude)
	}
	return img, nil
}

func findOneVertical(magnitude [][]float64, mode string) []int {
	if mode == "dynamics" {
		return findOneVerticalDynamics(magnitude)
	}
	if mode == "greedy" {
		return findOneVerticalGreedy(magnitude)
	}
	return findOneVerticalRandom(magnitude)
}

func findOneVerticalDynamics(magnitude [][]float64) []int {
	dyn := make([][]float64, len(magnitude))
	frm := make([][]int, len(magnitude))
	for x := 0; x < len(magnitude); x++ {
		dyn[x] = make([]float64, len(magnitude[x]))
		frm[x] = make([]int, len(magnitude[x]))
	}

	for x := 0; x < len(magnitude); x++ {
		dyn[x][0] = magnitude[x][0]
	}
	//fmt.Printf("magnitude len = %v\n", len(magnitude))
	for y := 1; y < len(magnitude[0]); y++ {
		for x := 0; x < len(magnitude); x++ {
			dyn[x][y] = dyn[x][y - 1] + magnitude[x][y]
			frm[x][y] = x
			if x != 0 && dyn[x - 1][y - 1] + magnitude[x][y] < dyn[x][y]{
				dyn[x][y] = dyn[x - 1][y - 1] + magnitude[x][y]
				frm[x][y] = x - 1
			}
			if x != len(magnitude) - 1 && dyn[x + 1][y - 1] + magnitude[x][y] < dyn[x][y] {
				dyn[x][y] = dyn[x + 1][y - 1] + magnitude[x][y]
				frm[x][y] = x + 1
			}
		}
	}

	lastP := 0

	for x := 1; x < len(magnitude); x ++ {
		if dyn[x][len(magnitude[0]) - 1] < dyn[lastP][len(magnitude[0]) - 1] {
			lastP = x
		}
	}

	vertical := []int{lastP}

	for y := len(magnitude[0]) - 1; y > 0; y -- {
		vertical = append([]int{frm[lastP][y]}, vertical...)
		lastP = frm[lastP][y]
	}
	return vertical
}

func findOneVerticalGreedy(magnitude [][]float64) []int {
	last := 0
	for x := 1; x < len(magnitude); x++ {
		if magnitude[x][0] < magnitude[last][0] {
			last = x
		}
	}
	vertical := []int{last}
	for y := 1; y < len(magnitude[0]); y++ {
		next := last
		if last != 0 && magnitude[last - 1][y] < magnitude[next][y] {
			next = last - 1
		}
		if last != len(magnitude) - 1 && magnitude[last + 1][y] < magnitude[next][y] {
			next = last + 1
		}
		vertical = append(vertical, next)
		last = next
	}
	return vertical
}

func findOneVerticalRandom(magnitude [][]float64) []int {
	last := rand.Intn(len(magnitude))
	vertical := []int{last}
	for y := 1; y < len(magnitude[0]); y++ {
		next := last + rand.Intn(3) - 1
		for next < 0 || next >= len(magnitude) {
			next = last + rand.Intn(3) - 1
		}
		vertical = append(vertical, next)
		last = next
	}
	return vertical
}

func deleteVertical(vertical []int, img image.Image, magnitude [][]float64) (image.Image, [][]float64)  {
	retMagnitude := make([][]float64, len(magnitude) - 1)
	for x := 0; x < len(magnitude) - 1; x++ {
		retMagnitude[x] = make([]float64, len(magnitude[x]))
	}

	retImg := image.NewRGBA(image.Rect(0,0, len(magnitude) - 1, len(magnitude[0])))
	for line, indexDel := range vertical {
		for p := 0; p < indexDel; p ++ {
			retImg.Set(p, line, img.At(p, line))
			retMagnitude[p][line] = magnitude[p][line]
		}

		for p := indexDel; p < len(magnitude) - 1; p++ {
			retImg.Set(p, line, img.At(p + 1, line))
			retMagnitude[p][line] = magnitude[p + 1][line]
		}
	}

	return retImg, retMagnitude
}

func printImage(finalImg image.Image, initImg image.Image, output string) error {
	outFile, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return errors.Wrapf(err, "could not create file at path '%v'", *outputPath)
	}
	defer outFile.Close()

	clasicImg := resize.Resize(uint(finalImg.Bounds().Dx()), uint(finalImg.Bounds().Dy()), initImg, resize.Lanczos3)

	newRect := image.Rectangle{
		Max: image.Point{
			X: max(initImg.Bounds().Dx(), finalImg.Bounds().Dx(), clasicImg.Bounds().Dx()),
			Y: initImg.Bounds().Dy() + pixelSpace + finalImg.Bounds().Dy() + pixelSpace + clasicImg.Bounds().Dy(),
		},
	}
	prtImage := image.NewRGBA(newRect)

	addImage(prtImage, initImg, 0, 0)
	addImage(prtImage, finalImg, 0, initImg.Bounds().Dy() + pixelSpace)
	addImage(prtImage, clasicImg, 0, initImg.Bounds().Dy() + pixelSpace + finalImg.Bounds().Dy() + pixelSpace)

	return jpeg.Encode(outFile, prtImage, nil)
}

func addImage(act *image.RGBA, appImage image.Image, xstart int, ystart int) {
	for x := 0; x < appImage.Bounds().Dx(); x++ {
		for y := 0; y < appImage.Bounds().Dy(); y++ {
			act.Set(x + xstart, y + ystart, appImage.At(x, y))
		}
	}
}

func max(x, y, z int) int {
	if x > y && x > z {
		return x
	}
	if y > x && y > z {
		return y
	}
	return z
}
