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

	openapiv3 "github.com/googleapis/gnostic/openapiv3"
)

// Collection of defined incompatibility reporters
var IncompatibilityReporters []IncompatibilityReporter = []IncompatibilityReporter{
	DocumentBaseSearch,
	PathsSearch,
	ComponentsSearch,
}

// A reporter takes in any openapiv3 document and returns incopatibilities
type IncompatibilityReporter func(*openapiv3.Document) []*Incompatibility

func aggregateIncompatibilityReporters(reporters ...IncompatibilityReporter) IncompatibilityReporter {
	return func(doc *openapiv3.Document) []*Incompatibility {
		var incompatibilities []*Incompatibility
		for _, reporter := range reporters {
			incompatibilities = append(incompatibilities, reporter(doc)...)
		}
		return incompatibilities
	}
}

//ReportOnDoc applies the given reporters on the given doc to produce an Incompatibilty Report
func ReportOnDoc(doc *openapiv3.Document, reporters ...IncompatibilityReporter) *IncompatibilityReport {
	return &IncompatibilityReport{Incompatibilities: aggregateIncompatibilityReporters(reporters...)(doc)}
}

// ======================== Defined Reporters  ====================== //

// DocumentBaseSearch is a reporter that scans for incompatibilities at the base of an OpenAPI doc
func DocumentBaseSearch(doc *openapiv3.Document) []*Incompatibility {
	var incompatibilities []*Incompatibility
	if doc.Security == nil {
		return incompatibilities
	}
	incompatibilities = append(incompatibilities,
		newIncompatibility(Severity_FAIL, IncompatibiltiyClassification_Security, "security"))
	return incompatibilities
}

// PathsSearch is a reporter that scans for incompatibilities in the paths component of an OpenAPI doc
func PathsSearch(doc *openapiv3.Document) []*Incompatibility {
	var incompatibilities []*Incompatibility
	pathsKey := []string{"paths"}
	if doc.Paths == nil {
		return incompatibilities
	}
	for _, pathItem := range doc.Paths.Path {
		pathKey := addKeyPath(pathsKey, pathItem.Name)
		pathValue := pathItem.Value
		if pathValue.Head != nil {
			incompatibilities = append(incompatibilities,
				newIncompatibility(Severity_FAIL, IncompatibiltiyClassification_InvalidOperation, addKeyPath(pathKey, "head")...))
		}
		if pathValue.Options != nil {
			incompatibilities = append(incompatibilities,
				newIncompatibility(Severity_FAIL, IncompatibiltiyClassification_InvalidOperation, addKeyPath(pathKey, "options")...))
		}
		if pathValue.Trace != nil {
			incompatibilities = append(incompatibilities,
				newIncompatibility(Severity_FAIL, IncompatibiltiyClassification_InvalidOperation, addKeyPath(pathKey, "trace")...))
		}
		incompatibilities = append(incompatibilities,
			validOperationSearch(pathValue.Get, addKeyPath(pathKey, "get"))...)
		incompatibilities = append(incompatibilities,
			validOperationSearch(pathValue.Put, addKeyPath(pathKey, "put"))...)
		incompatibilities = append(incompatibilities,
			validOperationSearch(pathValue.Post, addKeyPath(pathKey, "post"))...)
		incompatibilities = append(incompatibilities,
			validOperationSearch(pathValue.Delete, addKeyPath(pathKey, "delete"))...)
		incompatibilities = append(incompatibilities,
			validOperationSearch(pathValue.Patch, addKeyPath(pathKey, "patch"))...)

		for ind, paramOrRef := range pathValue.Parameters {
			incompatibilities = append(incompatibilities,
				parametersSearch(paramOrRef.GetParameter(), addKeyPath(pathKey, "parameters", strconv.Itoa(ind)))...)
		}
	}
	return incompatibilities
}

func ComponentsSearch(doc *openapiv3.Document) []*Incompatibility {
	var incompatibilities []*Incompatibility
	if doc.Components == nil {
		return incompatibilities
	}
	pathsKey := []string{"components"}

	if doc.Components.Callbacks != nil {
		incompatibilities = append(incompatibilities,
			newIncompatibility(Severity_WARNING, IncompatibiltiyClassification_ExternalTranscodingSupport, addKeyPath(pathsKey, "callbacks")...))
	}
	if doc.Components.Schemas != nil {
		for _, schemaRef := range doc.Components.Schemas.GetAdditionalProperties() {
			incompatibilities = append(incompatibilities, schemaSearch(schemaRef.Value.GetSchema(), addKeyPath(pathsKey, "schemas", schemaRef.Name))...)
		}
	}
	if doc.Components.Responses != nil {
		for _, resRef := range doc.Components.Responses.GetAdditionalProperties() {
			incompatibilities = append(incompatibilities,
				responseSearch(resRef.Value.GetResponse(), addKeyPath(pathsKey, "requestBodies", resRef.Name))...,
			)
		}
	}
	if doc.Components.Parameters != nil {
		for _, paramRef := range doc.Components.Parameters.GetAdditionalProperties() {
			incompatibilities = append(incompatibilities,
				parametersSearch(paramRef.GetValue().GetParameter(), addKeyPath(pathsKey, "parameters", paramRef.Name))...,
			)
		}
	}
	if doc.Components.RequestBodies != nil {
		for _, reqRef := range doc.Components.RequestBodies.GetAdditionalProperties() {
			incompatibilities = append(incompatibilities,
				requestBodySearch(reqRef.Value.GetRequestBody(), addKeyPath(pathsKey, "requestBodies", reqRef.Name))...,
			)
		}
	}
	if doc.Components.Headers != nil {
		for _, comRef := range doc.Components.Headers.GetAdditionalProperties() {
			incompatibilities = append(incompatibilities,
				headerSearch(comRef.Name, comRef.Value.GetHeader(), addKeyPath(pathsKey, "headers", comRef.Name))...,
			)
		}
	}

	return incompatibilities
}

// ========================= Helper Functions ======================== //

// validOperationSearch scans for incompatibilities within valid operations
func validOperationSearch(operation *openapiv3.Operation, keys []string) []*Incompatibility {
	var incompatibilities []*Incompatibility
	if operation == nil {
		return incompatibilities
	}
	if operation.Callbacks != nil {
		incompatibilities = append(incompatibilities,
			newIncompatibility(Severity_WARNING, IncompatibiltiyClassification_ExternalTranscodingSupport, addKeyPath(keys, "callbacks")...))
	}
	if operation.Security != nil {
		incompatibilities = append(incompatibilities,
			newIncompatibility(Severity_FAIL, IncompatibiltiyClassification_Security, addKeyPath(keys, "security")...))
	}
	for ind, paramOrRef := range operation.Parameters {
		incompatibilities = append(incompatibilities, parametersSearch(paramOrRef.GetParameter(), addKeyPath(keys, "parameters", strconv.Itoa(ind)))...)
	}
	return incompatibilities

}

// pathsSearch scans for incompatibilities within a parameters object
func parametersSearch(param *openapiv3.Parameter, keys []string) []*Incompatibility {
	var incompatibilities []*Incompatibility
	if param == nil {
		return incompatibilities
	}
	if param.Style != "" {
		incompatibilities = append(incompatibilities,
			newIncompatibility(Severity_WARNING, IncompatibiltiyClassification_ParameterStyling, addKeyPath(keys, "style")...))
	}
	if param.Explode {
		incompatibilities = append(incompatibilities,
			newIncompatibility(Severity_WARNING, IncompatibiltiyClassification_ParameterStyling, addKeyPath(keys, "explode")...))
	}
	if param.AllowReserved {
		incompatibilities = append(incompatibilities,
			newIncompatibility(Severity_WARNING, IncompatibiltiyClassification_ParameterStyling, addKeyPath(keys, "allowReserved")...))
	}
	if param.AllowEmptyValue {
		incompatibilities = append(incompatibilities,
			newIncompatibility(Severity_WARNING, IncompatibiltiyClassification_DataValidation, addKeyPath(keys, "allowEmptyValue")...))
	}
	if param.Schema != nil {
		incompatibilities = append(incompatibilities,
			schemaSearch(param.Schema.GetSchema(), addKeyPath(keys, "schema"))...)
	}
	return incompatibilities
}

// schemaSearch scans for incompatibilities within a schema object
func schemaSearch(schema *openapiv3.Schema, keys []string) []*Incompatibility {
	var incompatibilities []*Incompatibility
	if schema == nil {
		return incompatibilities
	}
	if schema.Nullable {
		incompatibilities = append(incompatibilities,
			newIncompatibility(Severity_FAIL, IncompatibiltiyClassification_InvalidDataState, addKeyPath(keys, "nullable")...))
	}
	if schema.Discriminator != nil {
		incompatibilities = append(incompatibilities,
			newIncompatibility(Severity_FAIL, IncompatibiltiyClassification_Inheritance, addKeyPath(keys, "discriminator")...))
	}
	if schema.ReadOnly {
		incompatibilities = append(incompatibilities,
			newIncompatibility(Severity_FAIL, IncompatibiltiyClassification_ParameterStyling, addKeyPath(keys, "readOnly")...))
	}
	if schema.WriteOnly {
		incompatibilities = append(incompatibilities,
			newIncompatibility(Severity_FAIL, IncompatibiltiyClassification_ParameterStyling, addKeyPath(keys, "writeOnly")...))
	}
	if schema.MultipleOf != 0 {
		incompatibilities = append(incompatibilities,
			newIncompatibility(Severity_WARNING, IncompatibiltiyClassification_DataValidation, addKeyPath(keys, "multipleOf")...))
	}
	if schema.Maximum != 0 {
		incompatibilities = append(incompatibilities,
			newIncompatibility(Severity_WARNING, IncompatibiltiyClassification_DataValidation, addKeyPath(keys, "maximum")...))
	}
	if schema.ExclusiveMaximum {
		incompatibilities = append(incompatibilities,
			newIncompatibility(Severity_WARNING, IncompatibiltiyClassification_DataValidation, addKeyPath(keys, "exclusiveMaximum")...))
	}
	if schema.Minimum != 0 {
		incompatibilities = append(incompatibilities,
			newIncompatibility(Severity_WARNING, IncompatibiltiyClassification_DataValidation, addKeyPath(keys, "minimum")...))
	}
	if schema.ExclusiveMinimum {
		incompatibilities = append(incompatibilities,
			newIncompatibility(Severity_WARNING, IncompatibiltiyClassification_DataValidation, addKeyPath(keys, "exclusiveMinimum")...))
	}
	if schema.MaxLength != 0 {
		incompatibilities = append(incompatibilities,
			newIncompatibility(Severity_WARNING, IncompatibiltiyClassification_DataValidation, addKeyPath(keys, "maxLength")...))
	}
	if schema.MinLength != 0 {
		incompatibilities = append(incompatibilities,
			newIncompatibility(Severity_WARNING, IncompatibiltiyClassification_DataValidation, addKeyPath(keys, "minimum")...))
	}
	if schema.Pattern != "" {
		incompatibilities = append(incompatibilities,
			newIncompatibility(Severity_WARNING, IncompatibiltiyClassification_DataValidation, addKeyPath(keys, "pattern")...))
	}
	if schema.MaxItems != 0 {
		incompatibilities = append(incompatibilities,
			newIncompatibility(Severity_WARNING, IncompatibiltiyClassification_DataValidation, addKeyPath(keys, "maxItems")...))
	}
	if schema.MinItems != 0 {
		incompatibilities = append(incompatibilities,
			newIncompatibility(Severity_WARNING, IncompatibiltiyClassification_DataValidation, addKeyPath(keys, "minItems")...))
	}
	if schema.UniqueItems {
		incompatibilities = append(incompatibilities,
			newIncompatibility(Severity_WARNING, IncompatibiltiyClassification_DataValidation, addKeyPath(keys, "uniqueItems")...))
	}
	if len(schema.AllOf) != 0 {
		incompatibilities = append(incompatibilities,
			newIncompatibility(Severity_FAIL, IncompatibiltiyClassification_Inheritance, addKeyPath(keys, "allOf")...))
	}
	if len(schema.OneOf) != 0 {
		incompatibilities = append(incompatibilities,
			newIncompatibility(Severity_FAIL, IncompatibiltiyClassification_Inheritance, addKeyPath(keys, "oneOf")...))
	}
	if len(schema.AnyOf) != 0 {
		incompatibilities = append(incompatibilities,
			newIncompatibility(Severity_FAIL, IncompatibiltiyClassification_Inheritance, addKeyPath(keys, "anyOf")...))
	}
	if schema.Items != nil {
		for ind, item := range schema.Items.SchemaOrReference {
			incompatibilities = append(incompatibilities, schemaSearch(item.GetSchema(), addKeyPath(keys, "items", strconv.Itoa(ind)))...)
		}
	}
	if schema.Properties != nil {
		for ind, prop := range schema.Properties.AdditionalProperties {
			incompatibilities = append(incompatibilities, schemaSearch(prop.Value.GetSchema(), addKeyPath(keys, "properties", strconv.Itoa(ind), prop.Name))...)
		}
	}
	if schema.AdditionalProperties != nil && schema.AdditionalProperties.GetSchemaOrReference() != nil {
		incompatibilities = append(incompatibilities,
			schemaSearch(schema.AdditionalProperties.GetSchemaOrReference().GetSchema(), addKeyPath(keys, "additionalProperties"))...)
	}

	return incompatibilities
}

// responseSearch scans for incompatibilities in a response object
func responseSearch(resp *openapiv3.Response, keys []string) []*Incompatibility {
	var incompatibilities []*Incompatibility
	if resp == nil {
		return incompatibilities
	}
	if resp.Headers != nil {
		for _, prop := range resp.Headers.AdditionalProperties {
			incompatibilities = append(incompatibilities,
				headerSearch(prop.Name, prop.GetValue().GetHeader(), addKeyPath(keys, prop.Name))...,
			)
		}
	}
	if resp.Content != nil {
		for _, prop := range resp.Content.AdditionalProperties {
			incompatibilities = append(incompatibilities,
				contentSearch(prop.Value, addKeyPath(keys, prop.Name))...,
			)
		}
	}
	return incompatibilities
}

//  headerSearch scans for incompatibilities in a header object
func headerSearch(headerName string, header *openapiv3.Header, keys []string) []*Incompatibility {
	var incompatibilities []*Incompatibility
	if header == nil {
		return incompatibilities
	}
	paramEquiv := header2Paramter(headerName, header)
	incompatibilities = append(incompatibilities,
		parametersSearch(paramEquiv, keys)...)
	return incompatibilities
}

// contentSearch scans for incompatibilities in a media object
func contentSearch(media *openapiv3.MediaType, keys []string) []*Incompatibility {
	var incompatibilities []*Incompatibility
	if media == nil {
		return incompatibilities
	}
	if media.Encoding != nil {
		incompatibilities = append(incompatibilities,
			newIncompatibility(Severity_WARNING, IncompatibiltiyClassification_ParameterStyling, addKeyPath(keys, "encoding")...))
	}
	if media.Schema != nil {
		incompatibilities = append(incompatibilities,
			schemaSearch(media.Schema.GetSchema(), addKeyPath(keys, "schema"))...,
		)
	}
	return incompatibilities
}

// requestBodySearch scans for incompatibilities in a restBody object
func requestBodySearch(req *openapiv3.RequestBody, keys []string) []*Incompatibility {
	var incompatibilities []*Incompatibility
	if req == nil {
		return incompatibilities
	}
	if req.Content != nil {
		for _, namedContent := range req.Content.GetAdditionalProperties() {
			incompatibilities = append(incompatibilities,
				contentSearch(namedContent.Value, addKeyPath(keys, namedContent.Name))...,
			)
		}
	}
	return incompatibilities
}

func newIncompatibility(severity Severity, classification IncompatibiltiyClassification, path ...string) *Incompatibility {
	return &Incompatibility{TokenPath: path, Classification: classification, Severity: severity}
}

// addKeyPath adds string to end of a copy of path
func addKeyPath(path []string, items ...string) (newPath []string) {
	newPath = make([]string, len(path))
	copy(newPath, path)
	newPath = append(newPath, items...)
	return
}

// header2Parameter creates an equivalent parameter object representation from a header
func header2Paramter(name string, header *openapiv3.Header) *openapiv3.Parameter {
	if header == nil {
		return &openapiv3.Parameter{}
	}
	return &openapiv3.Parameter{
		Name:                   name,
		In:                     "header",
		Description:            header.Description,
		Required:               header.Required,
		Deprecated:             header.Deprecated,
		AllowEmptyValue:        header.AllowEmptyValue,
		Style:                  header.Style,
		Explode:                header.Explode,
		AllowReserved:          header.AllowReserved,
		Schema:                 header.Schema,
		Example:                header.Example,
		Examples:               header.Examples,
		Content:                header.Content,
		SpecificationExtension: header.SpecificationExtension,
	}
}
