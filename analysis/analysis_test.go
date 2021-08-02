package main

import (
	"testing"

	"github.com/googleapis/gnostic-grpc/incompatibility"
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
		{"3Docs", "../incompatibility/oas-examples", 3},
		{"NoOpenAPIDocs", "../utils", 0},
	}
	for _, trial := range processOASTest {
		t.Run(trial.testName, func(tt *testing.T) {
			analysis := generateAnalysis(trial.dirPath)
			filesAnalyzed := analysis.OpenApiFiles
			if filesAnalyzed != int32(trial.expectedNumOASProcessed) {
				tt.Errorf("Incorrect number of openapi files analyzed: got %d, wanted %d\n",
					filesAnalyzed, trial.expectedNumOASProcessed)
			}
		})
	}
}

// TestFileInformationIncluded tests for the availability of
// file specific incompatibility information within a larger
// set analysis object
func TestFileInformationIncluded(t *testing.T) {
	var fileTest = []struct {
		testName string
		dirPath  string
		openAPI  []string
	}{
		{"ImmediateFile", "../incompatibility/oas-examples",
			[]string{
				"../incompatibility/oas-examples/petstore.yaml",
				"../incompatibility/oas-examples/openapi.yaml",
			}},
		{"deepFiles", "../examples/",
			[]string{
				"../examples/petstore/petstore.yaml",
				"../examples/bookstore/bookstore.yaml",
			}},
	}
	for _, trial := range fileTest {
		t.Run(trial.testName, func(tt *testing.T) {
			setAnalysis := generateAnalysis(trial.dirPath)
			for _, oasFilePath := range trial.openAPI {
				report, err := fileHandler(oasFilePath)
				if err != nil {
					tt.Fatalf(err.Error())
				}
				countFilePerClass :=
					incompatibility.CountIncompatibilities(report.Incompatibilities...).CountByClassification
				for class, count := range countFilePerClass {
					if count == 0 {
						continue
					}
					countFromAnalysis := getAnalysisIncompCount(setAnalysis, incompatibility.IncompatibiltiyClassification(class), oasFilePath)
					if countFromAnalysis != 1 {
						tt.Errorf("getAnalysisIncompCount(..., %v, %s), got %d, wanted %d",
							incompatibility.IncompatibiltiyClassification(class), oasFilePath, countFromAnalysis, count)
					}
				}
			}
		})
	}

}

func getAnalysisIncompCount(setAnalysis *incompatibility.ApiSetIncompatibility, class incompatibility.IncompatibiltiyClassification, oasFilePath string) int32 {
	classMapAnalysis := setAnalysis.AnalysisPerIncompatibility[class].CountPerFile
	countFromAnalysis := classMapAnalysis[oasFilePath].NumOccurrences
	return countFromAnalysis
}
