package cmd

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"image"
	"strconv"

	"computer_vision/lib"
)

func EraseObject() *cobra.Command {
	var command = &cobra.Command{
		Use: "erase <image path> [X1] [Y1] [X2] [Y2] [X3] [Y3] ...",
		Short: "Erase an object from an image by keeping the most interesting content.",
		Args: cobra.MinimumNArgs(7),
		RunE: func(_ *cobra.Command, args []string) error {
			imgPath := args[0]
			img, err := meta.GetImageFromPath(imgPath)
			if err != nil {
				return errors.Wrapf(err, "could not get an image obj from path '%v'", imgPath)
			}
			initImg := img

			var polyLine []meta.Point

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

				polyLine = append(polyLine, meta.Point{X: actX,Y: actY})
			}


			// Fake circularity.
			polyLine = append(polyLine, polyLine[0])
			polyLine = append(polyLine, polyLine[1])

			left := polyLine[0].X
			right := polyLine[0].X
			up := polyLine[0].Y
			down := polyLine[0].Y

			for _, pt := range polyLine {
				if pt.X < left {
					left = pt.X
				}
				if pt.X > right {
					right = pt.X
				}
				if pt.Y < up {
					up = pt.Y
				}
				if pt.Y > down {
					down = pt.Y
				}
			}

			noErasePixels := right - left

			if down - up < right - left {
				meta.RotateClockLine(img, polyLine)
				img = meta.RotateClock(img)
				noErasePixels = down - up
			}

			img, err = proceedObjectErase(img, noErasePixels, polyLine, *modeResize)
			if err != nil {
				return errors.Wrapf(err, "could not proceed object erase according to the received polyline")
			}

			if down - up < right - left {
				img = meta.RotateClock(img)
				img = meta.RotateClock(img)
				img = meta.RotateClock(img)
			}

			return printImage(img, initImg, *outputPath)
		},
	}
	return command
}

func proceedObjectErase(img image.Image, noPixelsToErase int, polyLine []meta.Point, mode string) (image.Image, error) {
	imgGray := meta.GetGrayImage(img)
	magnitude := meta.SobelFilter(imgGray)

	for x := range magnitude {
		for y := range magnitude[x] {
			if insidePolyLine(x, y, polyLine) {
				magnitude[x][y] = -10000000
			}
		}
	}

	if err := meta.PrintMagnitude(magnitude, "test_magnitude.jpeg"); err != nil {
		return nil, errors.Wrapf(err, "could not print magnitude image")
	}

	for i := 0; i < noPixelsToErase; i++ {
		vertical := findOneVertical(magnitude, mode)
		img, magnitude = deleteVertical(vertical, img, magnitude)
	}
	return img, nil
}

func insidePolyLine(x int, y int, polyLine []meta.Point) bool {
	for i := 2; i < len(polyLine); i ++ {
	//	Check if the point is always on the same part of the edge
		dir1 := (x - polyLine[i - 2].X) * (y - polyLine[i - 1].Y) -
			(y - polyLine[i - 2].Y) * (x - polyLine[i - 1].X)
		dir2 := (x - polyLine[i - 1].X) * (y - polyLine[i].Y) -
			(y - polyLine[i - 1].Y) * (x - polyLine[i].X)

		if dir1 * dir2 < 0 {
			return false
		}
	}
	return true
}
