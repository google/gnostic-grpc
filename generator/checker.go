// Copyright 2019 Google Inc. All Rights Reserved.
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

package generator

import (
	openapiv3 "github.com/google/gnostic/openapiv3"
	plugins "github.com/google/gnostic/plugins"
)

type GrpcChecker struct {
	// The document to be analyzed
	document *openapiv3.Document
	// The messages that are displayed to the user with information of what is not being processed by the generator.
	messages []*plugins.Message
}

// Creates a new checker.
func NewGrpcChecker(document *openapiv3.Document) *GrpcChecker {
	return &GrpcChecker{document: document, messages: make([]*plugins.Message, 0)}
}

// Runs the checker. It is a top-down approach.
func (c *GrpcChecker) Run() []*plugins.Message {
	c.analyzeOpenAPIDocument()
	return c.messages
}

// Analyzes the root object.
func (c *GrpcChecker) analyzeOpenAPIDocument() {
	fields := getNotSupportedOpenAPIDocumentFields(c.document)
	for _, f := range fields {
		text := "Field: '" + f + "' is not supported for the OpenAPI document with title: " + c.document.Info.Title
		msg := constructInfoMessage("DOCUMENTFIELDS", text, []string{f})
		c.messages = append(c.messages, &msg)
	}
	c.analyzeComponents()
	c.analyzePaths()
}

// Analyzes the components of a OpenAPI description.
func (c *GrpcChecker) analyzeComponents() {
	components := c.document.Components
	currentKeys := []string{"components"}

	fields := getNotSupportedComponentsFields(components)
	for _, f := range fields {
		text := "Field: '" + f + "' is not supported for the component"
		msg := constructInfoMessage("COMPONENTSFIELDS", text, append(copyKeys(currentKeys), f))
		c.messages = append(c.messages, &msg)
	}

	if schemas := components.GetSchemas(); schemas != nil {
		for _, pair := range schemas.AdditionalProperties {
			parentKeys := append(currentKeys, []string{"schemas", pair.Name}...)
			c.analyzeSchema(pair.Name, pair.Value, parentKeys)
		}
	}

	if responses := components.GetResponses(); responses != nil {
		for _, pair := range responses.AdditionalProperties {
			parentKeys := append(currentKeys, "responses")
			c.analyzeResponse(pair, parentKeys)
		}
	}

	if parameters := components.GetParameters(); parameters != nil {
		for _, pair := range parameters.AdditionalProperties {
			parentKeys := append(currentKeys, "parameters")
			c.analyzeParameter(pair.Value, parentKeys)
		}
	}

	if requestBodies := components.GetRequestBodies(); requestBodies != nil {
		for _, pair := range requestBodies.AdditionalProperties {
			parentKeys := append(currentKeys, []string{"requestBodies", pair.Name}...)
			c.analyzeRequestBody(pair, parentKeys)
		}
	}
}

// Analyzes all paths.
func (c *GrpcChecker) analyzePaths() {
	currentKeys := []string{"paths"}
	for _, pathItem := range c.document.Paths.Path {
		c.analyzePathItem(pathItem, currentKeys)
	}
}

// Analyzes one single path.
func (c *GrpcChecker) analyzePathItem(pair *openapiv3.NamedPathItem, parentKeys []string) {
	pathItem := pair.Value
	currentKeys := append(parentKeys, pair.Name)

	fields := getNotSupportedPathItemFields(pathItem)
	for _, f := range fields {
		text := "Field: '" + f + "' is not supported for path: " + pair.Name
		msg := constructInfoMessage("PATHFIELDS", text, append(copyKeys(currentKeys), f))
		c.messages = append(c.messages, &msg)
	}

	operations, operationType := getValidOperations(pathItem)
	for idx, op := range operations {
		pKeys := append(currentKeys, operationType[idx])
		c.analyzeOperation(op, pKeys)
	}
}

// Analyzes a single Operation.
func (c *GrpcChecker) analyzeOperation(operation *openapiv3.Operation, parentKeys []string) {
	currentKeys := parentKeys
	fields := getNotSupportedOperationFields(operation)

	if len(operation.OperationId) == 0 {
		text := "One of your operations does not have an 'operationId'. gnostic-grpc might produce an incorrect output file."
		msg := constructWarningMessage("OPERATION", text, currentKeys)
		c.messages = append(c.messages, &msg)
	}

	for _, f := range fields {
		text := "Field: '" + f + "' is not supported for operation: " + operation.OperationId
		msg := constructInfoMessage("OPERATIONFIELDS", text, append(copyKeys(currentKeys), f))
		c.messages = append(c.messages, &msg)
	}

	for _, param := range operation.Parameters {
		pKeys := append(currentKeys, "parameters")
		c.analyzeParameter(param, pKeys)
	}

	for _, response := range operation.Responses.GetResponseOrReference() {
		pKeys := append(currentKeys, "responses")
		c.analyzeResponse(response, pKeys)
	}

	if defaultResponse := operation.Responses.Default; defaultResponse != nil {
		wrap := &openapiv3.NamedResponseOrReference{Name: "default", Value: defaultResponse}
		pKeys := append(currentKeys, "responses")
		c.analyzeResponse(wrap, pKeys)
	}

	pKeys := append(currentKeys, "requestBody")
	wrap := &openapiv3.NamedRequestBodyOrReference{Name: operation.OperationId, Value: operation.RequestBody}
	c.analyzeRequestBody(wrap, pKeys)

}

// Analyzes the parameter.
func (c *GrpcChecker) analyzeParameter(paramOrRef *openapiv3.ParameterOrReference, parentKeys []string) {
	currentKeys := parentKeys

	if parameter := paramOrRef.GetParameter(); parameter != nil {
		fields := getNotSupportedParameterFields(parameter)
		for _, f := range fields {
			text := "Field: '" + f + "' is not supported for parameter: " + parameter.Name
			msg := constructInfoMessage("PARAMETERFIELDS", text, append(copyKeys(currentKeys), f))
			c.messages = append(c.messages, &msg)
		}

		pKeys := append(currentKeys, "schema")
		c.analyzeSchema(parameter.Name, parameter.Schema, pKeys)
	}
}

// Analyzes a response.
func (c *GrpcChecker) analyzeResponse(pair *openapiv3.NamedResponseOrReference, parentKeys []string) {
	currentKeys := append(parentKeys, pair.Name)

	if response := pair.Value.GetResponse(); response != nil {
		fields := getNotSupportedResponseFields(response)
		for _, f := range fields {
			text := "Field: '" + f + "' is not supported for response: " + pair.Name
			msg := constructInfoMessage("RESPONSEFIELDS", text, append(copyKeys(currentKeys), f))
			c.messages = append(c.messages, &msg)
		}
		if content := response.Content; content != nil {
			for _, pair := range content.AdditionalProperties {
				pKeys := append(currentKeys, []string{"content", pair.Name}...)
				c.analyzeContent(pair, pKeys)
			}
		}
	}
}

// Analyzes a request body.
func (c *GrpcChecker) analyzeRequestBody(pair *openapiv3.NamedRequestBodyOrReference, parentKeys []string) {
	currentKeys := parentKeys

	if requestBody := pair.Value.GetRequestBody(); requestBody != nil {
		if requestBody.Required {
			text := "Field: 'required' is not supported for the request: " + pair.Name
			msg := constructInfoMessage("REQUESTBODYFIELDS", text, append(copyKeys(currentKeys), "required"))
			c.messages = append(c.messages, &msg)
		}
		for _, pair := range requestBody.Content.AdditionalProperties {
			pKeys := append(currentKeys, []string{"content", pair.Name}...)
			c.analyzeContent(pair, pKeys)
		}
	}
}

// Analyzes the content of a response.
func (c *GrpcChecker) analyzeContent(pair *openapiv3.NamedMediaType, parentKeys []string) {
	currentKeys := parentKeys
	mediaType := pair.Value

	fields := getNotSupportedMediaTypeFields(mediaType)
	for _, f := range fields {
		text := "Field: '" + f + "' is not supported for the mediatype: " + pair.Name
		msg := constructInfoMessage("MEDIATYPEFIELDS", text, append(copyKeys(currentKeys), f))
		c.messages = append(c.messages, &msg)
	}

	if mediaType.Schema != nil {
		pKeys := append(currentKeys, "schema")
		c.analyzeSchema(pair.Name, mediaType.Schema, pKeys)
	}
}

// Analyzes the schema.
func (c *GrpcChecker) analyzeSchema(identifier string, schemaOrReference *openapiv3.SchemaOrReference, parentKeys []string) {
	currentKeys := parentKeys

	if schema := schemaOrReference.GetSchema(); schema != nil {
		fields := getNotSupportedSchemaFields(schema)
		for _, f := range fields {
			text := "Field: '" + f + "' is not supported for the schema: " + identifier
			msg := constructInfoMessage("SCHEMAFIELDS", text, append(copyKeys(currentKeys), f))
			c.messages = append(c.messages, &msg)
		}

		// Check for this: https://github.com/LorenzHW/gnostic-grpc-deprecated/issues/3#issuecomment-509348357
		if additionalProperties := schema.AdditionalProperties; additionalProperties != nil {
			if schema := additionalProperties.GetSchemaOrReference().GetSchema(); schema != nil {
				if schema.Type == "array" {
					text := "Field: 'additionalProperties' with type array is generated as empty message inside .proto."
					msg := constructInfoMessage("SCHEMAFIELDS", text, append(copyKeys(currentKeys), "additionalProperties"))
					c.messages = append(c.messages, &msg)
				}
			}
		}

		if items := schema.Items; items != nil {
			for _, schemaOrRef := range items.SchemaOrReference {
				pKeys := append(currentKeys, "items")
				c.analyzeSchema("Items of "+identifier, schemaOrRef, pKeys)
			}
		}

		if properties := schema.Properties; properties != nil {
			for _, pair := range properties.AdditionalProperties {
				pKeys := append(currentKeys, []string{"properties", pair.Name}...)
				c.analyzeSchema(pair.Name, pair.Value, pKeys)
			}
		}

		if additionalProperties := schema.AdditionalProperties; additionalProperties != nil {
			pKeys := append(currentKeys, "additionalProperties")
			c.analyzeSchema("AdditionalProperties of "+identifier, additionalProperties.GetSchemaOrReference(), pKeys)
		}
	}
}

// constructInfoMessage Constructs a info message which will be displayed to the user on the console
func constructInfoMessage(code string, text string, keys []string) plugins.Message {
	return plugins.Message{
		Code:  code,
		Level: plugins.Message_INFO,
		Text:  text,
		Keys:  keys,
	}
}

// constructWarningMessage constructs a warning message which will be displayed to the user on the console
func constructWarningMessage(code string, text string, keys []string) plugins.Message {
	return plugins.Message{
		Code:  code,
		Level: plugins.Message_WARNING,
		Text:  text,
		Keys:  keys,
	}
}

// Returns all valid operations that will be transcoded by the plugin.
func getValidOperations(pathItem *openapiv3.PathItem) (operations []*openapiv3.Operation, operationTypes []string) {
	operations = make([]*openapiv3.Operation, 0)
	operationTypes = make([]string, 0)
	if pathItem == nil {
		return operations, operationTypes
	}

	if pathItem.Get != nil {
		operations = append(operations, pathItem.Get)
		operationTypes = append(operationTypes, "get")
	}
	if pathItem.Put != nil {
		operations = append(operations, pathItem.Put)
		operationTypes = append(operationTypes, "put")
	}
	if pathItem.Post != nil {
		operations = append(operations, pathItem.Post)
		operationTypes = append(operationTypes, "post")
	}
	if pathItem.Delete != nil {
		operations = append(operations, pathItem.Delete)
		operationTypes = append(operationTypes, "delete")
	}
	if pathItem.Patch != nil {
		operations = append(operations, pathItem.Patch)
		operationTypes = append(operationTypes, "patch")
	}
	return operations, operationTypes
}

// Returns fields that the won't be considered by the plugin for document.
func getNotSupportedOpenAPIDocumentFields(document *openapiv3.Document) []string {
	fields := make([]string, 0)
	if document == nil {
		return fields
	}

	if document.Servers != nil {
		fields = append(fields, "servers")
	}
	if document.Security != nil {
		fields = append(fields, "security")
	}
	if document.Tags != nil {
		fields = append(fields, "tags")
	}
	if document.ExternalDocs != nil {
		fields = append(fields, "externalDocs")
	}
	return fields
}

// Returns fields that the won't be considered by the plugin for parameter.
func getNotSupportedParameterFields(parameter *openapiv3.Parameter) []string {
	fields := make([]string, 0)
	if parameter == nil {
		return fields
	}
	if parameter.Required {
		fields = append(fields, "required")
	}
	if parameter.Deprecated {
		fields = append(fields, "deprecated")
	}
	if parameter.AllowEmptyValue {
		fields = append(fields, "allowEmptyValue")
	}
	if parameter.Style != "" {
		fields = append(fields, "style")
	}
	if parameter.Explode {
		fields = append(fields, "explode")
	}
	if parameter.AllowReserved {
		fields = append(fields, "allowReserved")
	}
	if parameter.Example != nil {
		fields = append(fields, "example")
	}
	if parameter.Examples != nil {
		fields = append(fields, "examples")
	}
	if parameter.Content != nil {
		fields = append(fields, "content")
	}

	return fields
}

// Returns fields that the won't be considered by the plugin for schema.
func getNotSupportedSchemaFields(schema *openapiv3.Schema) []string {
	fields := make([]string, 0)
	if schema == nil {
		return fields
	}
	if schema.Nullable {
		fields = append(fields, "nullable")
	}
	if schema.Discriminator != nil {
		fields = append(fields, "discriminator")
	}
	if schema.ReadOnly {
		fields = append(fields, "readOnly")
	}
	if schema.WriteOnly {
		fields = append(fields, "writeOnly")
	}
	if schema.Xml != nil {
		fields = append(fields, "xml")
	}
	if schema.ExternalDocs != nil {
		fields = append(fields, "externalDocs")
	}
	if schema.Example != nil {
		fields = append(fields, "example")
	}
	if schema.Deprecated {
		fields = append(fields, "deprecated")
	}
	if schema.Title != "" {
		fields = append(fields, "title")
	}
	if schema.MultipleOf != 0 {
		fields = append(fields, "multipleOf")
	}
	if schema.Maximum != 0 {
		fields = append(fields, "maximum")
	}
	if schema.ExclusiveMaximum {
		fields = append(fields, "exclusiveMaximum")
	}
	if schema.Minimum != 0 {
		fields = append(fields, "minimum")
	}
	if schema.ExclusiveMinimum {
		fields = append(fields, "exclusiveMinimum")
	}
	if schema.MaxLength != 0 {
		fields = append(fields, "maxLength")
	}
	if schema.MinLength != 0 {
		fields = append(fields, "minLength")
	}
	if schema.Pattern != "" {
		fields = append(fields, "pattern")
	}
	if schema.MaxItems != 0 {
		fields = append(fields, "maxItems")
	}
	if schema.MinItems != 0 {
		fields = append(fields, "minItems")
	}
	if schema.UniqueItems {
		fields = append(fields, "uniqueItems")
	}
	if schema.MaxProperties != 0 {
		fields = append(fields, "maxProperties")
	}
	if schema.MinProperties != 0 {
		fields = append(fields, "minProperties")
	}
	if schema.Required != nil {
		fields = append(fields, "required")
	}
	if schema.Not != nil {
		fields = append(fields, "not")
	}
	if schema.Default != nil {
		fields = append(fields, "default")
	}
	return fields
}

// Returns fields that the won't be considered by the plugin for mediaType.
func getNotSupportedMediaTypeFields(mediaType *openapiv3.MediaType) []string {
	fields := make([]string, 0)
	if mediaType == nil {
		return fields
	}
	if mediaType.Examples != nil {
		fields = append(fields, "examples")
	}
	if mediaType.Example != nil {
		fields = append(fields, "example")
	}
	if mediaType.Encoding != nil {
		fields = append(fields, "encoding")
	}
	return fields
}

// Returns fields that the won't be considered by the plugin for operation.
func getNotSupportedOperationFields(operation *openapiv3.Operation) []string {
	fields := make([]string, 0)
	if operation == nil {
		return fields
	}
	if operation.Tags != nil {
		fields = append(fields, "tags")
	}
	if operation.ExternalDocs != nil {
		fields = append(fields, "externalDocs")
	}
	if operation.Callbacks != nil {
		fields = append(fields, "callbacks")
	}
	if operation.Deprecated {
		fields = append(fields, "deprecated")
	}
	if operation.Security != nil {
		fields = append(fields, "security")
	}
	if operation.Servers != nil {
		fields = append(fields, "servers")
	}
	return fields
}

// Returns fields that the won't be considered by the plugin for response.
func getNotSupportedResponseFields(response *openapiv3.Response) []string {
	fields := make([]string, 0)
	if response == nil {
		return nil
	}
	if response.Links != nil {
		fields = append(fields, "links")
	}
	if response.Headers != nil {
		fields = append(fields, "headers")
	}
	return fields
}

// Returns fields that the won't be considered by the plugin for pathItem.
func getNotSupportedPathItemFields(pathItem *openapiv3.PathItem) []string {
	fields := make([]string, 0)
	if pathItem == nil {
		return fields
	}
	if pathItem.Head != nil {
		fields = append(fields, "head")
	}
	if pathItem.Options != nil {
		fields = append(fields, "options")
	}
	if pathItem.Trace != nil {
		fields = append(fields, "trace")
	}
	if pathItem.Servers != nil {
		fields = append(fields, "servers")
	}
	if pathItem.Parameters != nil {
		fields = append(fields, "parameters")
	}
	return fields
}

// Returns fields that the won't be considered by the plugin for components.
func getNotSupportedComponentsFields(components *openapiv3.Components) []string {
	fields := make([]string, 0)
	if components == nil {
		return fields
	}

	if components.Examples != nil {
		fields = append(fields, "examples")
	}
	if components.Headers != nil {
		fields = append(fields, "headers")
	}
	if components.SecuritySchemes != nil {
		fields = append(fields, "securitySchemes")
	}
	if components.Links != nil {
		fields = append(fields, "links")
	}
	if components.Callbacks != nil {
		fields = append(fields, "callbacks")
	}
	return fields
}

// Returns a copy of arr.
func copyKeys(arr []string) []string {
	cpy := make([]string, len(arr))
	copy(cpy, arr)
	return cpy
}
