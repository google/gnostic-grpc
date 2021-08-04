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
	"testing"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v3"
)

type filePosition struct {
	Line int
	Col  int
}

func parseFile(t *testing.T, filePath string) *yaml.Node {
	baseNode, parseErr := MakeNode(filePath)
	if parseErr != nil {
		t.Fatalf(parseErr.Error())
	}
	return baseNode.Content[0]
}

func searchKey(t *testing.T, baseNode *yaml.Node, path ...string) (line, col int) {
	line, col, searchError := FindKey(baseNode, path...)
	if searchError != nil {
		t.Fatalf(searchError.Error())
	}
	return line, col
}

// Assert keysearch results in correct file position
func TestFindKey(t *testing.T) {

	shallowYaml := parseFile(t, "node-examples/yaml1.yaml")
	deepYaml := parseFile(t, "node-examples/yaml2.yaml")
	shallowJson := parseFile(t, "node-examples/json1.json")
	deepJson := parseFile(t, "node-examples/json2.json")

	var keyTest = []struct {
		testName             string
		baseNode             *yaml.Node
		path                 []string
		expectedFilePosition filePosition
	}{
		{
			"shallowYaml1",
			shallowYaml,
			[]string{"putting"},
			filePosition{1, 1},
		},
		{
			"shallowYaml2",
			shallowYaml,
			[]string{"trade", "poetry"},
			filePosition{3, 3},
		},
		{
			"deepYaml1",
			deepYaml,
			[]string{"0", "1", "neighbor", "1", "slightly", "group", "1", "development"},
			filePosition{54, 15},
		},
		{
			"deepYaml2",
			deepYaml,
			[]string{"0", "1", "neighbor", "1", "way"},
			filePosition{4, 9},
		},
		{
			"shallowjson1",
			shallowJson,
			[]string{"copper"},
			filePosition{2, 5},
		},
		{
			"deepjson2",
			deepJson,
			[]string{"0", "1", "0", "whole", "2", "1", "mistake"},
			filePosition{12, 17},
		},
	}
	for _, trial := range keyTest {
		t.Run(trial.testName, func(tt *testing.T) {
			foundLine, foundCol := searchKey(tt, trial.baseNode, trial.path...)
			foundFilePos := filePosition{
				Line: foundLine,
				Col:  foundCol,
			}
			diff := cmp.Diff(foundFilePos,
				trial.expectedFilePosition)
			if diff != "" {
				tt.Error("UnexpectedFilePosition: diff(-want +got):\n", diff)
			}
		})
	}
}

// Assert correct error reporting in FindKey
func TestFindKeyErrors(t *testing.T) {

	shallowYaml := parseFile(t, "node-examples/yaml1.yaml")
	deepYaml := parseFile(t, "node-examples/yaml2.yaml")
	shallowJson := parseFile(t, "node-examples/json1.json")
	deepJson := parseFile(t, "node-examples/json2.json")

	var keyTest = []struct {
		testName            string
		baseNode            *yaml.Node
		path                []string
		expectedErrorString string
	}{
		{
			"ExhaustiveSearch",
			shallowYaml,
			[]string{"putting", "invalidKey1"},
			"unable to find yaml node invalidKey1",
		},
		{
			"ExhaustiveSearch2",
			deepJson,
			[]string{"0", "1", "0", "whole", "2", "3", "invalidKey2"},
			"unable to find yaml node invalidKey2",
		},
		{
			"InvalidIndex",
			shallowJson,
			[]string{"corn", "worth", "7"},
			"invalid index parsed 7",
		},
		{
			"InvalidIndex2",
			deepYaml,
			[]string{"0", "1", "neighbor", "1", "slightly", "group", "1", "development", "search", "-1"},
			"invalid index parsed -1",
		},
	}
	for _, trial := range keyTest {
		t.Run(trial.testName, func(tt *testing.T) {
			_, _, err := FindKey(trial.baseNode, trial.path...)
			diff := cmp.Diff(trial.expectedErrorString, err.Error())
			if diff != "" {
				tt.Error("UnexpectedFilePosition: diff(-want +got):\n", diff)
			}
		})
	}
}
