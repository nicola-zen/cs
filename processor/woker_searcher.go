package processor

import (
	"bytes"
	"github.com/boyter/cs/processor/snippet"
	"strings"
	str "github.com/boyter/cs/string"
)

type SearcherWorker struct {
	input           chan *fileJob
	output          chan *fileJob
	searchParams    []searchParams
	FileCount       int64 // Count of the number of files that have been processed
	BinaryCount     int64 // Count the number of binary files
	MinfiedCount    int64
	SearchString    []string
	IncludeMinified bool
	IncludeBinary   bool
	CaseSensitive   bool
	MatchLimit      int
}

func NewSearcherWorker(input chan *fileJob, output chan *fileJob) SearcherWorker {
	return SearcherWorker{
		input:        input,
		output:       output,
		SearchString: []string{},
		MatchLimit:   100, // sensible default
	}
}

// Does the actual processing of stats and as such contains the hot path CPU call
// TODO add goroutines
func (f *SearcherWorker) Start() {

	// Build out the search params
	f.searchParams = parseArguments(f.SearchString)

	for res := range f.input {
		// Check for the presence of a nul byte indicating that this
		// is likely a binary file
		if !f.IncludeBinary {
			if bytes.IndexByte(res.Content, '\x00') != -1 {
				res.Binary = true
				continue
			}
		}

		// Check if the file is minified and if so ignore it
		if !f.IncludeMinified {
			split := bytes.Split(res.Content, []byte("\n"))
			sumLineLength := 0
			for _, s := range split {
				sumLineLength += len(s)
			}
			averageLineLength := sumLineLength / len(split)

			if averageLineLength > MinifiedLineByteLength {
				res.Minified = true
				continue
			}
		}

		for _, needle := range f.searchParams {
			switch needle.Type {
			case Default:
				str.IndexAll(string(res.Content), needle.Term, f.MatchLimit)
			}
		}

		if res.Score != 0 {
			f.output <- res
		}
	}

	close(f.output)
}

func (f *SearcherWorker) processMatches(res *fileJob, content []byte) bool {
	for i, term := range SearchBytes {
		// Currently only NOT does anything as the rest are just ignored
		if bytes.Equal(term, []byte("AND")) || bytes.Equal(term, []byte("OR")) || bytes.Equal(term, []byte("NOT")) {
			continue
		}

		if i != 0 && bytes.Equal(SearchBytes[i-1], []byte("NOT")) {
			index := bytes.Index(content, term)

			// If a negated term is found we bail out instantly as
			// this means we should not be matching at all
			if index != -1 {
				res.Score = 0
				return false
			}
		} else {

			if Fuzzy {
				if !bytes.HasSuffix(term, []byte("~1")) || !bytes.HasSuffix(term, []byte("~2")) {
					term = append(term, []byte("~1")...)
				}
			}

			// If someone supplies ~1 at the end of the term it means we want to expand out that
			// term to support fuzzy matches for that term where the number indicates a level
			// of fuzzyness
			res.Score = 0
			if bytes.HasSuffix(term, []byte("~1")) || bytes.HasSuffix(term, []byte("~2")) {
				terms := makeFuzzyDistanceOne(strings.TrimRight(string(term), "~1"))
				if bytes.HasSuffix(term, []byte("~2")) {
					terms = makeFuzzyDistanceTwo(strings.TrimRight(string(term), "~2"))
				}

				m := []int{}
				for _, t := range terms {
					m = append(m, snippet.ExtractLocation([]byte(t), content, MatchLimit)...)

					if len(m) != 0 {
						res.Locations[t] = m
						res.Score += float64(len(m))
					}
				}
			} else {
				// This is a regular search, not negated where we must try and find
				res.Locations[string(term)] = snippet.ExtractLocation(term, content, MatchLimit)

				if len(res.Locations[string(term)]) != 0 {
					res.Score += float64(len(res.Locations[string(term)]))
				} else {
					res.Score = 0
					return false
				}
			}
		}
	}

	return false
}
