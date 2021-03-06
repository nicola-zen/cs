// SPDX-License-Identifier: MIT
// SPDX-License-Identifier: Unlicense
package main

import (
	"github.com/boyter/cs/file"
	"github.com/boyter/cs/processor"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	//f, _ := os.Create("cs.pprof")
	//pprof.StartCPUProfile(f)
	//defer pprof.StopCPUProfile()

	rootCmd := &cobra.Command{
		Use: "cs",
		Long: "cs code search command line.\n" +
			"Version " + processor.Version + "\n" +
			"Ben Boyter <ben@boyter.org>",
		Version: processor.Version,
		Run: func(cmd *cobra.Command, args []string) {
			processor.SearchString = args
			p := processor.NewProcess(".")

			// If there are arguments we want to print straight out to the console
			// otherwise we should enter interactive tui mode
			if len(processor.SearchString) != 0 {
				p.StartProcess()
			} else {
				processor.Error = false // suppress writing errors in TUI mode
				processor.ProcessTui(true)
			}
		},
	}

	flags := rootCmd.PersistentFlags()

	flags.BoolVar(
		&processor.DisableCheckBinary,
		"binary",
		false,
		"disable binary file detection",
	)
	flags.BoolVar(
		&processor.Ignore,
		"no-ignore",
		false,
		"disables .ignore file logic",
	)
	flags.BoolVar(
		&processor.GitIgnore,
		"no-gitignore",
		false,
		"disables .gitignore file logic",
	)
	flags.BoolVar(
		&processor.Debug,
		"debug",
		false,
		"enable debug output",
	)
	flags.Int64VarP(
		&processor.ResultLimit,
		"limit",
		"l",
		100,
		"number of matching results to process",
	)
	flags.Int64VarP(
		&processor.SnippetLength,
		"snippet",
		"s",
		300,
		"number of matching results to process",
	)
	flags.StringSliceVar(
		&processor.PathDenylist,
		"exclude-dir",
		[]string{".git", ".hg", ".svn"},
		"directories to exclude",
	)
	flags.StringVarP(
		&processor.Format,
		"format",
		"f",
		"text",
		"set output format [text, json]",
	)
	flags.StringSliceVarP(
		&processor.AllowListExtensions,
		"include-ext",
		"i",
		[]string{},
		"limit to file extensions [comma separated list: e.g. go,java,js]",
	)
	flags.StringVarP(
		&processor.FileOutput,
		"output",
		"o",
		"",
		"output filename (default stdout)",
	)
	flags.BoolVarP(
		&processor.Trace,
		"trace",
		"t",
		false,
		"enable trace output. Not recommended when processing multiple files",
	)
	flags.BoolVarP(
		&processor.Verbose,
		"verbose",
		"v",
		false,
		"verbose output",
	)
	flags.BoolVar(
		&processor.IncludeMinified,
		"include-min",
		false,
		"include minified files",
	)
	flags.BoolVar(
		&file.IncludeHidden,
		"include-hidden",
		false,
		"include hidden files",
	)
	flags.IntVar(
		&processor.MinifiedLineByteLength,
		"min-line-length",
		255,
		"number of bytes per average line for file to be considered minified",
	)
	flags.BoolVar(
		&processor.Fuzzy,
		"fuzzy",
		false,
		"make the search by default fuzzy",
	)
	flags.StringSliceVarP(
		&processor.LocationExcludePattern,
		"exclude-pattern",
		"x",
		[]string{},
		"file locations matching this pattern ignoring case will be ignored",
	)
	flags.BoolVarP(
		&processor.CaseSensitive,
		"case-sensitive",
		"c",
		false,
		"make the search case sensitive",
	)
	flags.BoolVarP(
		&processor.FindRoot,
		"find-root",
		"r",
		false,
		"attempts to find the root of this repository recursively looking for .git or .hg",
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
