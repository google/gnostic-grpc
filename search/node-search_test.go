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

	"gopkg.in/yaml.v3"
)

func performSearch(t *testing.T, filePath string, seeking Seeking, yamlPaths ...[]string) []*yaml.Node {
	var foundNodes []*yaml.Node
	baseNode, parseNode := MakeNode(filePath)
	if parseNode != nil {
		t.Fatalf(parseNode.Error())
	}
	for _, yamlPath := range yamlPaths {
		node, searchError := findComponent(baseNode.Content[0], seeking, yamlPath...)
		if searchError != nil {
			t.Errorf(searchError.Error())
		} else {
			foundNodes = append(foundNodes, node)
		}
	}
	return foundNodes
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

// returns a slice of the last string from each path, order is perserved
func getLastTokensFromPaths(yamlPaths ...[]string) []string {
	var lastStrings []string
	for _, token := range yamlPaths {
		lastStrings = append(lastStrings, token[len(token)-1])
	}
	return lastStrings
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
				{"0", "1", "0"},
				{"0", "1", "0", "whole"},
				{"0", "1", "0", "whole", "2", "1", "mistake"},
			},
		},
	}
	for _, trial := range pathTest {
		t.Run(trial.testName, func(tt *testing.T) {
			foundNodes := performSearch(tt, trial.filePath, KEY, trial.yamlPaths...)
			if len(trial.yamlPaths) != len(foundNodes) {
				tt.Errorf("len(foundNodes) != len(yamlPaths), wanted: %d got: %d", len(trial.yamlPaths), len(foundNodes))
			}
			lastTokens := getLastTokensFromPaths(trial.yamlPaths...)
			for i := 0; i < len(foundNodes); i++ {
				foundNodeKey := foundNodes[i].Value
				ExpectedPathToken := lastTokens[i]
				if ind, numberConversionErr := strconv.Atoi(ExpectedPathToken); numberConversionErr == nil {
					originalPath := trial.yamlPaths[i]
					SequenceNode := getSequenceNode(tt, trial.filePath, originalPath)
					if err := validateIndex(SequenceNode, ind); err != nil {
						tt.Errorf(err.Error())
					}
				} else {
					if foundNodeKey != ExpectedPathToken {
						tt.Errorf("foundNode != lastToken, wanted: %s, got: %s", foundNodeKey, ExpectedPathToken)
					}
				}
			}
		})
	}
}

// given a valid path to an item in a sequnce, get the parent sequence node where item is located
func getSequenceNode(t *testing.T, filePath string, pathToItemInSequence []string) *yaml.Node {
	lenToParentOfSequenceNode := len(pathToItemInSequence) - 1
	pathtoParentOfSequenceNode := pathToItemInSequence[:lenToParentOfSequenceNode]
	SequenceNode := performSearch(t, filePath, VALUE, pathtoParentOfSequenceNode)[0]
	return SequenceNode
}
