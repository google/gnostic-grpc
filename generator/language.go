// Copyright 2019 Google Inc. All Rights Reserved.
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
package generator

import (
	"regexp"
	"strconv"
	"strings"

	surface_v1 "github.com/google/gnostic/surface"
)

type ProtoLanguageModel struct{}

func NewProtoLanguageModel() *ProtoLanguageModel {
	return &ProtoLanguageModel{}
}

// Prepare sets language-specific properties for all types and methods.
func (language *ProtoLanguageModel) Prepare(model *surface_v1.Model, inputDocumentType string) {
	for _, t := range model.Types {
		// determine the name of protocol buffer messages

		t.TypeName = protoTypeName(strings.Replace(t.Name, "Parameters", "Request", 1))

		for _, f := range t.Fields {
			f.FieldName = protoFieldName(f.Name, f.Type)
			f.NativeType = findNativeType(f.Type, f.Format)

			if f.EnumValues != nil {
				f.NativeType = strings.Title(f.Name)
			}
		}
	}

	for _, m := range model.Methods {
		m.HandlerName = protoTypeName(m.Name)
		m.ProcessorName = m.Name
		m.ClientName = m.Name
		m.ParametersTypeName = protoTypeName(strings.Replace(m.ParametersTypeName, "Parameters", "Request", 1))
		m.ResponsesTypeName = protoTypeName(m.ResponsesTypeName)
	}

	AdjustSurfaceModel(model, inputDocumentType)
}

// findNativeType maps OpenAPI data types (https://swagger.io/docs/specification/data-models/data-types/)
// to .proto types (https://developers.google.com/protocol-buffers/docs/proto3#scalar)
func findNativeType(fType string, fFormat string) string {
	switch fType {
	case "boolean":
		return "bool"
	case "number":
		switch fFormat {
		case "float":
			return "float"
		case "double":
			return "double"
		default:
			return "float"
		}
	case "integer":
		switch fFormat {
		case "uint32":
			return "uint32"
		case "int32":
			return "int32"
		case "uint64":
			return "uint64"
		case "int64":
			return "int64"
		default:
			return "int64"
		}
	case "object":
		return "message"
	case "string":
		switch fFormat {
		case "string":
			return "string"
		case "byte":
			return "bytes"
		default:
			return "string"
		}
	case "date":
		return "string"
	case "date-time":
		return "string"
	case "password":
		return "string"
	case "binary":
		return "bytes"
	case "email":
		return "string"
	case "uuid":
		return "string"
	case "uri":
		return "string"
	case "hostname":
		return "string"
	case "ipv4":
		return "string"
	case "ipv6":
		return "string"
	case "byte":
		return "string"
	default:
		if strings.Contains(fType, "map") {
			mapType := fType[11:]
			formattedType := map[string]bool{
				"int32": true,
				"int64": true,
			}
			if !formattedType[mapType] {
				return "map[string]" + findNativeType(mapType, "")
			}
			return fType
		}
		return protoTypeName(fType)
	}
}

// AdjustSurfaceModel simplifies and prettifies the types and fields of the surface model in order to get a better
// looking output file.
// Related to: https://github.com/google/gnostic-grpc/issues/11
func AdjustSurfaceModel(model *surface_v1.Model, inputDocumentType string) {
	if inputDocumentType == "openapi.v2.Document" {
		adjustV2Model(model)
	} else if inputDocumentType == "openapi.v3.Document" {
		adjustV3Model(model)
	} else if inputDocumentType == "discovery.v1.Document" {
		// TODO: We handle discovery format the same way like we handle v3 input files (which is probably wrong?).
		// Either fix this if someone complains or throw a warning in checker.go, since according to the README.md
		// gnnostic-grpc handles v3 schemas only. However, if other plugins also depend on this function this should be fixed!
		adjustV3Model(model)
	}

}

// adjustV3Model removes unnecessary types from the surface model. The original input file is an OpenAPI v2 file.
func adjustV3Model(model *surface_v1.Model) {
	nameToType, typesToDelete := initHashTables(model)
	for _, m := range model.Methods {
		if len(m.ParametersTypeName) > 0 {
			if parameters, ok := nameToType[m.ParametersTypeName]; ok {
				// For requestBodies we remove the intermediate type.
				for _, f := range parameters.Fields {
					if f.Name == "request_body" {
						reqBody := f
						if intermediateType, ok := nameToType[reqBody.NativeType]; ok {
							reqBody.FieldName = intermediateType.Fields[0].FieldName
							reqBody.NativeType = intermediateType.Fields[0].NativeType
							typesToDelete[intermediateType] = true
						}
					}
				}
			}
		}

		// We only render messages and types for the response with the lowest status code.
		if len(m.ResponsesTypeName) > 0 {
			if responses, ok := nameToType[m.ResponsesTypeName]; ok {
				// We remove the current response type which holds the responses for all status codes
				typesToDelete[nameToType[m.ResponsesTypeName]] = true

				// We remove all status codes as well
				for _, f := range responses.Fields {
					typesToDelete[nameToType[f.NativeType]] = true
				}

				lowestStatusCodeResponse := findLowestStatusCode(responses, nameToType)

				m.ResponsesTypeName = ""
				if lowestStatusCodeResponse != nil && lowestStatusCodeResponse.Fields[0].Kind != surface_v1.FieldKind_SCALAR {
					// We set the response with the lowest status code as response.
					m.ResponsesTypeName = lowestStatusCodeResponse.Fields[0].NativeType
				} else {
					// The nameToType hash map does not contain values from symbolic references. So if the OpenAPI
					// description we want to generate, references a response parameter inside another OpenAPI description
					// we end up here. Let's not render anything for now.
					m.ResponsesTypeName = ""
					typesToDelete[responses] = true
				}
			}
		}
	}

	// Remove types that we don't want to render
	filteredTypes := make([]*surface_v1.Type, 0)
	for _, t := range model.Types {
		if shouldDelete, ok := typesToDelete[t]; ok && !shouldDelete {
			filteredTypes = append(filteredTypes, t)
		}
	}
	model.Types = filteredTypes
}

// adjustV2Model removes types from the surface model. The original input file is an OpenAPI v2 file.
func adjustV2Model(model *surface_v1.Model) {
	nameToType, typesToDelete := initHashTables(model)
	for _, m := range model.Methods {
		// We only render messages and types for the response with the lowest status code.
		if len(m.ResponsesTypeName) > 0 {
			if responses, ok := nameToType[m.ResponsesTypeName]; ok {
				// We remove the current response type which holds the responses for all status codes
				typesToDelete[nameToType[m.ResponsesTypeName]] = true

				lowestStatusCodeResponse := findLowestStatusCode(responses, nameToType)
				m.ResponsesTypeName = ""
				if lowestStatusCodeResponse != nil {
					// We set the response with the lowest status code as response.
					m.ResponsesTypeName = lowestStatusCodeResponse.TypeName
				} else {
					// The nameToType hash map does not contain values from symbolic references. So if the OpenAPI
					// description we want to generate, references a response parameter inside another OpenAPI description
					// we end up here. Let's not render anything for now.
					m.ResponsesTypeName = ""
					typesToDelete[responses] = true
				}
			}
		}
	}

	// Remove types that we don't want to render
	filteredTypes := make([]*surface_v1.Type, 0)
	for _, t := range model.Types {
		if shouldDelete, ok := typesToDelete[t]; ok && !shouldDelete {
			filteredTypes = append(filteredTypes, t)
		}
	}
	model.Types = filteredTypes
}

// findLowestStatusCode returns a surface Type that represents the lowest status code for the given 'responses' type.
func findLowestStatusCode(responses *surface_v1.Type, nameToType map[string]*surface_v1.Type) *surface_v1.Type {
	if lowestStatusCodeResponse, ok := nameToType[responses.Fields[0].NativeType]; ok {
		lowestStatusCode, err := strconv.Atoi(responses.Fields[0].FieldName)
		if err == nil {
			for _, f := range responses.Fields {
				statusCode, err := strconv.Atoi(f.FieldName)
				if err == nil && statusCode < lowestStatusCode {
					lowestStatusCodeResponse = nameToType[f.NativeType]
					lowestStatusCode = statusCode
				}
			}
		}
		return lowestStatusCodeResponse
	}
	return nil
}

// initHashTables is a helper function to initialize two hash tables which are used in adjustV2Model and adjustV2Model
func initHashTables(model *surface_v1.Model) (map[string]*surface_v1.Type, map[*surface_v1.Type]bool) {
	nameToType := make(map[string]*surface_v1.Type)
	typesToDelete := make(map[*surface_v1.Type]bool)

	for _, t := range model.Types {
		nameToType[t.TypeName] = t
	}

	for _, t := range model.Types {
		typesToDelete[t] = false
	}
	return nameToType, typesToDelete
}

// protoFieldName returns the field names of proto messages according to
// https://developers.google.com/protocol-buffers/docs/style#message-and-field-names
func protoFieldName(originalName string, t string) string {
	name := CleanName(originalName)
	if len(name) == 0 {
		name = CleanName(t)
	}
	//name = toSnakeCase(name)
	return name
}

// protoTypeName returns the name of the proto message according to
// https://developers.google.com/protocol-buffers/docs/style#message-and-field-names
func protoTypeName(originalName string) (name string) {
	name = CleanName(originalName)
	name = toCamelCase(name)
	return name
}

// Removes characters which are not allowed for message names or field names inside .proto files.
func CleanName(name string) string {
	name = strings.Replace(name, "application/json", "", -1)
	name = strings.Replace(name, ".", "_", -1)
	name = strings.Replace(name, "-", "_", -1)
	name = strings.Replace(name, " ", "", -1)
	name = strings.Replace(name, "(", "", -1)
	name = strings.Replace(name, ")", "", -1)
	name = strings.Replace(name, "{", "", -1)
	name = strings.Replace(name, "}", "", -1)
	name = strings.Replace(name, "/", "_", -1)
	name = strings.Replace(name, "$", "", -1)
	return name
}

// toCamelCase converts str to CamelCase
func toCamelCase(str string) string {
	var link = regexp.MustCompile("(^[A-Za-z])|_([A-Za-z])")
	return link.ReplaceAllStringFunc(str, func(s string) string {
		return strings.ToUpper(strings.Replace(s, "_", "", -1))
	})
}

// toSnakeCase converts str to snake_case
func toSnakeCase(str string) string {
	var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func getRefName(str string) string {
	arr := strings.Split(str, "/")
	w := arr[len(arr)-1]
	return w
}
