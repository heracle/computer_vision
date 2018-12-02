package cmd

import (
	"computer_vision/lib"
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"image/jpeg"
	"os"
	"strconv"
)

var (
	alphaTexture = pflag.Float64("alpha-texture", 0.8, "This float is used for doing a weight sum between the real error of overlap and the difference between the initial image.")
	stepsTexture = pflag.IntP("steps", "s", 1, "This int is representing the number of steps of adding the texture (the previous resulted image) to the input image.")
	)

func AddTextureToImage() *cobra.Command {
	short := "Change the texture of an image."
	var command = &cobra.Command{
		Use: "add_texture <image path> <image texture>",
		Short: short,
		Args: cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			imgPath := args[0]
			img, err := meta.GetImageFromPath(imgPath)
			if err != nil {
				return errors.Wrapf(err, "could not get an image obj from path '%v'", imgPath)
			}

			imgPathTexture := args[1]
			imgTexture, err := meta.GetImageFromPath(imgPathTexture)
			if err != nil {
				return errors.Wrapf(err, "could not get an image obj from path '%v'", imgPathTexture)
			}

			for step := 0; step < *stepsTexture; step++ {
				fmt.Printf("begin step %v\n", step)

				blocks, err := getRandomBlocks(imgTexture, *noRandomBlocks, *lenBlockSquare, *lenOverlapSquares, *distanceFromBorder)
				if err != nil {
					return errors.Wrapf(err, "could not get the random blocks")
				}

				resultImg, err := createImage(
					blocks,
					img.Bounds().Dx(),
					img.Bounds().Dy(),
					*lenOverlapSquares,
					*alphaTexture,
					img,
				)
				if err != nil {
					return errors.Wrapf(err, "could not create the image from blocks")
				}

				// Set the resulted image as the texture for the future step.
				imgTexture = resultImg

				outFileName := strconv.Itoa(step) + (*outputPath)
				fmt.Printf("%v\n", outFileName)

				outFile, err := os.OpenFile(outFileName, os.O_WRONLY|os.O_CREATE, 0744)
				if err != nil {
					return errors.Wrapf(err, "could not create file at path '%v'", outFileName)
				}

				if err := jpeg.Encode(outFile, resultImg, nil); err != nil {
					return errors.Wrapf(err, "could not encode image in jpeg format")
				}
				outFile.Close()
				fmt.Printf("finished step %v\n", step)
			}
			return nil
		},
	}
	return command
}