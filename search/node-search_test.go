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
)

func performSearch(t *testing.T, filePath string, yamlPaths ...[]string) {
	baseNode, parseNode := MakeNode(filePath)
	if parseNode != nil {
		t.Fatalf(parseNode.Error())
	}
	for _, yamlPath := range yamlPaths {
		_, searchError := FindNode(baseNode.Content[0], yamlPath...)
		if searchError != nil {
			t.Errorf(searchError.Error())
		}
	}
}

// Assert that known paths are reported in Search
func TestHardcodedPaths(t *testing.T) {
	var pathTest = []struct {
		testName  string
		filePath  string
		yamlPaths [][]string
	}{
		{
			"shallowYaml",
			"node-examples/yaml1.yaml",
			[][]string{
				{"putting"},
				{"trade", "poetry"},
				{"addition"},
			},
		},
		{
			"deepYaml",
			"node-examples/yaml2.yaml",
			[][]string{
				{"2", "government"},
				{"0", "1", "neighbor", "1", "way"},
				{"0", "1", "neighbor", "1", "slightly", "group", "1", "development"},
			},
		},
		{
			"shallowjson",
			"node-examples/json1.json",
			[][]string{
				{"copper"},
				{"corn", "worth", "0"},
			},
		},
		{
			"deepjson",
			"node-examples/json2.json",
			[][]string{
				{"0", "1", "0", "whole"},
				{"0", "1", "0", "whole", "2", "1", "mistake"},
			},
		},
	}
	for _, trial := range pathTest {
		t.Run(trial.testName, func(tt *testing.T) {
			performSearch(tt, trial.filePath, trial.yamlPaths...)
		})
	}
}
