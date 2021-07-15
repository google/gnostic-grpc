package incompatibility

import (
	"reflect"
)

type TraverseError struct {
	desrciptorString string
}

func (e TraverseError) Error() string {
	return e.desrciptorString
}

// Structure for generic single-input single-output function with error indication
type GenericOperation struct {
	op         func(interface{}) (interface{}, error)
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
func makeGenericOperation(function interface{}) (GenericOperation, error) {
	switch inputType := reflect.TypeOf(function); inputType.Kind() {
	case reflect.Func:
		// Continue function
		validFunction := inputType.NumIn() == 1 && inputType.NumOut() == 1
		if validFunction {
			funcInputT := inputType.In(0)
			funcOutputT := inputType.Out(0)
			formattedGenericFunction := func(input interface{}) (interface{}, error) {
				switch t := reflect.TypeOf(input); t {
				case funcInputT:
					functionAsValue := reflect.ValueOf(function)
					returnAsInterface := functionAsValue.Call([]reflect.Value{reflect.ValueOf(input)})[0].Interface()
					returnType := reflect.TypeOf(returnAsInterface)
					if returnType != funcOutputT {
						return nil, TraverseError{"Bad Return Type in Invocation"}
					}
					return returnAsInterface, nil
				default:
					return nil, TraverseError{"Bad Input Type in Inovocation"}
				}
			}
			return GenericOperation{
				op:         formattedGenericFunction,
				inputType:  funcInputT,
				outputType: funcOutputT,
			}, nil
		}
		goto failGen
	default:
		// Invalid input
		goto failGen
	}
failGen:
	return GenericOperation{}, TraverseError{"input is not a single-input single-output function"}
}

// function should be a single-input single-output function whose output type is *IncompatibilityReport
// performs error checking for these requirements and casts into generic type
func makeGenericIncompatibilityReportFunc(function interface{}) (func(interface{}) (*IncompatibilityReport, error), error) {
	genGeneric, err := makeGenericOperation(function)
	badType := genGeneric.outputType != reflect.TypeOf(&IncompatibilityReport{})
	if err != nil || badType {
		return nil, err
	} else {
		return func(i interface{}) (*IncompatibilityReport, error) {
			res, err := genGeneric.op(i)
			return res.(*IncompatibilityReport), err
		}, nil
	}
}

// ComponentType - component type in OpenAPIDoc, i.e. Parameters, Components, ...
// ParentTraverse - lambda to tranverse from parent to component of type ComponentType
// OperationOnComponent - lambda which looks at current layer for incompatibilities
// ChildPaths - paths to check for deeper incompatibilities
type PathOperation struct {
	ComponentType        reflect.Type
	ParentTraverse       GenericOperation
	OperationOnComponent func(currentComponent interface{}) (*IncompatibilityReport, error)
	ChildPaths           []*PathOperation
}

// Merge Two Path Operations under the conditions:
// 1) Both Operations happen at same component
//   - Thus compose a new OperationOnComponent
// 2) p2's parent type is the same as p1's component type
//   - add p2 to p1 child paths
func (p1 PathOperation) compose(p2 PathOperation) (PathOperation, error) {
	sameComponentType := p1.ComponentType == p2.ComponentType
	sameTraversalType := p1.ParentTraverse.typeEquality(p2.ParentTraverse) == SameType
	if sameComponentType && sameTraversalType {
		// Assume operations take place along same component in path, thus merge
		var composedComponentOperation = func(currentComponent interface{}) (*IncompatibilityReport, error) {
			p1Report, err := p1.OperationOnComponent(currentComponent)
			if err != nil {
				return &IncompatibilityReport{}, err
			}
			p2Report, err := p2.OperationOnComponent(currentComponent)
			if err != nil {
				return &IncompatibilityReport{}, err
			}
			return &IncompatibilityReport{
				Incompatibilities: append(p1Report.GetIncompatibilities(), p2Report.GetIncompatibilities()...)}, nil
		}
		return PathOperation{
			ComponentType:        p1.ComponentType,
			ParentTraverse:       p1.ParentTraverse,
			OperationOnComponent: composedComponentOperation,
			ChildPaths:           append(p1.ChildPaths, p2.ChildPaths...),
		}, nil
	}
	//Try adding p2 to p1 children
	p2parentType := p2.ParentTraverse.inputType
	parent2childtraverse := p1.ComponentType == p2parentType
	if parent2childtraverse { //
		p1.ChildPaths = append(p1.ChildPaths, &p2)
		return p1, nil
	}
	return PathOperation{}, TraverseError{"Could not merge Path Operations"}
}

// compile Path Operations, return incompatibility report from paths
func (p PathOperation) compile(currentLevel interface{}) (*IncompatibilityReport, error) {
	var incomps []*Incompatibility = make([]*Incompatibility, 0)
	for _, childPath := range p.ChildPaths {
		childLevel, err := childPath.ParentTraverse.op(currentLevel)
		if err != nil {
			return &IncompatibilityReport{}, err
		}
		childReport, err := childPath.OperationOnComponent(childLevel)
		if err != nil {
			return &IncompatibilityReport{}, err
		}
		incomps = append(incomps, childReport.GetIncompatibilities()...)
	}
	incompatibilityReport, err := p.OperationOnComponent(currentLevel)
	if err != nil {
		return &IncompatibilityReport{}, err
	}
	incomps = append(incomps, incompatibilityReport.GetIncompatibilities()...)
	return &IncompatibilityReport{Incompatibilities: incomps}, nil
}
