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
	"fmt"
	"io/ioutil"
	"strconv"

	"gopkg.in/yaml.v3"
)

type keyValue struct {
	key   *yaml.Node
	value *yaml.Node
}

type Seeking int

// Indicates to search what component of key-value pairing findcomponent should return
const (
	KEY = iota
	VALUE
)

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

func FindKey(node *yaml.Node, path ...string) (*yaml.Node, error) {
	return findComponent(node, KEY, path...)
}

// Given a path in a yaml file return the node at the end of the path
func findComponent(node *yaml.Node, seeking Seeking, path ...string) (*yaml.Node, error) {
	if len(path) == 0 {
		return node, nil
	}
	// Sequence Index
	if node.Kind == yaml.SequenceNode {
		ind, err := strconv.Atoi(path[0])
		if err != nil || ind >= len(node.Content) {
			return nil, errors.New("invalid index parsed")
		}
		return findComponent(node.Content[ind], seeking, path[1:]...)
	}

	//Look for matching key
	if foundKeyVal, ok := mapKeyValuePairs(node.Content)[path[0]]; ok {
		return resolveMatchingPath(foundKeyVal.key, foundKeyVal.value, seeking, path...)
	}
	return nil, fmt.Errorf("unable to find yaml node %s in %s", path[0], node.Value)
}

func mapKeyValuePairs(content []*yaml.Node) map[string]keyValue {
	var keyNodePair map[string]keyValue = make(map[string]keyValue)
	for i := 0; i < len(content)-1; i += 2 {
		ky, val := keyValuePairFromContent(content, i)
		keyNodePair[ky.Value] = keyValue{key: ky, value: val}
	}
	return keyNodePair
}

// Helper in which path aligns with matching key node and needs
// further to determine further traversal
func resolveMatchingPath(key *yaml.Node, val *yaml.Node, seeking Seeking, path ...string) (*yaml.Node, error) {
	if len(path) == 1 {
		switch seeking {
		case VALUE:
			return val, nil
		default:
			return key, nil
		}
	}
	return findComponent(val, seeking, path[1:]...)
}

// Get key and value nodes from content array, bounds chechking should be
// done before function call
func keyValuePairFromContent(content []*yaml.Node, index int) (*yaml.Node, *yaml.Node) {
	key := content[index]
	val := content[index+1]
	return key, val
}
