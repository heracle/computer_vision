package cmd

import (
	"computer_vision/lib"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"image"
	"image/jpeg"
	"math"
	"math/rand"
	"os"
	"sort"
	"strconv"
)

var (
	outputPath = pflag.StringP("output", "o", "result.jpeg", "The path where to save the output jpeg picture.")
	noRandomBlocks = pflag.Int("no-blocks", 10000, "The number of random blocks which will fill the new image.")
	lenBlockSquare = pflag.Int("len-block-square", 36, "The number of pixels in length of each block square.")
	lenOverlapSquares = pflag.Int("len-overlap-blocks", 6, "The number of pixels in length representing the overlap between two consecutive blocks.")
	distanceFromBorder = pflag.Int("distance-border", 0, "The minimum distance of the random blocks from the border of the initial image.")
	)

type blockObj struct {
	complete image.RGBA
	xMin     [][]float64
	xMax     [][]float64
	yMin     [][]float64
	yMax     [][]float64
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

			resultImg, err := createImage(
				blocks,
				int(factorAmp * float64(img.Bounds().Dx())),
				int(factorAmp * float64(img.Bounds().Dy())),
				*lenOverlapSquares,
				)
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
		blocks[blockIndex].xMin = meta.GetGrayImage(defineBlockPart(up, left, overlap, sizeBlock, img))
		blocks[blockIndex].yMin = meta.GetGrayImage(defineBlockPart(up, left, sizeBlock, overlap, img))
		blocks[blockIndex].xMax = meta.GetGrayImage(defineBlockPart(up + sizeBlock - overlap, left, overlap, sizeBlock, img))
		blocks[blockIndex].yMax = meta.GetGrayImage(defineBlockPart(up, left + sizeBlock - overlap, sizeBlock, overlap, img))
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
			leftBlock = addBlockToImage(
				x,
				y,
				blockSize,
				imageBlockIndexPreviousLine[lenIndex],
				leftBlock,
				blocks,
				retImg,
				)
			imageBlockIndexPreviousLine[lenIndex] = leftBlock
			y += blockSize - overlap
			lenIndex++
		}
		x += blockSize - overlap
	}

	return retImg, nil
}

type pair struct {
	index int
	error float64
}

func addBlockToImage(
	xStart int,
	yStart int,
	blockSize int,
	upLastBlock int,
	leftLastBlock int,
	blocks []blockObj,
	img *image.RGBA,
	) int {
	if upLastBlock == -1 && leftLastBlock == -1 {
		firstBlock := rand.Intn(len(blocks))
		for x := 0; x < blockSize; x++ {
			for y := 0; y < blockSize; y++ {
				(*img).Set(x, y, blocks[firstBlock].complete.At(x, y))
			}
		}
		return firstBlock
	}

	minError := float64(math.MaxFloat64)
	minBlock := -1

	possibleBlocks := make([]pair, len(blocks))

	for indexBlock := 0; indexBlock < len(blocks); indexBlock++ {
		actualError := float64(0)
		if upLastBlock != -1 {
			for x := 0; x < len(blocks[upLastBlock].xMax); x++ {
				for y := 0; y < len(blocks[upLastBlock].xMax[x]); y++ {
					dif := blocks[upLastBlock].xMax[x][y] - blocks[indexBlock].xMin[x][y]
					actualError += dif*dif
				}
			}
		}
		if leftLastBlock != -1 {
			for x := 0; x < len(blocks[leftLastBlock].yMax); x++ {
				for y := 0; y < len(blocks[leftLastBlock].yMax[x]); y++ {
					dif := blocks[leftLastBlock].yMax[x][y] - blocks[indexBlock].yMin[x][y]
					actualError += dif*dif
				}
			}
		}
		if actualError < minError {
			minError = actualError
			minBlock = indexBlock
		}

		possibleBlocks[indexBlock] = pair{index: indexBlock, error: actualError}
	}

	sort.Slice(possibleBlocks, func(i, j int) bool {
		return possibleBlocks[i].error < possibleBlocks[j].error
	})

	foundOkBlocks := 0
	for foundOkBlocks < len(blocks) && possibleBlocks[foundOkBlocks].error <= 1.1 * minError {
		foundOkBlocks++
	}

	minBlock = possibleBlocks[rand.Intn(foundOkBlocks)].index

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

func findHorizontallySplit(img1 [][]float64, img2 [][]float64) []int {
	horizontal := findVerticallySplit(rotateClock(img1), rotateClock(img2))

	horizontalRev := make([]int, len(horizontal))
	for i := 0; i < len(horizontal); i++ {
		horizontalRev[i] = len(img1) - horizontal[i] - 1
	}

	return horizontalRev
}

func rotateClock(src [][]float64) [][]float64 {
	ret := make([][]float64, len(src[0]))
	for i := 0; i < len(src[0]); i++ {
		ret[i] = make([]float64, len(src))
	}

	for x := 0; x < len(src); x ++ {
		for y := 0; y < len(src[0]); y++ {
			ret[y][len(src) - 1 - x] = src[x][y]
		}
	}

	return ret
}

func findVerticallySplit(img1 [][]float64, img2 [][]float64) []int {
	dyn := make([][]float64, len(img1))
	frm := make([][]int, len(img1))
	for x := 0; x < len(img1); x++ {
		dyn[x] = make([]float64, len(img1[0]))
		frm[x] = make([]int, len(img1[0]))
	}

	for y := 0; y < len(img1[0]); y++ {
		dyn[0][y] = (img1[0][y] - img2[0][y]) * (img1[0][y] - img2[0][y])
	}

	for x := 1; x < len(img1); x++ {
		for y := 0; y < len(img1[0]); y++ {
			dyn[x][y] = dyn[x - 1][y]
			frm[x][y] = y
			if y != 0 && dyn[x - 1][y - 1] < dyn[x][y] {
				dyn[x][y] = dyn[x - 1][y - 1]
				frm[x][y] = y - 1
			}
			if y != len(img1[0]) - 1 && dyn[x - 1][y + 1] < dyn[x][y] {
				dyn[x][y] = dyn[x - 1][y + 1]
				frm[x][y] = y + 1
			}
			dyn[x][y] += (img1[x][y] - img2[x][y]) * (img1[x][y] - img2[x][y])
		}
	}

	lastP := 0
	for y := 0; y < len(img1[0]); y ++ {
		if dyn[len(img1) - 1][y] < dyn[len(img1) - 1][lastP] {
			lastP = y
		}
	}

	vertical := []int{lastP}

	for x := len(img1) - 1; x > 0; x -- {
		vertical = append([]int{frm[x][lastP]}, vertical...)
		lastP = frm[x][lastP]
	}
	return vertical
}
