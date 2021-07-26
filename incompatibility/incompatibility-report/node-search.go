package incompatibility

import (
	"errors"
	"strconv"

	"gopkg.in/yaml.v3"
)

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

	//Look for key value
	for i := 0; i < len(node.Content)-1; i += 2 {
		key := node.Content[i]
		val := node.Content[i+1]
		if key.Value != path[0] {
			continue
		}
		if len(path) == 1 {
			return key, nil
		}
		return findNode(val, path[1:]...)
	}

	return nil, errors.New("unable to find yaml node")
}

// func checkValue(node *yaml.Node, strCheck string) bool {

// }
