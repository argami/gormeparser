package pypdf2

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/argami/gormeparser/internal/models"
	"github.com/argami/gormeparser/internal/regex"
)

// PyPDF2Parser parses Section A BORME PDFs
type PyPDF2Parser struct {
	filename string
	data     *models.Borme
	actos    []models.BormeActo
}

// ParserState tracks the current parsing state
type ParserState struct {
	Cabecera   bool
	Texto      bool
	Fecha      bool
	Numero     bool
	Seccion    bool
	Provincia  bool
	CVE        bool
	CurrentActo string
	CurrentAnuncio *models.BormeAnuncio
}

// NewParser creates a new PyPDF2Parser
func NewParser(filename string) *PyPDF2Parser {
	return &PyPDF2Parser{
		filename: filename,
	}
}

// Parse parses a Section A PDF and returns a Borme object
func (p *PyPDF2Parser) Parse() (*models.Borme, error) {
	// Initialize Borme object
	borme := &models.Borme{
		Anuncios: make(map[int]*models.BormeAnuncio),
	}
	p.data = borme

	// Initialize actos slice
	p.actos = make([]models.BormeActo, 0)

	// For now, we'll try to read the PDF content
	// In production, we would use a proper PDF library
	// This is a placeholder that shows the structure

	// Try to read as text (some PDFs are text-based)
	text, err := p.readPDFText()
	if err != nil {
		log.Printf("Warning: Could not read PDF as text: %v", err)
	}

	if text != "" {
		state := &ParserState{}
		p.processText(text, state)
	}

	// Set announcement range
	if len(borme.Anuncios) > 0 {
		minID := -1
		maxID := -1
		for id := range borme.Anuncios {
			if minID == -1 || id < minID {
				minID = id
			}
			if maxID == -1 || id > maxID {
				maxID = id
			}
		}
		borme.SetAnunciosRango(minID, maxID)
	}

	return borme, nil
}

// readPDFText attempts to read PDF content
func (p *PyPDF2Parser) readPDFText() (string, error) {
	file, err := os.Open(p.filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Try to read as text first
	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	text := strings.Join(lines, "\n")

	// Check if it looks like a PDF
	if strings.Contains(text, "%PDF") {
		// It's a binary PDF, need proper library
		// For now, return empty
		return "", fmt.Errorf("binary PDF - requires proper PDF library")
	}

	return text, nil
}

// processText processes PDF text content
func (p *PyPDF2Parser) processText(content string, state *ParserState) {
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		// Check for markers
		switch {
		case strings.Contains(line, "Cabecera"):
			state.Cabecera = true
			state.Texto = false
			state.Fecha = false
			state.Numero = false
			state.Seccion = false
			state.Provincia = false
			state.CVE = false

		case strings.Contains(line, "Texto"):
			state.Texto = true
			state.Cabecera = false

		case strings.Contains(line, "Fecha"):
			state.Fecha = true
			state.Numero = false
			state.Seccion = false
			state.Provincia = false
			state.CVE = false

		case strings.HasPrefix(line, "/F1"):
			// Bold font - might be acto name
			name := extractAfterFont(line, "/F1")
			if name != "" && !strings.HasPrefix(name, "/") {
				state.CurrentActo = regex.CleanPDFText(name)
			}

		case strings.HasPrefix(line, "/F2"):
			// Normal font - acto value
			value := extractAfterFont(line, "/F2")
			if value != "" && state.CurrentActo != "" {
				p.parseActoValue(state.CurrentActo, regex.CleanPDFText(value))
				state.CurrentActo = ""
			}

		case strings.HasPrefix(line, "Núm.") || strings.HasPrefix(line, "Num."):
			// BORME number
			if match := regex.REGEX_BORME_NUM.FindStringSubmatch(line); match != nil {
				if num, err := strconv.Atoi(match[1]); err == nil {
					p.data.Num = num
				}
			}

		case strings.HasPrefix(line, "cve:") || strings.Contains(line, "CVE"):
			// CVE code
			if match := regex.REGEX_BORME_CVE.FindStringSubmatch(line); match != nil {
				p.data.SetCVE(match[1])
			}

		case state.Fecha && !state.Numero:
			// Parse date
			if t, err := regex.ParseFecha(line); err == nil && !t.IsZero() {
				p.data.Date = t
			}

		case state.Provincia && !state.CVE:
			// Parse province
			if prov := models.FromTitle(line); prov != nil {
				p.data.Provincia = prov
			}

		case state.CVE:
			// Parse CVE
			if match := regex.REGEX_BORME_CVE.FindStringSubmatch(line); match != nil {
				p.data.SetCVE(match[1])
			}

		case state.Cabecera:
			// Parse empresa header
			p.parseCabecera(line)

		case state.Texto && state.CurrentActo != "":
			// Parse acto text
			p.parseActoValue(state.CurrentActo, regex.CleanPDFText(line))
		}
	}
}

// extractAfterFont extracts text after font marker
func extractAfterFont(line, font string) string {
	idx := strings.Index(line, font)
	if idx == -1 {
		return ""
	}
	text := line[idx+len(font):]
	text = strings.TrimSpace(text)
	// Remove PDF operators
	text = strings.TrimSuffix(text, " Tj")
	text = strings.TrimSuffix(text, " Tj")
	return text
}

// parseCabecera parses the company header
func (p *PyPDF2Parser) parseCabecera(line string) {
	// Check if this line contains empresa info
	if strings.Contains(line, " - ") {
		id, name, registro := regex.ParseEmpresa(line)
		if id != "" {
			// Create new anuncio
			anuncio := &models.BormeAnuncio{
				ID:      len(p.data.Anuncios) + 1,
				Empresa: name,
			}
			if registro != nil {
				anuncio.Registro = registro["registro"]
			}
			p.data.Anuncios[anuncio.ID] = anuncio
		}
	}
}

// parseActoValue parses the value of an acto
func (p *PyPDF2Parser) parseActoValue(name, value string) {
	// Clean the value
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}

	// Create acto based on type
	if regex.IsActoCargo(name) {
		// Parse cargos
		cargos := regex.ParseCargos(value)

		acto := &models.BormeActoCargo{
			Name:  name,
			Value: cargos,
		}
		p.actos = append(p.actos, acto)

	} else if regex.IsActoBold(name) {
		// Bold acto
		acto := &models.BormeActoTexto{
			Name:  name,
			Value: &value,
		}
		p.actos = append(p.actos, acto)

	} else if regex.IsActoColon(name) {
		// Acto with colon argument
		acto := &models.BormeActoTexto{
			Name:  name,
			Value: &value,
		}
		p.actos = append(p.actos, acto)

	} else {
		// Regular acto
		acto := &models.BormeActoTexto{
			Name:  name,
			Value: &value,
		}
		p.actos = append(p.actos, acto)
	}
}

// ParseFilename extracts date and section from filename
// Expected format: BORME-{seccion}-{year}-{month}-{day}.pdf
// Example: BORME-A-2015-10-27.pdf
func ParseFilename(filename string) (time.Time, models.Seccion, int, error) {
	base := strings.TrimSuffix(filename, ".pdf")
	parts := strings.Split(base, "-")

	if len(parts) < 5 {
		return time.Time{}, "", 0, fmt.Errorf("invalid filename format: %s", filename)
	}

	seccion := models.Seccion(parts[1])

	year, err := strconv.Atoi(parts[2])
	if err != nil {
		return time.Time{}, "", 0, fmt.Errorf("invalid year in filename: %s", filename)
	}

	month, err := strconv.Atoi(parts[3])
	if err != nil {
		return time.Time{}, "", 0, fmt.Errorf("invalid month in filename: %s", filename)
	}

	day, err := strconv.Atoi(parts[4])
	if err != nil {
		return time.Time{}, "", 0, fmt.Errorf("invalid day in filename: %s", filename)
	}

	date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	nbo := date.YearDay()

	return date, seccion, nbo, nil
}

// PDFTextExtractor extracts text from PDF files
type PDFTextExtractor struct{}

// NewPDFTextExtractor creates a new extractor
func NewPDFTextExtractor() *PDFTextExtractor {
	return &PDFTextExtractor{}
}

// Extract extracts text from a PDF file
func (e *PDFTextExtractor) Extract(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Read first bytes to check if it's a text PDF
	header := make([]byte, 100)
	_, err = file.Read(header)
	if err != nil && err != io.EOF {
		return "", err
	}

	// Check if it's a text-based PDF
	if string(header[:4]) == "%PDF" {
		// Try to read as text
		file.Seek(0, 0)
		scanner := bufio.NewScanner(file)
		var lines []string
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		return strings.Join(lines, "\n"), nil
	}

	return "", fmt.Errorf("binary PDF - requires proper PDF parsing library")
}
