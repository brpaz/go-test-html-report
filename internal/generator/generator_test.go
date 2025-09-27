package generator_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/brpaz/go-test-html-report/internal/generator"
)

func TestGenerator_New(t *testing.T) {
	t.Parallel()

	t.Run("with valid options", func(t *testing.T) {
		t.Parallel()
		tempFile := filepath.Join(os.TempDir(), "valid.json")
		f, _ := os.Create(tempFile)
		_ = f.Close()
		defer func() { _ = os.Remove(tempFile) }()

		gen, err := generator.New(
			generator.WithInputFile(tempFile),
			generator.WithOutputFile("output.html"),
			generator.WithTitle("Test Report"),
		)

		assert.NoError(t, err)
		assert.NotNil(t, gen)
	})

	t.Run(("with missing input file"), func(t *testing.T) {
		t.Parallel()
		gen, err := generator.New(
			generator.WithOutputFile("output.html"),
			generator.WithTitle("Test Report"),
		)

		assert.Error(t, err)
		assert.ErrorIs(t, err, generator.ErrInputFileIsRequired)
		assert.Nil(t, gen)
	})

	t.Run(("with missing output file"), func(t *testing.T) {
		t.Parallel()
		tempFile := filepath.Join(os.TempDir(), "valid.json")
		f, _ := os.Create(tempFile)
		_ = f.Close()
		defer func() { _ = os.Remove(tempFile) }()

		gen, err := generator.New(
			generator.WithInputFile(tempFile),
			generator.WithTitle("Test Report"),
		)

		assert.Error(t, err)
		assert.ErrorIs(t, err, generator.ErrOutputFileIsRequired)
		assert.Nil(t, gen)
	})

	t.Run(("with non-existent input file"), func(t *testing.T) {
		t.Parallel()
		gen, err := generator.New(
			generator.WithInputFile("nonexistent.json"),
			generator.WithOutputFile("output.html"),
			generator.WithTitle("Test Report"),
		)
		assert.Error(t, err)
		assert.ErrorIs(t, err, generator.ErrInputFileDoesNotExist)
		assert.Nil(t, gen)
	})
}

func TestGenerator_GenerateReport(t *testing.T) {
	t.Run("successfully generates HTML report from valid JSON input", func(t *testing.T) {
		// Use internal testdata file instead of creating temporary file
		inputFile := "../testdata/testjson.json"
		outputFile := filepath.Join(os.TempDir(), "output.html")
		defer func() { _ = os.Remove(outputFile) }()

		// Create generator and generate report
		gen, err := generator.New(
			generator.WithInputFile(inputFile),
			generator.WithOutputFile(outputFile),
			generator.WithTitle("Test Report"),
		)
		require.NoError(t, err)

		err = gen.Generate(context.Background())
		assert.NoError(t, err)

		// Verify output file was created and contains expected content
		_, err = os.Stat(outputFile)
		assert.NoError(t, err)

		content, err := os.ReadFile(outputFile)
		assert.NoError(t, err)
		contentStr := string(content)

		// Check for expected HTML structure and content
		assert.Contains(t, contentStr, "<html")
		assert.Contains(t, contentStr, "Test Report")
		// Check for content from the testdata file
		assert.Contains(t, contentStr, "TestUserLogin")
		assert.Contains(t, contentStr, "github.com/myproject/internal/auth") // package name from testdata
		assert.Contains(t, contentStr, "TestDatabaseConnection")
		assert.Contains(t, contentStr, "github.com/myproject/internal/database") // another package from testdata
	})

	t.Run("successfully generates HTML report from stdin data", func(t *testing.T) {
		// Read the testdata file to use as stdin input
		testData, err := os.ReadFile("../testdata/testjson.json")
		require.NoError(t, err)

		outputFile := filepath.Join(os.TempDir(), "output_stdin.html")
		defer func() { _ = os.Remove(outputFile) }()

		// Create generator with input data instead of file
		gen, err := generator.New(
			generator.WithInputData(testData),
			generator.WithOutputFile(outputFile),
			generator.WithTitle("Stdin Test Report"),
		)
		require.NoError(t, err)

		err = gen.Generate(context.Background())
		assert.NoError(t, err)

		// Verify output file was created and contains expected content
		_, err = os.Stat(outputFile)
		assert.NoError(t, err)

		content, err := os.ReadFile(outputFile)
		assert.NoError(t, err)
		contentStr := string(content)

		// Check for expected HTML structure and content
		assert.Contains(t, contentStr, "<html")
		assert.Contains(t, contentStr, "Stdin Test Report")
		// Check for content from the testdata
		assert.Contains(t, contentStr, "TestUserLogin")
		assert.Contains(t, contentStr, "github.com/myproject/internal/auth")
	})
}
