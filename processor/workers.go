package processor

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
	"sync/atomic"
)

// Taken from https://en.wikipedia.org/wiki/Byte_order_mark#Byte_order_marks_by_encoding
// These indicate that we cannot count the file correctly so we can at least warn the user
var ByteOrderMarks = [][]byte{
	{254, 255},            // UTF-16 BE
	{255, 254},            // UTF-16 LE
	{0, 0, 254, 255},      // UTF-32 BE
	{255, 254, 0, 0},      // UTF-32 LE
	{43, 47, 118, 56},     // UTF-7
	{43, 47, 118, 57},     // UTF-7
	{43, 47, 118, 43},     // UTF-7
	{43, 47, 118, 47},     // UTF-7
	{43, 47, 118, 56, 45}, // UTF-7
	{247, 100, 76},        // UTF-1
	{221, 115, 102, 115},  // UTF-EBCDIC
	{14, 254, 255},        // SCSU
	{251, 238, 40},        // BOCU-1
	{132, 49, 149, 51},    // GB-18030
}

// Check if this file is binary by checking for nul byte and if so bail out
// this is how GNU Grep, git and ripgrep check for binary files
func isBinary(index int, currentByte byte) bool {
	if index < 10000 && !DisableCheckBinary && currentByte == 0 {
		return true
	}

	return false
}

// Check if we have any Byte Order Marks (BOM) in front of the file
func checkBomSkip(fileJob *FileJob) int {
	// UTF-8 BOM which if detected we should skip the BOM as we can then count correctly
	// []byte is UTF-8 BOM taken from https://en.wikipedia.org/wiki/Byte_order_mark#Byte_order_marks_by_encoding
	if bytes.HasPrefix(fileJob.Content, []byte{239, 187, 191}) {
		if Verbose {
			printWarn(fmt.Sprintf("UTF-8 BOM found for file %s skipping 3 bytes", fileJob.Filename))
		}
		return 3
	}

	// If we have one of the other BOM then we might not be able to count correctly so if verbose let the user know
	if Verbose {
		for _, v := range ByteOrderMarks {
			if bytes.HasPrefix(fileJob.Content, v) {
				printWarn(fmt.Sprintf("BOM found for file %s indicating it is not ASCII/UTF-8 and may be counted incorrectly or ignored as a binary file", fileJob.Filename))
			}
		}
	}

	return 0
}

// Reads entire file into memory and then pushes it onto the next queue
func fileReaderWorker(input chan *FileJob, output chan *FileJob) {
	var startTime int64
	var wg sync.WaitGroup

	for i := 0; i < FileReadJobWorkers; i++ {
		wg.Add(1)
		go func() {
			for res := range input {
				atomic.CompareAndSwapInt64(&startTime, 0, makeTimestampMilli())

				fileStartTime := makeTimestampNano()
				content, err := ioutil.ReadFile(res.Location)

				if Trace {
					printTrace(fmt.Sprintf("nanoseconds read into memory: %s: %d", res.Location, makeTimestampNano()-fileStartTime))
				}

				if err == nil {
					res.Content = content
					output <- res
				} else {
					if Verbose {
						printWarn(fmt.Sprintf("error reading: %s %s", res.Location, err))
					}
				}
			}

			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(output)

		if Debug {
			printDebug(fmt.Sprintf("milliseconds reading files into memory: %d", makeTimestampMilli()-startTime))
		}
	}()
}

// Does the actual processing of stats and as such contains the hot path CPU call
func fileProcessorWorker(input chan *FileJob, output chan *FileJob) {
	var startTime int64
	var wg sync.WaitGroup

	queryConcordance := BuildConcordance(strings.ToLower(strings.Join(SearchString, " ")))

	for i := 0; i < FileProcessJobWorkers; i++ {
		wg.Add(1)
		go func() {
			for res := range input {
				atomic.CompareAndSwapInt64(&startTime, 0, makeTimestampMilli())

				processingStartTime := makeTimestampNano()

				cleanCode := codeCleaner(string(res.Content))
				res.Score = Relation(queryConcordance, BuildConcordance(cleanCode))

				if Trace {
					printTrace(fmt.Sprintf("nanoseconds process: %s: %d", res.Location, makeTimestampNano()-processingStartTime))
				}

				if !res.Binary && res.Score != 0 {
					output <- res
				} else {
					if Verbose {
						printWarn(fmt.Sprintf("skipping file identified as binary: %s", res.Location))
					}
				}
			}

			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(output)
	}()

	if Debug {
		printDebug(fmt.Sprintf("milliseconds processing files: %d", makeTimestampMilli()-startTime))
	}
}