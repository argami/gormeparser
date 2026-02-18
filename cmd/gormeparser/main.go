package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/argami/gormeparser/internal/download"
	"github.com/argami/gormeparser/internal/models"
	"github.com/argami/gormeparser/internal/parser"
)

func main() {
	// File/directory mode flags
	file := flag.String("file", "", "BORME file or directory to parse")
	seccion := flag.String("seccion", "A", "Section to parse (A, B, or C)")
	output := flag.String("output", "", "Output directory for JSON files")
	pretty := flag.Bool("pretty", false, "Pretty-print JSON output")
	workers := flag.Int("workers", 4, "Number of parallel workers for batch processing")

	// Download + process mode flags
	startDate := flag.String("start-date", "", "Start date (YYYY-MM-DD) for download+process")
	endDate := flag.String("end-date", "", "End date (YYYY-MM-DD) for download+process")
	provincia := flag.String("provincia", "", "Province code or name (e.g., 'Madrid', 'Barcelona', '28')")
	downloadDir := flag.String("download-dir", "./downloads", "Directory to download PDFs")

	flag.Parse()

	// Check which mode to use
	hasDateRange := *startDate != "" && *endDate != ""
	hasFileOrDir := *file != ""

	if hasDateRange && hasFileOrDir {
		fmt.Println("Error: Cannot use both date range and file/directory modes")
		flag.Usage()
		os.Exit(1)
	}

	if hasDateRange {
		// Download + process mode
		downloadAndProcess(*startDate, *endDate, *provincia, *seccion, *downloadDir, *output, *pretty, *workers)
		return
	}

	if !hasFileOrDir {
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
		batchProcess(*file, *seccion, *output, *pretty, *workers)
	} else {
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

	if output != "" {
		if err := os.MkdirAll(output, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Printf("Processing %d files with %d workers...\n", len(files), workers)

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

func downloadAndProcess(startDate, endDate, provincia, seccion, downloadDir, output string, pretty bool, workers int) {
	// Parse dates
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid start date: %v\n", err)
		os.Exit(1)
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid end date: %v\n", err)
		os.Exit(1)
	}

	// Normalize province
	provCode := provincia
	if provincia != "" {
		provCode = normalizeProvincia(provincia)
		if provCode == "" {
			fmt.Fprintf(os.Stderr, "Unknown province: %s\n", provincia)
			os.Exit(1)
		}
	}

	// Create directories
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating download directory: %v\n", err)
		os.Exit(1)
	}

	if output != "" {
		if err := os.MkdirAll(output, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
			os.Exit(1)
		}
	}

	// Generate dates
	var dates []time.Time
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		dates = append(dates, d)
	}

	fmt.Printf("Downloading and processing BORME %s from %s to %s\n", seccion, startDate, endDate)
	if provincia != "" {
		fmt.Printf("Province filter: %s (%s)\n", provincia, provCode)
	}
	fmt.Printf("Processing %d dates with %d workers...\n", len(dates), workers)

	// Download and process in parallel
	var wg sync.WaitGroup
	sem := make(chan struct{}, workers)
	results := make(chan struct {
		date   time.Time
		file   string
		err    error
	}, len(dates))

	for _, date := range dates {
		wg.Add(1)
		sem <- struct{}{}

		go func(d time.Time) {
			defer wg.Done()
			defer func() { <-sem }()

			// Get URL
			url := download.GetURLPDF(d, seccion, provCode)

			// Generate filename
			filename := filepath.Join(downloadDir, fmt.Sprintf("BORME-%s-%s.pdf", seccion, d.Format("2006-01-02")))

			// Download
			if err := download.DownloadFile(url, filename); err != nil {
				results <- struct {
					date   time.Time
					file   string
					err    error
				}{d, "", err}
				return
			}

			// Parse
			var outFile string
			if output != "" {
				outFile = filepath.Join(output, fmt.Sprintf("BORME-%s-%s.json", seccion, d.Format("2006-01-02")))
			}

			err := processFile(filename, models.Seccion(seccion), outFile, pretty)
			results <- struct {
				date   time.Time
				file   string
				err    error
			}{d, filename, err}
		}(date)
	}

	wg.Wait()
	close(results)

	var success, failed int
	for r := range results {
		if r.err != nil {
			fmt.Printf("FAIL: %s - %v\n", r.date.Format("2006-01-02"), r.err)
			failed++
		} else {
			success++
		}
	}

	fmt.Printf("\nDone: %d dates processed, %d failed\n", success, failed)
}

func normalizeProvincia(prov string) string {
	// Try as code first
	codeMap := map[string]string{
		"28": "Madrid",
		"08": "Barcelona",
		"46": "Valencia",
		"41": "Sevilla",
		"30": "Murcia",
		"33": "Asturias",
		"07": "Illes Balears",
		"35": "Las Palmas",
		"38": "Santa Cruz de Tenerife",
	}

	if code, ok := codeMap[prov]; ok {
		return code
	}

	// Try as name
	nameMap := map[string]string{
		"madrid":              "Madrid",
		"barcelona":           "Barcelona",
		"valencia":            "Valencia",
		"sevilla":             "Sevilla",
		"murcia":              "Murcia",
		"asturias":            "Asturias",
		"balears":             "Illes Balears",
		"islas balears":       "Illes Balears",
		"las Palmas":          "Las Palmas",
		"gran canaria":        "Las Palmas",
		"tenerife":           "Santa Cruz de Tenerife",
		"canarias":           "Santa Cruz de Tenerife",
		"cadiz":               "Cadiz",
		"malaga":              "Malaga",
		"bizkaia":             "Bizkaia",
		"biscay":              "Bizkaia",
		"vizcaya":             "Bizkaia",
		"gipuzkoa":            "Gipuzkoa",
		"navarra":             "Navarra",
		"araba":               "Araba/Álava",
		"alava":               "Araba/Álava",
		"alicante":            "Alicante",
		"coruña":              "La Coruña",
		"a coruña":            "La Coruña",
		"pontevedra":          "Pontevedra",
		"galicia":             "La Coruña",
	}

	provLower := strings.ToLower(prov)
	for key, value := range nameMap {
		if strings.Contains(provLower, key) || provLower == key {
			return value
		}
	}

	return ""
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
