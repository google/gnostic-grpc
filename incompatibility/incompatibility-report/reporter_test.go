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
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	openapiv3 "github.com/googleapis/gnostic/openapiv3"
)

func makeIncompatibilityReport(incompatiblities ...*Incompatibility) *IncompatibilityReport {
	return &IncompatibilityReport{Incompatibilities: incompatiblities}
}

var ignoreUnexportedOption = cmpopts.IgnoreUnexported(IncompatibilityReport{}, Incompatibility{})
var incompatibilityOrderOption = cmpopts.SortSlices(func(l, r *Incompatibility) bool {
	if l.Classification != r.Classification {
		return l.Classification < r.Classification
	} else {
		minimumArrayLen := math.Min(float64(len(l.TokenPath)), float64(len(r.TokenPath)))
		for i := 0; i < int(minimumArrayLen); i++ {
			lstring := l.TokenPath[i]
			rstring := r.TokenPath[i]
			if lstring != rstring {
				return strings.Compare(lstring, rstring) < 0
			}
		}
	}
	return true
})

type InvalidOperationType int

const (
	OPTIONS = iota
	HEAD
	TRACE
)

func makePathsObject(pathsName string, operationType ...InvalidOperationType) *openapiv3.Paths {
	var pathItem *openapiv3.PathItem = &openapiv3.PathItem{}
	for _, opType := range operationType {
		switch opType {
		case OPTIONS:
			pathItem.Options = &openapiv3.Operation{}
		case HEAD:
			pathItem.Head = &openapiv3.Operation{}
		case TRACE:
			pathItem.Trace = &openapiv3.Operation{}
		}
	}
	return &openapiv3.Paths{Path: []*openapiv3.NamedPathItem{{Name: pathsName, Value: pathItem}}}
}

// Simple test for in-progress incompatibility chain coverage
func TestReporterCoverage(t *testing.T) {

	var reporterTest = []struct {
		givenDocument                 *openapiv3.Document
		expectedIncompatibilityReport *IncompatibilityReport
		incompatibilityReporters      IncompatibilityReporter
	}{
		{&openapiv3.Document{}, &IncompatibilityReport{}, aggregateIncompatibilityReporters()},
		{&openapiv3.Document{}, &IncompatibilityReport{}, aggregateIncompatibilityReporters(DocumentBaseSearch)},
		{&openapiv3.Document{}, &IncompatibilityReport{}, aggregateIncompatibilityReporters(PathsSearch, DocumentBaseSearch)},
		{
			&openapiv3.Document{
				Security: []*openapiv3.SecurityRequirement{{
					AdditionalProperties: []*openapiv3.NamedStringArray{},
				}}},
			makeIncompatibilityReport(newIncompatibility(Severity_FAIL, IncompatibiltiyClassification_Security, "security")),
			aggregateIncompatibilityReporters(DocumentBaseSearch),
		},
		{
			&openapiv3.Document{Security: []*openapiv3.SecurityRequirement{{
				AdditionalProperties: []*openapiv3.NamedStringArray{},
			}}},
			makeIncompatibilityReport(), // only includes path
			aggregateIncompatibilityReporters(PathsSearch),
		},
		{
			&openapiv3.Document{
				Security: []*openapiv3.SecurityRequirement{{
					AdditionalProperties: []*openapiv3.NamedStringArray{},
				}},
				Paths: makePathsObject("pathName", OPTIONS, HEAD, TRACE)},
			makeIncompatibilityReport(
				newIncompatibility(Severity_FAIL, IncompatibiltiyClassification_Security, "security"),
				newIncompatibility(Severity_FAIL, IncompatibiltiyClassification_InvalidOperation, "paths", "pathName", "options"),
				newIncompatibility(Severity_FAIL, IncompatibiltiyClassification_InvalidOperation, "paths", "pathName", "head"),
				newIncompatibility(Severity_FAIL, IncompatibiltiyClassification_InvalidOperation, "paths", "pathName", "trace"),
			),
			aggregateIncompatibilityReporters(DocumentBaseSearch, PathsSearch),
		},
	}
	for ind, tt := range reporterTest {
		testname := fmt.Sprintf("CoverageTest%d", ind)
		t.Run(testname, func(t *testing.T) {
			got := ReportOnDoc(tt.givenDocument, tt.incompatibilityReporters)
			if diff := cmp.Diff(tt.expectedIncompatibilityReport, got, ignoreUnexportedOption, incompatibilityOrderOption); diff != "" {
				t.Errorf("SearchChains(%v, %v): diff (-want +got):\n%v", tt.givenDocument, tt.incompatibilityReporters, diff)
			}
		})
	}
}
