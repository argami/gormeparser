package download

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Constants from Python's download.py
const (
	// URL_BASE is the base URL for BOE
	URLBase = "https://www.boe.es"

	// Protocol for downloads
	Protocol = "https"

	// Threads for parallel downloads
	Threads = 8

	// Download retry settings
	MaxRetries = 3
	RetryDelay = 1 * time.Second
)

// BORME PDF URL patterns
const (
	BormeABPDFURL = Protocol + "://boe.es/borme/dias/{year}/{month:02d}/{day:02d}/pdfs/BORME-{seccion}-{year}-{nbo}-{provincia}.pdf"
)

// BORME XML URL
const (
	BormeXMLURL = Protocol + "://www.boe.es/diario_borme/xml.php?id=BORME-S-{year}{month:02d}{day:02d}"
)

// BORME Section C URLs
const (
	BormeCHTMURL  = Protocol + "://boe.es/diario_borme/txt.php?id=BORME-C-{year}-{anuncio}"
	BormeCPDFURL  = Protocol + "://boe.es/borme/dias/{year}/{month:02d}/{day:02d}/pdfs/BORME-C-{year}-{anuncio}.pdf"
	BormeCXMLURL  = Protocol + "://boe.es/diario_borme/xml.php?id=BORME-C-{year}-{anuncio}"
)

// Error types
type DownloadError struct {
	Op  string
	URL string
	Err error
}

func (e *DownloadError) Error() string {
	return fmt.Sprintf("%s: %s: %v", e.Op, e.URL, e.Err)
}

func (e *DownloadError) Unwrap() error {
	return e.Err
}

// HTTPClient is the HTTP client for downloads
var HTTPClient = &http.Client{
	Timeout: 30 * time.Second,
}

// DownloadFile downloads a file from url to dest
func DownloadFile(urlStr, dest string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(dest)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return &DownloadError{Op: "mkdir", URL: urlStr, Err: err}
	}

	// Open file for writing
	out, err := os.Create(dest)
	if err != nil {
		return &DownloadError{Op: "create", URL: urlStr, Err: err}
	}
	defer out.Close()

	// Download with retries
	var lastErr error
	for attempt := 0; attempt < MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(RetryDelay * time.Duration(attempt))
		}

		resp, err := HTTPClient.Get(urlStr)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
			continue
		}

		_, err = io.Copy(out, resp.Body)
		if err != nil {
			lastErr = err
			continue
		}

		return nil
	}

	return &DownloadError{Op: "download", URL: urlStr, Err: lastErr}
}

// DownloadBytes downloads a URL and returns the body as bytes
func DownloadBytes(urlStr string) ([]byte, error) {
	var body []byte
	var lastErr error

	for attempt := 0; attempt < MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(RetryDelay * time.Duration(attempt))
		}

		resp, err := HTTPClient.Get(urlStr)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
			continue
		}

		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(resp.Body)
		if err != nil {
			lastErr = err
			continue
		}
		body = buf.Bytes()
		return body, nil
	}

	return nil, &DownloadError{Op: "download", URL: urlStr, Err: lastErr}
}

// GetURLPDF returns the URL for a BORME PDF
func GetURLPDF(date time.Time, seccion string, provincia string) string {
	// Get day of year (nbo)
	yearDay := date.YearDay()
	nbo := yearDay

	// Format URL
	urlStr := strings.ReplaceAll(BormeABPDFURL, "{year}", strconv.Itoa(date.Year()))
	urlStr = strings.ReplaceAll(urlStr, "{month}", strconv.Itoa(int(date.Month())))
	urlStr = strings.ReplaceAll(urlStr, "{day}", strconv.Itoa(date.Day()))
	urlStr = strings.ReplaceAll(urlStr, "{seccion}", seccion)
	urlStr = strings.ReplaceAll(urlStr, "{nbo}", strconv.Itoa(nbo))
	urlStr = strings.ReplaceAll(urlStr, "{provincia}", provincia)

	return urlStr
}

// GetURLXML returns the URL for the daily XML index
func GetURLXML(date time.Time) string {
	urlStr := strings.ReplaceAll(BormeXMLURL, "{year}", strconv.Itoa(date.Year()))
	urlStr = strings.ReplaceAll(urlStr, "{month}", strconv.Itoa(int(date.Month())))
	urlStr = strings.ReplaceAll(urlStr, "{day}", strconv.Itoa(date.Day()))

	return urlStr
}

// GetURLSeccionC returns URLs for Section C announcements
func GetURLSeccionC(date time.Time, format string) map[string]string {
	urls := make(map[string]string)

	switch format {
	case "htm":
		urls["url_txt"] = strings.ReplaceAll(BormeCHTMURL, "{year}", strconv.Itoa(date.Year()))
		urls["url_txt"] = strings.ReplaceAll(urls["url_txt"], "{anuncio}", "")
	case "pdf":
		urls["url_pdf"] = strings.ReplaceAll(BormeCPDFURL, "{year}", strconv.Itoa(date.Year()))
		urls["url_pdf"] = strings.ReplaceAll(urls["url_pdf"], "{month}", strconv.Itoa(int(date.Month())))
		urls["url_pdf"] = strings.ReplaceAll(urls["url_pdf"], "{day}", strconv.Itoa(date.Day()))
		urls["url_pdf"] = strings.ReplaceAll(urls["url_pdf"], "{anuncio}", "")
	case "xml":
		urls["url_xml"] = strings.ReplaceAll(BormeCXMLURL, "{year}", strconv.Itoa(date.Year()))
		urls["url_xml"] = strings.ReplaceAll(urls["url_xml"], "{anuncio}", "")
	}

	return urls
}

// DownloadXML downloads the daily XML index
func DownloadXML(date time.Time, filename string) error {
	urlStr := GetURLXML(date)
	return DownloadFile(urlStr, filename)
}

// DownloadPDF downloads a BORME PDF
func DownloadPDF(date time.Time, filename string, seccion string, provincia string) error {
	urlStr := GetURLPDF(date, seccion, provincia)
	return DownloadFile(urlStr, filename)
}

// BormeXMLIndex represents the XML index structure
type BormeXMLIndex struct {
	XMLName   xml.Name       `xml:"diario"`
	Date      string         `xml:"date"`
	NBO       int            `xml:"nbo"`
	Secciones []SeccionData  `xml:"seccion"`
}

// SeccionData represents a section in the XML index
type SeccionData struct {
	Letra   string     `xml:"letra,attr"`
	Empresa []EmpresaData `xml:"empresa"`
}

// EmpresaData represents a company in the XML index
type EmpresaData struct {
	IDCia      string     `xml:"id"`
	Provincia  string     `xml:"provincia"`
	NumBorme   string     `xml:"nbo"`
	NumAnuncio string     `xml:"num"`
	NumPagina  string     `xml:"pag"`
	URLPDF     string     `xml:"urlpdf"`
	URLCVE     string     `xml:"urlcve"`
}

// ParseXMLIndex parses the XML index and returns CVE URLs
func ParseXMLIndex(data []byte) ([]string, error) {
	var index BormeXMLIndex
	if err := xml.Unmarshal(data, &index); err != nil {
		return nil, err
	}

	var urls []string
	for _, seccion := range index.Secciones {
		for _, empresa := range seccion.Empresa {
			if empresa.URLCVE != "" {
				urls = append(urls, empresa.URLCVE)
			}
		}
	}

	return urls, nil
}

// DownloadURLs downloads multiple URLs in parallel
func DownloadURLs(urls []string, path string, names []string) map[string]string {
	results := make(map[string]string)
	sem := make(chan struct{}, Threads)

	for i, u := range urls {
		name := u
		if i < len(names) && names[i] != "" {
			name = names[i]
		}

		sem <- struct{}{}
		go func(urlStr, filename string) {
			defer func() { <-sem }()

			filepath := filepath.Join(path, filename)
			if err := DownloadFile(urlStr, filepath); err != nil {
				log.Printf("Error downloading %s: %v", urlStr, err)
				results[urlStr] = ""
			} else {
				results[urlStr] = filepath
			}
		}(u, name)
	}

	// Wait for all downloads to complete
	for i := 0; i < Threads; i++ {
		sem <- struct{}{}
	}

	return results
}

// GetNBOFromXML extracts the bulletin number from XML
func GetNBOFromXML(data []byte) (int, error) {
	var index BormeXMLIndex
	if err := xml.Unmarshal(data, &index); err != nil {
		return 0, err
	}
	return index.NBO, nil
}

// ValidateURL checks if a URL is valid
func ValidateURL(urlStr string) bool {
	u, err := url.Parse(urlStr)
	return err == nil && u.Scheme != "" && u.Host != ""
}
