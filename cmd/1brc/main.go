package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
)

var measurementFile = "measurements.txt"
var numReaders = 5

type FileChunk struct {
	Start int64
	End   int64
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
	fmt.Printf("Chunks to read: %v\n", fChunks)

}

func readers(fChunk FileChunk, readyCh <-chan bool) (*CalculatedPoints, error) {
	collectedData := make(CalculatedPoints, 0, 10000)
	f, err := os.Open(measurementFile)
	if err != nil {
		log.Fatalf("Unable to read file: %v", err)
	}
	for i, v := range []int64{fChunk.End, fChunk.Start} {
		// End should end at '\n'. Start will need to start at the next line
		r, err := alignChunkBoundaries(f, v, i)
		if err != nil {
			return nil, fmt.Errorf("unable to align the chunk. Offset: %d. Err: %v", v, err)
		}
	}

	return &collectedData, nil
}

func alignChunkBoundaries(f *os.File, offset int64, jump int) (int64, error) {
	_, err := f.Seek(offset, 0)
	if err != nil {
		return -1, fmt.Errorf("unable to change the file offset to %d. Err: %v", offset, err)
	}

	bu := bufio.NewReader(f)
	bs := make([]byte, 0, 30)
	_, err = bu.Read(bs)
	if err != nil {
		return -1, fmt.Errorf("unable to read byte from the reader. Err: %v", err)
	}
	// start from a newline
	s := bytes.IndexRune(bs, '\n')
	if s == -1 {
		return -1, fmt.Errorf("unable to find newline while readying it at Start: %d", offset)
	}
	// index start from zero and I need to start from the next byte that's why adding +1 always
	return int64(s + 1 + jump), nil
}
