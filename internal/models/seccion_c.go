package models

import "time"

// BormeC represents a Section C announcement (XML/HTML format)
type BormeC struct {
	Departamento       string    `json:"departamento"`
	Texto              string    `json:"texto"`
	DiarioNumero       int      `json:"diario_numero"`
	NumeroAnuncio      string    `json:"numero_anuncio"`
	IDAnuncio          string    `json:"id_anuncio"`
	PaginaInicial      int      `json:"pagina_inicial"`
	PaginaFinal        int      `json:"pagina_final"`
	Fecha              time.Time `json:"fecha"`
	Titulo             string    `json:"titulo"`
	Empresa            string    `json:"empresa"`
	EmpresasRelacionadas []string `json:"empresas_relacionadas,omitempty"`
	CIFs               []string  `json:"cifs,omitempty"`
	CVE                string    `json:"cve"`
	Seccion            Seccion   `json:"seccion"`
	Filename           *string   `json:"filename,omitempty"`
}

// BormeCSearchResult represents search results for Section C
type BormeCSearchResult struct {
	Anuncios []BormeC `json:"anuncios"`
	Total    int       `json:"total"`
}

// BormeXML represents the XML index file for a daily bulletin
type BormeXML struct {
	Date      time.Time `json:"date"`
	NBO       int       `json:"nbo"`
	PrevBorme *time.Time `json:"prev_borme,omitempty"`
	NextBorme *time.Time `json:"next_borme,omitempty"`
	IsFinal   bool      `json:"is_final"`
	URLs      []string  `json:"urls,omitempty"`
}

// NewBormeC creates a new Section C announcement
func NewBormeC() *BormeC {
	return &BormeC{
		EmpresasRelacionadas: make([]string, 0),
		CIFs:                make([]string, 0),
		Seccion:             SeccionC,
	}
}

// AddEmpresaRelacionada adds a related company (for mergers, acquisitions)
func (b *BormeC) AddEmpresaRelacionada(empresa string) {
	b.EmpresasRelacionadas = append(b.EmpresasRelacionadas, empresa)
}

// AddCIF adds a CIF/NIF number
func (b *BormeC) AddCIF(cif string) {
	b.CIFs = append(b.CIFs, cif)
}
