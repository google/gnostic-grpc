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

var IncompatibilityChains []ChainIncompatibilitySearch = []ChainIncompatibilitySearch{
	DocumentBaseSearch,
	PathsSearch,
}

type ChainIncompatibilitySearch func(*openapiv3.Document) (incompatibilities []*Incompatibility)

func SearchChains(doc *openapiv3.Document, chains ...ChainIncompatibilitySearch) *IncompatibilityReport {
	var incompatibilities []*Incompatibility
	for _, chain := range chains {
		incompatibilities = append(incompatibilities, chain(doc)...)
	}
	return &IncompatibilityReport{Incompatibilities: incompatibilities}
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
func PathsSearch(doc *openapiv3.Document) (incompatibilities []*Incompatibility) {
	pathsKey := []string{"paths"}
	if doc.Paths == nil {
		return
	}
	for _, pathItem := range doc.Paths.Path {
		pathKey := AddKeyPath(pathsKey, pathItem.Name)
		pathValue := pathItem.Value
		if pathValue.Head != nil {
			incompatibilities = append(incompatibilities, NewIncompatibility(AddKeyPath(pathKey, "head"), "HEAD"))
		}
		if pathValue.Options != nil {
			incompatibilities = append(incompatibilities, NewIncompatibility(AddKeyPath(pathKey, "options"), "OPTIONS"))
		}
		if pathValue.Trace != nil {
			incompatibilities = append(incompatibilities, NewIncompatibility(AddKeyPath(pathKey, "trace"), "TRACE"))
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
	return
}

// ========================= Helper Functions ======================== //

// ValidOperationSearch scans for incompatibilities within valid operations
func ValidOperationSearch(operation *openapiv3.Operation, keys []string) (incompatibilities []*Incompatibility) {
	if operation == nil {
		return
	}
	if operation.Callbacks != nil {
		incompatibilities = append(incompatibilities, NewIncompatibility(AddKeyPath(keys, "callbacks"), "CALLBACKS"))
	}
	if operation.Security != nil {
		incompatibilities = append(incompatibilities, NewIncompatibility(AddKeyPath(keys, "security"), "SECURITY"))
	}
	for ind, paramOrRef := range operation.Parameters {
		incompatibilities = append(incompatibilities, ParametersSearch(paramOrRef.GetParameter(), AddKeyPath(keys, "parameters", strconv.Itoa(ind)))...)
	}
	return

}

func ParametersSearch(param *openapiv3.Parameter, keys []string) (incompatibilities []*Incompatibility) {
	if param == nil {
		return
	}
	if param.Style != "" {
		incompatibilities = append(incompatibilities, NewIncompatibility(AddKeyPath(keys, "style"), "STYLE"))
	}
	if param.Explode {
		incompatibilities = append(incompatibilities, NewIncompatibility(AddKeyPath(keys, "explode"), "EXPLODE"))
	}
	if param.AllowReserved {
		incompatibilities = append(incompatibilities, NewIncompatibility(AddKeyPath(keys, "allowReserved"), "ALLOWRESERVED"))
	}
	if param.AllowEmptyValue {
		incompatibilities = append(incompatibilities, NewIncompatibility(AddKeyPath(keys, "allowEmptyValue"), "ALLOWEMPTYVALUE"))
	}
	if param.Schema != nil {
		incompatibilities = append(incompatibilities, SchemaSearch(param.Schema.GetSchema(), AddKeyPath(keys, "schema"))...)
	}
	return
}

func SchemaSearch(schema *openapiv3.Schema, keys []string) (incompatibilities []*Incompatibility) {
	if schema == nil {
		return
	}
	if schema.Nullable {
		incompatibilities = append(incompatibilities, NewIncompatibility(AddKeyPath(keys, "nullable"), "NULLABLE"))
	}
	if schema.Discriminator != nil {
		incompatibilities = append(incompatibilities, NewIncompatibility(AddKeyPath(keys, "discriminator"), "DISCRIMINATOR"))
	}
	return
}

func NewIncompatibility(path []string, classification string) *Incompatibility {
	return &Incompatibility{TokenPath: path, Classification: classification}
}

// AddKeyPath adds string to end of a copy of path
func AddKeyPath(path []string, items ...string) (newPath []string) {
	newPath = make([]string, len(path))
	copy(newPath, path)
	newPath = append(newPath, items...)
	println(path, newPath)
	return
}
