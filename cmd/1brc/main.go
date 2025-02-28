package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

var measurementFile = "measurements.txt"
var numReaders = 5

type FileChunk struct {
	Start        int64
	End          int64
	SkipStartEnd int64
}

type FileChunks []FileChunk

type MeasurementPoint struct {
	City string
	Temp int
}

type CalculatedPoint struct {
	Max   int
	Min   int
	Sum   int
	Count int
}

type CalculatedPoints []CalculatedPoint

func main() {
	var wg sync.WaitGroup
	// collectedData := make(map[string]CalculatedPoint)
	f, err := os.Open(measurementFile)
	if err != nil {
		log.Fatalf("Unable to read file: %v", err)
	}

	fStat, err := f.Stat()
	if err != nil {
		log.Fatalf("Unable to read file stat: %v", err)
	}
	fSize := fStat.Size()
	chunksSize := fSize / int64(numReaders)

	fChunks := make(FileChunks, 0, numReaders)
	for i := int64(0); i < fSize; i += chunksSize + 1 {
		fChunks = append(fChunks, FileChunk{
			Start: i,
			End:   i + chunksSize,
		})
	}
	// ensure the last FileChunk end is correct
	fChunks[len(fChunks)-1].End = fSize
	fChunks[len(fChunks)-1].SkipStartEnd = fSize
	fmt.Printf("Chunks to read: %v\n", fChunks)

	wg.Add(numReaders)
	for _, f := range fChunks {
		go reader(f, &wg)
	}
	wg.Wait()

}

func reader(fChunk FileChunk, wg *sync.WaitGroup) (*CalculatedPoints, error) {
	defer wg.Done()

	collectedData := make(CalculatedPoints, 0, 10000)
	f, err := os.Open(measurementFile)
	if err != nil {
		return nil, fmt.Errorf("Unable to read file: %v", err)
	}
	for i, v := range []int64{fChunk.End, fChunk.Start} {
		if v != fChunk.SkipStartEnd {
			// End should end at '\n'. Start will need to start at the next line
			r, err := alignChunkBoundaries(f, v, i)
			if err != nil {
				fmt.Printf("unable to align the chunk. Offset: %d. Err: %v\n", v, err)
				return nil, fmt.Errorf("unable to align the chunk. Offset: %d. Err: %v", v, err)
			}
			if i == 0 {
				fChunk.End = r
			} else {
				fChunk.Start = r
			}
		}
	}
	fmt.Printf("Start and End of the chunk after alignement: %v\n", fChunk)

	return &collectedData, nil
}

func alignChunkBoundaries(f *os.File, offset int64, jump int) (int64, error) {
	// start offset require starting one byte before because we might be perfectly aligned
	// and because of this we will skip a line
	seekOffset := offset - int64(jump)
	if _, err := f.Seek(seekOffset, io.SeekStart); err != nil {
		return -1, fmt.Errorf("unable to change the file offset to %d. Err: %v", offset, err)
	}

	reader := bufio.NewReader(f)
	line, err := reader.ReadBytes('\n')
	if err != nil && err != io.EOF {
		return 0, fmt.Errorf("read from %d: %v", seekOffset, err)
	}
	// index start from zero and I need to start from the next byte that's why adding +1 always
	return int64(seekOffset + int64(len(line)-1) + int64(jump)), nil
}
