package generator_test

import (
	"testing"

	"github.com/brpaz/go-test-html-report/internal/generator"
	"github.com/stretchr/testify/assert"
)

func TestNewHTMLReportGenerator(t *testing.T) {
	t.Parallel()
	opts := []generator.Option{
		generator.WithInputFile("input.txt"),
		generator.WithOutputFile("output.html"),
	}
	generator, err := generator.NewHTMLReportGenerator(opts...)
	assert.NoError(t, err)
	assert.NotNil(t, generator)
}
