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

// NewAnalysis initalizes and returns an apiset analysis object
func NewAnalysis() *ApiSetIncompatibility {
	incompatibilityMap := make(map[string]*IncompatibilityAnalysis)
	for _, incompClassString := range IncompatibiltiyClassification_name {
		incompatibilityMap[incompClassString] =
			&IncompatibilityAnalysis{FilesWithIncompatibility: 0}
	}
	return &ApiSetIncompatibility{
		OpenApiFiles:               0,
		IncompatibleFiles:          0,
		AnalysisPerIncompatibility: incompatibilityMap,
	}
}

// TODO
// AggregateReports aggregates incompatibility information from IncomaptibiltyReports
// into a new ApiSetIncompatiblity object.
func AggregateReports(reports ...*IncompatibilityReport) *ApiSetIncompatibility {
	analysis := NewAnalysis()
	for _, report := range reports {
		analysis = AggregateAnalysis(analysis, FormAnalysis(report))
	}
	return analysis
}

// TODO
// AggregateAnalysis aggregates incompatibility information from multiple ApiSetIncompatibility
// objects into one comprehensive ApiSetIncompatibility
func AggregateAnalysis(analysis ...*ApiSetIncompatibility) *ApiSetIncompatibility {
	aggAnalysis := NewAnalysis()
	for _, analysisObj := range analysis {
		aggAnalysis.OpenApiFiles += analysisObj.OpenApiFiles
	}
	return aggAnalysis
}

// FormAnalysis creates an analysis object from a single IncompatibilityReport
func FormAnalysis(report *IncompatibilityReport) *ApiSetIncompatibility {
	analysis := NewAnalysis()
	analysis.OpenApiFiles++
	return analysis
}

// TODO
// AggregateIncompatiblities aggrates information from individual incompatibilities into an
// existing analysis object
func AggregateIncompatibilities(analysis *ApiSetIncompatibility, incompatibilities ...*Incompatibility) *ApiSetIncompatibility {
	analysis.OpenApiFiles++
	return analysis
}
