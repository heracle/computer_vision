package cmd

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"image"
	"strconv"
)

type point struct {
	x 	int
	y 	int
}

func EraseObject() *cobra.Command {
	var command = &cobra.Command{
		Use: "erase <image path> [X1] [Y1] [X2] [Y2] [X3] [Y3] ...",
		Short: "Erase an object from an image by keeping the most interesting content.",
		Args: cobra.MinimumNArgs(7),
		RunE: func(_ *cobra.Command, args []string) error {
			imgPath := args[0]
			img, err := getImageFromPath(imgPath)
			if err != nil {
				return errors.Wrapf(err, "could not get an image obj from path '%v'", imgPath)
			}
			initImg := img

			var polyLine []point

			if len(args) % 2 != 1 {
				return errors.New("not an even number of numbers received")
			}

			for i := 1; i < len(args); i+=2 {
				actX, err := strconv.Atoi(args[i])
				if err != nil {
					return errors.Wrapf(err, "could not parse as integer arg received '%v'", args[1])
				}

				actY, err := strconv.Atoi(args[i + 1])
				if err != nil {
					return errors.Wrapf(err, "could not parse as integer arg received '%v'", args[1])
				}

				polyLine = append(polyLine, point{actX, actY})
			}


			// Fake circularity.
			polyLine = append(polyLine, polyLine[0])
			polyLine = append(polyLine, polyLine[1])

			left := polyLine[0].x
			right := polyLine[0].x
			up := polyLine[0].y
			down := polyLine[0].y

			for _, pt := range polyLine {
				if pt.x < left {
					left = pt.x
				}
				if pt.x > right {
					right = pt.x
				}
				if pt.y < up {
					up = pt.y
				}
				if pt.y > down {
					down = pt.y
				}
			}

			noErasePixels := right - left

			if down - up < right - left {
				rotateClockLine(img, polyLine)
				img = rotateClock(img)
				noErasePixels = down - up
			}

			img, err = proceedObjectErase(img, noErasePixels, polyLine, *modeResize)
			if err != nil {
				return errors.Wrapf(err, "could not proceed object erase according to the received polyline")
			}

			if down - up < right - left {
				img = rotateClock(img)
				img = rotateClock(img)
				img = rotateClock(img)
			}

			return printImage(img, initImg, *outputPath)
		},
	}
	return command
}

func proceedObjectErase(img image.Image, noPixelsToErase int, polyLine []point, mode string) (image.Image, error) {
	imgGray := getGrayImage(img)
	magnitude := sobelFilter(imgGray)

	for x := range magnitude {
		for y := range magnitude[x] {
			if insidePolyLine(x, y, polyLine) {
				magnitude[x][y] = -10000000
			}
		}
	}

	if err := printMagnitude(magnitude, "test_magnitude.jpeg"); err != nil {
		return nil, errors.Wrapf(err, "could not print magnitude image")
	}

	for i := 0; i < noPixelsToErase; i++ {
		vertical := findOneVertical(magnitude, mode)
		img, magnitude = deleteVertical(vertical, img, magnitude)
	}
	return img, nil
}

func insidePolyLine(x int, y int, polyLine []point) bool {
	for i := 2; i < len(polyLine); i ++ {
	//	Check if the point is always on the same part of the edge
		dir1 := (x - polyLine[i - 2].x) * (y - polyLine[i - 1].y) -
			(y - polyLine[i - 2].y) * (x - polyLine[i - 1].x)
		dir2 := (x - polyLine[i - 1].x) * (y - polyLine[i].y) -
			(y - polyLine[i - 1].y) * (x - polyLine[i].x)

		if dir1 * dir2 < 0 {
			return false
		}
	}
	return true
}
