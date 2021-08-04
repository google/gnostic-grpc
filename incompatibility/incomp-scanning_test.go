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
	"path/filepath"
	"testing"

	"github.com/googleapis/gnostic-grpc/search"
	"github.com/googleapis/gnostic-grpc/utils"
	openapiv3 "github.com/googleapis/gnostic/openapiv3"
	"gopkg.in/yaml.v3"
)

// Helper Function to check for single incompatibilty
func incompatibilityCheck(document *openapiv3.Document, path string, incompatibilityClass IncompatibiltiyClassification) bool {
	for _, incompatibility := range ScanIncompatibilities(document, path).Incompatibilities {
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
		{"../generator/testfiles/other.yaml", false},
		{"../examples/petstore/petstore.yaml", true},
	}
	for _, trial := range securityTest {
		t.Run(filepath.Base(trial.path)+"SecurityCheck", func(tt *testing.T) {
			if incompatibilityCheck(generateDoc(tt, trial.path), trial.path, IncompatibiltiyClassification_Security) != trial.expectSecurity {
				tt.Errorf("Incorrect security detection for file, got %t\n", trial.expectSecurity)
			}
		})
	}
}

// Assert that all reported incompatibility paths can be found when travering file
func TestIncompatibilityExistence(t *testing.T) {

	var existenceTest = []struct {
		path string
	}{
		{"../examples/petstore/petstore.yaml"},
		{"oas-examples/petstore.json"},
		{"../examples/bookstore/bookstore.yaml"},
		{"oas-examples/openapi.yaml"},
		{"oas-examples/adsense.yaml"},
	}

	for _, trial := range existenceTest {
		incompReport := createReport(t, trial.path)
		baseNode := createNodeFromFile(incompReport.ReportIdentifier, t)
		for _, incomp := range incompReport.GetIncompatibilities() {
			t.Run(filepath.Base(trial.path)+"IncompExistence", func(tt *testing.T) {
				searchForIncompatibiltiy(baseNode, incomp, t)
			})
		}

	}
}

// Test verifying detailing proces of an incompatibility report
func TestDetailing(t *testing.T) {
	var detailingTest = []struct {
		baseReport *IncompatibilityReport
	}{
		{createReport(t, "../examples/petstore/petstore.yaml")},
		{createReport(t, "oas-examples/petstore.json")},
		{createReport(t, "../examples/bookstore/bookstore.yaml")},
		{createReport(t, "oas-examples/openapi.yaml")},
		{createReport(t, "oas-examples/adsense.yaml")},
	}

	for _, trial := range detailingTest {
		t.Run(trial.baseReport.ReportIdentifier, func(tt *testing.T) {
			numIncompatibilitiesBaseReport := len(trial.baseReport.Incompatibilities)
			numIncompatibilitiesIDReport := len(detailReport(trial.baseReport).Incompatibilities)
			if numIncompatibilitiesBaseReport != numIncompatibilitiesIDReport {
				t.Errorf("len(IDReport(%s)): got: %d wanted: %d", trial.baseReport.ReportIdentifier,
					numIncompatibilitiesIDReport,
					numIncompatibilitiesBaseReport,
				)
			}
		})
	}
}

func createNodeFromFile(filePath string, t *testing.T) *yaml.Node {
	node, err := search.MakeNode(filePath)
	if err != nil {
		t.Fatalf(err.Error())
	}
	return node
}

func searchForIncompatibiltiy(node *yaml.Node, incomp *Incompatibility, t *testing.T) {
	_, _, searchErr := search.FindKey(node.Content[0], incomp.TokenPath...)
	if searchErr != nil {
		t.Errorf(searchErr.Error())
	}
}
