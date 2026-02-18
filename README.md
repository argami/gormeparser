# gormeparser

BORME parser written in Go - Port of [bormeparser](https://github.com/sincronia/bormeparser) from Python.

Parses Spanish Mercantile Registry (BORME) bulletins from PDF and XML formats.

## Installation

```bash
git clone https://github.com/argami/gormeparser.git
cd gormeparser
go build -o bin/gormeparser ./cmd/gormeparser
go build -o bin/compare ./cmd/compare
```

Or install globally:

```bash
go install github.com/argami/gormeparser/cmd/gormeparser@latest
go install github.com/argami/gormeparser/cmd/compare@latest
```

## Usage

### Parse a BORME PDF

```bash
# Parse Section A PDF
./bin/gormeparser -file examples/BORME-A-2015-27-10.pdf -pretty -output output.json

# Parse Section C XML
./bin/gormeparser -file examples/BORME-C-2011-20488.xml -seccion C -pretty -output output_c.json

# Parse from stdin
cat file.pdf | ./bin/gormeparser -file -
```

### Download BORME from BOE

```bash
# Using the download package programmatically
```

See `internal/download/download.go` for the download API.

### Compare Python vs Go Output

```bash
# First, generate JSON with Python
python -c "import bormeparser; b = bormeparser.parse('examples/BORME-A-2015-27-10.pdf', 'A'); b.to_json('output_python.json')"

# Generate JSON with Go
./bin/gormeparser -file examples/BORME-A-2015-27-10.pdf -pretty -output output_go.json

# Compare outputs
./bin/compare -python-json output_python.json -go-json output_go.json -v
```

### Batch Processing

Process multiple files in parallel:

```bash
# Process all PDFs in a directory (parallel, 4 workers by default)
./bin/gormeparser -file ./pdfs/ -output ./json_output/

# With 8 workers
./bin/gormeparser -file ./pdfs/ -output ./json_output/ -workers 8

# Pretty-printed output
./bin/gormeparser -file ./pdfs/ -output ./json_output/ -pretty

# Process XML files
./bin/gormeparser -file ./xml/ -seccion C -output ./json_output/
```

Output:
```
Processing 100 files with 4 workers...
Done: 98 successful, 2 failed
FAIL: file1.pdf - error message
FAIL: file2.pdf - error message
```

### Batch Processing API

```go
package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/argami/gormeparser/internal/models"
	"github.com/argami/gormeparser/internal/parser"
)

// BatchResult holds results from batch processing
type BatchResult struct {
	Total    int
	Success  int
	Failed   int
	Errors   map[string]error
}

// ProcessDirectory processes all PDF/XML files in a directory
func ProcessDirectory(dir string, seccion models.Seccion, workers int) (*BatchResult, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	// Find files
	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if ext == ".pdf" || ext == ".xml" || ext == ".html" {
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}

	result := &BatchResult{
		Total:  len(files),
		Errors: make(map[string]error),
	}

	// Process in parallel (use goroutines with sync.WaitGroup)
	// ... implementation with semaphore pattern

	return result, nil
}
```

### Download and Process by Date Range

```bash
# Download and process BORME for a date range
./bin/gormeparser -start-date 2024-01-01 -end-date 2024-01-31 \
  -provincia Madrid -output ./json/

# All provinces, Section A
./bin/gormeparser -start-date 2024-01-01 -end-date 2024-01-31 \
  -seccion A -output ./json/

# Section C, all provinces
./bin/gormeparser -start-date 2024-01-01 -end-date 2024-01-31 \
  -seccion C -output ./json/

# With 8 parallel workers
./bin/gormeparser -start-date 2024-01-01 -end-date 2024-01-31 \
  -workers 8 -output ./json/

# Pretty-printed output
./bin/gormeparser -start-date 2024-01-01 -end-date 2024-01-31 \
  -pretty -output ./json/
```

### Supported Provinces

```bash
# By name
-provincia Madrid
-provincia Barcelona
-provincia Valencia
-provincia Sevilla
-provincia Bizkaia  # or "Vizcaya", "Biscay"
-provincia Gipuzkoa
-provincia Asturias
-provincia Murcia

# By code
-provincia 28  # Madrid
-provincia 08  # Barcelona
-provincia 46  # Valencia
```

## API Usage

### Parse a PDF

```go
package main

import (
	"fmt"
	"log"

	"github.com/argami/gormeparser/internal/models"
	"github.com/argami/gormeparser/internal/parser"
)

func main() {
	result, err := parser.Parse("BORME-A-2015-27-10.pdf", models.SeccionA)
	if err != nil {
		log.Fatal(err)
	}

	borme := result.(*models.Borme)
	fmt.Printf("Date: %s\n", borme.Date)
	fmt.Printf("Section: %s\n", borme.Seccion)
	fmt.Printf("Announcements: %d\n", len(borme.Anuncios))

	for id, anuncio := range borme.Anuncios {
		fmt.Printf("\n[%d] %s\n", id, anuncio.Empresa)
		for _, acto := range anuncio.Actos {
			fmt.Printf("  - %s\n", acto.GetName())
		}
	}
}
```

### Download and Parse

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/argami/gormeparser/internal/download"
	"github.com/argami/gormeparser/internal/models"
	"github.com/argami/gormeparser/internal/parser"
)

func main() {
	// Download a BORME PDF
	date := time.Date(2015, 10, 27, 0, 0, 0, 0, time.UTC)
	url := download.GetURLPDF(date, "A", "Madrid")

	err := download.DownloadFile(url, "BORME-A-2015-27-10.pdf")
	if err != nil {
		log.Fatal(err)
	}

	// Parse the PDF
	result, err := parser.Parse("BORME-A-2015-27-10.pdf", models.SeccionA)
	if err != nil {
		log.Fatal(err)
	}

	borme := result.(*models.Borme)
	fmt.Printf("Parsed %d announcements\n", len(borme.Anuncios))
}
```

### Download and Process by Date Range (API)

```go
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/argami/gormeparser/internal/download"
	"github.com/argami/gormeparser/internal/models"
	"github.com/argami/gormeparser/internal/parser"
)

// DownloadAndProcess downloads and parses BORME for a date range
func DownloadAndProcess(startDate, endDate time.Time, provincia, seccion string, workers int) error {
	// Create directories
	downloadDir := "./downloads"
	jsonDir := "./json"
	os.MkdirAll(downloadDir, 0755)
	os.MkdirAll(jsonDir, 0755)

	// Process each date
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		// Generate filename
		filename := filepath.Join(downloadDir, fmt.Sprintf("BORME-%s-%s.pdf", seccion, d.Format("2006-01-02")))

		// Download
		url := download.GetURLPDF(d, seccion, provincia)
		if err := download.DownloadFile(url, filename); err != nil {
			log.Printf("Failed to download %s: %v", d.Format("2006-01-02"), err)
			continue
		}

		// Parse
		result, err := parser.Parse(filename, models.Seccion(seccion))
		if err != nil {
			log.Printf("Failed to parse %s: %v", filename, err)
			continue
		}

		// Save JSON
		jsonFile := filepath.Join(jsonDir, fmt.Sprintf("BORME-%s-%s.json", seccion, d.Format("2006-01-02")))
		var data []byte
		switch b := result.(type) {
		case *models.Borme:
			data, _ = models.BormeToJSON(b, true)
		case *models.BormeC:
			// Convert to JSON...
		}
		os.WriteFile(jsonFile, data, 0644)
	}

	return nil
}
```

### Serialize to JSON

```go
package main

import (
	"fmt"
	"log"

	"github.com/argami/gormeparser/internal/models"
	"github.com/argami/gormeparser/internal/parser"
)

func main() {
	result, err := parser.Parse("BORME-A-2015-27-10.pdf", models.SeccionA)
	if err != nil {
		log.Fatal(err)
	}

	borme := result.(*models.Borme)

	// Pretty print
	jsonData, err := models.BormeToJSON(borme, true)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(jsonData))

	// Or compact
	jsonData, err = models.BormeToJSON(borme, false)
	if err != nil {
		log.Fatal(err)
	}
}
```

## Output Format

### Section A (PDF)

```json
{
  "date": "2015-10-27T00:00:00Z",
  "seccion": "A",
  "provincia": {
    "code": 280,
    "name": "Madrid"
  },
  "num": 273,
  "cve": "BORME-A-2015-273-28",
  "anuncios": {
    "1": {
      "id": 1,
      "empresa": "ALDARA CATERING SL",
      "registro": "Madrid",
      "sucursal": false,
      "liquidacion": false,
      "actos": [
        {
          "name": "Constitucion",
          "value": "Texto del acto..."
        },
        {
          "name": "Nombramientos",
          "value": {
            "Adm. Solid.": ["RAMA SANCHEZ JOSE PEDRO", "RAMA SANCHEZ JAVIER"]
          }
        }
      ]
    }
  }
}
```

### Section C (XML/HTML)

```json
{
  "departamento": "CONVOCATORIAS DE JUNTAS",
  "texto": "Texto completo del anuncio...",
  "diario_numero": 20488,
  "numero_anuncio": "20488",
  "id_anuncio": "A110044738",
  "cve": "BORME-C-2011-20488",
  "titulo": "Convocatoria de junta...",
  "empresa": "EMPRESA EXAMPLE SA",
  "empresas_relacionadas": [],
  "cifs": ["A12345678"],
  "seccion": "C"
}
```

## Project Structure

```
gormeparser/
├── cmd/
│   ├── gormeparser/main.go    # CLI tool
│   └── compare/main.go        # Comparison tool
├── internal/
│   ├── models/
│   │   ├── borme.go          # Borme, BormeAnuncio, BormeActo
│   │   ├── seccion.go        # Section constants
│   │   └── seccion_c.go      # Section C models
│   ├── parser/
│   │   ├── parser.go         # Main router
│   │   ├── pypdf2/          # Section A (PDF)
│   │   └── seccion_c/        # Section C (XML/HTML)
│   ├── regex/                # Regular expressions
│   └── download/              # Download from BOE
├── examples/                  # Example files
└── go.mod
```

## Requirements

- Go 1.20+
- Git (for dependencies)

## Dependencies

- `github.com/antchfx/xmlquery` - XPath for XML parsing
- `github.com/rsc/pdf` - PDF text extraction

## License

MIT License - See LICENSE file.

## Credits

- Original Python implementation: [bormeparser](https://github.com/sincronia/bormeparser)
- BORME data: [Boletín Oficial del Registro Mercantil](https://www.boe.es/borme/)
