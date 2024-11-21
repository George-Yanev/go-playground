package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"sort"
	"strconv"
	"strings"
)

// const fileName = "/Users/gy/developers/go-fun/measurements.txt"
const fileName = "/Users/gy/developers/github.com/George-Yanev/1brc/measurements.txt"

type record struct {
	name string
	min  float32
	max  float32
	mean float32
	n    int
}

func (r record) String() string {
	return fmt.Sprintf("%s=%.2f/%.2f/%.2f", r.name, r.min, r.mean, r.max)
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

	records := make(map[string]record)
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

	ns := bufio.NewScanner(dh)
	for ns.Scan() {
		data := ns.Text()
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
		r, exists := records[city_name]
		if !exists {
			records[city_name] = record{
				name: city_name,
				min:  t,
				max:  t,
				mean: t,
				n:    1,
			}
		} else {
			// change min if needed
			if t < r.min {
				r.min = t
			}
			// change max if needed
			if t > r.max {
				r.max = t
			}
			// calculate mean
			r.mean = (r.mean*float32(r.n)+t)/float32(r.n) + 1

			// increment processed values for this city
			r.n++
		}
	}
	keys := make([]string, 0, len(records))
	for key := range records {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	// fmt.Println(records)

	// Print the map in the desired format directly
	fmt.Print("{")
	first := true
	for _, key := range keys {
		if !first {
			fmt.Print(", ")
		}
		fmt.Print(records[key].String())
		first = false
	}
	fmt.Println("}")
}
