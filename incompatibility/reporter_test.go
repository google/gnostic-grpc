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

	openapiv3 "github.com/google/gnostic/openapiv3"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func makeIncompatibilityReport(incompatiblities ...*Incompatibility) *IncompatibilityReport {
	return &IncompatibilityReport{Incompatibilities: incompatiblities}
}

var ignoreUnexportedOption = cmpopts.IgnoreUnexported(IncompatibilityReport{}, Incompatibility{}, IncompatibilityDescription{}, FileDescriptiveReport{})
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
	return false
})

type InvalidOperationType int

const (
	OPTIONS = iota
	HEAD
	TRACE
)

func makeShallowPathsObject(pathsName string, operationType ...InvalidOperationType) *openapiv3.Paths {
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
	pathItem.Summary = "Sudo Summary"
	pathItem.Description = "Sudo Description"
	return &openapiv3.Paths{Path: []*openapiv3.NamedPathItem{{Name: pathsName, Value: pathItem}}}
}

func testIncompatibilityReports(t *testing.T, formattedError string, want, got *IncompatibilityReport) {
	diff := cmp.Diff(want, got, ignoreUnexportedOption, incompatibilityOrderOption)
	if diff != "" {
		t.Errorf(formattedError+":\n+v", diff)
	}
}

func TestPathsSearch(t *testing.T) {
	var pathTest = []struct {
		testname                      string
		documentWithPaths             *openapiv3.Document
		expectedIncompatibilityReport *IncompatibilityReport
	}{
		{
			"EmptyPaths",
			&openapiv3.Document{
				Paths: &openapiv3.Paths{}},
			makeIncompatibilityReport(),
		},
		{
			"MetaDataInformation",
			&openapiv3.Document{
				Paths: makeShallowPathsObject("pathName")},
			makeIncompatibilityReport(),
		},
		{
			"AllInvalidOperationsInPaths",
			&openapiv3.Document{
				Paths: makeShallowPathsObject("pathName", OPTIONS, HEAD, TRACE)},
			makeIncompatibilityReport(
				newIncompatibility(IncompatibiltiyClassification_InvalidOperation, "paths", "pathName", "options"),
				newIncompatibility(IncompatibiltiyClassification_InvalidOperation, "paths", "pathName", "head"),
				newIncompatibility(IncompatibiltiyClassification_InvalidOperation, "paths", "pathName", "trace"),
			),
		},
	}
	for _, trial := range pathTest {
		got := ReportOnDoc(trial.documentWithPaths, "", PathsSearch)
		t.Run(trial.testname, func(tt *testing.T) {
			errorString := fmt.Sprintf("PathsSearch(%v): diff(-want +got):\n", trial.documentWithPaths)
			testIncompatibilityReports(tt, errorString, trial.expectedIncompatibilityReport, got)
		})
	}
}

func TestComponentSearch(t *testing.T) {
	var pathTest = []struct {
		testname                      string
		documentWithComponent         *openapiv3.Document
		expectedIncompatibilityReport *IncompatibilityReport
	}{
		{
			"emptycomponent",
			&openapiv3.Document{
				Components: &openapiv3.Components{}},
			makeIncompatibilityReport(),
		},
		{
			"MetaDataInformation",
			&openapiv3.Document{
				Components: &openapiv3.Components{
					Examples: &openapiv3.ExamplesOrReferences{},
					Links:    &openapiv3.LinksOrReferences{},
				}},
			makeIncompatibilityReport(),
		},
		{
			"AllInvalidOperationsInPaths",
			&openapiv3.Document{
				Components: &openapiv3.Components{
					Callbacks:       &openapiv3.CallbacksOrReferences{},
					SecuritySchemes: &openapiv3.SecuritySchemesOrReferences{},
				}},
			makeIncompatibilityReport(
				newIncompatibility(IncompatibiltiyClassification_ExternalTranscodingSupport, "components", "callbacks"),
				newIncompatibility(IncompatibiltiyClassification_Security, "components", "securitySchemes"),
			),
		},
	}
	for _, trial := range pathTest {
		got := ComponentsSearch(trial.documentWithComponent)
		t.Run(trial.testname, func(tt *testing.T) {
			errorString := fmt.Sprintf("componentsSearch(%v): diff(-want +got):\n", trial.documentWithComponent)
			testIncompatibilityReports(tt, errorString, trial.expectedIncompatibilityReport,
				&IncompatibilityReport{Incompatibilities: got})
		})
	}
}

func TestOperationSearch(t *testing.T) {
	var operationSearchTest = []struct {
		testname                      string
		operation                     *openapiv3.Operation
		expectedIncompatibilityReport *IncompatibilityReport
	}{
		{
			"emptyoperation",
			&openapiv3.Operation{},
			makeIncompatibilityReport(),
		},
		{
			"MetaDataFieldsandSupportedFields",
			&openapiv3.Operation{
				Deprecated:  true,
				OperationId: "id",
				Summary:     "sum",
				Description: "description",
			},
			makeIncompatibilityReport(),
		},
		{
			"InvalidFields",
			&openapiv3.Operation{
				Callbacks: &openapiv3.CallbacksOrReferences{},
				Security:  []*openapiv3.SecurityRequirement{},
			},
			makeIncompatibilityReport(
				newIncompatibility(IncompatibiltiyClassification_Security, "security"),
				newIncompatibility(IncompatibiltiyClassification_ExternalTranscodingSupport, "callbacks"),
			),
		},
	}
	for _, trial := range operationSearchTest {
		got := validOperationSearch(trial.operation, []string{})
		t.Run(trial.testname, func(tt *testing.T) {
			errorString := fmt.Sprintf("validOperationSearch(%v): diff(-want +got):\n", trial.operation)
			testIncompatibilityReports(tt, errorString, trial.expectedIncompatibilityReport,
				&IncompatibilityReport{Incompatibilities: got})
		})
	}
}

func TestParametersSearch(t *testing.T) {
	var parameterSearchTest = []struct {
		testname                      string
		parameter                     *openapiv3.Parameter
		expectedIncompatibilityReport *IncompatibilityReport
	}{
		{
			"emptyparameter",
			&openapiv3.Parameter{},
			makeIncompatibilityReport(),
		},
		{
			"MetaDataFieldsandSupportedFields",
			&openapiv3.Parameter{
				Name:     "name",
				Required: true,
			},
			makeIncompatibilityReport(),
		},
		{
			"InvalidFields",
			&openapiv3.Parameter{
				Style:           "sty",
				Explode:         true,
				AllowEmptyValue: true,
				AllowReserved:   true,
				Schema:          &openapiv3.SchemaOrReference{},
			},
			makeIncompatibilityReport(
				newIncompatibility(IncompatibiltiyClassification_ParameterStyling, "style"),
				newIncompatibility(IncompatibiltiyClassification_ParameterStyling, "explode"),
				newIncompatibility(IncompatibiltiyClassification_ParameterStyling, "allowReserved"),
				newIncompatibility(IncompatibiltiyClassification_DataValidation, "allowEmptyValue"),
			),
		},
	}
	for _, trial := range parameterSearchTest {
		got := parametersSearch(trial.parameter, []string{})
		t.Run(trial.testname, func(tt *testing.T) {
			errorString := fmt.Sprintf("parametersSearch(%v): diff(-want +got):\n", trial.parameter)
			testIncompatibilityReports(tt, errorString, trial.expectedIncompatibilityReport,
				&IncompatibilityReport{Incompatibilities: got})
		})
	}
}

func TestSchemaSearch(t *testing.T) {
	var schemaSearchTest = []struct {
		testname                      string
		schema                        *openapiv3.Schema
		expectedIncompatibilityReport *IncompatibilityReport
	}{
		{
			"emptyschema",
			&openapiv3.Schema{},
			makeIncompatibilityReport(),
		},
		{
			"MetaDataFieldsandSupportedFields",
			&openapiv3.Schema{
				Title:         "title",
				MaxProperties: 10,
				Not:           &openapiv3.Schema{},
				Type:          "type",
				Default:       &openapiv3.DefaultType{},
			},
			makeIncompatibilityReport(),
		},
		{
			"InvalidFields",
			&openapiv3.Schema{
				Nullable:         true,
				Discriminator:    &openapiv3.Discriminator{},
				ReadOnly:         true,
				WriteOnly:        true,
				MultipleOf:       11,
				Maximum:          11,
				ExclusiveMaximum: true,
				Minimum:          11,
				ExclusiveMinimum: true,
				MaxLength:        11,
				MinLength:        11,
				Pattern:          "pattern",
				MaxItems:         11,
				MinItems:         11,
				UniqueItems:      true,
				AllOf:            make([]*openapiv3.SchemaOrReference, 2),
				OneOf:            make([]*openapiv3.SchemaOrReference, 2),
				AnyOf:            make([]*openapiv3.SchemaOrReference, 2),
			},
			makeIncompatibilityReport(
				newIncompatibility(IncompatibiltiyClassification_InvalidDataState, "nullable"),
				newIncompatibility(IncompatibiltiyClassification_Inheritance, "discriminator"),
				newIncompatibility(IncompatibiltiyClassification_ParameterStyling, "readOnly"),
				newIncompatibility(IncompatibiltiyClassification_ParameterStyling, "writeOnly"),
				newIncompatibility(IncompatibiltiyClassification_DataValidation, "multipleOf"),
				newIncompatibility(IncompatibiltiyClassification_DataValidation, "maximum"),
				newIncompatibility(IncompatibiltiyClassification_DataValidation, "exclusiveMaximum"),
				newIncompatibility(IncompatibiltiyClassification_DataValidation, "minimum"),
				newIncompatibility(IncompatibiltiyClassification_DataValidation, "exclusiveMinimum"),
				newIncompatibility(IncompatibiltiyClassification_DataValidation, "maxLength"),
				newIncompatibility(IncompatibiltiyClassification_DataValidation, "minimum"),
				newIncompatibility(IncompatibiltiyClassification_DataValidation, "pattern"),
				newIncompatibility(IncompatibiltiyClassification_DataValidation, "maxItems"),
				newIncompatibility(IncompatibiltiyClassification_DataValidation, "minItems"),
				newIncompatibility(IncompatibiltiyClassification_DataValidation, "uniqueItems"),
				newIncompatibility(IncompatibiltiyClassification_Inheritance, "allOf"),
				newIncompatibility(IncompatibiltiyClassification_Inheritance, "oneOf"),
				newIncompatibility(IncompatibiltiyClassification_Inheritance, "anyOf"),
			),
		},
	}
	for _, trial := range schemaSearchTest {
		got := schemaSearch(trial.schema, []string{})
		t.Run(trial.testname, func(tt *testing.T) {
			errorString := fmt.Sprintf("schemaSearch(%v): diff(-want +got):\n", trial.schema)
			testIncompatibilityReports(tt, errorString, trial.expectedIncompatibilityReport,
				&IncompatibilityReport{Incompatibilities: got})
		})
	}
}

func TestResponseSearch(t *testing.T) {
	var responseSearchTest = []struct {
		testname                      string
		response                      *openapiv3.Response
		expectedIncompatibilityReport *IncompatibilityReport
	}{
		{
			"emptyschema",
			&openapiv3.Response{},
			makeIncompatibilityReport(),
		},
		{
			"MetaDataFieldsandSupportedFields",
			&openapiv3.Response{
				Description: "desc.",
				Links:       &openapiv3.LinksOrReferences{},
			},
			makeIncompatibilityReport(),
		},
		{
			"InvalidFields",
			&openapiv3.Response{
				Headers: &openapiv3.HeadersOrReferences{
					AdditionalProperties: []*openapiv3.NamedHeaderOrReference{
						{
							Name: "header",
							Value: &openapiv3.HeaderOrReference{
								Oneof: &openapiv3.HeaderOrReference_Header{
									Header: &openapiv3.Header{
										Style: "style",
									},
								},
							},
						},
					},
				},
			},
			makeIncompatibilityReport(
				newIncompatibility(IncompatibiltiyClassification_ParameterStyling, "header", "style"),
			),
		},
	}
	for _, trial := range responseSearchTest {
		got := responseSearch(trial.response, []string{})
		t.Run(trial.testname, func(tt *testing.T) {
			errorString := fmt.Sprintf("responseSearch(%v): diff(-want +got):\n", trial.response)
			testIncompatibilityReports(tt, errorString, trial.expectedIncompatibilityReport,
				&IncompatibilityReport{Incompatibilities: got})
		})
	}
}
