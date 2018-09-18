package autokeepout

// Kernel attributes, kernel size should always be odd and offset is the always kernel size minus
// one divide by two.
const (
	KernelSize = 5
	Offset     = (KernelSize - 1) / 2
)

// GaussKernel is used for applying Gaussian blur to an image.
var GaussKernel = [][]float64{
	{2.0, 4.0, 5.0, 4.0, 2.0},
	{4.0, 9.0, 12.0, 9.0, 4.0},
	{5.0, 12.0, 15.0, 12.0, 5.0},
	{4.0, 9.0, 12.0, 9.0, 4.0},
	{2.0, 4.0, 5.0, 4.0, 2.0},
}

func init() {
	var norm float64
	for i := 0; i < len(GaussKernel); i++ {
		for j := 0; j < len(GaussKernel[i]); j++ {
			norm += GaussKernel[i][j]
		}
	}

	for i := 0; i < len(GaussKernel); i++ {
		for j := 0; j < len(GaussKernel[i]); j++ {
			GaussKernel[i][j] /= norm
		}
	}
}

// GaussianMask applies Gaussian blur to an image matrix.
func GaussianMask(mat [][]float64) [][]float64 {
	maskedGrid := make([][]float64, len(mat))
	for i := 0; i < len(mat); i++ {
		maskedGrid[i] = make([]float64, len(mat[i]))
		for j := 0; j < len(mat[i]); j++ {
			maskedGrid[i][j] = gaussFilter(mat, i, j)
		}
	}

	return maskedGrid
}

// ParallelGaussianMask applies Gaussian blur to an image matrix using multiple subroutines to
// achieve parallelism.
func ParallelGaussianMask(mat [][]float64, numRoutines int) [][]float64 {
	rowsPerRoutine := len(mat) / numRoutines
	outputChan := make(chan *submask, numRoutines)

	n := 0
	for n < numRoutines-1 {
		go getGaussSubmask(mat, n, n*rowsPerRoutine, (n+1)*rowsPerRoutine, outputChan)
		n++
	}

	go getGaussSubmask(mat, n, n*rowsPerRoutine, len(mat), outputChan)

	n = 0
	submasks := make([]*submask, numRoutines)
	for submask := range outputChan {
		submasks[submask.Order] = submask
		n++

		if n == numRoutines {
			break
		}
	}

	mask := [][]float64{}
	for _, submask := range submasks {
		mask = append(mask, submask.Values...)
	}

	return mask
}

type submask struct {
	Order    int
	StartRow int
	Values   [][]float64
}

// getGaussSubmask is called in the optimized version of gaussian masking. It is called in
// multiple go routines to achieve parallel convolution operations.
func getGaussSubmask(mat [][]float64, n, startRow, endRow int, output chan *submask) {
	rowSize := endRow - startRow
	values := make([][]float64, rowSize)
	for i := 0; i < rowSize; i++ {
		colSize := len(mat[startRow+i])
		values[i] = make([]float64, colSize)
		for j := 0; j < colSize; j++ {
			values[i][j] = gaussFilter(mat, startRow+i, j)
		}
	}

	output <- &submask{
		Order:    n,
		StartRow: startRow,
		Values:   values,
	}
}

func gaussFilter(mat [][]float64, y, x int) float64 {
	return convolve(mat, y, x, 5, GaussKernel)
}
