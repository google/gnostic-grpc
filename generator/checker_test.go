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
	"testing"

	plugins "github.com/google/gnostic/plugins"

	"github.com/google/gnostic-grpc/utils"
)

func TestNewFeatureCheckerParameters(t *testing.T) {
	input := "testfiles/parameters.yaml"
	documentv3, err := utils.ParseOpenAPIDoc(input)
	if err != nil {
		t.Errorf("Error while parsing input file: %s", input)
		return
	}

	checker := NewGrpcChecker(documentv3)
	messages := checker.Run()
	expectedMessageKeys := [][]string{
		{"components", "parameters", "required"},
		{"paths", "/testParameterQueryEnum", "get", "parameters", "explode"},
		{"paths", "/testParameterQueryEnum", "get", "parameters", "schema", "items", "default"},
		{"paths", "/testParameterPathEnum/{param1}", "get", "parameters", "schema", "default"},
	}
	validateKeys(t, expectedMessageKeys, messages)
}

func TestFeatureCheckerRequestBodies(t *testing.T) {
	input := "testfiles/requestBodies.yaml"
	documentv3, err := utils.ParseOpenAPIDoc(input)
	if err != nil {
		t.Errorf("Error while parsing input file: %s", input)
		return
	}

	checker := NewGrpcChecker(documentv3)
	messages := checker.Run()
	expectedMessageKeys := [][]string{
		{"components", "schemas", "Person", "required"},
		{"components", "schemas", "Person", "properties", "name", "example"},
		{"components", "schemas", "Person", "properties", "photoUrls", "xml"},
		{"components", "requestBodies", "RequestBody", "required"},
	}
	validateKeys(t, expectedMessageKeys, messages)
}

func TestFeatureCheckerResponses(t *testing.T) {
	input := "testfiles/responses.yaml"
	documentv3, err := utils.ParseOpenAPIDoc(input)
	if err != nil {
		t.Errorf("Error while parsing input file: %s", input)
		return
	}

	checker := NewGrpcChecker(documentv3)
	messages := checker.Run()
	expectedMessageKeys := [][]string{
		{"components", "schemas", "Error", "required"},
		{"components", "schemas", "Person", "required"},
		{"components", "schemas", "Person", "properties", "name", "example"},
		{"components", "schemas", "Person", "properties", "photoUrls", "xml"},
	}
	validateKeys(t, expectedMessageKeys, messages)
}

func TestFeatureCheckerOther(t *testing.T) {
	input := "testfiles/other.yaml"
	documentv3, err := utils.ParseOpenAPIDoc(input)
	if err != nil {
		t.Errorf("Error while parsing input file: %s", input)
		return
	}

	checker := NewGrpcChecker(documentv3)
	messages := checker.Run()
	expectedMessageKeys := [][]string{
		{"components", "schemas", "Person", "required"},
		{"components", "schemas", "Person", "properties", "name", "example"},
		{"components", "schemas", "Person", "properties", "photoUrls", "xml"},
		{"paths", "/testAdditionalPropertiesArray", "get", "responses", "200", "content", "application/json", "schema", "additionalProperties"},
	}
	validateKeys(t, expectedMessageKeys, messages)
}

func validateKeys(t *testing.T, expectedKeys [][]string, messages []*plugins.Message) {
	if len(expectedKeys) != len(messages) {
		t.Errorf("Number of messages from GrpcChecker does not match expected number")
		return
	}
	for i, msg := range messages {
		for j, k := range msg.Keys {
			if k != expectedKeys[i][j] {
				t.Errorf("Key does not match expected key text: %s != %s", expectedKeys[i], msg.Keys)
			}
		}
	}
}
