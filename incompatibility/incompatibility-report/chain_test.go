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
	"reflect"
	"testing"

	openapiv3 "github.com/googleapis/gnostic/openapiv3"
)

// searchForIncompatibility looks for the i1 incompatibilty in the rp2 incompatibility report
func searchForIncompatibility(i1 *Incompatibility, rp2 *IncompatibilityReport) (found bool) {
	for _, rp2Item := range rp2.GetIncompatibilities() {
		if reflect.DeepEqual(i1, rp2Item) {
			found = true
			return
		}
	}
	return
}

// incompatibilityReportEquality checks equality for two incompatibility reports
func incompatibilityReportEquality(rp1 *IncompatibilityReport, rp2 *IncompatibilityReport) (equality bool) {
	equality = true
	if len(rp1.GetIncompatibilities()) != len(rp2.GetIncompatibilities()) {
		equality = false
		return
	}
	for _, incompatibility := range rp1.GetIncompatibilities() {
		if !searchForIncompatibility(incompatibility, rp2) {
			equality = false
			return
		}
	}
	return
}

func makeIncompatibilityReport(incompatiblities ...*Incompatibility) *IncompatibilityReport {
	return &IncompatibilityReport{Incompatibilities: incompatiblities}
}

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
			makeIncompatibilityReport(NewIncompatibility("SECURITY", "security")),
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
				NewIncompatibility("SECURITY", "security"),
				NewIncompatibility("OPTIONS", "paths", "pathName", "options"),
				NewIncompatibility("HEAD", "paths", "pathName", "head"),
				NewIncompatibility("TRACE", "paths", "pathName", "trace"),
			),
			aggregateIncompatibilityReporters(PathsSearch, DocumentBaseSearch),
		},
	}
	for ind, tt := range reporterTest {
		testname := fmt.Sprintf("Reporter Coverage Test at index %d", ind)
		t.Run(testname, func(t *testing.T) {
			if !incompatibilityReportEquality(SearchChains(tt.givenDocument, tt.incompatibilityReporters), tt.expectedIncompatibilityReport) {
				t.Errorf("Unexpected incompatibilty report at index %d", ind)
			}
		})
	}
}
