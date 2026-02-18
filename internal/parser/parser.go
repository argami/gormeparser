package parser

import (
	"fmt"
	"os"
	"strings"

	"github.com/argami/gormeparser/internal/models"
	"github.com/argami/gormeparser/internal/parser/pypdf2"
	"github.com/argami/gormeparser/internal/parser/seccion_c"
)

// Parse parses a BORME file and returns the appropriate object based on section
func Parse(filename string, seccion models.Seccion) (interface{}, error) {
	// Normalize section
	seccion = models.Seccion(strings.ToUpper(string(seccion)))

	switch seccion {
	case models.SeccionA:
		return ParseA(filename)
	case models.SeccionB:
		return ParseA(filename) // Section B uses same parser as A
	case models.SeccionC:
		return ParseC(filename)
	default:
		return nil, fmt.Errorf("sección no soportada: %s", seccion)
	}
}

// ParseA parses a Section A PDF file
func ParseA(filename string) (*models.Borme, error) {
	parser := pypdf2.NewParser(filename)
	return parser.Parse()
}

// ParseC parses a Section C XML/HTML file
func ParseC(filename string) (*models.BormeC, error) {
	parser := seccionc.NewParser(filename)
	return parser.Parse()
}

// ParseCFromURL parses a Section C file from a URL
func ParseCFromURL(url string) (*models.BormeC, error) {
	// Download the file first
	parser := seccionc.NewParser(url)
	return parser.Parse()
}

// ParseFromData parses a BORME file from byte data
func ParseFromData(data []byte, seccion models.Seccion, format string) (interface{}, error) {
	// Create temporary file
	tmpFile, err := os.CreateTemp("", "borme-*.tmp")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write temp file: %w", err)
	}
	tmpFile.Close()

	return Parse(tmpFile.Name(), seccion)
}
