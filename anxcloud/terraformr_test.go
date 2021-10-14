package anxcloud

import (
	"strings"
	"testing"

	"github.com/go-logr/logr"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Terraformr", func() {
	var writer *strings.Builder
	var logger logr.Logger

	BeforeEach(func() {
		writer = &strings.Builder{}
	})

	Context("Logger with base verbosity", func() {
		BeforeEach(func() {
			logger = NewTerraformr(writer)
		})

		It("logs errors with error prefix", func() {
			logger.Error(nil, "Something went wrong. Please contact support.")

			str := writer.String()
			Expect(str).To(HavePrefix("[ERROR] "))
			Expect(str).To(ContainSubstring("Something went wrong. Please contact support."))
		})

		It("logs verbosity 0 with warning prefix", func() {
			logger.V(0).Info("Strange things happen.")

			str := writer.String()
			Expect(str).To(HavePrefix("[WARN] "))
			Expect(str).To(ContainSubstring("Strange things happen."))
		})

		It("logs verbosity 1 with info prefix", func() {
			logger.V(1).Info("Everything is fine.")

			str := writer.String()
			Expect(str).To(HavePrefix("[INFO] "))
			Expect(str).To(ContainSubstring("Everything is fine."))
		})

		It("logs verbosity 2 with debug prefix", func() {
			logger.V(2).Info("What is happening here?!")

			str := writer.String()
			Expect(str).To(HavePrefix("[DEBUG] "))
			Expect(str).To(ContainSubstring("What is happening here?!"))
		})

		It("logs verbosity 3 with trace prefix", func() {
			logger.V(3).Info("Tracing the code")

			str := writer.String()
			Expect(str).To(HavePrefix("[TRACE] "))
			Expect(str).To(ContainSubstring("Tracing the code"))
		})

		It("logs verbosity 4 with trace+1 prefix", func() {
			logger.V(4).Info("Tracing the code")

			str := writer.String()
			Expect(str).To(HavePrefix("[TRACE]+1 "))
			Expect(str).To(ContainSubstring("Tracing the code"))
		})
	})

	Context("Logger with verbosity 1", func() {
		BeforeEach(func() {
			logger = NewTerraformr(writer).V(1)
		})

		It("logs info prefix", func() {
			logger.Info("Everything is fine.")

			str := writer.String()
			Expect(str).To(HavePrefix("[INFO] "))
			Expect(str).To(ContainSubstring("Everything is fine."))
		})
	})

	Context("Logger with verbosity 2", func() {
		BeforeEach(func() {
			logger = NewTerraformr(writer).V(2)
		})

		It("logs debug prefix", func() {
			logger.Info("What is happening here?!")

			str := writer.String()
			Expect(str).To(HavePrefix("[DEBUG] "))
			Expect(str).To(ContainSubstring("What is happening here?!"))
		})
	})

	Context("Logger with verbosity 3", func() {
		BeforeEach(func() {
			logger = NewTerraformr(writer).V(3)
		})

		It("logs trace prefix", func() {
			logger.Info("Tracing the code")

			str := writer.String()
			Expect(str).To(HavePrefix("[TRACE] "))
			Expect(str).To(ContainSubstring("Tracing the code"))
		})
	})

	Context("Logger with a name", func() {
		BeforeEach(func() {
			logger = NewTerraformr(writer).WithName("some_logger_name")
		})

		It("logs its name", func() {
			logger.Info("Hello world!")

			str := writer.String()
			Expect(str).To(HavePrefix("[WARN] "))
			Expect(str).To(ContainSubstring("Hello world!"))
			Expect(str).To(ContainSubstring("some_logger_name"))
		})
	})

	Context("Logger with values", func() {
		BeforeEach(func() {
			logger = NewTerraformr(writer).WithValues("some", "value")
		})

		It("logs its name", func() {
			logger.Info("Hello world!")

			str := writer.String()
			Expect(str).To(HavePrefix("[WARN] "))
			Expect(str).To(ContainSubstring("Hello world!"))
			Expect(str).To(ContainSubstring(`"some"="value"`))
		})
	})
})

func TestTerraformrSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "anxcloud suite")
}
