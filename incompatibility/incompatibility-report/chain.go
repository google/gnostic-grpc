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
	openapiv3 "github.com/googleapis/gnostic/openapiv3"
)

type ChainIncompatibilitySearch func(*openapiv3.Document) (incompatibilities []*Incompatibility)

const DefinedChains []ChainIncompatibilitySearch = []ChainIncompatibilitySearch{DocumentBaseSearch, PathsSearch}

func SearchChains(doc *openapiv3.Document, chains ...ChainIncompatibilitySearch) *IncompatibilityReport {
	var incompatibilities []*Incompatibility
	for _, chain := range chains {
		incompatibilities = append(incompatibilities, chain(doc)...)
	}
	return &IncompatibilityReport{Incompatibilities: incompatibilities}
}

// GetKnownIncompatibilityPaths combines hardcoded chains
func GetKnownIncompatibilityPaths() (chains []ChainIncompatibilitySearch) {
	chains = append(chains, DocumentBaseSearch, PathsSearch)
	return
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
			incompatibilities = append(incompatibilities, &Incompatibility{TokenPath: AddKeyPath(pathKey, "head"), Classification: "HEAD"})
		}
		if pathValue.Options != nil {
			incompatibilities = append(incompatibilities, &Incompatibility{TokenPath: AddKeyPath(pathKey, "options"), Classification: "OPTIONS"})
		}
		if pathValue.Trace != nil {
			incompatibilities = append(incompatibilities, &Incompatibility{TokenPath: AddKeyPath(pathKey, "trace"), Classification: "TRACE"})
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
		incompatibilities = append(incompatibilities, &Incompatibility{TokenPath: AddKeyPath(keys, "callbacks")})
	}
	if operation.Security != nil {
		incompatibilities = append(incompatibilities, &Incompatibility{TokenPath: AddKeyPath(keys, "security")})
	}
	return

}

// AddKeyPath adds string to end of a copy of path
func AddKeyPath(path []string, item string) (newPath []string) {
	newPath = make([]string, len(path))
	copy(newPath, path)
	newPath = append(newPath, item)
	println(path, newPath)
	return
}
