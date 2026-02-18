package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/argami/gormeparser/internal/models"
	"github.com/argami/gormeparser/internal/parser"
)

func main() {
	filename := flag.String("file", "", "BORME file to parse")
	seccion := flag.String("seccion", "A", "Section to parse (A, B, or C)")
	output := flag.String("output", "", "Output JSON file")
	pretty := flag.Bool("pretty", false, "Pretty-print JSON output")
	flag.Parse()

	if *filename == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Parse the file
	result, err := parser.Parse(*filename, models.Seccion(*seccion))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing file: %v\n", err)
		os.Exit(1)
	}

	// Handle different result types
	var data []byte
	var parseErr error

	switch b := result.(type) {
	case *models.Borme:
		data, parseErr = models.BormeToJSON(b, *pretty)
	case *models.BormeC:
		data, parseErr = bormeCToJSON(b, *pretty)
	default:
		fmt.Fprintf(os.Stderr, "Unknown result type: %T\n", result)
		os.Exit(1)
	}

	if parseErr != nil {
		fmt.Fprintf(os.Stderr, "Error serializing JSON: %v\n", parseErr)
		os.Exit(1)
	}

	// Write output
	if *output != "" {
		if err := os.WriteFile(*output, data, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Output written to %s\n", *output)
	} else {
		fmt.Println(string(data))
	}
}

// bormeCToJSON serializes BormeC to JSON
func bormeCToJSON(b *models.BormeC, pretty bool) ([]byte, error) {
	if pretty {
		data, err := json.MarshalIndent(b, "", "  ")
		if err != nil {
			return nil, err
		}
		data = append(data, '\n')
		return data, nil
	}
	return json.Marshal(b)
}
