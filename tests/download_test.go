package gormeparser_test

import (
	"net/url"
	"time"

	"github.com/argami/gormeparser/internal/download"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Download", func() {
	ginkgo.Describe("GetURLPDF", func() {
		ginkgo.It("should generate correct Section A URL", func() {
			date := time.Date(2015, 10, 27, 0, 0, 0, 0, time.UTC)
			urlStr := download.GetURLPDF(date, "A", "Madrid")
			gomega.Expect(urlStr).To(gomega.ContainSubstring("BORME-A-2015-"))
			gomega.Expect(urlStr).To(gomega.ContainSubstring("Madrid"))
			gomega.Expect(urlStr).To(gomega.ContainSubstring("boe.es"))
		})

		ginkgo.It("should handle Barcelona province", func() {
			date := time.Date(2015, 10, 27, 0, 0, 0, 0, time.UTC)
			urlStr := download.GetURLPDF(date, "A", "Barcelona")
			gomega.Expect(urlStr).To(gomega.ContainSubstring("Barcelona"))
		})

		ginkgo.It("should generate valid URL", func() {
			date := time.Date(2015, 10, 27, 0, 0, 0, 0, time.UTC)
			urlStr := download.GetURLPDF(date, "A", "Madrid")
			parsedURL, err := url.Parse(urlStr)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			gomega.Expect(parsedURL.Scheme).To(gomega.Equal("https"))
			gomega.Expect(parsedURL.Host).To(gomega.ContainSubstring("boe.es"))
		})
	})

	ginkgo.Describe("GetURLXML", func() {
		ginkgo.It("should generate XML index URL", func() {
			date := time.Date(2015, 9, 24, 0, 0, 0, 0, time.UTC)
			urlStr := download.GetURLXML(date)
			gomega.Expect(urlStr).To(gomega.ContainSubstring("BORME-S-2015"))
			gomega.Expect(urlStr).To(gomega.ContainSubstring("xml.php"))
		})
	})

	ginkgo.Describe("DownloadFile", func() {
		ginkgo.It("should handle invalid URL gracefully", func() {
			tempFile := "/tmp/test_download_invalid.pdf"
			err := download.DownloadFile("http://invalid.invalid.invalid/file.pdf", tempFile)
			gomega.Expect(err).To(gomega.HaveOccurred())
		})
	})

	ginkgo.Describe("DownloadBytes", func() {
		ginkgo.It("should handle invalid URL gracefully", func() {
			_, err := download.DownloadBytes("http://invalid.invalid.invalid/file.pdf")
			gomega.Expect(err).To(gomega.HaveOccurred())
		})
	})
})
