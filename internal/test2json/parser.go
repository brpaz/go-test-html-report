package test2json

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"

	"github.com/brpaz/go-test-html-report/internal/domain"
)

type TestEvent struct {
	Time    time.Time // encodes as an RFC3339-format string
	Action  string
	Package string
	Test    string
	Elapsed float64 // seconds
	Output  string
}

// Parse parses the given JSON data in test2json format and returns a struct representing the test results.
func Parse(data []byte) ([]domain.TestSuite, error) {
	// Map to store test suites by package name
	suiteMap := make(map[string]*domain.TestSuite)

	// Map to store test cases by package and test name for accumulating output
	testMap := make(map[string]map[string]*domain.TestCase)

	// Parse multiple JSON objects from the data
	decoder := json.NewDecoder(bytes.NewReader(data))

	for decoder.More() {
		var event TestEvent
		if err := decoder.Decode(&event); err != nil {
			return nil, err
		}		// Skip events without a package or test
		if event.Package == "" {
			continue
		}

		// Initialize suite if not exists
		if suiteMap[event.Package] == nil {
			suiteMap[event.Package] = &domain.TestSuite{
				Name:      event.Package,
				TestCases: []domain.TestCase{},
			}
			testMap[event.Package] = make(map[string]*domain.TestCase)
		}

		// Skip if no test name (package-level events)
		if event.Test == "" {
			continue
		}

		// Initialize test case if not exists
		if testMap[event.Package][event.Test] == nil {
			testCase := &domain.TestCase{
				Name:     event.Test,
				Status:   "running",
				Duration: 0,
				Output:   []string{},
			}
			testMap[event.Package][event.Test] = testCase
			suiteMap[event.Package].TestCases = append(suiteMap[event.Package].TestCases, *testCase)
		}

		testCase := testMap[event.Package][event.Test]

		// Handle different actions
		switch event.Action {
		case "output":
			if event.Output != "" {
				testCase.Output = append(testCase.Output, strings.TrimSuffix(event.Output, "\n"))
			}
		case "pass":
			testCase.Status = "pass"
			testCase.Duration = event.Elapsed
		case "fail":
			testCase.Status = "fail"
			testCase.Duration = event.Elapsed
		case "skip":
			testCase.Status = "skip"
			testCase.Duration = event.Elapsed
		}

		// Update the test case in the suite
		for i := range suiteMap[event.Package].TestCases {
			if suiteMap[event.Package].TestCases[i].Name == event.Test {
				suiteMap[event.Package].TestCases[i] = *testCase
				break
			}
		}
	}

	// Convert map to slice
	var suites []domain.TestSuite
	for _, suite := range suiteMap {
		suites = append(suites, *suite)
	}

	return suites, nil
}
