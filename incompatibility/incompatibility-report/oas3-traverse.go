package incompatibility

import (
	"reflect"
)

// Structure for generic single-input single-output function with error indication
type GenericOperation struct {
	op         func(interface{}) (interface{}, int)
	inputType  reflect.Type
	outputType reflect.Type
}

type transverseComp int

const (
	DiffType = iota
	SameType
)

func (op1 GenericOperation) typeEquality(op2 GenericOperation) transverseComp {
	if op1.inputType == op2.inputType && op1.outputType == op2.outputType {
		return SameType
	}
	return DiffType
}

// arg function should be a single-input single-output function
// creates generic wrapper around function, indicates error on return
func makeGenericOperation(function interface{}) (GenericOperation, int) {
	switch inputType := reflect.TypeOf(function); inputType.Kind() {
	case reflect.Func:
		// Continue function
		validFunction := inputType.NumIn() == 1 && inputType.NumOut() == 1
		if validFunction {
			funcInputT := inputType.In(0)
			funcOutputT := inputType.Out(0)
			formattedGenericFunction := func(input interface{}) (interface{}, int) {
				switch t := reflect.TypeOf(input); t {
				case funcInputT:
					functionAsValue := reflect.ValueOf(function)
					returnAsInterface := functionAsValue.Call([]reflect.Value{reflect.ValueOf(input)})[0].Interface()
					returnType := reflect.TypeOf(returnAsInterface)
					if returnType != funcOutputT {
						return nil, 1
					}
					return returnAsInterface, 0
				default:
					return nil, 1
				}
			}
			return GenericOperation{
				op:         formattedGenericFunction,
				inputType:  funcInputT,
				outputType: funcOutputT,
			}, 0
		}
		goto error
	default:
		// Invalid input
		goto error
	}
error:
	return GenericOperation{}, 1
}

// function should be a single-input single-output function whose output type is *IncompatibilityReport
// performs error checking for these requirements and casts into generic type
func makeGenericIncompatibilityReportFunc(function interface{}) (func(interface{}) (*IncompatibilityReport, int), int) {
	genGeneric, err := makeGenericOperation(function)
	badType := genGeneric.outputType != reflect.TypeOf(&IncompatibilityReport{})
	if err == 1 || badType {
		return nil, 1
	} else {
		return func(i interface{}) (*IncompatibilityReport, int) {
			res, err := genGeneric.op(i)
			return res.(*IncompatibilityReport), err
		}, 0
	}
}

// ComponentType - component type in OpenAPIDoc, i.e. Parameters, Components, ...
// ParentTraverse - lambda to tranverse from parent to component of type ComponentType
// OperationOnComponent - lambda which looks at current layer for incompatibilities
// ChildPaths - paths to check for deeper incompatibilities
type PathOperation struct {
	ComponentType        reflect.Type
	ParentTraverse       GenericOperation
	OperationOnComponent func(currentComponent interface{}) (*IncompatibilityReport, int)
	ChildPaths           []*PathOperation
}

// Merge Two Path Operations under the conditions:
// 1) Both Operations happen at same component
//   - Thus compose a new OperationOnComponent
// 2) p2's parent type is the same as p1's component type
//   - add p2 to p1 child paths
func (p1 PathOperation) compose(p2 PathOperation) (PathOperation, int) {
	sameComponentType := p1.ComponentType == p2.ComponentType
	sameTraversalType := p1.ParentTraverse.typeEquality(p2.ParentTraverse) == SameType
	if sameComponentType && sameTraversalType {
		// Assume operations take place along same component in path, thus merge
		var composedComponentOperation = func(currentComponent interface{}) (*IncompatibilityReport, int) {
			p1Report, err := p1.OperationOnComponent(currentComponent)
			if err == 1 {
				return &IncompatibilityReport{}, 1
			}
			p2Report, err := p2.OperationOnComponent(currentComponent)
			if err == 1 {
				return &IncompatibilityReport{}, 1
			}
			return &IncompatibilityReport{
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
	p2parentType := p2.ParentTraverse.inputType
	parent2childtraverse := p1.ComponentType == p2parentType
	if parent2childtraverse { //
		p1.ChildPaths = append(p1.ChildPaths, &p2)
		return p1, 0
	}
	return PathOperation{}, 1
}

// compile Path Operations, return incompatibility report from paths
func (p PathOperation) compile(currentLevel interface{}) (*IncompatibilityReport, int) {
	var incomps []*Incompatibility = make([]*Incompatibility, 0)
	for _, childPath := range p.ChildPaths {
		childLevel, err := childPath.ParentTraverse.op(currentLevel)
		if err == 1 {
			return &IncompatibilityReport{}, 1
		}
		childReport, err := childPath.OperationOnComponent(childLevel)
		if err == 1 {
			return &IncompatibilityReport{}, 1
		}
		incomps = append(incomps, childReport.GetIncompatibilities()...)
	}
	incompatibilityReport, err := p.OperationOnComponent(currentLevel)
	if err == 1 {
		return &IncompatibilityReport{}, 1
	}
	incomps = append(incomps, incompatibilityReport.GetIncompatibilities()...)
	return &IncompatibilityReport{Incompatibilities: incomps}, 0
}
