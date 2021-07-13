package traverse

import (
	"reflect"

	"github.com/googleapis/gnostic-grpc/incompatibility/incompatibility-report"
)

// ComponentType - component type in OpenAPIDoc, i.e. Parameters, Components, ...
// ParentTraverse - lambda to tranverse from parent to component of type ComponentType
// OperationOnComponent - lambda which looks at current layer for incompatibilities
// ChildPaths - paths to check for deeper incompatibilities
type PathOperation struct {
	ComponentType        reflect.Type
	ParentTraverse       func(parentComponent interface{}) (interface{}, int)
	OperationOnComponent func(currentComponent interface{}) (*incompatibility.IncompatibilityReport, int)
	ChildPaths           []*PathOperation
}

// Merge Two Path Operations under the conditions:
// 1) Both Operations happen at same component
//   - Thus compose a new OperationOnComponent
// 2) p2's parent type is the same as p1's component type
// 	 - add p2 to p1 child paths
func (p1 PathOperation) compose(p2 PathOperation) (PathOperation, int) {
	sameComponentType := p1.ComponentType == p2.ComponentType
	sameTraversalType := reflect.TypeOf(p1.ParentTraverse) == reflect.TypeOf(p2.ParentTraverse)
	if sameComponentType && sameTraversalType {
		// Assume operations take place along same component in path, thus merge
		var composedComponentOperation = func(currentComponent interface{}) (*incompatibility.IncompatibilityReport, int) {
			p1Report, err := p1.OperationOnComponent(currentComponent)
			if err == 1 {
				return &incompatibility.IncompatibilityReport{}, 1
			}
			p2Report, err := p2.OperationOnComponent(currentComponent)
			if err == 1 {
				return &incompatibility.IncompatibilityReport{}, 1
			}
			return &incompatibility.IncompatibilityReport{
				Incompatibilities: append(p1Report.GetIncompatibilities(), p2Report.GetIncompatibilities()...)}, 0
		}
		return PathOperation{
			ComponentType:        p1.ComponentType,
			ParentTraverse:       p1.ParentTraverse,
			OperationOnComponent: composedComponentOperation,
			ChildPaths:           append(p1.ChildPaths, p2.ChildPaths...),
		}, 0
	}
	//Try adding p2 to p1 children
	p2parentType := reflect.TypeOf(p2.ParentTraverse).In(0)
	parent2childtraverse := p1.ComponentType == p2parentType
	if parent2childtraverse { //
		p1.ChildPaths = append(p1.ChildPaths, &p2)
		return p1, 0
	}
	return PathOperation{}, 1
}

// Compile Path Operations, return incompatibility report from paths
func (p PathOperation) compile(currentLevel interface{}) (*incompatibility.IncompatibilityReport, int) {
	var incomps []*incompatibility.Incompatibility = make([]*incompatibility.Incompatibility, 0)
	for _, childPath := range p.ChildPaths {
		childLevel, err := childPath.ParentTraverse(currentLevel)
		if err == 1 {
			return &incompatibility.IncompatibilityReport{}, 1
		}
		childReport, err := childPath.OperationOnComponent(childLevel)
		if err == 1 {
			return &incompatibility.IncompatibilityReport{}, 1
		}
		incomps = append(incomps, childReport.GetIncompatibilities()...)
	}
	return &incompatibility.IncompatibilityReport{Incompatibilities: incomps}, 0
}
