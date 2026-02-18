package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Seccion represents the BORME section (A, B, or C)
type Seccion string

const (
	SeccionA Seccion = "A"
	SeccionB Seccion = "B"
	SeccionC Seccion = "C"
)

// Provincia represents a Spanish province
type Provincia struct {
	Code int    `json:"code"`
	Name string `json:"name"`
}

var Provincias = map[string]Provincia{
	"ALAVA":              {1, "Alava"},
	"ALBACETE":           {2, "Albacete"},
	"ALICANTE":           {3, "Alicante"},
	"ALMERIA":            {4, "Almeria"},
	"ARABA":              {1, "Araba/Álava"},
	"ASTURIAS":           {330, "Asturias"},
	"AVILA":              {50, "Avila"},
	"BADAJOZ":            {60, "Badajoz"},
	"BARCELONA":          {80, "Barcelona"},
	"BISCAY":             {48, "Bizkaia"},
	"BURGOS":             {90, "Burgos"},
	"CACERES":            {100, "Caceres"},
	"CADIZ":              {110, "Cadiz"},
	"CANTABRIA":          {390, "Cantabria"},
	"CASTELLON":          {120, "Castellon"},
	"CEUTA":             {510, "Ceuta"},
	"CORDOBA":           {140, "Cordoba"},
	"CUENCA":            {160, "Cuenca"},
	"GIPUZCOA":          {200, "Gipuzkoa"},
	"GIRONA":            {170, "Girona"},
	"GRANADA":           {180, "Granada"},
	"GUADALAJARA":       {190, "Guadalajara"},
	"HUELVA":            {210, "Huelva"},
	"HUESCA":            {220, "Huesca"},
	"ILLES BALEARS":     {70, "Illes Balears"},
	"JAEN":              {230, "Jaen"},
	"LA CORUÑA":         {150, "La Coruña"},
	"LA RIOJA":          {260, "La Rioja"},
	"LAS PALMAS":        {350, "Las Palmas"},
	"LEON":              {240, "Leon"},
	"LLEIDA":            {250, "Lleida"},
	"LUGO":              {270, "Lugo"},
	"MADRID":            {280, "Madrid"},
	"MALAGA":            {290, "Malaga"},
	"MELILLA":           {520, "Melilla"},
	"MURCIA":            {300, "Murcia"},
	"NAVARRA":           {310, "Navarra"},
	"OURENSE":           {320, "Ourense"},
	"PALENCIA":          {340, "Palencia"},
	"PONTEVEDRA":        {360, "Pontevedra"},
	"SALAMANCA":         {370, "Salamanca"},
	"SEGOVIA":           {400, "Segovia"},
	"SEVILLA":           {410, "Sevilla"},
	"SORIA":             {420, "Soria"},
	"TARRAGONA":         {430, "Tarragona"},
	"TERUEL":            {440, "Teruel"},
	"TOLEDO":            {450, "Toledo"},
	"VALENCIA":          {460, "Valencia"},
	"VALLADOLID":        {470, "Valladolid"},
	"ZAMORA":            {490, "Zamora"},
	"ZARAGOZA":          {500, "Zaragoza"},
}

// FromTitle returns the Provincia matching a title (case-insensitive partial match)
func FromTitle(title string) *Provincia {
	title = strings.ToUpper(title)
	for _, p := range Provincias {
		if strings.Contains(title, p.Name) || strings.Contains(title, strings.ReplaceAll(p.Name, "Á", "A")) {
			return &p
		}
	}
	return nil
}

// BormeActo is the base interface for act types
type BormeActo interface {
	GetName() string
	GetValue() interface{}
}

// BormeActoTexto represents a text-only act (e.g., "Constitución", "Disolución")
type BormeActoTexto struct {
	Name  string   `json:"name"`
	Value *string  `json:"value,omitempty"`
}

func (a *BormeActoTexto) GetName() string   { return a.Name }
func (a *BormeActoTexto) GetValue() interface{} {
	if a.Value == nil {
		return nil
	}
	return *a.Value
}

// BormeActoCargo represents an act with cargo assignments (e.g., "Nombramientos", "Ceses")
type BormeActoCargo struct {
	Name  string              `json:"name"`
	Value map[string][]string `json:"value"` // cargo type -> list of person names
}

func (a *BormeActoCargo) GetName() string   { return a.Name }
func (a *BormeActoCargo) GetValue() interface{} { return a.Value }

// GetNombresCargos returns the cargo types (keys of the value map)
func (a *BormeActoCargo) GetNombresCargos() []string {
	if a.Value == nil {
		return nil
	}
	keys := make([]string, 0, len(a.Value))
	for k := range a.Value {
		keys = append(keys, k)
	}
	return keys
}

// BormeAnuncio represents a single announcement in the BORME
type BormeAnuncio struct {
	ID              int          `json:"id"`
	Empresa         string       `json:"empresa"`
	Registro        string       `json:"registro,omitempty"`
	Sucursal        bool         `json:"sucursal,omitempty"`
	Liquidacion     bool         `json:"liquidacion,omitempty"`
	DatosRegistrales string      `json:"datos_registrales,omitempty"`
	Actos           []BormeActo  `json:"actos"`
}

func (a *BormeAnuncio) GetBormeActos() []BormeActo {
	return a.Actos
}

func (a *BormeAnuncio) String() string {
	return fmt.Sprintf("BormeAnuncio(ID=%d, Empresa=%s)", a.ID, a.Empresa)
}

// Borme represents a complete BORME bulletin
type Borme struct {
	Date           time.Time      `json:"date"`
	Seccion        Seccion        `json:"seccion"`
	Provincia      *Provincia     `json:"provincia,omitempty"`
	Num            int            `json:"num"`
	CVE            string         `json:"cve,omitempty"`
	Filename       *string        `json:"filename,omitempty"`
	Anuncios       map[int]*BormeAnuncio `json:"anuncios"`
	AnunciosRango [2]int         `json:"anuncios_rango,omitempty"`
}

// NewBorme creates a new Borme instance
func NewBorme(date time.Time, seccion Seccion, provincia *Provincia, num int) *Borme {
	return &Borme{
		Date:      date,
		Seccion:   seccion,
		Provincia: provincia,
		Num:       num,
		Anuncios:  make(map[int]*BormeAnuncio),
	}
}

// SetCVE sets the CVE (Código de Verificación Electrónica)
func (b *Borme) SetCVE(cve string) {
	b.CVE = cve
}

// SetFilename sets the source filename
func (b *Borme) SetFilename(filename string) {
	b.Filename = &filename
}

// AddAnuncio adds an announcement to the bulletin
func (b *Borme) AddAnuncio(a *BormeAnuncio) {
	b.Anuncios[a.ID] = a
}

// SetAnunciosRango sets the min/max announcement IDs
func (b *Borme) SetAnunciosRango(minID, maxID int) {
	b.AnunciosRango = [2]int{minID, maxID}
}

// BormeToJSON serializes Borme to JSON
func BormeToJSON(b *Borme, pretty bool) ([]byte, error) {
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

// BormeFromJSON deserializes JSON to Borme
func BormeFromJSON(data []byte) (*Borme, error) {
	var b Borme
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, err
	}
	if b.Anuncios == nil {
		b.Anuncios = make(map[int]*BormeAnuncio)
	}
	return &b, nil
}

// BormeAnuncioList is a helper for JSON serialization of announcements
type BormeAnuncioList struct {
	Anuncios []*BormeAnuncio `json:"anuncios"`
}

// ToJSON serializes BormeAnuncio list to JSON
func BormeAnuncioListToJSON(anuncios []*BormeAnuncio, pretty bool) ([]byte, error) {
	list := BormeAnuncioList{Anuncios: anuncios}
	if pretty {
		return json.MarshalIndent(list, "", "  ")
	}
	return json.Marshal(list)
}
