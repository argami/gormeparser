package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/argami/gormeparser/internal/models"
	"github.com/argami/gormeparser/internal/parser"
)

func main() {
	// CLI flags
	file := flag.String("file", "", "BORME file to parse (or directory for batch)")
	seccion := flag.String("seccion", "A", "Section to parse (A, B, or C)")
	output := flag.String("output", "", "Output directory for JSON files")
	pretty := flag.Bool("pretty", false, "Pretty-print JSON output")
	workers := flag.Int("workers", 4, "Number of parallel workers for batch processing")
	flag.Parse()

	if *file == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Check if input is a file or directory
	info, err := os.Stat(*file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error accessing path: %v\n", err)
		os.Exit(1)
	}

	if info.IsDir() {
		// Batch processing
		batchProcess(*file, *seccion, *output, *pretty, *workers)
	} else {
		// Single file processing
		singleProcess(*file, *seccion, *output, *pretty)
	}
}

func singleProcess(filename, seccion, output string, pretty bool) {
	result, err := parser.Parse(filename, models.Seccion(seccion))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing %s: %v\n", filename, err)
		os.Exit(1)
	}

	var data []byte
	var jsonErr error

	switch b := result.(type) {
	case *models.Borme:
		data, jsonErr = models.BormeToJSON(b, pretty)
	case *models.BormeC:
		data, jsonErr = bormeCToJSON(b, pretty)
	default:
		fmt.Fprintf(os.Stderr, "Unknown result type: %T\n", result)
		os.Exit(1)
	}

	if jsonErr != nil {
		fmt.Fprintf(os.Stderr, "Error serializing JSON: %v\n", jsonErr)
		os.Exit(1)
	}

	if output != "" {
		baseName := filepath.Base(filename)
		outFile := filepath.Join(output, strings.TrimSuffix(baseName, filepath.Ext(baseName))+".json")
		if err := os.WriteFile(outFile, data, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Written: %s\n", outFile)
	} else {
		fmt.Println(string(data))
	}
}

func batchProcess(dir, seccion, output string, pretty bool, workers int) {
	// Find all PDF/XML files
	var files []string
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading directory: %v\n", err)
		os.Exit(1)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if ext == ".pdf" || ext == ".xml" || ext == ".html" {
			files = append(files, filepath.Join(dir, name))
		}
	}

	if len(files) == 0 {
		fmt.Println("No PDF/XML files found in directory")
		return
	}

	// Create output directory if needed
	if output != "" {
		if err := os.MkdirAll(output, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Printf("Processing %d files with %d workers...\n", len(files), workers)

	// Process in parallel
	var wg sync.WaitGroup
	sem := make(chan struct{}, workers)
	results := make(chan struct {
		file string
		err  error
	}, len(files))

	for _, f := range files {
		wg.Add(1)
		sem <- struct{}{}

		go func(filename string) {
			defer wg.Done()
			defer func() { <-sem }()

			baseName := filepath.Base(filename)
			var outFile string
			if output != "" {
				outFile = filepath.Join(output, strings.TrimSuffix(baseName, filepath.Ext(baseName))+".json")
			}

			err := processFile(filename, models.Seccion(seccion), outFile, pretty)
			results <- struct {
				file string
				err  error
			}{filename, err}
		}(f)
	}

	wg.Wait()
	close(results)

	// Report results
	var success, failed int
	for r := range results {
		if r.err != nil {
			fmt.Printf("FAIL: %s - %v\n", r.file, r.err)
			failed++
		} else {
			success++
		}
	}

	fmt.Printf("\nDone: %d successful, %d failed\n", success, failed)
}

func processFile(filename string, seccion models.Seccion, outputFile string, pretty bool) error {
	result, err := parser.Parse(filename, seccion)
	if err != nil {
		return err
	}

	var data []byte
	var jsonErr error

	switch b := result.(type) {
	case *models.Borme:
		data, jsonErr = models.BormeToJSON(b, pretty)
	case *models.BormeC:
		data, jsonErr = bormeCToJSON(b, pretty)
	default:
		return fmt.Errorf("unknown result type: %T", result)
	}

	if jsonErr != nil {
		return jsonErr
	}

	if outputFile != "" {
		return os.WriteFile(outputFile, data, 0644)
	}

	return nil
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
