package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

// const fileName = "/Users/gy/developers/go-fun/measurements_100k.txt"

const fileName = "/Users/gy/developers/github.com/George-Yanev/1brc/measurements.txt"

type record struct {
	name  string
	min   float32
	max   float32
	mean  float32
	total int
}

func (r record) String() string {
	return fmt.Sprintf("%s=%.2f/%.2f/%.2f", r.name, r.min, r.max, r.mean)
}

func main() {
	// CPUs to use
	cpus := runtime.GOMAXPROCS(runtime.NumCPU() - 1)

	// Start CPU profiling
	// cpuProfile, err := os.Create("cpu.prof")
	// if err != nil {
	// 	log.Fatal("could not create CPU profile: ", err)
	// }
	// defer cpuProfile.Close() // Ensure the file is closed after profiling

	// if err := pprof.StartCPUProfile(cpuProfile); err != nil {
	// 	log.Fatal("could not start CPU profile: ", err)
	// }
	// defer pprof.StopCPUProfile()

	// 1. read the file
	// 2. Create Goroutines that do the following:
	// - read chunk of the file and add to a data structure that can handle it. First variant - 2 routines trying to update the same record.
	// Both are getting it with the same n but at the time of trying to write it the n has changed for one of them so it will read the new value and
	// try to write again and if fail because n has again changed it will try loop until success. Add-on to that approach is
	// leave it in the buffer for later and save it at the end
	// 2. Divide it on chunks based on the cpus (fix at the beginning). Each CPU to read specific part of the file
	// 3. go routines

	dh, err := os.Open(fileName)
	defer dh.Close()
	if err != nil {
		log.Fatal("Cannot read the file: ", err)
	}

	fi, err := dh.Stat()
	if err != nil {
		log.Fatal("Cannot get file info", err)
	}

	fileSize := fi.Size()
	fmt.Println("fileSize is ", fileSize)
	b := int64(1024) * 64
	if b > fileSize {
		b = fileSize
	}
	chunks := fileSize / b
	fmt.Println("chunks are: ", chunks)

	var wg sync.WaitGroup
	jobs := make(chan int64, chunks)
	resultsCh := make(chan map[string][]float32, chunks)
	progress := atomic.Int32{}
	for i := 0; i < cpus; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for start := range jobs {
				// fmt.Println("Start is: ", start)
				chunkRecords := readChunk(start, b, dh)
				// fmt.Println("chunkRecords are ", chunkRecords)
				resultsCh <- chunkRecords

				currentProgress := int32(float32(start) / float32(fileSize) * 100)
				if currentProgress/10 > progress.Load()/10 {
					progress.Store(currentProgress)
					fmt.Printf("%d%%...", currentProgress)
				}
			}
		}()
	}

	for i := 0; i < int(chunks); i++ {
		start := int64(i) * b
		jobs <- start
	}

	var intermediateResults []record
	// fmt.Println("results are: ", results)
	var wg2 sync.WaitGroup
	for i := 0; i < cpus; i++ {
		go func() {
			wg2.Add(1)
			defer wg2.Done()
			for m := range resultsCh {
				for n, r := range m {
					var record record
					var sum float32
					for f := range r {
						sum += float32(f)
					}
					record.name = n
					record.min = r[0]
					record.max = r[len(r)-1]
					record.mean = sum / float32(len(r))
					record.total = len(r)

					intermediateResults = append(intermediateResults, record)
				}
			}
		}()
	}

	close(jobs)
	wg.Wait()
	close(resultsCh)
	wg2.Wait()

	// sort
	sort.Slice(intermediateResults, func(i, j int) bool { return intermediateResults[i].name < intermediateResults[j].name })
	var finalResults []record
	for i := 1; i < len(intermediateResults); i++ {
		c := intermediateResults[i]
		p := intermediateResults[i-1]
		if p.name == c.name {
			// merge in the previous record and make the calculation
			if p.min > c.min {
				p.min = c.min
			}
			if p.max < c.max {
				p.max = c.max
			}
			p.mean = (float32(p.total)*p.mean+c.mean*float32(c.total))/float32(p.total) + float32(c.total)
		} else {
			finalResults = append(finalResults, c)
		}
	}
	// print
	fmt.Println(finalResults)
}

func alignStartOffset(start int64, file *os.File) (int64, error) {
	if start == 0 {
		return 0, nil
	}

	bufferSize := 1024 * 64
	buffer := make([]byte, bufferSize)

	for {
		_, err := file.ReadAt(buffer, start)
		if err != nil {
			log.Fatalf("Error reading file at offset %d: %v", start, err)
		}

		idx := bytes.IndexByte(buffer, '\n')
		if idx != -1 {
			return start + int64(idx) + 1, nil
		}
		return -1, errors.New("Didn't find newline when adjusting the start offset")
	}
}

func alignEndOffset(end int64, file *os.File) (int64, error) {
	fileStat, err := file.Stat()
	if err != nil {
		log.Fatal("Cannot get file stat")
	}
	size := fileStat.Size()
	if end >= size {
		return fileStat.Size(), nil
	}

	// fmt.Println("end is ", end)
	bufferSize := int64(1024) * 64
	if end+int64(bufferSize) > size {
		bufferSize = size - end
	}
	buffer := make([]byte, bufferSize)

	for {
		_, err := file.ReadAt(buffer, end)
		if err != nil {
			log.Fatalf("Error reading file at offset %d: %v", end, err)
		}

		idx := bytes.IndexByte(buffer, '\n')
		if idx != -1 {
			return end + int64(idx), nil
		}
		return -1, errors.New("Didn't find newline when adjusting the start offset")
	}

}

func readChunk(start int64, size int64, dh *os.File) map[string][]float32 {
	dataMap := make(map[string][]float32)

	s, err := alignStartOffset(start, dh)
	if err != nil {
		log.Fatal("Cannot align start offset", err)
	}
	end := start + size
	e, err := alignEndOffset(end, dh)

	newSize := e - s
	buffer := make([]byte, newSize)
	_, err = dh.ReadAt(buffer, s)
	if err != nil && err != io.EOF {
		log.Fatalf("Couldn't ReadAt starting from %d: %v", start, err)
	}
	scanner := bufio.NewScanner(strings.NewReader(string(buffer)))
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		data := scanner.Text()
		parts := strings.Split(data, ";")
		if len(parts) > 2 {
			fmt.Println("Exit as the following records has more than 2 parts", parts)
		}
		cityName := parts[0]
		temp, err := strconv.ParseFloat(parts[1], 32)
		if err != nil {
			fmt.Println("Cannot convert string temp to float. Exit", err)
			os.Exit(1)
		}
		dataMap[cityName] = append(dataMap[cityName], float32(temp))
	}
	return dataMap
}
