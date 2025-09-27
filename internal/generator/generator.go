package generator

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"os"
	"time"

	"github.com/brpaz/go-test-html-report/internal/domain"
	"github.com/brpaz/go-test-html-report/internal/test2json"
)

type HTMLGenerator struct {
	InputFile  string
	OutputFile string
}

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

// NewHTMLReportGenerator creates a new HTMLReportGenerator with the provided options.
func NewHTMLReportGenerator(opts ...Option) (*HTMLGenerator, error) {
	g := &HTMLGenerator{}
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
	TestSuites   []domain.TestSuite
	GeneratedAt  string
	TotalTests   int
	PassedTests  int
	FailedTests  int
	SkippedTests int
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

	// Calculate statistics
	templateData := g.calculateStats(testSuites)

	// Create and parse template
	tmpl, err := template.New("report").Funcs(template.FuncMap{
		"statusClass": g.getStatusClass,
		"statusIcon":  g.getStatusIcon,
		"formatDuration": g.formatDuration,
		"title": func(s string) string {
			if len(s) == 0 {
				return s
			}
			return fmt.Sprintf("%c%s", s[0]-32, s[1:])
		},
	}).Parse(htmlTemplate)
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

// calculateStats calculates test statistics for the template
func (g *HTMLGenerator) calculateStats(testSuites []domain.TestSuite) TemplateData {
	var totalTests, passedTests, failedTests, skippedTests int

	for _, suite := range testSuites {
		for _, testCase := range suite.TestCases {
			totalTests++
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
		TestSuites:   testSuites,
		GeneratedAt:  time.Now().Format("2006-01-02 15:04:05"),
		TotalTests:   totalTests,
		PassedTests:  passedTests,
		FailedTests:  failedTests,
		SkippedTests: skippedTests,
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

// htmlTemplate contains the HTML template with Tailwind CSS
const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Go Test Report</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script>
        tailwind.config = {
            theme: {
                extend: {
                    fontFamily: {
                        mono: ['JetBrains Mono', 'Consolas', 'Monaco', 'Courier New', 'monospace']
                    }
                }
            }
        }
    </script>
</head>
<body class="bg-gray-50 font-sans">
    <div class="min-h-screen">
        <!-- Header -->
        <header class="bg-white shadow-sm border-b border-gray-200">
            <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                <div class="py-6">
                    <h1 class="text-3xl font-bold text-gray-900">Go Test Report</h1>
                    <p class="mt-2 text-sm text-gray-600">Generated on {{.GeneratedAt}}</p>
                </div>
            </div>
        </header>

        <!-- Statistics -->
        <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
            <div class="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
                <div class="bg-white overflow-hidden shadow rounded-lg">
                    <div class="p-5">
                        <div class="flex items-center">
                            <div class="flex-shrink-0">
                                <div class="w-8 h-8 bg-blue-500 rounded-full flex items-center justify-center">
                                    <span class="text-white text-sm font-semibold">T</span>
                                </div>
                            </div>
                            <div class="ml-5 w-0 flex-1">
                                <dl>
                                    <dt class="text-sm font-medium text-gray-500 truncate">Total Tests</dt>
                                    <dd class="text-lg font-medium text-gray-900">{{.TotalTests}}</dd>
                                </dl>
                            </div>
                        </div>
                    </div>
                </div>

                <div class="bg-white overflow-hidden shadow rounded-lg">
                    <div class="p-5">
                        <div class="flex items-center">
                            <div class="flex-shrink-0">
                                <div class="w-8 h-8 bg-green-500 rounded-full flex items-center justify-center">
                                    <span class="text-white text-sm font-semibold">✓</span>
                                </div>
                            </div>
                            <div class="ml-5 w-0 flex-1">
                                <dl>
                                    <dt class="text-sm font-medium text-gray-500 truncate">Passed</dt>
                                    <dd class="text-lg font-medium text-green-600">{{.PassedTests}}</dd>
                                </dl>
                            </div>
                        </div>
                    </div>
                </div>

                <div class="bg-white overflow-hidden shadow rounded-lg">
                    <div class="p-5">
                        <div class="flex items-center">
                            <div class="flex-shrink-0">
                                <div class="w-8 h-8 bg-red-500 rounded-full flex items-center justify-center">
                                    <span class="text-white text-sm font-semibold">✗</span>
                                </div>
                            </div>
                            <div class="ml-5 w-0 flex-1">
                                <dl>
                                    <dt class="text-sm font-medium text-gray-500 truncate">Failed</dt>
                                    <dd class="text-lg font-medium text-red-600">{{.FailedTests}}</dd>
                                </dl>
                            </div>
                        </div>
                    </div>
                </div>

                <div class="bg-white overflow-hidden shadow rounded-lg">
                    <div class="p-5">
                        <div class="flex items-center">
                            <div class="flex-shrink-0">
                                <div class="w-8 h-8 bg-yellow-500 rounded-full flex items-center justify-center">
                                    <span class="text-white text-sm font-semibold">⚠</span>
                                </div>
                            </div>
                            <div class="ml-5 w-0 flex-1">
                                <dl>
                                    <dt class="text-sm font-medium text-gray-500 truncate">Skipped</dt>
                                    <dd class="text-lg font-medium text-yellow-600">{{.SkippedTests}}</dd>
                                </dl>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Test Suites -->
            <div class="space-y-8">
                {{range .TestSuites}}
                <div class="bg-white shadow rounded-lg overflow-hidden">
                    <div class="bg-gray-50 px-6 py-4 border-b border-gray-200">
                        <h2 class="text-lg font-medium text-gray-900 font-mono">{{.Name}}</h2>
                        <p class="text-sm text-gray-500 mt-1">{{len .TestCases}} test(s)</p>
                    </div>

                    <div class="divide-y divide-gray-200">
                        {{range .TestCases}}
                        <div class="p-6">
                            <div class="flex items-start justify-between">
                                <div class="flex items-start space-x-3 flex-1">
                                    <div class="flex-shrink-0">
                                        <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium border {{statusClass .Status}}">
                                            {{statusIcon .Status}} {{.Status | title}}
                                        </span>
                                    </div>
                                    <div class="flex-1 min-w-0">
                                        <h3 class="text-sm font-medium text-gray-900 font-mono">{{.Name}}</h3>
                                        <div class="mt-2 flex items-center space-x-4">
                                            <span class="text-sm text-gray-500">
                                                Duration: {{formatDuration .Duration}}
                                            </span>
                                        </div>

                                        {{if .Output}}
                                        <div class="mt-4">
                                            <details class="group">
                                                <summary class="cursor-pointer text-sm font-medium text-gray-700 hover:text-gray-900">
                                                    <span class="group-open:hidden">Show output ({{len .Output}} line(s))</span>
                                                    <span class="hidden group-open:inline">Hide output</span>
                                                </summary>
                                                <div class="mt-3 bg-gray-900 rounded-md p-4 overflow-x-auto">
                                                    <pre class="text-green-400 text-xs font-mono whitespace-pre-wrap">{{range .Output}}{{.}}
{{end}}</pre>
                                                </div>
                                            </details>
                                        </div>
                                        {{end}}
                                    </div>
                                </div>
                            </div>
                        </div>
                        {{end}}
                    </div>
                </div>
                {{end}}
            </div>
        </div>

        <!-- Footer -->
        <footer class="bg-white border-t border-gray-200 mt-12">
            <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
                <p class="text-center text-sm text-gray-500">
                    Generated by go-test-html-report on {{.GeneratedAt}}
                </p>
            </div>
        </footer>
    </div>

    <script>
        // Add some interactivity
        document.addEventListener('DOMContentLoaded', function() {
            // Smooth scrolling for anchor links
            document.querySelectorAll('a[href^="#"]').forEach(anchor => {
                anchor.addEventListener('click', function (e) {
                    e.preventDefault();
                    document.querySelector(this.getAttribute('href')).scrollIntoView({
                        behavior: 'smooth'
                    });
                });
            });
        });
    </script>
</body>
</html>`

