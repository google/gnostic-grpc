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
	openapiv3 "github.com/googleapis/gnostic/OpenAPIv3"
	plugins "github.com/googleapis/gnostic/plugins"
	"strings"
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
	if len(fields) > 0 {
		text := "Fields: " + strings.Join(fields, ", ") + " are not supported for Document with title: " + c.document.Info.Title
		msg := constructMessage("DOCUMENTFIELDS", text, []string{"Document"})
		c.messages = append(c.messages, msg)
	}
	c.analyzeComponents()
	c.analyzePaths()
}

// Analyzes the components of a OpenAPI description.
func (c *GrpcChecker) analyzeComponents() {
	components := c.document.Components

	fields := getNotSupportedComponentsFields(components)
	if len(fields) > 0 {
		text := "Fields: " + strings.Join(fields, ", ") + " are not supported for the component"
		msg := constructMessage("COMPONENTSFIELDS", text, []string{"Component"})
		c.messages = append(c.messages, msg)
	}

	if schemas := components.GetSchemas(); schemas != nil {
		for _, pair := range schemas.AdditionalProperties {
			c.analyzeSchema(pair.Name, pair.Value)
		}
	}

	if responses := components.GetResponses(); responses != nil {
		for _, pair := range responses.AdditionalProperties {
			c.analyzeResponse(pair)
		}
	}

	if parameters := components.GetParameters(); parameters != nil {
		for _, pair := range parameters.AdditionalProperties {
			c.analyzeParameter(pair.Value)
		}
	}

	if requestBodies := components.GetRequestBodies(); requestBodies != nil {
		for _, pair := range requestBodies.AdditionalProperties {
			c.analyzeRequestBody(pair)
		}
	}
}

// Analyzes all paths.
func (c *GrpcChecker) analyzePaths() {
	for _, pathItem := range c.document.Paths.Path {
		c.analyzePathItem(pathItem)
	}
}

// Analyzes one single path.
func (c *GrpcChecker) analyzePathItem(pair *openapiv3.NamedPathItem) {
	pathItem := pair.Value

	fields := getNotSupportedPathItemFields(pathItem)
	if len(fields) > 0 {
		text := "Fields: " + strings.Join(fields, ", ") + " are not supported for path: " + pair.Name
		msg := constructMessage("PATHFIELDS", text, []string{"Paths", pair.Name, "Operation"})
		c.messages = append(c.messages, msg)
	}

	operations := getValidOperations(pathItem)
	for _, op := range operations {
		c.analyzeOperation(op)
	}
}

// Analyzes a single Operation.
func (c *GrpcChecker) analyzeOperation(operation *openapiv3.Operation) {
	fields := getNotSupportedOperationFields(operation)
	if len(fields) > 0 {
		text := "Fields:  " + strings.Join(fields, ", ") + " are not supported for operation: " + operation.OperationId
		msg := constructMessage("OPERATIONFIELDS", text, []string{"Operation", operation.OperationId, "Callbacks"})
		c.messages = append(c.messages, msg)
	}

	for _, param := range operation.Parameters {
		c.analyzeParameter(param)
	}

	for _, response := range operation.Responses.GetResponseOrReference() {
		c.analyzeResponse(response)
	}

	if defaultResponse := operation.Responses.Default; defaultResponse != nil {
		wrap := &openapiv3.NamedResponseOrReference{Name: operation.OperationId + " Default response", Value: defaultResponse}
		c.analyzeResponse(wrap)
	}

	wrap := &openapiv3.NamedRequestBodyOrReference{Name: operation.OperationId, Value: operation.RequestBody}
	c.analyzeRequestBody(wrap)

}

// Analyzes the parameter.
func (c *GrpcChecker) analyzeParameter(paramOrRef *openapiv3.ParameterOrReference) {
	if parameter := paramOrRef.GetParameter(); parameter != nil {
		fields := getNotSupportedParameterFields(parameter)
		if len(fields) > 0 {
			text := "Fields: " + strings.Join(fields, ", ") + " are not supported for parameter: " + parameter.Name
			msg := constructMessage("PARAMATERFIELDS", text, []string{"Parameter", parameter.Name})
			c.messages = append(c.messages, msg)
		}
		c.analyzeSchema(parameter.Name, parameter.Schema)
	}
}

// Analyzes the schema.
func (c *GrpcChecker) analyzeSchema(identifier string, schemaOrReference *openapiv3.SchemaOrReference) {
	if schema := schemaOrReference.GetSchema(); schema != nil {
		fields := getNotSupportedSchemaFields(schema)
		if len(fields) > 0 {
			text := "Fields: " + strings.Join(fields, ", ") + " are not supported for the schema: " + identifier
			msg := constructMessage("SCHEMAFIELDS", text, []string{identifier, "Schema"})
			c.messages = append(c.messages, msg)
		}

		if enum := schema.Enum; enum != nil {
			text := "Field: Enum is not generated as enum in .proto for schema: " + identifier
			msg := constructMessage("SCHEMAFIELDS", text, []string{identifier, "Schema"})
			c.messages = append(c.messages, msg)
		}

		// Check for this: https://github.com/LorenzHW/gnostic-grpc/issues/3#issuecomment-509348357
		if additionalProperties := schema.AdditionalProperties; additionalProperties != nil {
			if schema := additionalProperties.GetSchemaOrReference().GetSchema(); schema != nil {
				if schema.Type == "array" {
					text := "Field: additionalProperties with type array is generated as empty message inside .proto."
					msg := constructMessage("SCHEMAFIELDS", text, []string{identifier, "Schema"})
					c.messages = append(c.messages, msg)
				}
			}
		}

		if items := schema.Items; items != nil {
			for _, schemaOrRef := range items.SchemaOrReference {
				c.analyzeSchema("Items of "+identifier, schemaOrRef)
			}
		}

		if properties := schema.Properties; properties != nil {
			for _, pair := range properties.AdditionalProperties {
				c.analyzeSchema(pair.Name, pair.Value)
			}
		}

		if additionalProperties := schema.AdditionalProperties; additionalProperties != nil {
			c.analyzeSchema("AdditionalProperties of "+identifier, additionalProperties.GetSchemaOrReference())
		}
	}
}

// Analyzes a response.
func (c *GrpcChecker) analyzeResponse(pair *openapiv3.NamedResponseOrReference) {
	if response := pair.Value.GetResponse(); response != nil {
		fields := getNotSupportedResponseFields(response)
		if len(fields) > 0 {
			text := "Fields:" + strings.Join(fields, ", ") + " are not supported for response: " + pair.Name
			msg := constructMessage("RESPONSEFIELDS", text, []string{"Response", pair.Name})
			c.messages = append(c.messages, msg)
		}
		if content := response.Content; content != nil {
			for _, pair := range content.AdditionalProperties {
				c.analyzeContent(pair)
			}
		}
	}
}

// Analyzes a request body.
func (c *GrpcChecker) analyzeRequestBody(pair *openapiv3.NamedRequestBodyOrReference) {
	if requestBody := pair.Value.GetRequestBody(); requestBody != nil {
		if requestBody.Required {
			text := "Fields: Required are not supported for the request: " + pair.Name
			msg := constructMessage("REQUESTBODYFIELDS", text, []string{"RequestBody", pair.Name})
			c.messages = append(c.messages, msg)
		}
		for _, pair := range requestBody.Content.AdditionalProperties {
			c.analyzeContent(pair)
		}
	}
}

// Analyzes the content of a response.
func (c *GrpcChecker) analyzeContent(pair *openapiv3.NamedMediaType) {
	mediaType := pair.Value

	fields := getNotSupportedMediaTypeFields(mediaType)
	if len(fields) > 0 {
		text := "Fields:" + strings.Join(fields, ", ") + " are not supported for the mediatype: " + pair.Name
		msg := constructMessage("MEDIATYPEFIELDS", text, []string{"MediaType", pair.Name})
		c.messages = append(c.messages, msg)
	}

	if mediaType.Schema != nil {
		c.analyzeSchema(pair.Name, mediaType.Schema)
	}
}

// Constructs a message which the end user will see on the console.
func constructMessage(code string, text string, keys []string) *plugins.Message {
	return &plugins.Message{
		Code:  code,
		Level: plugins.Message_WARNING,
		Text:  text,
		Keys:  keys,
	}
}

// Returns all valid operations that will be transcoded by the plugin.
func getValidOperations(pathItem *openapiv3.PathItem) []*openapiv3.Operation {
	operations := make([]*openapiv3.Operation, 0)
	if pathItem == nil {
		return operations
	}

	if pathItem.Get != nil {
		operations = append(operations, pathItem.Get)
	}
	if pathItem.Put != nil {
		operations = append(operations, pathItem.Put)
	}
	if pathItem.Post != nil {
		operations = append(operations, pathItem.Post)
	}
	if pathItem.Delete != nil {
		operations = append(operations, pathItem.Delete)
	}
	if pathItem.Patch != nil {
		operations = append(operations, pathItem.Patch)
	}
	return operations
}

// Returns fields that the won't be considered by the plugin for document.
func getNotSupportedOpenAPIDocumentFields(document *openapiv3.Document) []string {
	fields := make([]string, 0)
	if document == nil {
		return fields
	}

	if document.Servers != nil {
		fields = append(fields, "Servers")
	}
	if document.Security != nil {
		fields = append(fields, "Security")
	}
	if document.Tags != nil {
		fields = append(fields, "Tags")
	}
	if document.ExternalDocs != nil {
		fields = append(fields, "ExternalDocs")
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
		fields = append(fields, "Required")
	}
	if parameter.Deprecated {
		fields = append(fields, "Deprecated")
	}
	if parameter.AllowEmptyValue {
		fields = append(fields, "AllowEmptyValue")
	}
	if parameter.Style != "" {
		fields = append(fields, "Style")
	}
	if parameter.Explode {
		fields = append(fields, "Explode")
	}
	if parameter.AllowReserved {
		fields = append(fields, "AllowReserved")
	}
	if parameter.Example != nil {
		fields = append(fields, "Example")
	}
	if parameter.Examples != nil {
		fields = append(fields, "Examples")
	}
	if parameter.Content != nil {
		fields = append(fields, "Content")
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
		fields = append(fields, "Nullable")
	}
	if schema.Discriminator != nil {
		fields = append(fields, "Discriminator")
	}
	if schema.ReadOnly {
		fields = append(fields, "ReadOnly")
	}
	if schema.WriteOnly {
		fields = append(fields, "WriteOnly")
	}
	if schema.Xml != nil {
		fields = append(fields, "Xml")
	}
	if schema.ExternalDocs != nil {
		fields = append(fields, "ExternalDocs")
	}
	if schema.Example != nil {
		fields = append(fields, "Example")
	}
	if schema.Deprecated {
		fields = append(fields, "Deprecated")
	}
	if schema.Title != "" {
		fields = append(fields, "Title")
	}
	if schema.MultipleOf != 0 {
		fields = append(fields, "MultipleOf")
	}
	if schema.Maximum != 0 {
		fields = append(fields, "Maximum")
	}
	if schema.ExclusiveMaximum {
		fields = append(fields, "ExclusiveMaximum")
	}
	if schema.Minimum != 0 {
		fields = append(fields, "Minimum")
	}
	if schema.ExclusiveMinimum {
		fields = append(fields, "ExclusiveMinimum")
	}
	if schema.MaxLength != 0 {
		fields = append(fields, "MaxLength")
	}
	if schema.MinLength != 0 {
		fields = append(fields, "MinLength")
	}
	if schema.Pattern != "" {
		fields = append(fields, "Pattern")
	}
	if schema.MaxItems != 0 {
		fields = append(fields, "MaxItems")
	}
	if schema.MinItems != 0 {
		fields = append(fields, "MinItems")
	}
	if schema.UniqueItems {
		fields = append(fields, "UniqueItems")
	}
	if schema.MaxProperties != 0 {
		fields = append(fields, "MaxProperties")
	}
	if schema.MinProperties != 0 {
		fields = append(fields, "MinProperties")
	}
	if schema.Required != nil {
		fields = append(fields, "Required")
	}
	if schema.AllOf != nil {
		fields = append(fields, "AllOf")
	}
	if schema.OneOf != nil {
		fields = append(fields, "OneOf")
	}

	if schema.AnyOf != nil {
		fields = append(fields, "AnyOf")
	}
	if schema.Not != nil {
		fields = append(fields, "Not")
	}
	if schema.Default != nil {
		fields = append(fields, "Default")
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
		fields = append(fields, "Examples")
	}
	if mediaType.Example != nil {
		fields = append(fields, "Example")
	}
	if mediaType.Encoding != nil {
		fields = append(fields, "Encoding")
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
		fields = append(fields, "Tags")
	}
	if operation.ExternalDocs != nil {
		fields = append(fields, "ExternalDocs")
	}
	if operation.Callbacks != nil {
		fields = append(fields, "Callbacks")
	}
	if operation.Deprecated {
		fields = append(fields, "Deprecated")
	}
	if operation.Security != nil {
		fields = append(fields, "Security")
	}
	if operation.Servers != nil {
		fields = append(fields, "Servers")
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
		fields = append(fields, "Links")
	}
	if response.Headers != nil {
		fields = append(fields, "Headers")
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
		fields = append(fields, "Head")
	}
	if pathItem.Options != nil {
		fields = append(fields, "Options")
	}
	if pathItem.Trace != nil {
		fields = append(fields, "Trace")
	}
	if pathItem.Servers != nil {
		fields = append(fields, "Servers")
	}
	if pathItem.Parameters != nil {
		fields = append(fields, "Parameters")
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
		fields = append(fields, "Examples")
	}
	if components.Headers != nil {
		fields = append(fields, "Headers")
	}
	if components.SecuritySchemes != nil {
		fields = append(fields, "SecuritySchemes")
	}
	if components.Links != nil {
		fields = append(fields, "Links")
	}
	if components.Callbacks != nil {
		fields = append(fields, "Callbacks")
	}
	return fields
}
