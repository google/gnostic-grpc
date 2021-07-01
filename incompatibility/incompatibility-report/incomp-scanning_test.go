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

package incompatibility

import (
	"testing"

	"github.com/googleapis/gnostic-grpc/utils"
	openapiv3 "github.com/googleapis/gnostic/openapiv3"
)

// Helper Function to check for single incompatibilty
func incompatibilityCheck(document *openapiv3.Document, incompatibilityClass string) bool {
	for _, incompatibility := range ScanIncompatibilities(document).Incompatibilities {
		if incompatibility.Classification == incompatibilityClass {
			return true
		}
	}
	return false
}

// Helper Test Function generate OpenAPI representation or Error
func generateDoc(t *testing.T, path string) *openapiv3.Document {
	document, err := utils.ParseOpenAPIDoc(path)
	if err != nil {
		t.Fatalf("Error while parsing input file: %s\n", path)
	}
	return document
}

// Simple test for servers incompatibility
func TestBasicServerIncompatibility(t *testing.T) {
	noServerPath := "../../generator/testfiles/other.yaml"
	if incompatibilityCheck(generateDoc(t, noServerPath), "SERVERS") {
		t.Errorf("Reporting false servers incompatibility for file at %s\n", noServerPath)
	}

	serverPath := "../oas-examples/petstore.yaml"
	if incompatibilityCheck(generateDoc(t, serverPath), "SERVERS") {
		t.Errorf("Failed to report server incompatibility at %s", serverPath)
	}
}
