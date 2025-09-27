package generator

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"os"
	"time"

	"github.com/brpaz/go-test-html-report/internal/report"
	"github.com/brpaz/go-test-html-report/internal/test2json"
)

//go:embed templates/*.html
var templateFS embed.FS

const (
	ReportTemplate = "templates/report.html"
	ReportTitle    = "Go Test Report"
)

// HTMLGenerator is the main struct for generating HTML reports.
type HTMLGenerator struct {
	InputFile  string
	OutputFile string
	Title      string
}

// Validate checks if the HTMLGenerator has valid configuration.
func (g *HTMLGenerator) Validate() error {
	var errs []error
	if g.InputFile == "" {
		errs = append(errs, errors.New("input file is required"))
	}
	if g.OutputFile == "" {
		errs = append(errs, errors.New("output file is required"))
	}

	// Check if input file exists
	if _, err := os.Stat(g.InputFile); os.IsNotExist(err) {
		errs = append(errs, fmt.Errorf("input file does not exist: %s", g.InputFile))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

type Option func(*HTMLGenerator) error

// WithInputFile sets the input file for the HTML report generator.
func WithInputFile(inputFile string) Option {
	return func(g *HTMLGenerator) error {
		g.InputFile = inputFile
		return nil
	}
}

// WithOutputFile sets the output file for the HTML report generator.
func WithOutputFile(outputFile string) Option {
	return func(g *HTMLGenerator) error {
		g.OutputFile = outputFile
		return nil
	}
}

// WithTitle sets the title for the HTML report.
func WithTitle(title string) Option {
	return func(g *HTMLGenerator) error {
		g.Title = title
		return nil
	}
}

// NewHTMLReportGenerator creates a new HTMLReportGenerator with the provided options.
func NewHTMLReportGenerator(opts ...Option) (*HTMLGenerator, error) {
	g := &HTMLGenerator{
		Title: ReportTitle,
	}
	for _, opt := range opts {
		opt(g)
	}

	if err := g.Validate(); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}

	return g, nil
}

// TemplateData holds the data passed to the HTML template
type TemplateData struct {
	TestSuites    []report.TestSuite
	GeneratedAt   string
	Title         string
	TotalTests    int
	PassedTests   int
	FailedTests   int
	SkippedTests  int
	TotalDuration float64
}

// Generate generates the HTML report from the input file to the output file.
func (g *HTMLGenerator) Generate(ctx context.Context) error {
	fmt.Printf("Generating HTML report from %s to %s\n", g.InputFile, g.OutputFile)

	data, err := os.ReadFile(g.InputFile)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	testSuites, err := test2json.Parse(data)
	if err != nil {
		return fmt.Errorf("failed to parse input file: %w", err)
	}

	templateData := g.prepareTemplateData(testSuites)

	// Create and parse template
	tmpl, err := template.New("report.html").Funcs(template.FuncMap{
		"statusClass": g.getStatusClass,
		"statusIcon":  g.getStatusIcon,
		"formatDuration": g.formatDuration,
		"title": func(s string) string {
			if len(s) == 0 {
				return s
			}
			return fmt.Sprintf("%c%s", s[0]-32, s[1:])
		},
	}).ParseFS(templateFS, ReportTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Create output file
	outFile, err := os.Create(g.OutputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Execute template
	if err := tmpl.Execute(outFile, templateData); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	fmt.Printf("HTML report generated successfully: %s\n", g.OutputFile)
	return nil
}

// prepareTemplateData calculates test statistics for the template
func (g *HTMLGenerator) prepareTemplateData(testSuites []report.TestSuite) TemplateData {
	var totalTests, passedTests, failedTests, skippedTests int
	var totalDuration float64

	for _, suite := range testSuites {
		for _, testCase := range suite.TestCases {
			totalTests++
			totalDuration += testCase.Duration
			switch testCase.Status {
			case "pass":
				passedTests++
			case "fail":
				failedTests++
			case "skip":
				skippedTests++
			}
		}
	}

	return TemplateData{
		TestSuites:    testSuites,
		GeneratedAt:   time.Now().Format(time.RFC3339),
		Title:         g.Title,
		TotalTests:    totalTests,
		PassedTests:   passedTests,
		FailedTests:   failedTests,
		SkippedTests:  skippedTests,
		TotalDuration: totalDuration,
	}
}

// getStatusClass returns Tailwind CSS classes for test status
func (g *HTMLGenerator) getStatusClass(status string) string {
	switch status {
	case "pass":
		return "bg-green-100 text-green-800 border-green-200"
	case "fail":
		return "bg-red-100 text-red-800 border-red-200"
	case "skip":
		return "bg-yellow-100 text-yellow-800 border-yellow-200"
	default:
		return "bg-gray-100 text-gray-800 border-gray-200"
	}
}

// getStatusIcon returns an icon for test status
func (g *HTMLGenerator) getStatusIcon(status string) string {
	switch status {
	case "pass":
		return "✓"
	case "fail":
		return "✗"
	case "skip":
		return "⚠"
	default:
		return "?"
	}
}

// formatDuration formats duration in seconds to a readable string
func (g *HTMLGenerator) formatDuration(seconds float64) string {
	if seconds < 0.001 {
		return "< 1ms"
	}
	if seconds < 1 {
		return fmt.Sprintf("%.0fms", seconds*1000)
	}
	return fmt.Sprintf("%.2fs", seconds)
}

