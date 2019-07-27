package main

import (
	"github.com/boyter/sc/processor"
	"github.com/spf13/cobra"
	"os"
	"runtime/pprof"
)

func main() {
	f, _ := os.Create("sc.pprof")
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	rootCmd := &cobra.Command{
		Use:     "sc",
		Short:   "sc [DIRECTORY]",
		Long:    "sc searchcode command line search.\nBen Boyter <ben@boyter.org>",
		Version: "0.0.1",
		Run: func(cmd *cobra.Command, args []string) {
			processor.SearchString = args
			processor.Process()
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
	flags.StringSliceVar(
		&processor.PathBlacklist,
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
		&processor.WhiteListExtensions,
		"include-ext",
		"i",
		[]string{},
		"limit to file extensions [comma separated list: e.g. go,java,js]",
	)
	flags.StringArrayVarP(
		&processor.Exclude,
		"not-match",
		"M",
		[]string{},
		"ignore files and directories matching regular expression",
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

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}