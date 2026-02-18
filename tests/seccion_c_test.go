package gormeparser_test

import (
	seccionc "github.com/argami/gormeparser/internal/parser/seccion_c"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Section C Parser", func() {
	ginkgo.Describe("NewParser", func() {
		ginkgo.It("should create new parser", func() {
			parser := seccionc.NewParser("testdata/BORME-C-2011-20488.xml")
			gomega.Expect(parser).ToNot(gomega.BeNil())
		})

		ginkgo.It("should return error for non-existent file", func() {
			parser := seccionc.NewParser("testdata/nonexistent.xml")
			_, err := parser.Parse()
			gomega.Expect(err).To(gomega.HaveOccurred())
		})
	})

	ginkgo.Describe("ParseMultipleXML", func() {
		ginkgo.It("should return error for non-existent file", func() {
			_, err := seccionc.ParseMultipleXML("testdata/nonexistent.xml")
			gomega.Expect(err).To(gomega.HaveOccurred())
		})
	})
})
