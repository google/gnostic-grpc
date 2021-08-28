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

// IntermediateReport counts the number of incompatibility occurrences
type IntermediateReport struct {
	countByClassification []int32
	countBySeverity       []int32
}

func (iReport IntermediateReport) GetCountByClass() []int32 {
	return iReport.countByClassification
}
func (iReport IntermediateReport) GetCountBySeverity() []int32 {
	return iReport.countBySeverity
}

// NewAnalysis initalizes and returns an apiset analysis object
func NewAnalysis() *ApiSetIncompatibility {
	var incompatibilityByClass []*IncompatibilityAnalysis = make([]*IncompatibilityAnalysis, len(IncompatibiltiyClassification_value))
	for class := range IncompatibiltiyClassification_name {
		incompatibilityByClass[class] = &IncompatibilityAnalysis{
			IncompatibilityClass: IncompatibiltiyClassification(class),
			CountPerFile:         make(map[string]*FileIncompatibilityClassificationAnalysis),
		}
	}
	return &ApiSetIncompatibility{
		OpenApiFiles:               0,
		IncompatibleFiles:          0,
		AnalysisPerIncompatibility: incompatibilityByClass,
	}
}

// AggregateAnalysis aggregates incompatibility information from multiple ApiSetIncompatibility
// objects into one comprehensive ApiSetIncompatibility
func aggregateAnalysis(analysis ...*ApiSetIncompatibility) *ApiSetIncompatibility {
	aggAnalysis := NewAnalysis()
	for _, analysisObj := range analysis {
		aggAnalysis.OpenApiFiles += analysisObj.OpenApiFiles
		aggAnalysis.IncompatibleFiles += analysisObj.IncompatibleFiles
		aggAnalysis.AnalysisPerIncompatibility =
			mergeIncompatibilityAnalysis(aggAnalysis.AnalysisPerIncompatibility, analysisObj.AnalysisPerIncompatibility)
	}
	return aggAnalysis
}

// FormAnalysis creates an analysis object from a single IncompatibilityReport
func formAnalysis(report *IncompatibilityReport) *ApiSetIncompatibility {
	analysis := NewAnalysis()
	analysis.OpenApiFiles++
	intermedReport := CountIncompatibilities(report.GetIncompatibilities()...)
	for class, count := range intermedReport.countByClassification {
		if count == 0 {
			continue
		}
		fileOccurMap := analysis.AnalysisPerIncompatibility[class].CountPerFile
		fileOccurMap[report.ReportIdentifier] =
			&FileIncompatibilityClassificationAnalysis{
				NumOccurrences: count}
	}
	failIncompatibilitiesCount := intermedReport.countBySeverity[Severity_FAIL]
	if failIncompatibilitiesCount > 0 {
		analysis.IncompatibleFiles++
	}
	return analysis
}

func AggregateReports(reports ...*IncompatibilityReport) *ApiSetIncompatibility {
	analysis := NewAnalysis()
	for _, report := range reports {
		analysis = aggregateAnalysis(analysis, formAnalysis(report))
	}
	return analysis
}

// merge incompatibily analysis by classification
func mergeIncompatibilityAnalysis(a1, a2 []*IncompatibilityAnalysis) []*IncompatibilityAnalysis {
	var a3 []*IncompatibilityAnalysis = make([]*IncompatibilityAnalysis, len(IncompatibiltiyClassification_name))
	for _, class := range IncompatibiltiyClassification_value {
		mapCount1 := a1[class].CountPerFile
		mapCount2 := a2[class].CountPerFile
		a3[class] = &IncompatibilityAnalysis{
			IncompatibilityClass: IncompatibiltiyClassification(class),
			CountPerFile:         appendFileInfomation(mapCount1, mapCount2),
		}
	}
	return a3
}

// adds incompatibility counts by file in m2 to m1 each map should correspond to the same
// incompatibility classification
func appendFileInfomation(m1, m2 map[string]*FileIncompatibilityClassificationAnalysis) map[string]*FileIncompatibilityClassificationAnalysis {
	for fileName, incompAnalysis := range m2 {
		m1[fileName] = incompAnalysis
	}
	return m1
}

func CountIncompatibilities(incompatibilities ...*Incompatibility) IntermediateReport {
	var countByClass []int32 = make([]int32, len(IncompatibiltiyClassification_value))
	var countBySev []int32 = make([]int32, len(Severity_value))
	for _, incomp := range incompatibilities {
		countByClass[incomp.Classification]++
		countBySev[incomp.Severity]++
	}
	return IntermediateReport{
		countByClassification: countByClass,
		countBySeverity:       countBySev,
	}
}
