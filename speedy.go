// Speedy Core

package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	// "strconv"
	"runtime"
	"strings"
)

// Slices that contain the IOCs and Rules
var (
	filenameIOCComments map[*regexp.Regexp]string
	counter             int
)

// SpeedyCore is the main class that runs all checks
type SpeedyCore struct {
	maxSize int
	debug   bool
	path    string
}

// initializeSpeedy intitializes some values
func (s *SpeedyCore) initialize() {
	// Filename IOCs
	filenameIOCComments = make(map[*regexp.Regexp]string)
}

// runFileScan performs a file system scan
func (s *SpeedyCore) runFileScan() {
	fmt.Println("Starting scan ...")
	err := filepath.Walk(s.path, s.scanFile)
	// Result -------------------------------------------------------------------
	if err != nil {
		fmt.Printf("filepath.Walk() returned %v\n", err)
	}
}

func (s *SpeedyCore) scanFile(filePath string, f os.FileInfo, err error) (err_scan error) {
	// File Name Checks ------------------------------------------------------
	for regex, comment := range filenameIOCComments {
		match := regex.FindString(filePath)
		if match != "" {
			fmt.Printf("MATCH REGEX: %s COMMENT: %s", regex.String(), comment)
		}
	}

	// Memory profile
	memStats := &runtime.MemStats{}
	runtime.ReadMemStats(memStats)

	if counter%1000 == 0 {
		fmt.Println((memStats.Alloc/1024)/1024, " MB")
	}
	counter += 1

	return err_scan
}

// readCSVIOCFile reads IOCs from a Simple IOC file (CSV)
func (s *SpeedyCore) processCSVIOC() {

	fmt.Println("Processing IOC file ...")
	fileReader, err := os.Open("./filename-iocs.txt")
	if err != nil {
		fmt.Sprintf("Cannot open Filename IOC file %s", err)
	}

	// Create a new CSV reader
	r := bufio.NewReader(fileReader)

	// Predefined comment string
	var comment string = ""
	for {
		line, err := r.ReadString('\n')
		line = strings.TrimSuffix(line, "\n")
		line = strings.TrimSuffix(line, "\r")
		record := strings.Split(line, ";")

		// Stop at EOF
		if err == io.EOF || len(record) < 1 {
			break
		}

		// Save comment from comment line
		if strings.Index(record[0], "# ") == 0 {
			elements := strings.Split(record[0], "# ")
			if len(elements) > 1 {
				comment = elements[1]
			}
		} else if strings.Index(record[0], "#") == 0 {
			elements := strings.Split(record[0], "#")
			if len(elements) > 1 {
				comment = elements[1]
			}
		}

		// Not a value line of CSV IOC files
		if len(record) < 2 {
			continue
		}

		// First column is regex
		regex_string := fmt.Sprintf(`%s`, record[0])
		// Type 1 - second column contains score
		// score, err_score := strconv.ParseInt(record[1], 10, 32)

		// Compile Regex
		regex, err_regex := regexp.Compile(regex_string)
		if err_regex != nil {
			fmt.Println("Error compiling regex %s", regex_string)
		} else {
			// Create a filename IOC
			filenameIOCComments[regex] = comment
		}
	}
	// Info
	fmt.Printf("%d file name IOCs processed\n", len(filenameIOCComments))
}

// main
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: speedy scanpath")
		os.Exit(1)
	}
	speedy := SpeedyCore{path: os.Args[1]}
	speedy.initialize()
	speedy.processCSVIOC()
	speedy.runFileScan()
}
