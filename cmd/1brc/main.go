package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"sync"
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
	fmt.Printf("Start and End of the chunk after alignement: %v\n", fChunk)

	return &collectedData, nil
}

func alignChunkBoundaries(f *os.File, offset int64, jump int) (int64, error) {
	_, err := f.Seek(offset, 0)
	if err != nil {
		return -1, fmt.Errorf("unable to change the file offset to %d. Err: %v", offset, err)
	}

	bu := bufio.NewReader(f)
	bs := make([]byte, 100)
	_, err = bu.Read(bs)
	if err != nil {
		return -1, fmt.Errorf("unable to read byte from the reader. Err: %v", err)
	}
	// start from a newline
	s := bytes.IndexRune(bs, '\n')
	if s == -1 {
		return -1, fmt.Errorf("unable to find newline while reading it offset: %d", offset)
	}
	// index start from zero and I need to start from the next byte that's why adding +1 always
	return int64(s + 1 + jump), nil
}
