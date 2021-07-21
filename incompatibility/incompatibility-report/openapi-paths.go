package incompatibility

import openapiv3 "github.com/googleapis/gnostic/openapiv3"

type OpenAPIPathRule struct {
	path     *openapiv3.PathItem
	ruleFunc func(*openapiv3.PathItem) *Incompatibility
}

func NewParamRule(scanner func(*openapiv3.PathItem) *Incompatibility) OpenAPIPathRule {
	return OpenAPIPathRule{ruleFunc: scanner}
}

func (paramRule *OpenAPIPathRule) SetInputDoc(path *openapiv3.PathItem) {
	paramRule.path = path
}

func (paramRule OpenAPIPathRule) ObjectKind() ObjectID {
	return Parameters
}

func (paramRule OpenAPIPathRule) ScanIncompatibility() *Incompatibility {
	return paramRule.ruleFunc(paramRule.path)
}

type ParamTraverse struct {
	rules []OpenAPIPathRule
	paths []OpenAPITraversal
}

func NewParamTraverse(paths ...OpenAPITraversal) DocumentTraverse {
	return DocumentTraverse{paths: paths}
}

func (p ParamTraverse) acceptIncompatibilityRule(incompatibilityRules ...OpenAPIIncompatibilityRule) {
	for _, rule := range incompatibilityRules {
		if rule.ObjectKind() == Path {
			p.rules = append(p.rules, rule.(OpenAPIPathRule))
		}
	}
}

func (p ParamTraverse) traverse(i interface{}) []*Incompatibility {
	var incompatibilities []*Incompatibility
	switch i := i.(type) {
	case *openapiv3.Document:
		// At document level, move down to paths
		for _, path := range i.GetPaths().GetPath() {
			incompatibilities = append(incompatibilities, p.traverseTyped(path.Value)...)
		}
	case *openapiv3.PathItem:
		incompatibilities = append(incompatibilities, p.traverseTyped(i)...)

	}
	return incompatibilities
}

func (p ParamTraverse) traverseTyped(path *openapiv3.PathItem) []*Incompatibility {
	var incompatibilities []*Incompatibility
	//Scan for incompatibilities
	for _, rule := range p.rules {
		rule.SetInputDoc(path)
		incompatibilities = append(incompatibilities, rule.ScanIncompatibility())
	}
	// traverse document
	for _, childTraverse := range p.paths {
		incompatibilities = append(incompatibilities, childTraverse.traverse(path)...)
	}
	return incompatibilities
}
