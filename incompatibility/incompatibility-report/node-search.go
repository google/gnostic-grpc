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
	for _, keyValue := range groupKeyValuePairs(node.Content) {
		key := keyValue.key
		val := keyValue.value
		if key.Value != path[0] {
			continue
		}
		return resolveMatchingPath(key, val, path...)
	}
	return nil, errors.New("unable to find yaml node")
}

func groupKeyValuePairs(content []*yaml.Node) []keyValue {
	var keyValuePairs []keyValue
	for i := 0; i < len(content)-1; i += 2 {
		ky, val := keyValuePairFromContent(content, i)
		keyValuePairs = append(keyValuePairs, keyValue{key: ky, value: val})
	}
	return keyValuePairs
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
