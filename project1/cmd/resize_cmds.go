package cmd

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"strconv"
)

var (
	outputPath = pflag.StringP("output", "o", "result.jpeg", "The path where to save the output jpeg picture.")
	modeResize = pflag.StringP("mode", "m", "dynamics", "The mode of erasing one column of pixels.\n1.'dynamics' for doing a dynamic programming approach\n2. 'greedy' for doing a greedy approach\n3. (anything else) for doing a random approach\n")
	maxIncreaseDiv = pflag.Int("max-increase-div", 2, "No more than image_size/<value> pixels will be added in the same time for increasing size commands.")
	)

func DecreaseSizeImage() *cobra.Command {
	var command = &cobra.Command{
		Use: "decrease <image path> <no pixels width> <no pixels height>",
		Short: "Decrease the number of pixels in width and height while keeping the same content of interest.",
		Args: cobra.ExactArgs(3),
		RunE: func(_ *cobra.Command, args []string) error {
			imgPath := args[0]
			img, err := getImageFromPath(imgPath)
			if err != nil {
				return errors.Wrapf(err, "could not get an image obj from path '%v'", imgPath)
			}
			initImg := img

			noPixelsWidthToErase, err := strconv.Atoi(args[1])
			if err != nil {
				return errors.Wrapf(err, "could not parse as integer arg received '%v'", args[1])
			}

			noPixelsHeightToErase, err := strconv.Atoi(args[2])
			if err != nil {
				return errors.Wrapf(err, "could not parse as integer arg received '%v'", args[1])
			}

			img, err = proceedErase(img, noPixelsWidthToErase, noPixelsHeightToErase, *modeResize)
			if err != nil {
				return errors.Wrapf(err, "failed to process the erase of %vx%v pixels", noPixelsWidthToErase, noPixelsHeightToErase)
			}

			return printImage(img, initImg, *outputPath)
		},
	}
	return command
}

func IncreaseSizeImage() *cobra.Command {
	var command = &cobra.Command{
		Use: "increase <image path> <no pixels width> <no pixels height>",
		Short: "Increase the number of pixels in width and height while keeping the same content of interest.",
		Args: cobra.ExactArgs(3),
		RunE: func(_ *cobra.Command, args []string) error {
			imgPath := args[0]
			img, err := getImageFromPath(imgPath)
			if err != nil {
				return errors.Wrapf(err, "could not get an image obj from path '%v'", imgPath)
			}
			initImg := img

			noPixelsWidthToIncrease, err := strconv.Atoi(args[1])
			if err != nil {
				return errors.Wrapf(err, "could not parse as integer arg received '%v'", args[1])
			}

			noPixelsHeightToIncrease, err := strconv.Atoi(args[2])
			if err != nil {
				return errors.Wrapf(err, "could not parse as integer arg received '%v'", args[1])
			}

			for noPixelsWidthToIncrease > 0 {
				maxPixelsErase := img.Bounds().Dy() / *maxIncreaseDiv
				pixelsToErase := noPixelsWidthToIncrease
				if pixelsToErase > maxPixelsErase {
					pixelsToErase = maxPixelsErase
				}
				img, err = processVerticalIncrease(img, pixelsToErase, *modeResize)
				if err != nil {
					return errors.Wrapf(err, "could not process the vertical increase of %v pixels on received image '%v'", args[1], imgPath)
				}
				noPixelsWidthToIncrease -= pixelsToErase
			}

			img = rotateClock(img)

			for noPixelsHeightToIncrease > 0 {
				maxPixelsErase := img.Bounds().Dx() / *maxIncreaseDiv

				pixelsToErase := noPixelsHeightToIncrease
				if pixelsToErase > maxPixelsErase {
					pixelsToErase = maxPixelsErase
				}
				img, err = processVerticalIncrease(img, pixelsToErase, *modeResize)
				if err != nil {
					return errors.Wrapf(err, "could not process the vertical increase of %v pixels on the rotated received image '%v'", args[1], imgPath)
				}
				noPixelsHeightToIncrease -= pixelsToErase
			}

			img = rotateClock(img)
			img = rotateClock(img)
			img = rotateClock(img)

			return printImage(img, initImg, *outputPath)
		},
	}
	return command
}