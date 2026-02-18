package gormeparser_test

import (
	"time"

	"github.com/argami/gormeparser/internal/regex"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Regex", func() {
	ginkgo.Describe("ParseEmpresa", func() {
		ginkgo.It("should parse empresa without registro", func() {
			id, name, registro := regex.ParseEmpresa("57344 - ALDARA CATERING SL")
			gomega.Expect(id).To(gomega.Equal("57344"))
			gomega.Expect(name).To(gomega.Equal("ALDARA CATERING SL"))
			gomega.Expect(registro).To(gomega.BeNil())
		})

		ginkgo.It("should parse empresa with registro", func() {
			id, name, registro := regex.ParseEmpresa("57344 - ALDARA CATERING SL(R.M. Madrid)")
			gomega.Expect(id).To(gomega.Equal("57344"))
			gomega.Expect(name).To(gomega.Equal("ALDARA CATERING SL"))
			gomega.Expect(registro).ToNot(gomega.BeNil())
			gomega.Expect(registro["registro"]).To(gomega.Equal("Madrid"))
		})

		ginkgo.It("should parse first empresa from PDF sample", func() {
			id, name, _ := regex.ParseEmpresa("1 - ALDARA CATERING SL")
			gomega.Expect(id).To(gomega.Equal("1"))
			gomega.Expect(name).To(gomega.Equal("ALDARA CATERING SL"))
		})
	})

	ginkgo.Describe("ParseCargos", func() {
		ginkgo.It("should parse cargo with single person", func() {
			cargos := regex.ParseCargos("Adm. Solid.: JUAN PEREZ")
			gomega.Expect(cargos).To(gomega.HaveKey("Adm. Solid."))
			gomega.Expect(cargos["Adm. Solid."]).To(gomega.ContainElement("JUAN PEREZ"))
		})
	})

	ginkgo.Describe("ParseFecha", func() {
		ginkgo.It("should parse Spanish date", func() {
			t, err := regex.ParseFecha("Martes 27 de octubre de 2015")
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			gomega.Expect(t.Year()).To(gomega.Equal(2015))
			gomega.Expect(t.Month()).To(gomega.Equal(time.October))
			gomega.Expect(t.Day()).To(gomega.Equal(27))
		})

		ginkgo.It("should parse lowercase month", func() {
			t, _ := regex.ParseFecha("lunes 2 de junio de 2015")
			gomega.Expect(t.Month()).To(gomega.Equal(time.June))
			gomega.Expect(t.Day()).To(gomega.Equal(2))
		})

		ginkgo.It("should parse full month names", func() {
			t, _ := regex.ParseFecha("Viernes 15 de enero de 2021")
			gomega.Expect(t.Month()).To(gomega.Equal(time.January))
			gomega.Expect(t.Year()).To(gomega.Equal(2021))
		})
	})

	ginkgo.Describe("IsCompany", func() {
		ginkgo.It("should identify SL suffix", func() {
			gomega.Expect(regex.IsCompany("ACME SL")).To(gomega.BeTrue())
			gomega.Expect(regex.IsCompany("ACME, SL")).To(gomega.BeTrue())
		})

		ginkgo.It("should identify SA suffix", func() {
			gomega.Expect(regex.IsCompany("ACME SA")).To(gomega.BeTrue())
		})

		ginkgo.It("should reject non-company names", func() {
			gomega.Expect(regex.IsCompany("Juan Perez")).To(gomega.BeFalse())
			gomega.Expect(regex.IsCompany("Calle Mayor 123")).To(gomega.BeFalse())
			gomega.Expect(regex.IsCompany("")).To(gomega.BeFalse())
		})
	})

	ginkgo.Describe("CleanPDFText", func() {
		ginkgo.It("should unescape parentheses", func() {
			cleaned := regex.CleanPDFText("Constitucion \\(Sociedad Limitada\\)")
			gomega.Expect(cleaned).To(gomega.Equal("Constitucion (Sociedad Limitada)"))
		})
	})

	ginkgo.Describe("ParseBoldActo", func() {
		ginkgo.It("should handle input", func() {
			name, _ := regex.ParseBoldActo("Test Acto")
			gomega.Expect(name).To(gomega.Equal(""))
		})
	})

	ginkgo.Describe("IsActoCargo", func() {
		ginkgo.It("should identify cargo acto types", func() {
			gomega.Expect(regex.IsActoCargo("Nombramientos")).To(gomega.BeTrue())
			gomega.Expect(regex.IsActoCargo("Ceses/Dimisiones")).To(gomega.BeTrue())
		})

		ginkgo.It("should reject non-cargo acto types", func() {
			gomega.Expect(regex.IsActoCargo("Constitucion")).To(gomega.BeFalse())
		})
	})

	ginkgo.Describe("IsActoNoArg", func() {
		ginkgo.It("should return false for unknown types", func() {
			gomega.Expect(regex.IsActoNoArg("Unknown")).To(gomega.BeFalse())
		})
	})

	ginkgo.Describe("IsActoColon", func() {
		ginkgo.It("should return false for unknown types", func() {
			gomega.Expect(regex.IsActoColon("Unknown")).To(gomega.BeFalse())
		})
	})

	ginkgo.Describe("IsActoBold", func() {
		ginkgo.It("should return false for unknown types", func() {
			gomega.Expect(regex.IsActoBold("Unknown")).To(gomega.BeFalse())
		})
	})

	ginkgo.Describe("CapitalizeSentence", func() {
		ginkgo.It("should capitalize first letter", func() {
			result := regex.CapitalizeSentence("hola mundo")
			gomega.Expect(result).To(gomega.Equal("Hola mundo"))
		})
	})
})
