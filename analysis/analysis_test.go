package main

import (
	"testing"
)

// TestProcessOpenAPIDocs is a simple test that tests that filehander
// is able to correctly process the OpenAPIdocuments in a given
// directory
func TestProcessOpenAPIDocs(t *testing.T) {
	var processOASTest = []struct {
		testName    string
		dirPath     string
		openAPIDocs int
	}{
		{"3Docs", "../incompatibility/oas-examples", 3},
		{"0Docs", "../utils", 0},
	}

	for _, trial := range processOASTest {
		t.Run(trial.testName, func(t *testing.T) {
			analysis := generateAnalysis(trial.dirPath)
			filesAnalyzed := analysis.OpenApiFiles
			if filesAnalyzed != int32(trial.openAPIDocs) {
				t.Errorf("Incorrect number of openapi files analyzed: got %d, wanted %d\n",
					filesAnalyzed, trial.openAPIDocs)
			}
		})
	}

}
