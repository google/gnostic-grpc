// Copyright 2021 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package incompatibility

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var ignoreUnexportedSetAnalysis = cmpopts.IgnoreUnexported(
	IncompatibilityAnalysis{},
	ApiSetIncompatibility{},
	FileIncompatibilityClassificationAnalysis{},
)

// Counts incompatibility from set analysis by classification
func countIncompSetAnalysis(setAnalysis *ApiSetIncompatibility) []int32 {
	var counts []int32 = make([]int32, len(IncompatibiltiyClassification_value))
	for class, incompatiblityGroup := range setAnalysis.AnalysisPerIncompatibility {
		for _, count := range incompatiblityGroup.CountPerFile {
			counts[class] += count.NumOccurrences
		}
	}
	return counts
}

// create a slice of incompatibilites populated from all given reports
func groupIncompatibilities(incompatibilityReports []*IncompatibilityReport) []*Incompatibility {
	var allIncompatibilities []*Incompatibility
	for _, report := range incompatibilityReports {
		allIncompatibilities = append(allIncompatibilities, report.GetIncompatibilities()...)
	}
	return allIncompatibilities
}

func createReport(t *testing.T, path string) *IncompatibilityReport {
	return ScanIncompatibilities(generateDoc(t, path), path)
}

// TestIncompatibilityCount tests for transferring incompatibily counts from reports to set
// analysis.
func TestIncompatibilityCount(t *testing.T) {
	var googleReport = createReport(t, "oas-examples/openapi.yaml")
	var petStoreReport = createReport(t, "oas-examples/petstore.yaml")
	var bookStoreReport = createReport(t, "../examples/bookstore/bookstore.yaml")

	var countTest = []struct {
		testName               string
		incompatibilityReports []*IncompatibilityReport
	}{
		{"noIncompatibilities", make([]*IncompatibilityReport, 10)},
		{"OneReport",
			[]*IncompatibilityReport{googleReport}},
		{"MultipleReports",
			[]*IncompatibilityReport{
				googleReport,
				petStoreReport,
				bookStoreReport,
			},
		},
	}
	for _, trial := range countTest {
		t.Run(trial.testName, func(tt *testing.T) {
			countFromReport := CountIncompatibilities(
				groupIncompatibilities(trial.incompatibilityReports)...).GetCountByClass()
			countFromAnalysis := countIncompSetAnalysis(
				AggregateReports(trial.incompatibilityReports...))
			diff := cmp.Diff(countFromReport, countFromAnalysis)
			if diff != "" {
				tt.Errorf("IncompatibilityCount : diff (-want +got):\n %v", diff)
			}
		})
	}
}

// TestAggregatingSameFileIncompatibility checks file associated information is not duplicated in set analysis
// object.
func TestAggregatingSameFileIncompatibility(t *testing.T) {
	var googleReport = createReport(t, "oas-examples/openapi.yaml")
	var petStoreReport = createReport(t, "oas-examples/petstore.yaml")
	var bookStoreReport = createReport(t, "../examples/bookstore/bookstore.yaml")

	var aggregationTest = []struct {
		testName   string
		reportset1 []*IncompatibilityReport // fileset v1 and v2 should have an intersection of core files
		reportset2 []*IncompatibilityReport // their content can have repeated items from this intersection
	}{
		{"Base1to1", []*IncompatibilityReport{googleReport}, []*IncompatibilityReport{googleReport}},
		{"SingleRepeated", []*IncompatibilityReport{googleReport, googleReport}, []*IncompatibilityReport{googleReport}},
		{"MultipleRepeated",
			[]*IncompatibilityReport{
				bookStoreReport,
				googleReport,
				petStoreReport,
				bookStoreReport,
				googleReport,
			},
			[]*IncompatibilityReport{
				bookStoreReport,
				googleReport,
				petStoreReport,
			},
		},
		{"RepeatedRearranged",
			[]*IncompatibilityReport{
				bookStoreReport,
				googleReport,
				petStoreReport,
				petStoreReport,
				googleReport,
				bookStoreReport,
			},
			[]*IncompatibilityReport{
				bookStoreReport,
				petStoreReport,
				googleReport,
				bookStoreReport,
				petStoreReport,
				googleReport,
			},
		},
	}
	for _, trial := range aggregationTest {
		t.Run(trial.testName, func(tt *testing.T) {
			diff := cmp.Diff(
				AggregateReports(trial.reportset1...).AnalysisPerIncompatibility,
				AggregateReports(trial.reportset2...).AnalysisPerIncompatibility,
				ignoreUnexportedSetAnalysis,
			)
			if diff != "" {
				tt.Errorf("IncompatibilityMapping: diff (-want +got):\n %v", diff)
			}
		})
	}
}
