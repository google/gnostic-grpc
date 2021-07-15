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
	"reflect"

	openapiv3 "github.com/googleapis/gnostic/openapiv3"
	plugins "github.com/googleapis/gnostic/plugins"
)

type Report int

const (
	BaseIncompatibility_Report = iota
	ID_Report
)

// Create a report of incompatibilities, write protobuf message to
// environment, bool indicates if detailed incompatibility report is
// desired.
func CreateIncompReport(env *plugins.Environment, reportType Report) {

	// Generate Base Incompatibility Report

	// If indicated by report type associate incompatibilities with line
	// references, etc. and generate an ID_REPORT

}

// Scan for incompatibilities in an OpenAPI document
func ScanIncompatibilities(document *openapiv3.Document) *IncompatibilityReport {

	paths, err := knownIncompatibilityPaths()
	if err != nil {
		return &IncompatibilityReport{}
	}
	IncompReport, err := paths.compile(document)
	if err != nil {
		return &IncompatibilityReport{}
	}
	return IncompReport
}

// Function to get path representations of
func knownIncompatibilityPaths() (PathOperation, error) {
	reportServersTyped := func(document *openapiv3.Document) *IncompatibilityReport {
		var incompatibilities []*Incompatibility
		if document.Servers != nil {
			incompatibilities = append(incompatibilities, &Incompatibility{Classification: "SERVERS"})
		}
		return &IncompatibilityReport{Incompatibilities: incompatibilities}
	}
	reportServersGR, err := makeGenericIncompatibilityReportFunc(reportServersTyped)
	parentP := PathOperation{
		ComponentType:        reflect.TypeOf(openapiv3.Document{}),
		OperationOnComponent: reportServersGR,
	}
	return parentP, err

}
