package report

type TestSuite struct {
	Name      string
	TestCases []TestCase
}

type TestCase struct {
	Name     string
	Status   string
	Duration float64
	Output   []string
}
