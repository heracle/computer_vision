package meta

import (
	"github.com/pkg/errors"
	"image"
	"image/color"
	"image/jpeg"

	"math"
	"os"
)

type Point struct {
	X 	int
	Y 	int
}

func GetImageFromPath(path string) (image.Image, error) {
	imgFile, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrapf(err, "could not open file '%v'", path)
	}
	defer imgFile.Close()

	img, _, err := image.Decode(imgFile)
	if err != nil {
		return nil, errors.Wrapf(err, "could not decode file '%v'", path)
	}
	return img, nil
}

func GetGrayImage(img image.RGBA) [][]float64 {
	bounds := img.Bounds()

	grayScale := make([][]float64, bounds.Dx())

	for i := range grayScale {
		grayScale[i] = make([]float64, bounds.Dy())
	}

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// A color's RGBA method returns values in the range [0, 65535].
			r, g, b, _ := img.At(x, y).RGBA()
			grayScale[x - bounds.Min.X][y - bounds.Min.Y] = 0.2989 * float64(r) + 0.5870 * float64(g) + 0.1140 * float64(b)
		}
	}
	return  grayScale
}

func SobelFilter(gray [][]float64) [][]float64 {
	gx := [][]float64{
		{-1, 0, 1},
		{-2, 0, 2},
		{-1, 0, 1},
	}
	gy := [][]float64{
		{-1, -2, -1},
		{0, 0, 0},
		{1, 2, 1},
	}
	magnitude := make([][]float64, len(gray))

	for x := range magnitude {
		magnitude[x] = make([]float64, len(gray[x]))
	}

	for x := 1; x < len(gray) - 1; x ++ {
		for y := 1; y < len(gray[x]) - 1; y ++ {
			sx := CartesianProductSum(gx, gray, x, y)
			sy := CartesianProductSum(gy, gray, x, y)

			magnitude[x][y] = math.Sqrt(sx * sx + sy * sy)
		}
	}
	return magnitude
}

func CartesianProductSum(g [][]float64, img [][]float64, x int, y int) float64 {
	var res float64

	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			res += g[i + 1][j + 1] * img[x +i][y + j]
		}
	}

	return res
}

func PrintMagnitude(magnitude [][]float64, path string) error {
	img := image.NewRGBA(image.Rect(0,0, len(magnitude), len(magnitude[0])))
	for x := 0; x < len(magnitude); x ++ {
		if len(magnitude[x]) != len(magnitude[0]) {
			return errors.Wrapf (nil, "magnitude at line %v has value %v, when line 0 has %v", x, magnitude[x], magnitude[0])
		}
		for y := 0; y < len(magnitude[0]); y++ {
			pixel := color.Gray16{uint16(magnitude[x][y])}
			img.Set(x, y, pixel)
		}
	}

	outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return errors.Wrapf(err, "could not create file at path '%v'", path)
	}
	defer outFile.Close()
	return jpeg.Encode(outFile, img, nil)
}

func RotateClockLine(srcImg image.Image, poliLyne []Point) {
	for i := 0; i < len(poliLyne); i++ {
		oldX := poliLyne[i].X
		poliLyne[i].X = poliLyne[i].Y
		poliLyne[i].Y = srcImg.Bounds().Dx() - 1 - oldX
	}
}

func RotateClock(srcImg image.RGBA) image.RGBA {
	srcDim := srcImg.Bounds()
	dstImage := image.NewRGBA(image.Rect(0, 0, srcDim.Dy(), srcDim.Dx()))

	for x := 0; x < srcDim.Dx(); x ++ {
		for y := 0; y < srcDim.Dy(); y++ {
			dstImage.Set(y, srcDim.Dx() - 1 - x, srcImg.At(x, y))
		}
	}

	return *dstImage
}

