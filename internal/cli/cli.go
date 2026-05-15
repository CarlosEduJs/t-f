// Package cli implements the command-line interface logic.
package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"t-f/internal/cssreader"
	"t-f/internal/tokens"
	"t-f/internal/version"
)

// Run executes the CLI logic and returns the exit code.
func Run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("t-f", flag.ContinueOnError)
	flags.SetOutput(stderr)

	figmaMode := flags.Bool("figma", false, "convert OKLCH colors to HEX for Figma compatibility")
	showVersion := flags.Bool("version", false, "show version information")

	if err := flags.Parse(args); err != nil {
		return 1
	}

	if *showVersion {
		fmt.Fprintf(stdout, "t-f version %s (commit: %s, date: %s)\n", version.Version, version.Commit, version.Date)
		return 0
	}

	if flags.NArg() < 2 {
		fmt.Fprintf(stderr, "Usage: t-f [--figma] input.css output.json\n")
		return 1
	}

	inputPath := flags.Arg(0)
	outputPath := flags.Arg(1)

	f, err := os.Open(inputPath)
	if err != nil {
		fmt.Fprintf(stderr, "Error opening input: %v\n", err)
		return 1
	}
	defer func() {
		if cerr := f.Close(); cerr != nil {
			fmt.Fprintf(stderr, "Warning: closing input file: %v\n", cerr)
		}
	}()

	parser := cssreader.New(f)
	vars, err := parser.Parse()
	if err != nil {
		fmt.Fprintf(stderr, "Error parsing CSS: %v\n", err)
		return 1
	}

	gen := tokens.NewGenerator()
	gen.FigmaMode = *figmaMode

	data, err := gen.Generate(vars)
	if err != nil {
		fmt.Fprintf(stderr, "Error generating tokens: %v\n", err)
		return 1
	}

	var pretty any
	if err := json.Unmarshal(data, &pretty); err != nil {
		fmt.Fprintf(stderr, "Error formatting JSON: %v\n", err)
		return 1
	}

	clean, err := json.MarshalIndent(pretty, "", "  ")
	if err != nil {
		fmt.Fprintf(stderr, "Error marshaling JSON: %v\n", err)
		return 1
	}

	if err := os.WriteFile(outputPath, clean, 0644); err != nil {
		fmt.Fprintf(stderr, "Error writing output: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "Wrote %d variables to %s\n", len(vars), outputPath)
	return 0
}
