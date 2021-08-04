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

type Component struct {
	key   *yaml.Node
	value *yaml.Node
}

func (comp *Component) getKey() (*yaml.Node, error) {
	if comp.key == nil {
		return nil, errors.New("invalid key")
	}
	return comp.key, nil
}

func (comp *Component) getValue() (*yaml.Node, error) {
	if comp.value == nil {
		return nil, errors.New("invalid value")
	}
	return comp.value, nil
}

func newComponent(ky, val *yaml.Node) *Component {
	return &Component{
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

// Returns the key at the end of the search, valid for paths ending in key-value mappings
func FindKey(node *yaml.Node, path ...string) (*yaml.Node, error) {
	comp, err := findComponent(node, path...)
	if err != nil {
		return nil, err
	}
	return comp.getKey()
}

// Returns the value at the end of the search, valid for paths ending in
// sequence items, single objects, values of ke
func FindValue(node *yaml.Node, path ...string) (*yaml.Node, error) {
	comp, err := findComponent(node, path...)
	if err != nil {
		return nil, err
	}
	return comp.getValue()
}

// Given a path in a yaml file return a pairing of nodes at the end of the path
func findComponent(node *yaml.Node, path ...string) (*Component, error) {
	if len(path) == 0 {
		return newComponent(nil, node), nil
	}
	// Sequence Index
	if node.Kind == yaml.SequenceNode {
		ind, err := strconv.Atoi(path[0])
		if err != nil || ind >= len(node.Content) {
			return nil, errors.New("invalid index parsed")
		}
		return findComponent(node.Content[ind], path[1:]...)
	}

	//Look for matching key
	if foundKeyVal, ok := mapKeyValuePairs(node.Content)[path[0]]; ok {
		return resolveMatchingPath(foundKeyVal, path...)
	}
	return nil, fmt.Errorf("unable to find yaml node %s in %s", path[0], node.Value)
}

func mapKeyValuePairs(content []*yaml.Node) map[string]*Component {
	var keyNodePair map[string]*Component = make(map[string]*Component)
	for i := 0; i < len(content)-1; i += 2 {
		ky, val := keyValuePairFromContent(content, i)
		keyNodePair[ky.Value] = newComponent(ky, val)
	}
	return keyNodePair
}

// Helper in which path aligns with matching key node and needs
// further to determine further traversal
func resolveMatchingPath(keyVal *Component, path ...string) (*Component, error) {
	if len(path) == 1 {
		return keyVal, nil
	}
	val, valExistsError := keyVal.getValue()
	if valExistsError != nil {
		return nil, valExistsError
	}
	return findComponent(val, path[1:]...)
}

// Get key and value nodes from content array, bounds chechking should be
// done before function call
func keyValuePairFromContent(content []*yaml.Node, index int) (*yaml.Node, *yaml.Node) {
	key := content[index]
	val := content[index+1]
	return key, val
}
