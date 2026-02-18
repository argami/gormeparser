package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/argami/gormeparser/internal/models"
)

// ComparisonResult holds the results of comparing two JSON files
type ComparisonResult struct {
	PythonCount int
	GoCount     int
	MatchCount  int
	Differences []Difference
}

// Difference represents a difference between two announcements
type Difference struct {
	ID        int
	Field     string
	PythonVal interface{}
	GoVal     interface{}
	Severity  string // "critical", "warning", "info"
	Message   string
}

func main() {
	pythonJSON := flag.String("python-json", "", "Path to Python JSON output")
	goJSON := flag.String("go-json", "", "Path to Go JSON output")
	tolerance := flag.Float64("tolerance", 0.01, "Tolerance for floating point comparison")
	verbose := flag.Bool("v", false, "Verbose output")
	flag.Parse()

	if *pythonJSON == "" || *goJSON == "" {
		flag.Usage()
		os.Exit(1)
	}

	result, err := compare(*pythonJSON, *goJSON, *tolerance)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Print results
	fmt.Println("=== BORME Parser Comparison Results ===")
	fmt.Printf("Python announcements: %d\n", result.PythonCount)
	fmt.Printf("Go announcements: %d\n", result.GoCount)
	fmt.Printf("Matching: %d\n", result.MatchCount)
	fmt.Printf("Differences: %d\n", len(result.Differences))

	if len(result.Differences) > 0 {
		fmt.Println("\n=== Differences ===")
		for _, diff := range result.Differences {
			severityIcon := "ℹ️"
			switch diff.Severity {
			case "critical":
				severityIcon = "🚨"
			case "warning":
				severityIcon = "⚠️"
			}
			fmt.Printf("%s [ID=%d] %s: %s\n", severityIcon, diff.ID, diff.Field, diff.Message)
			if *verbose {
				fmt.Printf("  Python: %v\n", diff.PythonVal)
				fmt.Printf("  Go: %v\n", diff.GoVal)
			}
		}
	}

	// Exit with error code if there are critical differences
	hasCritical := false
	for _, diff := range result.Differences {
		if diff.Severity == "critical" {
			hasCritical = true
			break
		}
	}

	if hasCritical {
		fmt.Println("\n❌ Critical differences found!")
		os.Exit(1)
	}

	if len(result.Differences) > 0 {
		fmt.Println("\n⚠️  Warnings found, but parsing may still be valid.")
		os.Exit(0)
	}

	fmt.Println("\n✅ Parsing results match!")
}

func compare(pythonPath, goPath string, tolerance float64) (*ComparisonResult, error) {
	result := &ComparisonResult{
		Differences: make([]Difference, 0),
	}

	// Read Python JSON
	pythonData, err := os.ReadFile(pythonPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read Python JSON: %w", err)
	}

	var pythonBorme models.Borme
	if err := json.Unmarshal(pythonData, &pythonBorme); err != nil {
		return nil, fmt.Errorf("failed to parse Python JSON: %w", err)
	}

	// Read Go JSON
	goData, err := os.ReadFile(goPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read Go JSON: %w", err)
	}

	var goBorme models.Borme
	if err := json.Unmarshal(goData, &goBorme); err != nil {
		return nil, fmt.Errorf("failed to parse Go JSON: %w", err)
	}

	// Compare counts
	result.PythonCount = len(pythonBorme.Anuncios)
	result.GoCount = len(goBorme.Anuncios)

	// Compare common IDs
	for id, pythonAnuncio := range pythonBorme.Anuncios {
		goAnuncio, exists := goBorme.Anuncios[id]
		if !exists {
			result.Differences = append(result.Differences, Difference{
				ID:        id,
				Field:     "exists",
				PythonVal: true,
				GoVal:     false,
				Severity:  "critical",
				Message:   "Announcement exists in Python but not in Go",
			})
			continue
		}

		result.MatchCount++

		// Compare fields
		compareAnuncio(result, id, pythonAnuncio, goAnuncio)
	}

	// Check for announcements in Go but not in Python
	for id := range goBorme.Anuncios {
		if _, exists := pythonBorme.Anuncios[id]; !exists {
			result.Differences = append(result.Differences, Difference{
				ID:        id,
				Field:     "exists",
				PythonVal: false,
				GoVal:     true,
				Severity:  "critical",
				Message:   "Announcement exists in Go but not in Python",
			})
		}
	}

	// Compare metadata
	compareMetadata(result, &pythonBorme, &goBorme)

	return result, nil
}

func compareAnuncio(result *ComparisonResult, id int, python, goVal *models.BormeAnuncio) {
	// Compare empresa
	if python.Empresa != goVal.Empresa {
		result.Differences = append(result.Differences, Difference{
			ID:        id,
			Field:     "empresa",
			PythonVal: python.Empresa,
			GoVal:     goVal.Empresa,
			Severity:  "critical",
			Message:   "Company name mismatch",
		})
	}

	// Compare registro
	if python.Registro != goVal.Registro {
		result.Differences = append(result.Differences, Difference{
			ID:        id,
			Field:     "registro",
			PythonVal: python.Registro,
			GoVal:     goVal.Registro,
			Severity:  "warning",
			Message:   "Register name differs",
		})
	}

	// Compare sucursal
	if python.Sucursal != goVal.Sucursal {
		result.Differences = append(result.Differences, Difference{
			ID:        id,
			Field:     "sucursal",
			PythonVal: python.Sucursal,
			GoVal:     goVal.Sucursal,
			Severity:  "warning",
			Message:   "Sucursal flag differs",
		})
	}

	// Compare liquidacion
	if python.Liquidacion != goVal.Liquidacion {
		result.Differences = append(result.Differences, Difference{
			ID:        id,
			Field:     "liquidacion",
			PythonVal: python.Liquidacion,
			GoVal:     goVal.Liquidacion,
			Severity:  "warning",
			Message:   "Liquidacion flag differs",
		})
	}

	// Compare datos_registrales
	if python.DatosRegistrales != goVal.DatosRegistrales {
		result.Differences = append(result.Differences, Difference{
			ID:        id,
			Field:     "datos_registrales",
			PythonVal: python.DatosRegistrales,
			GoVal:     goVal.DatosRegistrales,
			Severity:  "info",
			Message:   "Datos registrales differs",
		})
	}

	// Compare actos count
	if len(python.Actos) != len(goVal.Actos) {
		result.Differences = append(result.Differences, Difference{
			ID:        id,
			Field:     "actos_count",
			PythonVal: len(python.Actos),
			GoVal:     len(goVal.Actos),
			Severity:  "warning",
			Message:   "Number of acts differs",
		})
	}

	// Compare actos content
	compareActos(result, id, python.Actos, goVal.Actos)
}

func compareActos(result *ComparisonResult, id int, python, goVal []models.BormeActo) {
	minLen := len(python)
	if len(goVal) < minLen {
		minLen = len(goVal)
	}
	for i := 0; i < minLen; i++ {
		pythonActo := python[i]
		goActo := goVal[i]

		if pythonActo.GetName() != goActo.GetName() {
			result.Differences = append(result.Differences, Difference{
				ID:        id,
				Field:     fmt.Sprintf("acto[%d].name", i),
				PythonVal: pythonActo.GetName(),
				GoVal:     goActo.GetName(),
				Severity:  "warning",
				Message:   "Act name differs",
			})
		}
	}
}

func compareMetadata(result *ComparisonResult, python, goVal *models.Borme) {
	// Compare section
	if python.Seccion != goVal.Seccion {
		result.Differences = append(result.Differences, Difference{
			Field:     "seccion",
			PythonVal: python.Seccion,
			GoVal:     goVal.Seccion,
			Severity:  "critical",
			Message:   "Section mismatch",
		})
	}

	// Compare number
	if python.Num != goVal.Num {
		result.Differences = append(result.Differences, Difference{
			Field:     "num",
			PythonVal: python.Num,
			GoVal:     goVal.Num,
			Severity:  "warning",
			Message:   "BORME number differs",
		})
	}

	// Compare CVE
	if python.CVE != goVal.CVE {
		result.Differences = append(result.Differences, Difference{
			Field:     "cve",
			PythonVal: python.CVE,
			GoVal:     goVal.CVE,
			Severity:  "info",
			Message:   "CVE differs",
		})
	}
}
