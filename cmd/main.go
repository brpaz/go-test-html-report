package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/brpaz/go-test-html-report/internal/generator"
)

var (
	version = "0.0.0-dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd := &cli.Command{
		Name:    "go-test-html-report",
		Usage:   "generate HTML report from Go test results",
		Version: fmt.Sprintf("%s (commit: %s, date: %s)", version, commit, date),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output file for the HTML report",
				Value:   "report.html",
			},
			&cli.StringFlag{
				Name:    "input",
				Aliases: []string{"i"},
				Usage:   "Input file containing Go test results (use '-' for stdin)",
				Value:   "test_results.json",
			},
			&cli.StringFlag{
				Name:    "title",
				Aliases: []string{"t"},
				Usage:   "Title for the HTML report",
				Value:   "Go Test Report",
			},
		},
		EnableShellCompletion: true,
		Description: `
			go-test-html-report is a CLI tool that generates an HTML report from Go test results.
			It processes the output of 'go test' and creates a user-friendly HTML report.

			Examples:
			  go test -json ./... | go-test-html-report -i - -o report.html
			  go-test-html-report -i test_results.json -o report.html`,
		Action: actionGenerateReport,
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		_, _ = cmd.ErrWriter.Write([]byte(err.Error() + "\n"))
	}
}

func actionGenerateReport(ctx context.Context, cmd *cli.Command) error {
	inputFile := cmd.String("input")
	outputFile := cmd.String("output")
	title := cmd.String("title")

	var opts []generator.Option
	if inputFile == "-" {
		// Read from stdin
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}
		opts = []generator.Option{
			generator.WithInputData(data),
			generator.WithOutputFile(outputFile),
			generator.WithTitle(title),
		}
	} else {
		opts = []generator.Option{
			generator.WithInputFile(inputFile),
			generator.WithOutputFile(outputFile),
			generator.WithTitle(title),
		}
	}

	reporter, err := generator.New(opts...)
	if err != nil {
		return err
	}
	return reporter.Generate(ctx)
}
