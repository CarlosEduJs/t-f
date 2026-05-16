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
	if err := execute(args, stdout, stderr); err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}
	return 0
}

func execute(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("t-f", flag.ContinueOnError)
	flags.SetOutput(stderr)

	figmaMode := flags.Bool("figma", false, "convert OKLCH colors to HEX for Figma compatibility")
	showVersion := flags.Bool("version", false, "show version information")

	if err := flags.Parse(args); err != nil {
		return err
	}

	if *showVersion {
		fmt.Fprintf(stdout, "t-f version %s (commit: %s, date: %s)\n", version.Version, version.Commit, version.Date)
		return nil
	}

	if flags.NArg() < 2 {
		return fmt.Errorf("usage: t-f [--figma] input.css output.json")
	}

	inputPath := flags.Arg(0)
	outputPath := flags.Arg(1)

	f, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("opening input file %q: %w", inputPath, err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil {
			fmt.Fprintf(stderr, "Warning: closing input file: %v\n", cerr)
		}
	}()

	parser := cssreader.New(f)
	vars, err := parser.Parse()
	if err != nil {
		return fmt.Errorf("parsing CSS: %w", err)
	}

	gen := tokens.NewGenerator()
	gen.FigmaMode = *figmaMode

	data, err := gen.Generate(vars)
	if err != nil {
		return fmt.Errorf("generating tokens: %w", err)
	}

	var pretty any
	if err := json.Unmarshal(data, &pretty); err != nil {
		return fmt.Errorf("formatting JSON: %w", err)
	}

	clean, err := json.MarshalIndent(pretty, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}

	if err := os.WriteFile(outputPath, clean, 0644); err != nil {
		return fmt.Errorf("writing output file %q: %w", outputPath, err)
	}

	fmt.Fprintf(stdout, "Wrote %d variables to %s\n", len(vars), outputPath)
	return nil
}
