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

	openapiv3 "github.com/google/gnostic/openapiv3"
	plugins "github.com/google/gnostic/plugins"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"gopkg.in/yaml.v3"

	"github.com/google/gnostic-grpc/search"
	"github.com/google/gnostic-grpc/utils"
)

// Helper Test Function generate OpenAPI representation or Error
func generateDoc(t *testing.T, path string) *openapiv3.Document {
	document, err := utils.ParseOpenAPIDoc(path)
	if err != nil {
		t.Fatalf("Error while parsing input file: %s\n", path)
	}
	return document
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
		{sudoGnosticFlowBaseReport(t, "../examples/petstore/petstore.yaml")},
		{sudoGnosticFlowBaseReport(t, "oas-examples/petstore.json")},
		{sudoGnosticFlowBaseReport(t, "../examples/bookstore/bookstore.yaml")},
		{sudoGnosticFlowBaseReport(t, "oas-examples/openapi.yaml")},
		{sudoGnosticFlowBaseReport(t, "oas-examples/adsense.yaml")},
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

// Test process of writing an incompatibility report to plugin
// file object and then extract from object back to incompatibility
// report, also trests
func TestCreateIncompReports(t *testing.T) {
	var detailingTest = []struct {
		oasFilePath string
	}{
		{"../examples/petstore/petstore.yaml"},
		{"oas-examples/petstore.json"},
		{"../examples/bookstore/bookstore.yaml"},
		{"oas-examples/openapi.yaml"},
		{"oas-examples/adsense.yaml"},
	}

	for _, trial := range detailingTest {
		t.Run(trial.oasFilePath, func(tt *testing.T) {
			baseReportFromGnosticFlow := sudoGnosticFlowBaseReport(tt, trial.oasFilePath)
			descriptiveReportFromGnosticFlow := sudoGnosticFlowFDReport(tt, trial.oasFilePath)
			descDiff := cmp.Diff(detailReport(baseReportFromGnosticFlow), descriptiveReportFromGnosticFlow, ignoreUnexportedOption)
			if descDiff != "" {
				tt.Error("DetailIncompReport(...) : diff(-want +got)\n", descDiff)
			}

		})
	}
}

// Creates a sudo environment for incompatibility testing
func createSudoEnvironment(t *testing.T, filePath string) *plugins.Environment {
	doc, parseError := utils.ParseOpenAPIDoc(filePath)
	if parseError != nil {
		t.Fatal("parse error")
	}
	docBytes, marshalErr := proto.Marshal(doc)
	if marshalErr != nil {
		t.Fatal("unable to create sudo environment from marshal error")
	}
	sudoEnvironment := &plugins.Environment{
		Request: &plugins.Request{
			SourceName: filePath,
			Models: []*anypb.Any{
				{TypeUrl: "openapi.v3.Document", Value: docBytes},
			},
		},
		Response: &plugins.Response{
			Files: make([]*plugins.File, 0),
		},
	}
	return sudoEnvironment
}

// Formatting and ErrorHandling for Base Report creation GnosticIncompatibiltyScanning
func sudoGnosticFlowBaseReport(t *testing.T, filePath string) *IncompatibilityReport {
	var baseReport IncompatibilityReport
	sudoEnvironment := createSudoEnvironment(t, filePath)
	GnosticIncompatibiltyScanning(sudoEnvironment, BaseIncompatibility_Report)
	if len(sudoEnvironment.Response.Files) != 1 {
		t.Fatalf("Did not store singular base incompatibility report")
	}
	binData := sudoEnvironment.Response.Files[0].Data
	if prototext.Unmarshal(binData, &baseReport) != nil {
		t.Fatalf("Failed to create fd reoprt from gnostic")
	}
	return &baseReport
}

// Formatting and ErrorHandling for FileDescriptive Report creation in GnosticIncompatibiltyScanning
func sudoGnosticFlowFDReport(t *testing.T, filePath string) *FileDescriptiveReport {
	var fdReport FileDescriptiveReport
	sudoEnvironment := createSudoEnvironment(t, filePath)
	GnosticIncompatibiltyScanning(sudoEnvironment, FileDescriptive_Report)
	binData := sudoEnvironment.Response.Files[0].Data
	if len(sudoEnvironment.Response.Files) != 1 {
		t.Fatalf("Did not store singular base incompatibility report")
	}
	if prototext.Unmarshal(binData, &fdReport) != nil {
		t.Fatalf("Failed to create fd report from gnostic")
	}
	return &fdReport
}

// Error handling for makeNode
func createNodeFromFile(filePath string, t *testing.T) *yaml.Node {
	node, err := search.MakeNode(filePath)
	if err != nil {
		t.Fatalf(err.Error())
	}
	return node
}

// Error handling for findKey
func searchForIncompatibiltiy(node *yaml.Node, incomp *Incompatibility, t *testing.T) {
	_, _, searchErr := search.FindKey(node.Content[0], incomp.TokenPath...)
	if searchErr != nil {
		t.Errorf(searchErr.Error())
	}
}
