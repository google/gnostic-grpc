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
	"errors"
	"strconv"

	"gopkg.in/yaml.v3"
)

type keyValue struct {
	key   *yaml.Node
	value *yaml.Node
}

// Given a path in a yaml file return the node at the end of the path
func findNode(node *yaml.Node, path ...string) (*yaml.Node, error) {
	if len(path) == 0 {
		return node, nil
	}
	// Sequence Index
	if node.Kind == yaml.SequenceNode {
		ind, err := strconv.Atoi(path[0])
		if err != nil || ind >= len(node.Content) {
			return nil, errors.New("invalid index parsed")
		}
		return findNode(node.Content[ind], path[1:]...)
	}

	//Look for matching key
	if foundKeyVal, ok := MapKeyValuePairs(node.Content)[path[0]]; ok {
		return resolveMatchingPath(foundKeyVal.key, foundKeyVal.value, path...)
	}
	return nil, errors.New("unable to find yaml node")
}

func MapKeyValuePairs(content []*yaml.Node) map[string]keyValue {
	var keyNodePair map[string]keyValue = make(map[string]keyValue)
	for i := 0; i < len(content)-1; i += 2 {
		ky, val := keyValuePairFromContent(content, i)
		keyNodePair[ky.Value] = keyValue{key: ky, value: val}
	}
	return keyNodePair
}

// Helper in which path aligns with matching key node and needs
// further to determine further traversal
func resolveMatchingPath(key *yaml.Node, val *yaml.Node, path ...string) (*yaml.Node, error) {
	if len(path) == 1 {
		return key, nil
	}
	return findNode(val, path[1:]...)
}

// Get key and value nodes from content array, bounds chechking should be
// done before function call
func keyValuePairFromContent(content []*yaml.Node, index int) (*yaml.Node, *yaml.Node) {
	key := content[index]
	val := content[index+1]
	return key, val
}
