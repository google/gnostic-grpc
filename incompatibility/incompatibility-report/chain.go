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

var IncompatibilityReporters []IncompatibilityReporter = []IncompatibilityReporter{
	DocumentBaseSearch,
	PathsSearch,
}

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

func SearchChains(doc *openapiv3.Document, reporters ...IncompatibilityReporter) *IncompatibilityReport {
	return &IncompatibilityReport{Incompatibilities: aggregateIncompatibilityReporters(reporters...)(doc)}
}

// ======================== Hard Coded Chains ====================== //

// DocumentBaseSearch is a chain that scans for incompatibilities at the base of an OpenAPI doc
func DocumentBaseSearch(doc *openapiv3.Document) (incompatibilities []*Incompatibility) {
	if doc.Security != nil {
		incompatibilities = append(incompatibilities,
			&Incompatibility{TokenPath: []string{"security"}, Classification: "SECURITY"})
	}
	return
}

// PathsSearch is a chain that scans for incompatibilities in the paths component of an OpenAPI doc
func PathsSearch(doc *openapiv3.Document) []*Incompatibility {
	var incompatibilities []*Incompatibility
	pathsKey := []string{"paths"}
	if doc.Paths == nil {
		return incompatibilities
	}
	for _, pathItem := range doc.Paths.Path {
		pathKey := AddKeyPath(pathsKey, pathItem.Name)
		pathValue := pathItem.Value
		if pathValue.Head != nil {
			incompatibilities = append(incompatibilities, NewIncompatibility("HEAD", AddKeyPath(pathKey, "head")...))
		}
		if pathValue.Options != nil {
			incompatibilities = append(incompatibilities, NewIncompatibility("OPTIONS", AddKeyPath(pathKey, "options")...))
		}
		if pathValue.Trace != nil {
			incompatibilities = append(incompatibilities, NewIncompatibility("TRACE", AddKeyPath(pathKey, "trace")...))
		}
		incompatibilities = append(incompatibilities,
			ValidOperationSearch(pathValue.Get, AddKeyPath(pathKey, "get"))...)
		incompatibilities = append(incompatibilities,
			ValidOperationSearch(pathValue.Put, AddKeyPath(pathKey, "put"))...)
		incompatibilities = append(incompatibilities,
			ValidOperationSearch(pathValue.Post, AddKeyPath(pathKey, "post"))...)
		incompatibilities = append(incompatibilities,
			ValidOperationSearch(pathValue.Delete, AddKeyPath(pathKey, "delete"))...)
		incompatibilities = append(incompatibilities,
			ValidOperationSearch(pathValue.Patch, AddKeyPath(pathKey, "patch"))...)
	}
	return incompatibilities
}

// ========================= Helper Functions ======================== //

// ValidOperationSearch scans for incompatibilities within valid operations
func ValidOperationSearch(operation *openapiv3.Operation, keys []string) []*Incompatibility {
	var incompatibilities []*Incompatibility
	if operation != nil {
		if operation.Callbacks != nil {
			incompatibilities = append(incompatibilities, NewIncompatibility("CALLBACKS", AddKeyPath(keys, "callbacks")...))
		}
		if operation.Security != nil {
			incompatibilities = append(incompatibilities, NewIncompatibility("SECURITY", AddKeyPath(keys, "security")...))
		}
		for ind, paramOrRef := range operation.Parameters {
			incompatibilities = append(incompatibilities, ParametersSearch(paramOrRef.GetParameter(), AddKeyPath(keys, "parameters", strconv.Itoa(ind)))...)
		}
	}
	return incompatibilities

}

func ParametersSearch(param *openapiv3.Parameter, keys []string) []*Incompatibility {
	var incompatibilities []*Incompatibility
	if param != nil {
		if param.Style != "" {
			incompatibilities = append(incompatibilities, NewIncompatibility("STYLE", AddKeyPath(keys, "style")...))
		}
		if param.Explode {
			incompatibilities = append(incompatibilities, NewIncompatibility("EXPLODE", AddKeyPath(keys, "explode")...))
		}
		if param.AllowReserved {
			incompatibilities = append(incompatibilities, NewIncompatibility("ALLOWRESERVED", AddKeyPath(keys, "allowReserved")...))
		}
		if param.AllowEmptyValue {
			incompatibilities = append(incompatibilities, NewIncompatibility("ALLOWEMPTYVALUE", AddKeyPath(keys, "allowEmptyValue")...))
		}
		if param.Schema != nil {
			incompatibilities = append(incompatibilities, SchemaSearch(param.Schema.GetSchema(), AddKeyPath(keys, "schema"))...)
		}
	}
	return incompatibilities
}

func SchemaSearch(schema *openapiv3.Schema, keys []string) []*Incompatibility {
	var incompatibilities []*Incompatibility
	if schema != nil {
		if schema.Nullable {
			incompatibilities = append(incompatibilities, NewIncompatibility("NULLABLE", AddKeyPath(keys, "nullable")...))
		}
		if schema.Discriminator != nil {
			incompatibilities = append(incompatibilities, NewIncompatibility("DISCRIMINATOR", AddKeyPath(keys, "discriminator")...))
		}
	}
	return incompatibilities
}

func NewIncompatibility(classification string, path ...string) *Incompatibility {
	return &Incompatibility{TokenPath: path, Classification: classification}
}

// AddKeyPath adds string to end of a copy of path
func AddKeyPath(path []string, items ...string) (newPath []string) {
	newPath = make([]string, len(path))
	copy(newPath, path)
	newPath = append(newPath, items...)
	return
}
