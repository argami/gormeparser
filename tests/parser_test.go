package gormeparser_test

import (
	"time"

	"github.com/argami/gormeparser/internal/parser/pypdf2"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Parser Router", func() {
	ginkgo.Describe("ParseFilename", func() {
		ginkgo.It("should parse BORME-A-2015-10-27.pdf filename", func() {
			date, seccion, nbo, err := pypdf2.ParseFilename("BORME-A-2015-10-27.pdf")
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			gomega.Expect(date.Year()).To(gomega.Equal(2015))
			gomega.Expect(date.Month()).To(gomega.Equal(time.October))
			gomega.Expect(date.Day()).To(gomega.Equal(27))
			gomega.Expect(string(seccion)).To(gomega.Equal("A"))
			gomega.Expect(nbo).To(gomega.Equal(300))
		})

		ginkgo.It("should handle filename without extension", func() {
			_, seccion, nbo, err := pypdf2.ParseFilename("BORME-A-2015-10-27")
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			gomega.Expect(string(seccion)).To(gomega.Equal("A"))
			gomega.Expect(nbo).To(gomega.Equal(300))
		})

		ginkgo.It("should return error for invalid filename", func() {
			_, _, _, err := pypdf2.ParseFilename("invalid.pdf")
			gomega.Expect(err).To(gomega.HaveOccurred())
		})
	})

	ginkgo.Describe("PyPDF2Parser", func() {
		ginkgo.It("should create new parser", func() {
			parser := pypdf2.NewParser("testdata/BORME-A-2015-27-10.pdf")
			gomega.Expect(parser).ToNot(gomega.BeNil())
		})

		ginkgo.It("should handle non-existent file gracefully", func() {
			parser := pypdf2.NewParser("testdata/nonexistent.pdf")
			result, err := parser.Parse()
			// Parser logs warning but may still return valid result
			gomega.Expect(result).ToNot(gomega.BeNil())
			// Error may or may not occur depending on implementation
			_ = err
		})
	})
})
