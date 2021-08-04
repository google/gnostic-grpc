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

package search

import (
	"errors"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/googleapis/gnostic-grpc/utils"
	"gopkg.in/yaml.v3"
)

func parseFile(t *testing.T, filePath string) *yaml.Node {
	baseNode, parseErr := MakeNode(filePath)
	if parseErr != nil {
		t.Fatalf(parseErr.Error())
	}
	return baseNode.Content[0]
}

func searchKey(t *testing.T, baseNode *yaml.Node, path ...string) *yaml.Node {
	node, searchError := FindKey(baseNode, path...)
	if searchError != nil {
		t.Fatalf(searchError.Error())
	}
	return node
}

func searchValue(t *testing.T, baseNode *yaml.Node, path ...string) *yaml.Node {
	node, searchError := FindValue(baseNode, path...)
	if searchError != nil {
		t.Fatalf(searchError.Error())
	}
	return node
}

func validateIndex(sequenceNode *yaml.Node, ind int) error {
	if sequenceNode.Kind != yaml.SequenceNode {
		println(sequenceNode.Value)
		return errors.New("not a sequence node")
	}
	if ind < 0 || ind >= len(sequenceNode.Content) {
		return errors.New("not in sequence bounds")
	}
	return nil
}

// Assert that from key value pairs, keys are returned
func TestKeySearch(t *testing.T) {

	shallowYaml := parseFile(t, "node-examples/yaml1.yaml")
	deepYaml := parseFile(t, "node-examples/yaml2.yaml")
	shallowJson := parseFile(t, "node-examples/json1.json")
	deepJson := parseFile(t, "node-examples/json2.json")

	var keyTest = []struct {
		testName    string
		baseNode    *yaml.Node
		path        []string
		expectedKey string
	}{
		{
			"shallowYaml1",
			shallowYaml,
			[]string{"putting"},
			"putting",
		},
		{
			"shallowYaml2",
			shallowYaml,
			[]string{"trade", "poetry"},
			"poetry",
		},
		{
			"deepYaml1",
			deepYaml,
			[]string{"0", "1", "neighbor", "1", "slightly", "group", "1", "development"},
			"development",
		},
		{
			"deepYaml2",
			deepYaml,
			[]string{"0", "1", "neighbor", "1", "way"},
			"way",
		},
		{
			"shallowjson1",
			shallowJson,
			[]string{"copper"},
			"copper",
		},
		{
			"deepjson2",
			deepJson,
			[]string{"0", "1", "0", "whole", "2", "1", "mistake"},
			"mistake",
		},
	}
	for _, trial := range keyTest {
		t.Run(trial.testName, func(tt *testing.T) {
			foundNode := searchKey(tt, trial.baseNode, trial.path...)
			if trial.expectedKey != foundNode.Value {
				tt.Errorf("foundNode != lastToken, wanted: %s, got: %s", trial.expectedKey, foundNode.Value)
			}
		})
	}
}

// Assert values
func TestValueSearch(t *testing.T) {
	shallowYaml := parseFile(t, "node-examples/yaml1.yaml")
	deepYaml := parseFile(t, "node-examples/yaml2.yaml")
	shallowJson := parseFile(t, "node-examples/json1.json")
	deepJson := parseFile(t, "node-examples/json2.json")

	var valueTest = []struct {
		testName      string
		baseNode      *yaml.Node
		path          []string
		expectedValue string
	}{
		{
			"shallowYaml1",
			shallowYaml,
			[]string{"putting"},
			"false",
		},
		{
			"shallowYaml2",
			shallowYaml,
			[]string{"trade", "poetry"},
			"below",
		},
		{
			"deepYaml1",
			deepYaml,
			[]string{"0", "1", "neighbor", "1", "slightly", "group", "1", "view"},
			"gate",
		},
		{
			"deepYaml2",
			deepYaml,
			[]string{"0", "1", "neighbor", "0"},
			"1577599595",
		},
		{
			"shallowjson1",
			shallowJson,
			[]string{"copper"},
			"false",
		},
		{
			"deepjson2",
			deepJson,
			[]string{"0", "1", "0", "whole", "2", "1", "mistake"},
			"-1611512731",
		},
	}
	for _, trial := range valueTest {
		t.Run(trial.testName, func(tt *testing.T) {
			foundNode := searchKey(tt, trial.baseNode, trial.path...)
			if trial.expectedValue != foundNode.Value {
				tt.Errorf("foundNode != lastToken, wanted: %s, got: %s", trial.expectedValue, foundNode.Value)
			}
		})
	}
}

// Assert that paths ending in sequence indexes are valid
// in the sequence nodes
func TestSequenceValidation(t *testing.T) {
	shallowJson := parseFile(t, "node-examples/json1.json")
	deepJson := parseFile(t, "node-examples/json2.json")

	var pathTest = []struct {
		testName             string
		baseNode             *yaml.Node
		pathToSequence       []string
		knownLengthOfSequnce int
	}{
		{
			"shallowJson",
			shallowJson,
			[]string{"corn", "worth"},
			6,
		},
		{
			"deepJson",
			deepJson,
			[]string{"0", "1"},
			6,
		},
		{
			"deepJson2",
			deepJson,
			[]string{"0", "1", "0", "whole"},
			6,
		},
	}
	for _, trial := range pathTest {
		t.Run(trial.testName, func(tt *testing.T) {
			foundSequenceNode := searchValue(tt, trial.baseNode, trial.pathToSequence...)
			for i := 0; i < trial.knownLengthOfSequnce; i++ {
				//Validate Index
				if validationError := validateIndex(foundSequenceNode, i); validationError != nil {
					tt.Error(validationError.Error())
				}

				//Check expected sequence entry vs recomputed sequence entry
				expectedSequenceEntry := foundSequenceNode.Content[i]
				pathToSequenceEntry := utils.ExtendPath(trial.pathToSequence, strconv.Itoa(i))
				foundSequenceEntry := searchValue(t, trial.baseNode, pathToSequenceEntry...)

				diff := cmp.Diff(expectedSequenceEntry, foundSequenceEntry)
				if diff != "" {
					tt.Error("expectedSequenceEntry != reComputedSequenceEntry: diff(-want +got):\n", diff)
				}
			}
			if err := validateIndex(foundSequenceNode, trial.knownLengthOfSequnce); err.Error() != "not in sequence bounds" {
				tt.Errorf("Expected out of bounds error, got %v", err)
			}
		})
	}
}
