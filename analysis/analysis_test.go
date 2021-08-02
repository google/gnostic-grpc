package main

import (
	"testing"
)

// TestGenerateAnalysis is a simple test that tests if
// generateAnlysis produces a correct analysis object
func TestGenerateAnalysis(t *testing.T) {
	var processOASTest = []struct {
		testName                string
		dirPath                 string
		expectedNumOASProcessed int
	}{
		{"falseDir", "fake", 0},
		{"dirWSubDir", "../generator/testfiles", 6},
		{"dirWMalFile", "../incompatibility/oas-examples/malformed", 0},
		{"3Docs", "../incompatibility/oas-examples", 4},
		{"NoOpenAPIDocs", "../utils", 0},
	}
	for _, trial := range processOASTest {
		t.Run(trial.testName, func(t *testing.T) {
			analysis := generateAnalysis(trial.dirPath)
			filesAnalyzed := analysis.OpenApiFiles
			if filesAnalyzed != int32(trial.expectedNumOASProcessed) {
				t.Errorf("Incorrect number of openapi files analyzed: got %d, wanted %d\n",
					filesAnalyzed, trial.expectedNumOASProcessed)
			}
		})
	}

}
