package cmd

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"strconv"

	"github.com/nfnt/resize"

	"computer_vision/lib"
)

func AmplificationImageContent() *cobra.Command {
	short := "Decrease the number of pixels in height while keeping the same content of interest. "
	var command = &cobra.Command{
		Use: "amplification <image path> <procent increase factor>",
		Short: short,
		Long: short + "Example usage 'amplification data/praga.jpg 20' will increase the size with 20% and delete after the same number of added pixels for an amplification of the content.",
		Args: cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			imgPath := args[0]
			img, err := meta.GetImageFromPath(imgPath)
			if err != nil {
				return errors.Wrapf(err, "could not get an image obj from path '%v'", imgPath)
			}
			initImg := img

			factorAmp, err := strconv.Atoi(args[1])
			if err != nil {
				return errors.Wrapf(err, "could not parse as integer arg received '%v'", args[1])
			}

			surpDimX := img.Bounds().Dx() * factorAmp / 100
			surpDimY := img.Bounds().Dy() * factorAmp / 100

			img = resize.Resize(uint(img.Bounds().Dx() + surpDimX), uint(img.Bounds().Dy() + surpDimY), img, resize.Lanczos3)

			img, err = proceedErase(img, surpDimX, surpDimY, *modeResize)
			if err != nil {
				return errors.Wrapf(err, "failed to process the erase of %vx%v pixels", surpDimX, surpDimY)
			}

			return printImage(img, initImg, *outputPath)
		},
	}
	return command
}
