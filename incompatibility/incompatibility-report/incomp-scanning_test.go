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
	"github.com/googleapis/gnostic-grpc/utils"
	openapiv3 "github.com/googleapis/gnostic/openapiv3"

	// openapiv3 "github.com/googleapis/gnostic/openapiv3"
	"testing"
)

// Helper Function to check for single incompatibilty
func IncompatibilityCheck(document *openapiv3.Document, incompatibilityClass string) bool {
	for _, incompatibility := range ScanIncompatibilities(document).Incompatibilities {
		if incompatibility.Classification == incompatibilityClass {
			return true
		}
	}
	return false
}

// Simple test for servers incompatibility
func TestBasicServerIncompatibility(t *testing.T) {
	noServerPath := "../../generator/testfiles/other.yaml"
	serverAbsentDocument, err := utils.ParseOpenAPIDoc(noServerPath)
	if err != nil {
		t.Fatalf("Error while parsing input file: %s\n", noServerPath)
	}
	if IncompatibilityCheck(serverAbsentDocument, "SERVERS") {
		t.Errorf("Reporting false servers incompatibility for file at %s\n", noServerPath)
	}

	serverPath := "../oas-examples/petstore.yaml"
	serverPresentDocument, err := utils.ParseOpenAPIDoc(serverPath)
	if err != nil {
		t.Fatalf("Error while parsing input file: %s\n", serverPath)
	}
	if !IncompatibilityCheck(serverPresentDocument, "SERVERS") {
		t.Errorf("Failed to report server incompatibility at %s", serverPath)
	}
}
