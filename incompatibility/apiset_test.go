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
	"strconv"
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

// Treats each report as if they were created from a unique file
// and creates an set analysis object
func uniqueReportAnalysis(reports []*IncompatibilityReport) *ApiSetIncompatibility {
	setAnalysis := NewAnalysis()
	for ind, report := range reports {
		setAnalysis = AggregateAnalysis(setAnalysis, FormAnalysis(report, strconv.Itoa(ind)))
	}
	return setAnalysis
}

// create a slice of incompatibilites populated from all given reports
func groupIncompatibilities(incompatibilityReports []*IncompatibilityReport) []*Incompatibility {
	var allIncompatibilities []*Incompatibility
	for _, report := range incompatibilityReports {
		allIncompatibilities = append(allIncompatibilities, report.GetIncompatibilities()...)
	}
	return allIncompatibilities
}

// generate an set analysis from the given paths
func genAnalysisFromFiles(t *testing.T, filepaths []string) *ApiSetIncompatibility {
	setAnalysis := NewAnalysis()
	for _, file := range filepaths {
		setAnalysis = AggregateAnalysis(setAnalysis,
			FormAnalysis(ScanIncompatibilities(generateDoc(t, file)), file))
	}
	return setAnalysis
}

// TestIncompatibilityCount tests for transferring incompatibily counts from reports to set
// analysis.
func TestIncompatibilityCount(t *testing.T) {
	var countTest = []struct {
		testName               string
		incompatibilityReports []*IncompatibilityReport
	}{
		{"noIncompatibilities", make([]*IncompatibilityReport, 10)},
		{"OneReport",
			[]*IncompatibilityReport{ScanIncompatibilities(generateDoc(t, "oas-examples/openapi.yaml"))}},
		{"MultipleReports",
			[]*IncompatibilityReport{
				ScanIncompatibilities(generateDoc(t, "oas-examples/openapi.yaml")),
				ScanIncompatibilities(generateDoc(t, "oas-examples/petstore.yaml")),
				ScanIncompatibilities(generateDoc(t, "../examples/bookstore/bookstore.yaml")),
			},
		},
	}
	for _, trial := range countTest {
		t.Run(trial.testName, func(tt *testing.T) {
			countFromReport := CountIncompatibilities(
				groupIncompatibilities(trial.incompatibilityReports)...).CountByClassification
			countFromAnalysis := countIncompSetAnalysis(
				uniqueReportAnalysis(trial.incompatibilityReports))
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
	var aggregationTest = []struct {
		testName  string
		filesetv1 []string // fileset v1 and v2 should have an intersection of core files
		filesetv2 []string // their content can have repeated items from this intersection
	}{
		{"Base1to1", []string{"oas-examples/openapi.yaml"}, []string{"oas-examples/openapi.yaml"}},
		{"SingleRepeated", []string{"oas-examples/openapi.yaml", "oas-examples/openapi.yaml"}, []string{"oas-examples/openapi.yaml"}},
		{"MultipleRepeated",
			[]string{
				"../examples/bookstore/bookstore.yaml",
				"oas-examples/openapi.yaml",
				"../examples/bookstore/bookstore.yaml",
				"../examples/petstore/petstore.yaml",
				"oas-examples/openapi.yaml",
			},
			[]string{
				"../examples/bookstore/bookstore.yaml",
				"../examples/petstore/petstore.yaml",
				"oas-examples/openapi.yaml",
			},
		},
		{"RepeatedRearranged",
			[]string{
				"../examples/bookstore/bookstore.yaml",
				"../examples/petstore/petstore.yaml",
				"oas-examples/openapi.yaml",
				"oas-examples/openapi.yaml",
				"../examples/petstore/petstore.yaml",
				"../examples/bookstore/bookstore.yaml",
			},
			[]string{
				"../examples/bookstore/bookstore.yaml",
				"../examples/petstore/petstore.yaml",
				"oas-examples/openapi.yaml",
				"../examples/bookstore/bookstore.yaml",
				"../examples/petstore/petstore.yaml",
				"oas-examples/openapi.yaml",
			},
		},
	}
	for _, trial := range aggregationTest {
		t.Run(trial.testName, func(tt *testing.T) {
			diff := cmp.Diff(
				genAnalysisFromFiles(tt, trial.filesetv1).AnalysisPerIncompatibility,
				genAnalysisFromFiles(tt, trial.filesetv2).AnalysisPerIncompatibility,
				ignoreUnexportedSetAnalysis,
			)
			if diff != "" {
				tt.Errorf("IncompatibilityMapping: diff (-want +got):\n %v", diff)
			}
		})
	}
}
