package regex

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Compiled regex patterns from Python's regex.py

// REGEX_EMPRESA matches "57344 - ALDARA CATERING SL"
var REGEX_EMPRESA = regexp.MustCompile(`^(\d+) - (.*?)\.?$`)

// REGEX_EMPRESA_REGISTRO matches company with register like "57344 - ALDARA CATERING SL(R.M. Madrid)"
var REGEX_EMPRESA_REGISTRO = regexp.MustCompile(`^(\d+) - (.*)\(R\.M\. (.*)\)\.?$`)

// REGEX_PDF_TEXT matches PDF text markers like "(...)Tj"
var REGEX_PDF_TEXT = regexp.MustCompile(`^\((.*)\)Tj$`)

// REGEX_BORME_NUM matches BORME number like "Núm. 57344"
var REGEX_BORME_NUM = regexp.MustCompile(`^Núm\. (\d+)`)

// REGEX_BORME_FECHA matches date format like "Martes 2 de junio de 2015"
var REGEX_BORME_FECHA = regexp.MustCompile(`^\w+ (\d+) de (\w+) de (\d+)`)

// REGEX_BORME_CVE matches CVE identifier like "cve: BORME-A-2015-101-29"
var REGEX_BORME_CVE = regexp.MustCompile(`^cve: (.*)$`)

// REGEX_ARGCOLON matches acto with colon argument (e.g., "Capital: 3.000 EUR")
var REGEX_ARGCOLON = regexp.MustCompile(`^(.*):\s*(.*)$`)

// REGEX_NOARG matches acto without arguments (simple bold text)
var REGEX_NOARG = regexp.MustCompile(`^(.+)$`)

// REGEX_BOLD matches bold actos with arguments
var REGEX_BOLD = regexp.MustCompile(`^(.*?):\s*(.+)$`)

// RE_CARGOS_MATCH matches cargo assignments like "Adm. Solid.: JUAN PEREZ;MARIA GARCIA"
var RE_CARGOS_MATCH = regexp.MustCompile(`([A-Za-z\.\s]+):\s*([^;]+);?`)

// RE_CARGOS_SEPARADOR separates cargo assignments
var RE_CARGOS_SEPARADOR = regexp.MustCompile(`;\s*`)

// Regex empresa result
type EmpresaMatch struct {
	ID        string
	Name      string
	Extra     string
	Registro  string
}

// RegexCargosResult represents parsed cargo assignments
type RegexCargosResult struct {
	Cargos map[string][]string
}

// RegexFechaResult represents parsed date
type RegexFechaResult struct {
	Year   int
	Month  int
	Day    int
}

// RegexBoldActoResult represents parsed bold acto
type RegexBoldActoResult struct {
	Name  string
	Value string
}

// RegexArgColonResult represents parsed acto with colon
type RegexArgColonResult struct {
	Name  string
	Value string
}

// ParseEmpresa parses company string like "57344 - ALDARA CATERING SL"
// Returns: id, name, extra_dict (registro info)
func ParseEmpresa(s string) (id string, name string, registro map[string]string) {
	// First try REGEX_EMPRESA_REGISTRO
	if match := REGEX_EMPRESA_REGISTRO.FindStringSubmatch(s); match != nil {
		return match[1], match[2], map[string]string{"registro": match[3]}
	}

	// Then try REGEX_EMPRESA
	if match := REGEX_EMPRESA.FindStringSubmatch(s); match != nil {
		return match[1], match[2], nil
	}

	return "", s, nil
}

// ParseCargos parses cargo assignments like "Adm. Solid.: JUAN PEREZ;MARIA GARCIA"
// Returns: map[cargo_type][]person_names
func ParseCargos(s string) map[string][]string {
	result := make(map[string][]string)

	matches := RE_CARGOS_MATCH.FindAllStringSubmatch(s, -1)
	for _, match := range matches {
		if len(match) >= 3 {
			cargo := strings.TrimSpace(match[1])
			personas := RE_CARGOS_SEPARADOR.Split(strings.TrimSpace(match[2]), -1)

			for _, p := range personas {
				p = strings.TrimSpace(p)
				if p != "" {
					result[cargo] = append(result[cargo], p)
				}
			}
		}
	}

	return result
}

// ParseBoldActo parses bold acto with arguments
// Returns: (name, value) or ("", original if no match)
func ParseBoldActo(s string) (name string, value string) {
	match := REGEX_BOLD.FindStringSubmatch(s)
	if match != nil && len(match) >= 3 {
		return strings.TrimSpace(match[1]), strings.TrimSpace(match[2])
	}
	return "", s
}

// ParseArgColon parses acto with colon argument
// Returns: (name, value) or ("", original if no match)
func ParseArgColon(s string) (name string, value string) {
	match := REGEX_ARGCOLON.FindStringSubmatch(s)
	if match != nil && len(match) >= 3 {
		name = strings.TrimSpace(match[1])
		value = strings.TrimSpace(match[2])
		// Check if name is an acto that doesn't take arguments
		if IsActoNoArg(name) {
			return "", s
		}
		return name, value
	}
	return "", s
}

// ParseNoArg parses acto without arguments
func ParseNoArg(s string) string {
	match := REGEX_NOARG.FindStringSubmatch(s)
	if match != nil && len(match) >= 2 {
		return strings.TrimSpace(match[1])
	}
	return s
}

// ParseFecha parses Spanish date format like "Martes 2 de junio de 2015"
func ParseFecha(s string) (time.Time, error) {
	match := REGEX_BORME_FECHA.FindStringSubmatch(s)
	if match == nil || len(match) < 4 {
		return time.Time{}, nil
	}

	day := match[1]
	monthStr := match[2]
	year := match[3]

	// Parse month name to number
	monthMap := map[string]int{
		"enero":      1,
		"febrero":    2,
		"marzo":      3,
		"abril":      4,
		"mayo":       5,
		"junio":      6,
		"julio":      7,
		"agosto":     8,
		"septiembre": 9,
		"setiembre":  9,
		"octubre":    10,
		"noviembre":  11,
		"diciembre":  12,
	}

	month, ok := monthMap[strings.ToLower(monthStr)]
	if !ok {
		return time.Time{}, nil
	}

	// Parse day and year
	var dayInt int
	_, err := fmt.Sscanf(day, "%d", &dayInt)
	if err != nil {
		return time.Time{}, nil
	}

	var yearInt int
	_, err = fmt.Sscanf(year, "%d", &yearInt)
	if err != nil {
		return time.Time{}, nil
	}

	return time.Date(yearInt, time.Month(month), dayInt, 0, 0, 0, 0, time.UTC), nil
}

// Acto types that take cargo arguments
var actosConCargo = map[string]bool{
	"Nombramientos":            true,
	"Revocaciones":             true,
	"Ceses/Dimisiones":         true,
	"Nombramiento":             true,
	"Reelecciones":             true,
	"Socio único":              true,
	"Socio profesional":         true,
	"Otro cargo":               true,
	"Disolución":               false,
	"Constitución":             false,
}

// IsActoCargo returns true if the acto type takes cargo arguments
func IsActoCargo(actoType string) bool {
	return actosConCargo[actoType]
}

// Acto types that don't take arguments
var actosSinArg = map[string]bool{
	"Crédito incobrable":       true,
	"Sociedad unipersonal":     true,
	"Extinción":               true,
	"Cuadro de cargos":        true,
	"Cambio de objeto social":  true,
	"Otro acto":               true,
}

// IsActoNoArg returns true if the acto type doesn't take arguments
func IsActoNoArg(actoType string) bool {
	return actosSinArg[actoType]
}

// Acto types with colon arguments
var actosConColon = map[string]bool{
	"Modificación de duración": true,
	"Fe de erratas":           true,
	"Domicilio":               true,
	"Objeto":                  true,
	"Capital":                 true,
	"Estatutos":               true,
	"Denominación":            true,
}

// IsActoColon returns true if the acto type has colon arguments
func IsActoColon(actoType string) bool {
	return actosConColon[actoType]
}

// Bold acto types
var actosBold = map[string]bool{
	"Declaración de unipersonalidad": true,
	"Sociedad unipersonal":          true,
	"Escisión total":                true,
	"Fusión":                       true,
}

// IsActoBold returns true if the acto type is a bold acto
func IsActoBold(actoType string) bool {
	return actosBold[actoType]
}

// IsCompany returns true if the entity is a company (has SL/SA suffix)
func IsCompany(name string) bool {
	upper := strings.ToUpper(name)
	return strings.HasSuffix(upper, " SL") ||
		strings.HasSuffix(upper, ", SL") ||
		strings.HasSuffix(upper, " S.L.") ||
		strings.HasSuffix(upper, " SA") ||
		strings.HasSuffix(upper, ", SA") ||
		strings.HasSuffix(upper, " S.A.") ||
		strings.HasSuffix(upper, " S.L.L.") ||
		strings.HasSuffix(upper, " S.A.L.")
}

// CleanPDFText removes PDF encoding artifacts from text
func CleanPDFText(s string) string {
	// Remove Tj markers if present
	if match := REGEX_PDF_TEXT.FindStringSubmatch(s); match != nil {
		s = match[1]
	}

	// Unescape PDF special characters
	s = strings.ReplaceAll(s, "\\(", "(")
	s = strings.ReplaceAll(s, "\\)", ")")
	s = strings.ReplaceAll(s, "\\ ", " ")

	// Remove double spaces
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}

	return strings.TrimSpace(s)
}

// CapitalizeSentence capitalizes the first letter of a sentence
func CapitalizeSentence(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}
