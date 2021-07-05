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
	serverPath := "../oas-examples/petstore.yaml"

	var serversTest = []struct {
		path            string
		serversDetected bool
	}{
		{noServerPath, false},
		{serverPath, true},
	}
	for _, tt := range serversTest {
		if incompatibilityCheck(generateDoc(t, tt.path), "SERVERS") != tt.serversDetected {
			t.Errorf("Incorrect server detection for file at %s, got %t\n", noServerPath, tt.serversDetected)
		}
	}
}
