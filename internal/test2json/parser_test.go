package test2json_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/brpaz/go-test-html-report/internal/report"
	"github.com/brpaz/go-test-html-report/internal/test2json"
)

func findTestCaseByName(t *testing.T, testCase []report.TestCase, s string) *report.TestCase {
	t.Helper()
	for i := range testCase {
		if testCase[i].Name == s {
			return &testCase[i]
		}
	}
	return nil
}

func TestParse(t *testing.T) {
	t.Parallel()
	t.Run("successfully parses valid Go test JSON output", func(t *testing.T) {
		t.Parallel()
		testData := `{"Time":"2025-09-27T14:30:00Z","Action":"start","Package":"github.com/example/pkg"}
{"Time":"2025-09-27T14:30:00Z","Action":"run","Package":"github.com/example/pkg","Test":"TestPass"}
{"Time":"2025-09-27T14:30:01Z","Action":"pass","Package":"github.com/example/pkg","Test":"TestPass","Elapsed":0.5}
{"Time":"2025-09-27T14:30:01Z","Action":"run","Package":"github.com/example/pkg","Test":"TestFail"}
{"Time":"2025-09-27T14:30:02Z","Action":"fail","Package":"github.com/example/pkg","Test":"TestFail","Elapsed":1.0}
`
		suites, err := test2json.Parse([]byte(testData))
		assert.NoError(t, err)
		require.Len(t, suites, 1)

		suite := suites[0]
		assert.Equal(t, "github.com/example/pkg", suite.Name)
		assert.Len(t, suite.TestCases, 2)

		// Verify both tests were parsed correctly
		passedTest := findTestCaseByName(t, suite.TestCases, "TestPass")
		failedTest := findTestCaseByName(t, suite.TestCases, "TestFail")

		require.NotNil(t, passedTest)
		assert.Equal(t, "pass", passedTest.Status)
		assert.Equal(t, 0.5, passedTest.Duration)

		require.NotNil(t, failedTest)
		assert.Equal(t, "fail", failedTest.Status)
		assert.Equal(t, 1.0, failedTest.Duration)
	})

	t.Run("handles multiple packages correctly", func(t *testing.T) {
		t.Parallel()
		testData := `{"Time":"2025-09-27T14:30:00Z","Action":"start","Package":"pkg1"}
{"Time":"2025-09-27T14:30:00Z","Action":"run","Package":"pkg1","Test":"Test1"}
{"Time":"2025-09-27T14:30:01Z","Action":"pass","Package":"pkg1","Test":"Test1","Elapsed":0.5}
{"Time":"2025-09-27T14:30:01Z","Action":"start","Package":"pkg2"}
{"Time":"2025-09-27T14:30:01Z","Action":"run","Package":"pkg2","Test":"Test2"}
{"Time":"2025-09-27T14:30:02Z","Action":"fail","Package":"pkg2","Test":"Test2","Elapsed":1.0}
`

		suites, err := test2json.Parse([]byte(testData))
		assert.NoError(t, err)
		assert.Len(t, suites, 2)

		// Verify each package has its tests
		packageNames := make(map[string]bool)
		for _, suite := range suites {
			packageNames[suite.Name] = true
			assert.Len(t, suite.TestCases, 1)
		}

		assert.True(t, packageNames["pkg1"])
		assert.True(t, packageNames["pkg2"])
	})

	t.Run("accumulates test output correctly", func(t *testing.T) {
		t.Parallel()
		testData := `{"Time":"2025-09-27T14:30:00Z","Action":"start","Package":"test"}
{"Time":"2025-09-27T14:30:00Z","Action":"run","Package":"test","Test":"TestOutput"}
{"Time":"2025-09-27T14:30:00Z","Action":"output","Package":"test","Test":"TestOutput","Output":"Line 1"}
{"Time":"2025-09-27T14:30:00Z","Action":"output","Package":"test","Test":"TestOutput","Output":"Line 2"}
{"Time":"2025-09-27T14:30:01Z","Action":"pass","Package":"test","Test":"TestOutput","Elapsed":1.0}
`

		suites, err := test2json.Parse([]byte(testData))
		assert.NoError(t, err)
		require.Len(t, suites, 1)
		require.Len(t, suites[0].TestCases, 1)

		testCase := suites[0].TestCases[0]
		assert.Equal(t, "TestOutput", testCase.Name)
		assert.Equal(t, "pass", testCase.Status)
		assert.Contains(t, testCase.Output, "Line 1")
		assert.Contains(t, testCase.Output, "Line 2")
	})

	t.Run("handles empty input gracefully", func(t *testing.T) {
		t.Parallel()
		suites, err := test2json.Parse([]byte(""))
		assert.NoError(t, err)
		assert.Empty(t, suites)
	})

	t.Run("rejects invalid JSON input", func(t *testing.T) {
		t.Parallel()
		testData := `invalid json content`
		suites, err := test2json.Parse([]byte(testData))
		assert.Error(t, err)
		assert.Nil(t, suites)
	})

	t.Run("rejects malformed JSON lines", func(t *testing.T) {
		t.Parallel()
		testData := `{"Time":"2025-09-27T14:30:00Z","Action":"start","Package":"test"}
{invalid json line}
`
		suites, err := test2json.Parse([]byte(testData))
		assert.Error(t, err)
		assert.Nil(t, suites)
	})
}
