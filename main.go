package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

// const fileName = "/Users/gy/developers/go-fun/measurements_100k.txt"

const fileName = "/Users/gy/developers/go-fun/measurements_100M.txt"

// const fileName = "/Users/gy/developers/go-fun/measurements.txt"

// const fileName = "/Users/gy/developers/github.com/George-Yanev/1brc/measurements.txt"

var cpus = runtime.NumCPU() - 1

type record struct {
	min, max, mean, sum float32
	count               int
}

func (r record) String() string {
	return fmt.Sprintf("%.2f/%.2f/%.2f", r.min, r.max, r.mean)
}

func main() {
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
	fmt.Println("fileSize is ", fileSize)
	chunksSize := fileSize / int64(cpus)
	fmt.Println("chunksSize are: ", chunksSize)
	type chunk struct {
		start int64
		end   int64
	}
	chunks := make([]chunk, 0, cpus)
	for i := 0; i < cpus; i++ {
		chunks = append(chunks, chunk{
			start: int64(i) * int64(chunksSize),
			end:   int64(i+1) * int64(chunksSize),
		})
	}
	chunks[len(chunks)-1].end = fileSize
	fmt.Println("chunks are ", chunks)
	// os.Exit(1)

	data, err := syscall.Mmap(int(dh.Fd()), 0, int(fileSize), syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		log.Fatalf("Mmap: %v", err)
	}

	defer func() {
		if err := syscall.Munmap(data); err != nil {
			log.Fatalf("Munmap: %v", err)
		}
	}()

	var wg sync.WaitGroup
	var results = make([]map[string]*record, cpus)
	// progress := atomic.Int32{}
	for i, c := range chunks {
		wg.Add(1)
		go func() {
			defer wg.Done()
			chunkRecords := readChunk(data, c.start, c.end)
			// fmt.Println("chunkRecords are ", chunkRecords)
			results[i] = chunkRecords

			// currentProgress := int32(float32(start) / float32(fileSize) * 100)
			// if currentProgress/10 > progress.Load()/10 {
			// 	progress.Store(currentProgress)
			// 	fmt.Printf("%d%%...", currentProgress)
			// }
		}()
	}
	wg.Wait()

	// fmt.Println("results are: ", results)
	finalMap := results[len(results)-1]
	for i := 0; i < len(results)-1; i++ {
		currMap := results[i]
		for s, r := range currMap {
			if fRecord, exists := finalMap[s]; exists {
				if r.min < fRecord.min {
					fRecord.min = r.min
				}
				if r.max > fRecord.max {
					fRecord.max = r.max
				}
				fRecord.count += r.count
				fRecord.sum += r.sum
				fRecord.mean = r.sum / float32(r.count)
				// delete(currMap, s)
			} else {
				finalMap[s] = r
			}
		}

	}

	// get only the names
	var names []string
	for c := range finalMap {
		names = append(names, c)
	}
	sort.Strings(names)

	for _, n := range names {
		fmt.Printf("%s=%s\n", n, finalMap[n])
	}
	// fmt.Println(fResults)
}

func readChunk(data []byte, start, end int64) map[string]*record {
	dataRecords := make(map[string]*record)

	if start != 0 {
		lookAhead := 50
		d := data[start : start+int64(lookAhead)]
		for i, s := range d {
			if s == '\n' {
				start += int64(i + 1)
				break
			}
		}
	}

	fmt.Printf("start - %d, end - %d\n", start, end)
	if end != int64(len(data)) {
		lookAhead := 50
		d := data[end : end+int64(lookAhead)]
		for i, s := range d {
			if s == '\n' {
				end += int64(i)
				break
			}
		}
	}

	var newStart int
	chunkContent := data[newStart:end]
	for i, b := range chunkContent {
		if b == '\n' {
			line := chunkContent[newStart:i]
			ct := strings.Split(string(line), ";")
			city := ct[0]
			temp := ct[1]

			t, err := strconv.ParseFloat(temp, 32)
			if err != nil {
				fmt.Println("Error parsing float:", err)
			}
			tf := float32(t)

			if r, exists := dataRecords[city]; !exists {
				dataRecords[city] = &record{
					sum:   tf,
					count: 1,
					min:   tf,
					max:   tf,
				}
			} else {
				r.sum += tf
				r.count += 1
				if r.min > tf {
					r.min = tf
				}
				if r.max < tf {
					r.max = tf
				}
			}
			newStart = i + 1
		}
	}
	return dataRecords
}
