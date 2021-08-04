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
	"fmt"
	"io/ioutil"
	"strconv"

	"gopkg.in/yaml.v3"
)

type KeyValue struct {
	key   *yaml.Node
	value *yaml.Node
}

func newPair(ky, val *yaml.Node) *KeyValue {
	return &KeyValue{
		key:   ky,
		value: val,
	}
}

// Parses filePath and attempts to create Node Object
func MakeNode(filePath string) (*yaml.Node, error) {
	var node yaml.Node
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	marshErr := yaml.Unmarshal(data, &node)
	return &node, marshErr
}

// Returns the fileposition of the key at the end of path
func FindKey(node *yaml.Node, path ...string) (int, int, error) {
	var line, col int
	node, searchErr := findComponent(node, path...)
	if searchErr != nil {
		return line, col, searchErr
	}
	line = node.Line
	col = node.Column
	return line, col, nil
}

// Given a path in a yaml file return a pairing of nodes at the end of the path
func findComponent(node *yaml.Node, path ...string) (*yaml.Node, error) {
	if len(path) == 0 {
		return node, nil
	}
	// Sequence Index
	if node.Kind == yaml.SequenceNode {
		ind, err := strconv.Atoi(path[0])
		if err != nil || ind >= len(node.Content) || ind < 0 {
			return nil, fmt.Errorf("invalid index parsed %s", path[0])
		}
		return findComponent(node.Content[ind], path[1:]...)
	}

	//Look for matching key
	if foundKeyVal, ok := mapKeyValuePairs(node.Content)[path[0]]; ok {
		return resolveMatchingPath(foundKeyVal, path...)
	}

	return nil, fmt.Errorf("unable to find yaml node %s", path[0])
}

func mapKeyValuePairs(content []*yaml.Node) map[string]*KeyValue {
	var keyNodePair map[string]*KeyValue = make(map[string]*KeyValue)
	for i := 0; i < len(content)-1; i += 2 {
		ky, val := keyValuePairFromContent(content, i)
		keyNodePair[ky.Value] = newPair(ky, val)
	}
	return keyNodePair
}

// Helper in which path aligns with matching key node and needs
// to determine further traversal
func resolveMatchingPath(keyVal *KeyValue, path ...string) (*yaml.Node, error) {
	if len(path) == 1 { // Key last item in path so return key
		return keyVal.key, nil
	}
	return findComponent(keyVal.value, path[1:]...)
}

// Get key and value nodes from content array, bounds chechking should be
// done before function call
func keyValuePairFromContent(content []*yaml.Node, index int) (*yaml.Node, *yaml.Node) {
	key := content[index]
	val := content[index+1]
	return key, val
}
