package gormeparser_test

import (
	"github.com/argami/gormeparser/internal/models"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Section Constants", func() {
	ginkgo.Describe("Section Types", func() {
		ginkgo.It("should have correct SeccionA value", func() {
			gomega.Expect(string(models.SeccionA)).To(gomega.Equal("A"))
		})

		ginkgo.It("should have correct SeccionB value", func() {
			gomega.Expect(string(models.SeccionB)).To(gomega.Equal("B"))
		})

		ginkgo.It("should have correct SeccionC value", func() {
			gomega.Expect(string(models.SeccionC)).To(gomega.Equal("C"))
		})
	})

	ginkgo.Describe("ActoCargo Type Constants", func() {
		ginkgo.It("should have ActoCargoNombramientos constant", func() {
			gomega.Expect(string(models.ActoCargoNombramientos)).To(gomega.Equal("Nombramientos"))
		})

		ginkgo.It("should have ActoCargoCesesDimisiones constant", func() {
			gomega.Expect(string(models.ActoCargoCesesDimisiones)).To(gomega.Equal("Ceses/Dimisiones"))
		})

		ginkgo.It("should have ActoCargoConstitucion constant", func() {
			gomega.Expect(string(models.ActoCargoConstitucion)).To(gomega.Equal("Constitucion"))
		})

		ginkgo.It("should have ActoCargoDisolucion constant", func() {
			gomega.Expect(string(models.ActoCargoDisolucion)).To(gomega.Equal("Disolucion"))
		})
	})

	ginkgo.Describe("ActoNoArg Type Constants", func() {
		ginkgo.It("should have ActoNoArgOtro constant", func() {
			gomega.Expect(string(models.ActoNoArgOtro)).To(gomega.Equal("Otro"))
		})

		ginkgo.It("should have ActoNoArgExtincion constant", func() {
			gomega.Expect(string(models.ActoNoArgExtincion)).To(gomega.Equal("Extincion"))
		})
	})

	ginkgo.Describe("ActoColon Type Constants", func() {
		ginkgo.It("should have ActoColonModificacionDuracion constant", func() {
			gomega.Expect(string(models.ActoColonModificacionDuracion)).To(gomega.Equal("Modificacion de duracion"))
		})

		ginkgo.It("should have ActoColonCapital constant", func() {
			gomega.Expect(string(models.ActoColonCapital)).To(gomega.Equal("Capital"))
		})
	})

	ginkgo.Describe("ActoBold Type Constants", func() {
		ginkgo.It("should have ActoBoldDisolucion constant", func() {
			gomega.Expect(string(models.ActoBoldDisolucion)).To(gomega.Equal("Disolucion"))
		})

		ginkgo.It("should have ActoBoldFusion constant", func() {
			gomega.Expect(string(models.ActoBoldFusion)).To(gomega.Equal("Fusion"))
		})
	})

	ginkgo.Describe("Subseccion Constants", func() {
		ginkgo.It("should have SubseccionActosInscritos constant", func() {
			gomega.Expect(string(models.SubseccionActosInscritos)).To(gomega.Equal("A"))
		})

		ginkgo.It("should have SubseccionOtrosActos constant", func() {
			gomega.Expect(string(models.SubseccionOtrosActos)).To(gomega.Equal("B"))
		})
	})
})
