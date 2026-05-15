package main

import (
	"encoding/json"
	"fmt"
	"os"

	"t-f/internal/cssreader"
	"t-f/internal/tokens"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: t-f input.css output.json\n")
		os.Exit(1)
	}

	inputPath := os.Args[1]
	outputPath := os.Args[2]

	f, err := os.Open(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening input: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	parser := cssreader.New(f)
	vars, err := parser.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing CSS: %v\n", err)
		os.Exit(1)
	}

	gen := tokens.NewGenerator()
	data, err := gen.Generate(vars)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating tokens: %v\n", err)
		os.Exit(1)
	}

	var pretty interface{}
	json.Unmarshal(data, &pretty)
	clean, _ := json.MarshalIndent(pretty, "", "  ")

	if err := os.WriteFile(outputPath, clean, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Wrote %d variables to %s\n", len(vars), outputPath)
}
