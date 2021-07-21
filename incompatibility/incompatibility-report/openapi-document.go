package incompatibility

import openapiv3 "github.com/googleapis/gnostic/openapiv3"

type OpenAPIDocumentRule struct {
	doc      *openapiv3.Document
	ruleFunc func(*openapiv3.Document) *Incompatibility
}

func NewDocRule(scanner func(*openapiv3.Document) *Incompatibility) OpenAPIDocumentRule {
	return OpenAPIDocumentRule{ruleFunc: scanner}
}

func (docRules *OpenAPIDocumentRule) SetInputDoc(doc *openapiv3.Document) {
	docRules.doc = doc
}

func (docRules OpenAPIDocumentRule) ObjectKind() ObjectID {
	return Document
}

func (docRules OpenAPIDocumentRule) ScanIncompatibility() *Incompatibility {
	return docRules.ruleFunc(docRules.doc)
}

type DocumentTraverse struct {
	rules []OpenAPIDocumentRule
	paths []OpenAPITraversal
}

func NewDocTraverse(paths ...OpenAPITraversal) DocumentTraverse {
	return DocumentTraverse{paths: paths}
}

func (d DocumentTraverse) acceptIncompatibilityRule(incompatibilityRules ...OpenAPIIncompatibilityRule) {
	for _, rule := range incompatibilityRules {
		if rule.ObjectKind() == Document {
			d.rules = append(d.rules, rule.(OpenAPIDocumentRule))
		}
	}
}

func (d DocumentTraverse) traverse(i interface{}) []*Incompatibility {
	var incompatibilities []*Incompatibility
	doc, conversion := i.(*openapiv3.Document)
	if conversion {
		//Scan for incompatibilities
		for _, rule := range d.rules {
			rule.SetInputDoc(doc)
			incompatibilities = append(incompatibilities, rule.ScanIncompatibility())
		}
		// traverse document
		for _, childTraverse := range d.paths {
			incompatibilities = append(incompatibilities, childTraverse.traverse(doc)...)
		}
	}
	return incompatibilities
}
