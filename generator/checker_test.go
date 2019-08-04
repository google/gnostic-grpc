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

package generator

import (
	openapiv3 "github.com/googleapis/gnostic/OpenAPIv3"
	plugins "github.com/googleapis/gnostic/plugins"
	"os/exec"
	"testing"
)

func TestNewFeatureCheckerParameters(t *testing.T) {
	input := "testfiles/parameters.yaml"
	documentv3 := readOpenAPIBinary(input)

	checker := NewGrpcChecker(documentv3)
	messages := checker.Run()
	expectedMessageTexts := []string{
		"Fields: Explode are not supported for parameter: param2",
		"Fields: Default are not supported for the schema: Items of param2",
		"Field: Enum is not generated as enum in .proto for schema: Items of param2",
		"Fields: Default are not supported for the schema: param4",
		"Field: Enum is not generated as enum in .proto for schema: param4",
	}
	validateMessages(t, expectedMessageTexts, messages)
}

func TestFeatureCheckerRequestBodies(t *testing.T) {
	input := "testfiles/requestBodies.yaml"
	documentv3 := readOpenAPIBinary(input)

	checker := NewGrpcChecker(documentv3)
	messages := checker.Run()
	expectedMessageTexts := []string{
		"Fields: Required are not supported for the schema: Person",
		"Fields: Example are not supported for the schema: name",
		"Fields: Xml are not supported for the schema: photoUrls",
		"Fields: Required are not supported for the request: RequestBody",
	}
	validateMessages(t, expectedMessageTexts, messages)
}

func TestFeatureCheckerResponses(t *testing.T) {
	input := "testfiles/responses.yaml"
	documentv3 := readOpenAPIBinary(input)

	checker := NewGrpcChecker(documentv3)
	messages := checker.Run()
	expectedMessageTexts := []string{
		"Fields: Required are not supported for the schema: Error",
		"Fields: Required are not supported for the schema: Person",
		"Fields: Example are not supported for the schema: name",
		"Fields: Xml are not supported for the schema: photoUrls",
	}
	validateMessages(t, expectedMessageTexts, messages)
}

func TestFeatureCheckerOther(t *testing.T) {
	input := "testfiles/other.yaml"
	documentv3 := readOpenAPIBinary(input)

	checker := NewGrpcChecker(documentv3)
	messages := checker.Run()
	expectedMessageTexts := []string{
		"Fields: Required are not supported for the schema: Person",
		"Fields: Example are not supported for the schema: name",
		"Fields: Xml are not supported for the schema: photoUrls",
		"Field: additionalProperties with type array is generated as empty message inside .proto.",
	}
	validateMessages(t, expectedMessageTexts, messages)
}

func validateMessages(t *testing.T, expectedMessageTexts []string, messages []*plugins.Message) {
	if len(expectedMessageTexts) != len(messages) {
		t.Errorf("Number of messages from GrpcChecker does not match expected number")
		return
	}
	for i, msg := range messages {
		if msg.Text != expectedMessageTexts[i] {
			t.Errorf("Message text does not match expected message text: %s != %s", msg.Text, expectedMessageTexts[i])
		}
	}
}

func readOpenAPIBinary(input string) *openapiv3.Document {
	cmd := exec.Command("gnostic", "--pb-out=-", input)
	b, _ := cmd.Output()
	documentv3, _ := createOpenAPIDocFromGnosticOutput(b)
	return documentv3
}
