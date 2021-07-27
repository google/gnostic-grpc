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
	"path/filepath"
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

	var securityTest = []struct {
		path           string
		expectSecurity bool
	}{
		{"../../generator/testfiles/other.yaml", false},
		{"../../examples/petstore/petstore.yaml", true},
	}
	for _, trial := range securityTest {
		t.Run(filepath.Base(trial.path)+"SecurityCheck", func(tt *testing.T) {
			if incompatibilityCheck(generateDoc(tt, trial.path), IncompatibiltiyClassification_Security) != trial.expectSecurity {
				tt.Errorf("Incorrect security detection for file, got %t\n", trial.expectSecurity)
			}
		})
	}
}

func TestIncompatibilityExistence(t *testing.T) {

	var existenceTest = []struct {
		path string
	}{
		{"../../examples/petstore/petstore.yaml"},
		{"../oas-examples/petstore.json"},
		{"../../examples/bookstore/bookstore.yaml"},
	}

	for _, trial := range existenceTest {
		var node yaml.Node
		incompReport := ScanIncompatibilities(generateDoc(t, trial.path))
		data, _ := ioutil.ReadFile(trial.path)
		marshErr := yaml.Unmarshal(data, &node)
		if marshErr != nil {
			t.Fatalf("Unable to marshal file<%s>", trial.path)
		}
		for _, incomp := range incompReport.GetIncompatibilities() {
			t.Run(filepath.Base(trial.path)+"IncompExistence", func(tt *testing.T) {
				_, searchErr :=
					findNode(node.Content[0], incomp.GetTokenPath()...)
				if searchErr != nil {
					tt.Errorf(searchErr.Error())
				}
			})
		}

	}
}
