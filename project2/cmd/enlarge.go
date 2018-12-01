package cmd

import (
	"computer_vision/lib"
	"image"
	"image/color"
	"image/jpeg"
	"math"
	"math/rand"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"strconv"
)

var (
	outputPath = pflag.StringP("output", "o", "result.jpeg", "The path where to save the output jpeg picture.")
	noRandomBlocks = pflag.Int("no-blocks", 5000, "The number of random blocks which will fill the new image.")
	lenBlockSquare = pflag.Int("len-block-square", 36, "The number of pixels in length of each block square.")
	lenOverlapSquares = pflag.Int("len-overlap-blocks", 6, "The number of pixels in length representing the overlap between two consecutive blocks.")
	distanceFromBorder = pflag.Int("distance-border", 0, "The minimum distance of the random blocks from the border of the initial image.")
	)

type blockObj struct {
	complete image.RGBA
	xMin     image.RGBA
	xMax     image.RGBA
	yMin     image.RGBA
	yMax     image.RGBA
}

func EnlargeImage() *cobra.Command {
	short := "Enlarge the image by multiplying the content."
	var command = &cobra.Command{
		Use: "enlarge <image path> <percent increase factor>",
		Short: short,
		Long: short + "Example usage 'enlarge data/prague.jpg 3.5' will increase both length and width with 3.5 of the initial size.",
		Args: cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			imgPath := args[0]
			img, err := meta.GetImageFromPath(imgPath)
			if err != nil {
				return errors.Wrapf(err, "could not get an image obj from path '%v'", imgPath)
			}

			factorAmp, err := strconv.ParseFloat(args[1], 64)
			if err != nil {
				return errors.Wrapf(err, "could not parse as integer arg received '%v'", args[1])
			}

			blocks, err := getRandomBlocks(img, *noRandomBlocks, *lenBlockSquare, *lenOverlapSquares, *distanceFromBorder)
			if err != nil {
				return errors.Wrapf(err, "could not get the random blocks")
			}

			resultImg, err := createImage(blocks, int(factorAmp * float64(img.Bounds().Dx())), int(factorAmp * float64(img.Bounds().Dy())), *lenOverlapSquares)
			if err != nil {
				return errors.Wrapf(err, "could not create the image from blocks")
			}

			outFile, err := os.OpenFile(*outputPath, os.O_WRONLY|os.O_CREATE, 0600)
			if err != nil {
				return errors.Wrapf(err, "could not create file at path '%v'", *outputPath)
			}
			defer outFile.Close()

			return jpeg.Encode(outFile, resultImg, nil)
		},
	}
	return command
}

func getRandomBlocks(img image.Image, noBlocks int, sizeBlock int, overlap int, distanceBorder int) ([]blockObj, error) {
	blocks := make([]blockObj, noBlocks)

	for blockIndex := 0; blockIndex < noBlocks; blockIndex++ {
		up := rand.Intn(img.Bounds().Dx() - sizeBlock - 2 * distanceBorder) + distanceBorder
		left := rand.Intn(img.Bounds().Dy() - sizeBlock - 2* distanceBorder) + distanceBorder

		blocks[blockIndex].complete = defineBlockPart(up, left, sizeBlock, sizeBlock, img)
		blocks[blockIndex].xMin = defineBlockPart(up, left, overlap, sizeBlock, img)
		blocks[blockIndex].yMin = defineBlockPart(up, left, sizeBlock, overlap, img)
		blocks[blockIndex].xMax = defineBlockPart(up + sizeBlock - overlap, left, overlap, sizeBlock, img)
		blocks[blockIndex].yMax = defineBlockPart(up, left + sizeBlock - overlap, sizeBlock, overlap, img)
	}
	return blocks, nil
}

func defineBlockPart(up int, left int, width int, length int, img image.Image) image.RGBA{
	ret := *image.NewRGBA(image.Rect(0, 0, width, length))
	for x := 0; x < width; x++ {
		for y := 0; y < length; y++ {
			ret.Set(x, y, img.At(up + x, left + y))
		}
	}
	return ret
}

func createImage(blocks []blockObj, width int, length int, overlap int) (image.Image, error){
	retImg := image.NewRGBA(image.Rect(0,0, width, length))
	blockSize := blocks[0].complete.Rect.Dx()

	imageBlockIndexPreviousLine := make([]int, width)
	// Initiate all fields with -1 to say that there is no block above our position.
	for i := 0; i < len(imageBlockIndexPreviousLine); i++ {
		imageBlockIndexPreviousLine[i] = -1
	}

	x := 0
	y := 0
	for x < width {
		y = 0
		lenIndex := 0
		leftBlock := -1
		for y < length {
			leftBlock = addBlockToImage(x, y, blockSize, imageBlockIndexPreviousLine[lenIndex], leftBlock, blocks, retImg)
			imageBlockIndexPreviousLine[lenIndex] = leftBlock
			y += blockSize - overlap
			lenIndex++
		}
		x += blockSize - overlap
	}

	return retImg, nil
}

func addBlockToImage(xStart int, yStart int, blockSize int, upLastBlock int, leftLastBlock int, blocks []blockObj, img *image.RGBA) int {
	if upLastBlock == -1 && leftLastBlock == -1 {
		firstBlock := rand.Intn(len(blocks))
		for x := 0; x < blockSize; x++ {
			for y := 0; y < blockSize; y++ {
				(*img).Set(x, y, blocks[firstBlock].complete.At(x, y))
			}
		}
		return firstBlock
	}

	minError := float64(math.MaxInt32)
	minBlock := -1

	for indexBlock := 0; indexBlock < len(blocks); indexBlock++ {
		actualError := float64(0)
		if upLastBlock != -1 {
			grayXMax := meta.GetGrayImage(blocks[upLastBlock].xMax)
			grayXMin := meta.GetGrayImage(blocks[indexBlock].xMin)

			for x := 0; x < len(grayXMax); x++ {
				for y := 0; y < len(grayXMax[x]); y++ {
					actualError += math.Abs(grayXMax[x][y] - grayXMin[x][y])
				}
			}

			//for i := 0; i < len(blocks[upLastBlock].xMax.Pix); i++ {
			//	actualError += sqDiffUInt8(blocks[upLastBlock].xMax.Pix[i], blocks[indexBlock].xMin.Pix[i])
			//}
		}
		if leftLastBlock != -1 {
			grayYMax := meta.GetGrayImage(blocks[leftLastBlock].yMax)
			grayYMin := meta.GetGrayImage(blocks[indexBlock].yMin)

			for x := 0; x < len(grayYMax); x++ {
				for y := 0; y < len(grayYMax[x]); y++ {
					actualError += math.Abs(grayYMax[x][y] - grayYMin[x][y])
				}
			}

			//for i := 0; i < len(blocks[leftLastBlock].yMax.Pix); i++ {
			//	actualError += sqDiffUInt8(blocks[leftLastBlock].yMax.Pix[i], blocks[indexBlock].yMin.Pix[i])
			//}
		}
		if actualError < minError {
			minError = actualError
			minBlock = indexBlock
		}
	}
	var verticallySplit []int
	var horizontallySplit []int

	if leftLastBlock != -1 {
		verticallySplit = findVerticallySplit(blocks[leftLastBlock].yMax, blocks[minBlock].yMin)
	} else {
		verticallySplit = emptySplitSlice(blockSize)
	}
	if upLastBlock != -1 {
		horizontallySplit = findHorizontallySplit(blocks[upLastBlock].xMax, blocks[minBlock].xMin)
	} else {
		horizontallySplit = emptySplitSlice(blockSize)
	}

	for x := 0; x < blockSize; x++ {
		for y := verticallySplit[x] + 1; y < blockSize; y++ {
			if x <= horizontallySplit[y] {
				continue
			}
			(*img).Set(xStart + x, yStart + y, blocks[minBlock].complete.At(x, y))
		}
	}

	return minBlock
}

func emptySplitSlice(len int) []int {
	ret := make([]int, len)
	for i := 0; i < len; i++ {
		ret[i] = -1
	}
	return ret
}

func findHorizontallySplit(img1 image.RGBA, img2 image.RGBA) []int {
	horizontal := findVerticallySplit(rotateClock(img1), rotateClock(img2))

	horizontalRev := make([]int, len(horizontal))
	for i := 0; i < len(horizontal); i++ {
		horizontalRev[i] = img1.Rect.Dx() - horizontal[i] - 1
	}

	return horizontalRev
}

func rotateClock(srcImg image.RGBA) image.RGBA {
	srcDim := srcImg.Bounds()
	dstImage := image.NewRGBA(image.Rect(0, 0, srcDim.Dy(), srcDim.Dx()))

	for x := 0; x < srcDim.Dx(); x ++ {
		for y := 0; y < srcDim.Dy(); y++ {
			dstImage.Set(y, srcDim.Dx() - 1 - x, srcImg.At(x, y))
		}
	}

	return *dstImage
}

func findVerticallySplit(img1 image.RGBA, img2 image.RGBA) []int {
	dyn := make([][]float64, img1.Rect.Dx())
	frm := make([][]int, img1.Rect.Dx())
	for x := 0; x < img1.Rect.Dx(); x++ {
		dyn[x] = make([]float64, img1.Rect.Dy())
		frm[x] = make([]int, img1.Rect.Dx())
	}

	for y := 0; y < img1.Rect.Dy(); y++ {
		dyn[0][y] = getDifferencePixels(img1.At(0, y), img2.At(0, y))
	}

	for x := 1; x < img1.Rect.Dx(); x++ {
		for y := 0; y < img1.Rect.Dy(); y++ {
			dyn[x][y] = dyn[x - 1][y]
			frm[x][y] = y
			if y != 0 && dyn[x - 1][y - 1] < dyn[x][y] {
				dyn[x][y] = dyn[x - 1][y - 1]
				frm[x][y] = y - 1
			}
			if y != img1.Rect.Dy() - 1 && dyn[x - 1][y + 1] < dyn[x][y] {
				dyn[x][y] = dyn[x - 1][y + 1]
				frm[x][y] = y + 1
			}
			dyn[x][y] += getDifferencePixels(img1.At(x, y), img2.At(x, y))
		}
	}

	lastP := 0
	for y := 0; y < img1.Rect.Dy(); y ++ {
		if dyn[img1.Rect.Dx() - 1][y] < dyn[img1.Rect.Dx() - 1][lastP] {
			lastP = y
		}
	}

	vertical := []int{lastP}

	for x := img1.Rect.Dx() - 1; x > 0; x -- {
		vertical = append([]int{frm[x][lastP]}, vertical...)
		lastP = frm[x][lastP]
	}
	return vertical
}

func getDifferencePixels(c1 color.Color, c2 color.Color) float64 {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()
	return math.Sqrt(float64((r1-r2)*(r1-r2) + (g1-g2)*(g1-g2) + (b1-b2)*(b1-b2) + (a1-a2)*(a1-a2)))
}

func sqDiffUInt8(x, y uint8) int64 {
	d := int64(x) - int64(y)
	return d * d
}