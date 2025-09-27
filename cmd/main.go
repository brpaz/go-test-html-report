package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/brpaz/go-test-html-report/internal/generator"
	"github.com/urfave/cli/v3"
)

var (
	version = "0.0.0-dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd := &cli.Command{
        Name:  "go-test-html-report",
        Usage: "generate HTML report from Go test results",
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
				Usage:   "Input file containing Go test results",
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
			It processes the output of 'go test' and creates a user-friendly HTML report.`,
        Action: func(ctx context.Context, cmd *cli.Command) error {
			inputFile := cmd.String("input")
			outputFile := cmd.String("output")
			title := cmd.String("title")

            reporter, err := generator.NewHTMLReportGenerator(
				generator.WithInputFile(inputFile),
				generator.WithOutputFile(outputFile),
				generator.WithTitle(title),
			)
            if err != nil {
                return err
            }
            return reporter.Generate(ctx)
        },
    }

    if err := cmd.Run(context.Background(), os.Args); err != nil {
        log.Fatal(err)
    }
}
