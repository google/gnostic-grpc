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
	"io/ioutil"
	"testing"

	"github.com/googleapis/gnostic-grpc/utils"
	openapiv3 "github.com/googleapis/gnostic/openapiv3"
	"gopkg.in/yaml.v3"
)

// Helper Function to check for single incompatibilty
func incompatibilityCheck(document *openapiv3.Document, incompatibilityClass IncompatibiltiyClassification) bool {
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

// Simple test for security incompatibility
func TestBasicSecurityIncompatibility(t *testing.T) {
	path1 := "../../generator/testfiles/other.yaml"
	path2 := "../../examples/petstore/petstore.yaml"

	var securityTest = []struct {
		path           string
		expectSecurity bool
	}{
		{path1, false},
		{path2, true},
	}
	for _, tt := range securityTest {
		if incompatibilityCheck(generateDoc(t, tt.path), IncompatibiltiyClassification_Security) != tt.expectSecurity {
			t.Errorf("Incorrect security detection for file at %s, got %t\n", path1, tt.expectSecurity)
		}
	}
}

func TestIncompatibilityExistence(t *testing.T) {
	petstorePath := "../../examples/petstore/petstore.yaml"
	petstoreJson := "../oas-examples/petstore.json"
	bookstorePath := "../../examples/bookstore/bookstore.yaml"

	var existenceTest = []struct {
		path string
	}{
		{petstorePath},
		{petstoreJson},
		{bookstorePath},
	}

	for _, tt := range existenceTest {
		var node yaml.Node
		incompReport := ScanIncompatibilities(generateDoc(t, tt.path))
		data, _ := ioutil.ReadFile(tt.path)
		marshErr := yaml.Unmarshal(data, &node)
		if marshErr != nil {
			t.Fatalf("Unable to marshal file<%s>", tt.path)
		}
		for _, incomp := range incompReport.GetIncompatibilities() {
			_, searchErr :=
				findNode(node.Content[0], incomp.GetTokenPath()...)
			if searchErr != nil {
				t.Errorf(searchErr.Error())
			}
			// fmt.Printf("ln: %d, col: %d, incomp: %+v\n\n", fnde.Line, fnde.Column, incomp.TokenPath)
		}

	}
}
