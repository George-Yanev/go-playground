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
	"runtime/pprof"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// const fileName = "/Users/gy/developers/go-fun/measurements.txt"
const fileName = "/Users/gy/developers/github.com/George-Yanev/1brc/measurements.txt"

type record struct {
	name   string
	values []float32
	min    float32
	max    float32
	mean   float32
}

func (r *record) calculateMean() float32 {
	var sum float32
	for _, v := range r.values {
		sum += v
	}
	return sum / float32(len(r.values))
}

func (r record) String() string {
	return fmt.Sprintf("%s=%.2f/%.2f/%.2f", r.name, r.min, r.max, r.mean)
}

func main() {
	// CPUs to use
	cpus := runtime.GOMAXPROCS(runtime.NumCPU() - 1)

	// Start CPU profiling
	cpuProfile, err := os.Create("cpu.prof")
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	defer cpuProfile.Close() // Ensure the file is closed after profiling

	if err := pprof.StartCPUProfile(cpuProfile); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}
	defer pprof.StopCPUProfile()

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
	b := int64(1024) * 64
	chunks := fileSize / b

	var wg sync.WaitGroup
	results := make([]record, 0, chunks)
	for i := 0; i < 2*cpus; i++ {
		start := int64(i) * int64(b)
		wg.Add(1)
		go func() {
			defer wg.Done()
			results = append(results, readChunk(start, b, dh)...)
		}()
	}
	wg.Wait()

	sort.Slice(results, func(i, j int) bool { return results[i].name < results[j].name })
	var p record
	var t record
	for i, r := range results {
		// first index, set the total to the r and set p to r
		if i == 0 {
			t = r
			p = r
			continue
		}

		if p.name != r.name {
			// calculate
			slices.Sort(t.values)
			t.min = t.values[0]
			t.max = t.values[len(t.values)-1]
			t.mean = t.calculateMean()

			// print
			fmt.Println(t)

			// set total starting with the new record
			t = r
		} else {
			t.values = append(t.values, r.values...)
		}
		p = r
	}

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
	if end == fileStat.Size() {
		return end, nil
	}

	bufferSize := 1024 * 64
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

func readChunk(start int64, size int64, dh *os.File) []record {
	records := make([]record, 0, 10000)

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
		city_name := parts[0]
		temp, err := strconv.ParseFloat(parts[1], 32)
		if err != nil {
			fmt.Println("Cannot convert string temp to float. Exit", err)
			os.Exit(1)
		}
		t := float32(temp)
		i := sort.Search(len(records), func(i int) bool { return records[i].name == city_name })
		if i < len(records) && records[i].name == city_name {
			// append to values
			records[i].values = append(records[i].values, t)
		} else {
			records = append(records[:i], append([]record{{name: city_name, values: []float32{t}}}, records[i:]...)...)

		}
	}
	return records
}
