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

	openapiv3 "github.com/google/gnostic/openapiv3"
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
func ReportOnDoc(doc *openapiv3.Document, reportIdentifier string, reporters ...IncompatibilityReporter) *IncompatibilityReport {
	return &IncompatibilityReport{
		ReportIdentifier:  reportIdentifier,
		Incompatibilities: aggregateIncompatibilityReporters(reporters...)(doc)}
}

// ======================== Defined Reporters  ====================== //

// DocumentBaseSearch is a reporter that scans for incompatibilities at the base of an OpenAPI doc
func DocumentBaseSearch(doc *openapiv3.Document) []*Incompatibility {
	var incompatibilities []*Incompatibility
	if doc.Security == nil || len(doc.Security) == 0 {
		return incompatibilities
	}
	incompatibilities = append(incompatibilities,
		newIncompatibility(IncompatibiltiyClassification_Security, "security"))
	return incompatibilities
}

// PathsSearch is a reporter that scans for incompatibilities in the paths component of an OpenAPI doc
func PathsSearch(doc *openapiv3.Document) []*Incompatibility {
	var incompatibilities []*Incompatibility
	path := []string{"paths"}
	if doc.Paths == nil {
		return incompatibilities
	}
	for _, pathItem := range doc.Paths.Path {
		pathKey := extendPath(path, pathItem.Name)
		path := pathItem.Value
		if path.Head != nil {
			incompatibilities = append(incompatibilities,
				newIncompatibility(IncompatibiltiyClassification_InvalidOperation, extendPath(pathKey, "head")...))
		}
		if path.Options != nil {
			incompatibilities = append(incompatibilities,
				newIncompatibility(IncompatibiltiyClassification_InvalidOperation, extendPath(pathKey, "options")...))
		}
		if path.Trace != nil {
			incompatibilities = append(incompatibilities,
				newIncompatibility(IncompatibiltiyClassification_InvalidOperation, extendPath(pathKey, "trace")...))
		}
		incompatibilities = append(incompatibilities,
			validOperationSearch(path.Get, extendPath(pathKey, "get"))...)
		incompatibilities = append(incompatibilities,
			validOperationSearch(path.Put, extendPath(pathKey, "put"))...)
		incompatibilities = append(incompatibilities,
			validOperationSearch(path.Post, extendPath(pathKey, "post"))...)
		incompatibilities = append(incompatibilities,
			validOperationSearch(path.Delete, extendPath(pathKey, "delete"))...)
		incompatibilities = append(incompatibilities,
			validOperationSearch(path.Patch, extendPath(pathKey, "patch"))...)

		for ind, paramOrRef := range path.Parameters {
			incompatibilities = append(incompatibilities,
				parametersSearch(paramOrRef.GetParameter(), extendPath(pathKey, "parameters", strconv.Itoa(ind)))...)
		}
	}
	return incompatibilities
}

// ComponentsSearch is a reporter that scans for incompatibilities in the components object within an OpenAPI document
func ComponentsSearch(doc *openapiv3.Document) []*Incompatibility {
	var incompatibilities []*Incompatibility
	if doc.Components == nil {
		return incompatibilities
	}
	path := []string{"components"}

	if doc.Components.Callbacks != nil {
		incompatibilities = append(incompatibilities,
			newIncompatibility(IncompatibiltiyClassification_ExternalTranscodingSupport, extendPath(path, "callbacks")...))
	}
	if doc.Components.Schemas != nil {
		for _, schemaRef := range doc.Components.Schemas.GetAdditionalProperties() {
			incompatibilities = append(incompatibilities, schemaSearch(schemaRef.Value.GetSchema(), extendPath(path, "schemas", schemaRef.Name))...)
		}
	}
	if doc.Components.Responses != nil {
		for _, resRef := range doc.Components.Responses.GetAdditionalProperties() {
			incompatibilities = append(incompatibilities,
				responseSearch(resRef.Value.GetResponse(), extendPath(path, "requestBodies", resRef.Name))...,
			)
		}
	}
	if doc.Components.Parameters != nil {
		for _, paramRef := range doc.Components.Parameters.GetAdditionalProperties() {
			incompatibilities = append(incompatibilities,
				parametersSearch(paramRef.GetValue().GetParameter(), extendPath(path, "parameters", paramRef.Name))...,
			)
		}
	}
	if doc.Components.RequestBodies != nil {
		for _, reqRef := range doc.Components.RequestBodies.GetAdditionalProperties() {
			incompatibilities = append(incompatibilities,
				requestBodySearch(reqRef.Value.GetRequestBody(), extendPath(path, "requestBodies", reqRef.Name))...,
			)
		}
	}
	if doc.Components.Headers != nil {
		for _, comRef := range doc.Components.Headers.GetAdditionalProperties() {
			incompatibilities = append(incompatibilities,
				headerSearch(comRef.Name, comRef.Value.GetHeader(), extendPath(path, "headers", comRef.Name))...,
			)
		}
	}
	if doc.Components.SecuritySchemes != nil {
		incompatibilities = append(incompatibilities, newIncompatibility(IncompatibiltiyClassification_Security, extendPath(path, "securitySchemes")...))
	}

	return incompatibilities
}

// ========================= Helper Functions ======================== //

// validOperationSearch scans for incompatibilities within valid operations
func validOperationSearch(operation *openapiv3.Operation, path []string) []*Incompatibility {
	var incompatibilities []*Incompatibility
	if operation == nil {
		return incompatibilities
	}
	if operation.Callbacks != nil {
		incompatibilities = append(incompatibilities,
			newIncompatibility(IncompatibiltiyClassification_ExternalTranscodingSupport, extendPath(path, "callbacks")...))
	}
	if operation.Security != nil || len(operation.Security) != 0 {
		incompatibilities = append(incompatibilities,
			newIncompatibility(IncompatibiltiyClassification_Security, extendPath(path, "security")...))
	}
	for ind, paramOrRef := range operation.Parameters {
		incompatibilities = append(incompatibilities, parametersSearch(paramOrRef.GetParameter(), extendPath(path, "parameters", strconv.Itoa(ind)))...)
	}
	return incompatibilities

}

// pathsSearch scans for incompatibilities within a parameters object
func parametersSearch(param *openapiv3.Parameter, path []string) []*Incompatibility {
	var incompatibilities []*Incompatibility
	if param == nil {
		return incompatibilities
	}
	if param.Style != "" {
		incompatibilities = append(incompatibilities,
			newIncompatibility(IncompatibiltiyClassification_ParameterStyling, extendPath(path, "style")...))
	}
	if param.Explode {
		incompatibilities = append(incompatibilities,
			newIncompatibility(IncompatibiltiyClassification_ParameterStyling, extendPath(path, "explode")...))
	}
	if param.AllowReserved {
		incompatibilities = append(incompatibilities,
			newIncompatibility(IncompatibiltiyClassification_ParameterStyling, extendPath(path, "allowReserved")...))
	}
	if param.AllowEmptyValue {
		incompatibilities = append(incompatibilities,
			newIncompatibility(IncompatibiltiyClassification_DataValidation, extendPath(path, "allowEmptyValue")...))
	}
	if param.Schema != nil {
		incompatibilities = append(incompatibilities,
			schemaSearch(param.Schema.GetSchema(), extendPath(path, "schema"))...)
	}
	return incompatibilities
}

// schemaSearch scans for incompatibilities within a schema object
func schemaSearch(schema *openapiv3.Schema, path []string) []*Incompatibility {
	var incompatibilities []*Incompatibility
	if schema == nil {
		return incompatibilities
	}
	if schema.Nullable {
		incompatibilities = append(incompatibilities,
			newIncompatibility(IncompatibiltiyClassification_InvalidDataState, extendPath(path, "nullable")...))
	}
	if schema.Discriminator != nil {
		incompatibilities = append(incompatibilities,
			newIncompatibility(IncompatibiltiyClassification_Inheritance, extendPath(path, "discriminator")...))
	}
	if schema.ReadOnly {
		incompatibilities = append(incompatibilities,
			newIncompatibility(IncompatibiltiyClassification_ParameterStyling, extendPath(path, "readOnly")...))
	}
	if schema.WriteOnly {
		incompatibilities = append(incompatibilities,
			newIncompatibility(IncompatibiltiyClassification_ParameterStyling, extendPath(path, "writeOnly")...))
	}
	if schema.MultipleOf != 0 {
		incompatibilities = append(incompatibilities,
			newIncompatibility(IncompatibiltiyClassification_DataValidation, extendPath(path, "multipleOf")...))
	}
	if schema.Maximum != 0 {
		incompatibilities = append(incompatibilities,
			newIncompatibility(IncompatibiltiyClassification_DataValidation, extendPath(path, "maximum")...))
	}
	if schema.ExclusiveMaximum {
		incompatibilities = append(incompatibilities,
			newIncompatibility(IncompatibiltiyClassification_DataValidation, extendPath(path, "exclusiveMaximum")...))
	}
	if schema.Minimum != 0 {
		incompatibilities = append(incompatibilities,
			newIncompatibility(IncompatibiltiyClassification_DataValidation, extendPath(path, "minimum")...))
	}
	if schema.ExclusiveMinimum {
		incompatibilities = append(incompatibilities,
			newIncompatibility(IncompatibiltiyClassification_DataValidation, extendPath(path, "exclusiveMinimum")...))
	}
	if schema.MaxLength != 0 {
		incompatibilities = append(incompatibilities,
			newIncompatibility(IncompatibiltiyClassification_DataValidation, extendPath(path, "maxLength")...))
	}
	if schema.MinLength != 0 {
		incompatibilities = append(incompatibilities,
			newIncompatibility(IncompatibiltiyClassification_DataValidation, extendPath(path, "minimum")...))
	}
	if schema.Pattern != "" {
		incompatibilities = append(incompatibilities,
			newIncompatibility(IncompatibiltiyClassification_DataValidation, extendPath(path, "pattern")...))
	}
	if schema.MaxItems != 0 {
		incompatibilities = append(incompatibilities,
			newIncompatibility(IncompatibiltiyClassification_DataValidation, extendPath(path, "maxItems")...))
	}
	if schema.MinItems != 0 {
		incompatibilities = append(incompatibilities,
			newIncompatibility(IncompatibiltiyClassification_DataValidation, extendPath(path, "minItems")...))
	}
	if schema.UniqueItems {
		incompatibilities = append(incompatibilities,
			newIncompatibility(IncompatibiltiyClassification_DataValidation, extendPath(path, "uniqueItems")...))
	}
	if schema.AllOf != nil || len(schema.AllOf) != 0 {
		incompatibilities = append(incompatibilities,
			newIncompatibility(IncompatibiltiyClassification_Inheritance, extendPath(path, "allOf")...))
	}
	if schema.OneOf != nil || len(schema.OneOf) != 0 {
		incompatibilities = append(incompatibilities,
			newIncompatibility(IncompatibiltiyClassification_Inheritance, extendPath(path, "oneOf")...))
	}
	if schema.AnyOf != nil || len(schema.AnyOf) != 0 {
		incompatibilities = append(incompatibilities,
			newIncompatibility(IncompatibiltiyClassification_Inheritance, extendPath(path, "anyOf")...))
	}
	if schema.Items != nil {
		for ind, item := range schema.Items.SchemaOrReference {
			incompatibilities = append(incompatibilities, schemaSearch(item.GetSchema(), extendPath(path, "items", strconv.Itoa(ind)))...)
		}
	}
	if schema.Properties != nil {
		for _, prop := range schema.Properties.AdditionalProperties {
			incompatibilities = append(incompatibilities, schemaSearch(prop.Value.GetSchema(), extendPath(path, "properties", prop.Name))...)
		}
	}
	if schema.AdditionalProperties != nil && schema.AdditionalProperties.GetSchemaOrReference() != nil {
		incompatibilities = append(incompatibilities,
			schemaSearch(schema.AdditionalProperties.GetSchemaOrReference().GetSchema(), extendPath(path, "additionalProperties"))...)
	}

	return incompatibilities
}

// responseSearch scans for incompatibilities in a response object
func responseSearch(resp *openapiv3.Response, path []string) []*Incompatibility {
	var incompatibilities []*Incompatibility
	if resp == nil {
		return incompatibilities
	}
	if resp.Headers != nil {
		for _, prop := range resp.Headers.AdditionalProperties {
			incompatibilities = append(incompatibilities,
				headerSearch(prop.Name, prop.GetValue().GetHeader(), extendPath(path, prop.Name))...,
			)
		}
	}
	if resp.Content != nil {
		for _, prop := range resp.Content.AdditionalProperties {
			incompatibilities = append(incompatibilities,
				contentSearch(prop.Value, extendPath(path, prop.Name))...,
			)
		}
	}
	return incompatibilities
}

//  headerSearch scans for incompatibilities in a header object
func headerSearch(headerName string, header *openapiv3.Header, path []string) []*Incompatibility {
	var incompatibilities []*Incompatibility
	if header == nil {
		return incompatibilities
	}
	paramEquiv := header2Paramter(headerName, header)
	incompatibilities = append(incompatibilities,
		parametersSearch(paramEquiv, path)...)
	return incompatibilities
}

// contentSearch scans for incompatibilities in a media object
func contentSearch(media *openapiv3.MediaType, path []string) []*Incompatibility {
	var incompatibilities []*Incompatibility
	if media == nil {
		return incompatibilities
	}
	if media.Encoding != nil {
		incompatibilities = append(incompatibilities,
			newIncompatibility(IncompatibiltiyClassification_ParameterStyling, extendPath(path, "encoding")...))
	}
	if media.Schema != nil {
		incompatibilities = append(incompatibilities,
			schemaSearch(media.Schema.GetSchema(), extendPath(path, "schema"))...,
		)
	}
	return incompatibilities
}

// requestBodySearch scans for incompatibilities in a restBody object
func requestBodySearch(req *openapiv3.RequestBody, path []string) []*Incompatibility {
	var incompatibilities []*Incompatibility
	if req == nil {
		return incompatibilities
	}
	if req.Content != nil {
		for _, namedContent := range req.Content.GetAdditionalProperties() {
			incompatibilities = append(incompatibilities,
				contentSearch(namedContent.Value, extendPath(path, namedContent.Name))...,
			)
		}
	}
	return incompatibilities
}

func newIncompatibility(classification IncompatibiltiyClassification, path ...string) *Incompatibility {
	return &Incompatibility{
		TokenPath:      path,
		Classification: classification,
		Severity:       classificationSeverity(classification),
	}
}

// extendPath adds string to end of a copy of path
func extendPath(path []string, items ...string) (newPath []string) {
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
