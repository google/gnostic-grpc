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
