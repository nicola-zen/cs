package processor

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
)

// Flags set via the CLI which control how the output is displayed

// Files indicates if there should be file output or not when formatting
var Files = false

// Languages indicates if the command line should print out the supported languages
var Languages = false

// Verbose enables verbose logging output
var Verbose = false

// Debug enables debug logging output
var Debug = false

// Trace enables trace logging output which is extremely verbose
var Trace = false

// Disables .gitignore checks
var GitIgnore = false

// Disables ignore file checks
var Ignore = false

// DisableCheckBinary toggles checking for binary files using NUL bytes
var DisableCheckBinary = false

// Exclude is a regular expression which is used to exclude files from being processed
var Exclude = []string{}

// Format sets the output format of the formatter
var Format = ""

// FileOutput sets the file that output should be written to
var FileOutput = ""

// PathBlacklist sets the paths that should be skipped
var PathBlacklist = []string{}

// FileListQueueSize is the queue of files found and ready to be read into memory
var FileListQueueSize = runtime.NumCPU()

// FileReadJobWorkers is the number of processes that read files off disk into memory
var FileReadJobWorkers = runtime.NumCPU() * 4

// FileReadContentJobQueueSize is a queue of files ready to be processed
var FileReadContentJobQueueSize = runtime.NumCPU()

// FileProcessJobWorkers is the number of workers that process the file collecting stats
var FileProcessJobWorkers = runtime.NumCPU() * 4

// FileSummaryJobQueueSize is the queue used to hold processed file statistics before formatting
var FileSummaryJobQueueSize = runtime.NumCPU()

// WhiteListExtensions is a list of extensions which are whitelisted to be processed
var WhiteListExtensions = []string{}

// Search string if set to anything is what we want to run the search for against the current directory
var SearchString = []string{}

// Process is the main entry point of the command line it sets everything up and starts running
func Process() {
	if Debug {
		printDebug(fmt.Sprintf("White List: %v", WhiteListExtensions))
		printDebug(fmt.Sprintf("File Output: %t", Files))
		printDebug(fmt.Sprintf("Verbose: %t", Verbose))
		printDebug(fmt.Sprintf("NumCPU: %d", runtime.NumCPU()))
		printDebug(fmt.Sprintf("PathBlacklist: %v", PathBlacklist))
	}

	fileListQueue := make(chan *FileJob, FileListQueueSize)                     // Files ready to be read from disk
	fileReadContentJobQueue := make(chan *FileJob, FileReadContentJobQueueSize) // Files ready to be processed
	fileSummaryJobQueue := make(chan *FileJob, FileSummaryJobQueueSize)         // Files ready to be summarised

	go walkDirectoryParallel(filepath.Clean("."), fileListQueue)
	go fileReaderWorker(fileListQueue, fileReadContentJobQueue)
	go fileProcessorWorker(fileReadContentJobQueue, fileSummaryJobQueue)


	result := fileSummarize(fileSummaryJobQueue)

	if FileOutput == "" {
		fmt.Println(result)
	} else {
		_ = ioutil.WriteFile(FileOutput, []byte(result), 0600)
		fmt.Println("results written to " + FileOutput)
	}
}