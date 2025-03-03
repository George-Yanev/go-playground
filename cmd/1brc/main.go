package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"runtime/pprof"
	"sort"
	"sync"
)

var measurementFile = "measurements.txt"
var numReaders = 16

type FileChunk struct {
	Start     int64
	End       int64
	SkipStart int64
	SkipEnd   int64
}

type FileChunks []FileChunk

type MeasurementPoint struct {
	City string
	Temp int
}

type CalculatedPoint struct {
	Max, Min, Sum, Count int
}

type FinalPoint struct {
	Max, Min, Mean int
}

type CalculatedPoints map[string]CalculatedPoint

type Result struct {
	Data *CalculatedPoints
	Err  error
}

func main() {
	var wg sync.WaitGroup

	f, err := os.Create("cpu.prof")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	// collectedData := make(map[string]CalculatedPoint)
	f, err = os.Open(measurementFile)
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

	fChunks[0].SkipStart = 0
	// ensure the last FileChunk end is correct
	fChunks[len(fChunks)-1].End = fSize
	fChunks[len(fChunks)-1].SkipEnd = fSize
	fmt.Printf("Chunks to read: %v\n", fChunks)

	resultCh := make(chan Result, numReaders)
	wg.Add(numReaders)
	for _, f := range fChunks {
		go reader(f, resultCh, &wg)
	}
	// wg.Wait()
	// close(resultCh)

	finish := numReaders
	finalResult := make(map[string]CalculatedPoint, 10000)
	for data := range resultCh {
		finish -= 1
		if finish == 0 {
			close(resultCh)
		}
		if err != nil {
			log.Fatalf("stop because of an error in a reader goroutine. Err: %v", data.Err)
		}
		for c, p := range *data.Data {
			if _, ok := finalResult[c]; !ok {
				finalResult[c] = p
			} else {
				// need to merge the temp points
				f := finalResult[c]
				f.Count += p.Count
				f.Sum += p.Sum
				if f.Max < p.Max {
					f.Max = p.Max
				}
				if f.Min > p.Min {
					f.Min = p.Min
				}
				finalResult[c] = f
			}
		}
	}
	// sort the result and print it
	cities := make([]string, 0, len(finalResult))
	for k, _ := range finalResult {
		cities = append(cities, k)
	}
	// slices.SortFunc(keys, func(a, b string) int { return strings.Compare(a, b) })
	sort.Strings(cities)
	fmt.Printf("{")
	for _, city := range cities {
		c := finalResult[city]
		fmt.Printf("%s=%.1f/%.1f/%.1f, ", city, float64(c.Min)/10, float64(c.Sum)/float64(c.Count)/10, float64(c.Max)/10)
	}
	fmt.Printf("}")
}

func reader(fChunk FileChunk, resultCh chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()

	collectedData := make(CalculatedPoints, 10000)
	f, err := os.Open(measurementFile)
	if err != nil {
		resultCh <- Result{
			Data: nil,
			Err:  fmt.Errorf("Unable to read file: %v", err),
		}
		return
	}

	if fChunk.Start != fChunk.SkipStart {
		r, err := alignChunkBoundaries(f, fChunk.Start, 1)
		if err != nil {
			fmt.Printf("unable to align the chunk. Offset: %d. Err: %v\n", 1, err)
			resultCh <- Result{
				Data: nil,
				Err:  fmt.Errorf("unable to align the chunk. Offset: %d. Err: %v", 1, err),
			}
			return
		}
		fChunk.Start = r
	}

	if fChunk.End != fChunk.SkipEnd {
		r, err := alignChunkBoundaries(f, fChunk.End, 0)
		if err != nil {
			fmt.Printf("unable to align the chunk. Offset: %d. Err: %v\n", 0, err)
			resultCh <- Result{
				Data: nil,
				Err:  fmt.Errorf("unable to align the chunk. Offset: %d. Err: %v", 0, err),
			}
			return
		}
		fChunk.End = r
	}
	fmt.Printf("Start and End of the chunk after alignement: %v\n", fChunk)
	// let's read :)
	_, err = f.Seek(fChunk.Start, io.SeekStart)
	if err != nil {
		resultCh <- Result{
			Data: nil,
			Err:  fmt.Errorf("cannot put the file seek position to offset: %d. Err: %v", fChunk.Start, err),
		}
		return
	}

	reader := bufio.NewReaderSize(f, 1024*4024)
	pos := fChunk.Start
	for pos < fChunk.End {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			fmt.Printf("unable to read a line from reader: %v\n", err)
		}
		pos += int64(len(line))

		city, tempBytes, hasSemi := bytes.Cut(line, []byte(";"))
		if !hasSemi {
			continue
		}

		newLen := len(tempBytes)
		// first element will never be the dot, skip it
		// last element will never be the dot, skip it
		for i := 1; i < newLen-1; i++ {
			if tempBytes[i] == 46 {
				copy(tempBytes[i:], tempBytes[i+1:])
				newLen--
				break
			}
		}
		if tempBytes[newLen] == 10 {
			newLen--
		}
		tempBytes = tempBytes[:newLen]

		temp, err := parseTempFromBytes(tempBytes)
		if err != nil {
			fmt.Printf("error parsing temp from string to int: %s. Err: %v", city, err)
		}

		if val, ok := collectedData[string(city)]; ok {
			if val.Min > temp {
				val.Min = temp
			}
			if val.Max < temp {
				val.Max = temp
			}
			val.Sum += temp
			val.Count += 1
			collectedData[string(city)] = val
		} else {
			collectedData[string(city)] = CalculatedPoint{
				Min:   temp,
				Max:   temp,
				Sum:   temp,
				Count: 1,
			}
		}
	}

	// fmt.Printf("goroutine collected data %v\n", collectedData)
	resultCh <- Result{
		Data: &collectedData,
		Err:  nil,
	}
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

func parseTempFromBytes(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, fmt.Errorf("data byte is zero\n")
	}

	start := 0
	if data[0] == 45 {
		start = 1
	}

	var result int
	for i := start; i < len(data); i++ {
		result = result*10 + int(data[i]-48)
	}

	if start == 1 {
		result = -result
	}
	return result, nil
}
