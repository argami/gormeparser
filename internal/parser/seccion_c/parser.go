package seccionc

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/argami/gormeparser/internal/models"
	"github.com/antchfx/xmlquery"
)

// LxmlBormeCParser parses Section C XML/HTML announcements
type LxmlBormeCParser struct {
	filename string
}

// NewParser creates a new Section C parser
func NewParser(filename string) *LxmlBormeCParser {
	return &LxmlBormeCParser{
		filename: filename,
	}
}

// Parse parses a Section C file (XML or HTML) and returns a BormeC object
func (p *LxmlBormeCParser) Parse() (*models.BormeC, error) {
	// Open file
	file, err := os.Open(p.filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read content
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Detect format
	contentStr := string(content)
	if strings.Contains(contentStr, "<?xml") || strings.Contains(contentStr, "<xml") {
		return p.parseXML(content)
	}

	return p.parseHTML(content)
}

// parseXML parses XML format
func (p *LxmlBormeCParser) parseXML(content []byte) (*models.BormeC, error) {
	doc, err := xmlquery.Parse(strings.NewReader(string(content)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	borme := models.NewBormeC()

	// Extract fields from XML structure
	// Common XML structure for BORME-C:
	// <borme ...> or <diario ...>

	// Extract departamento (department)
	depto := xmlquery.FindOne(doc, "//departamento|//Departamento|//department")
	if depto != nil {
		borme.Departamento = strings.TrimSpace(depto.Data)
	}

	// Extract texto (announcement text)
	texto := xmlquery.FindOne(doc, "//texto|//Texto|//announcement_text")
	if texto != nil {
		borme.Texto = strings.TrimSpace(texto.Data)
	}

	// Extract diario_numero (daily bulletin number)
	nbo := xmlquery.FindOne(doc, "//diario_numero|//DiarioNumero|//nbo")
	if nbo != nil {
		fmt.Sscanf(nbo.Data, "%d", &borme.DiarioNumero)
	}

	// Extract numero_anuncio (announcement number)
	numAnuncio := xmlquery.FindOne(doc, "//numero_anuncio|//NumeroAnuncio|//num")
	if numAnuncio != nil {
		borme.NumeroAnuncio = strings.TrimSpace(numAnuncio.Data)
	}

	// Extract id_anuncio (full ID like "A110044738")
	idAnuncio := xmlquery.FindOne(doc, "//id_anuncio|//IdAnuncio|//id")
	if idAnuncio != nil {
		borme.IDAnuncio = strings.TrimSpace(idAnuncio.Data)
	}

	// Extract CVE
	cve := xmlquery.FindOne(doc, "//cve|//CVE|//verificacion")
	if cve != nil {
		borme.CVE = strings.TrimSpace(cve.Data)
	}

	// Extract titulo (title)
	titulo := xmlquery.FindOne(doc, "//titulo|//Titulo|//title")
	if titulo != nil {
		borme.Titulo = strings.TrimSpace(titulo.Data)
	}

	// Extract empresa (company name)
	empresa := xmlquery.FindOne(doc, "//empresa|//Empresa|//company")
	if empresa != nil {
		borme.Empresa = strings.TrimSpace(empresa.Data)
	}

	// Extract CIFs
	cifs := xmlquery.Find(doc, "//cif|//CIF|//nif")
	for _, cif := range cifs {
		if cif.Data != "" {
			borme.AddCIF(strings.TrimSpace(cif.Data))
		}
	}

	// Extract empresas_relacionadas (related companies for mergers)
	relacionadas := xmlquery.Find(doc, "//empresas_relacionadas|//relacionada|//related_company")
	for _, rel := range relacionadas {
		if rel.Data != "" {
			borme.AddEmpresaRelacionada(strings.TrimSpace(rel.Data))
		}
	}

	// Extract pages
	paginaIni := xmlquery.FindOne(doc, "//pagina_inicial|//pagina")
	if paginaIni != nil {
		fmt.Sscanf(paginaIni.Data, "%d", &borme.PaginaInicial)
	}

	// Extract fecha (date)
	fecha := xmlquery.FindOne(doc, "//fecha|//Fecha|//date")
	if fecha != nil {
		if t, err := time.Parse("2006-01-02", fecha.Data); err == nil {
			borme.Fecha = t
		}
	}

	// Set filename
	filename := p.filename
	borme.Filename = &filename

	return borme, nil
}

// parseHTML parses HTML format
func (p *LxmlBormeCParser) parseHTML(content []byte) (*models.BormeC, error) {
	doc, err := xmlquery.Parse(strings.NewReader(string(content)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	borme := models.NewBormeC()

	// Extract from HTML structure
	// Common elements in BORME-C HTML

	// Extract titulo
	titulo := xmlquery.FindOne(doc, "//h1|//h2|//h3|//title")
	if titulo != nil {
		borme.Titulo = strings.TrimSpace(titulo.Data)
	}

	// Extract texto from paragraph elements
	paras := xmlquery.Find(doc, "//p|//div[@class='texto']")
	for _, para := range paras {
		borme.Texto += " " + strings.TrimSpace(para.Data)
	}
	borme.Texto = strings.TrimSpace(borme.Texto)

	// Extract empresa from headers or specific elements
	empresa := xmlquery.FindOne(doc, "//strong|//b|//span[@class='empresa']")
	if empresa != nil {
		borme.Empresa = strings.TrimSpace(empresa.Data)
	}

	// Set filename
	filename := p.filename
	borme.Filename = &filename

	return borme, nil
}

// ParseMultipleXML parses multiple announcements from an XML file
func ParseMultipleXML(filename string) ([]models.BormeC, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Try to parse as XML
	doc, err := xmlquery.Parse(strings.NewReader(string(content)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	var results []models.BormeC

	// Find all announcement elements
	anuncios := xmlquery.Find(doc, "//anuncio|//Anuncio|//announcement")

	for _, anuncio := range anuncios {
		borme := models.NewBormeC()

		// Extract fields from each announcement
		depto := xmlquery.FindOne(anuncio, "./departamento|./Departamento")
		if depto != nil {
			borme.Departamento = strings.TrimSpace(depto.Data)
		}

		texto := xmlquery.FindOne(anuncio, "./texto|./Texto")
		if texto != nil {
			borme.Texto = strings.TrimSpace(texto.Data)
		}

		empresa := xmlquery.FindOne(anuncio, "./empresa|./Empresa")
		if empresa != nil {
			borme.Empresa = strings.TrimSpace(empresa.Data)
		}

		cve := xmlquery.FindOne(anuncio, "./cve|./CVE")
		if cve != nil {
			borme.CVE = strings.TrimSpace(cve.Data)
		}

		results = append(results, *borme)
	}

	return results, nil
}
