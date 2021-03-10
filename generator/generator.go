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
	"google.golang.org/protobuf/types/descriptorpb"
	"log"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/golang/protobuf/descriptor"
	"github.com/golang/protobuf/proto"
	dpb "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/golang/protobuf/ptypes/empty"
	openapiv3 "github.com/googleapis/gnostic/openapiv3"
	surface_v1 "github.com/googleapis/gnostic/surface"
	"google.golang.org/genproto/googleapis/api/annotations"
)

var protoBufScalarTypes = getProtobufTypes()

// Gathers all symbolic references we generated in recursive calls.
var generatedSymbolicReferences = make(map[string]bool, 0)

// Gathers all messages that have been generated from symbolic references in recursive calls.
var generatedMessages = make(map[string]string, 0)

// Uses the output of gnostic to return a dpb.FileDescriptorSet (in bytes). 'renderer' contains
// the 'model' (surface model) which has all the relevant data to create the dpb.FileDescriptorSet.
// There are four main steps:
//		1. buildAllMessageDescriptors is called to create all messages which will be rendered in .proto
//		2. buildAllServiceDescriptors is called to create a RPC service which will be rendered in .proto
// 		3. buildSymbolicReferences 	recursively executes this plugin to generate all FileDescriptorSet based on symbolic
// 									references. A symbolic reference is an URL to another OpenAPI description inside of
//									current description.
// 		4. buildDependencies to build all static FileDescriptorProto we need.
func (renderer *Renderer) runFileDescriptorSetGenerator() (fdSet *dpb.FileDescriptorSet, err error) {
	syntax := "proto3"
	n := renderer.Package + ".proto"

	protoToBeRendered := &dpb.FileDescriptorProto{
		Name:    &n,
		Package: &renderer.Package,
		Syntax:  &syntax,
	}

	allMessages, err := buildAllMessageDescriptors(renderer)
	if err != nil {
		return nil, err
	}
	protoToBeRendered.MessageType = allMessages

	allServices, err := buildAllServiceDescriptors(protoToBeRendered.MessageType, renderer)
	if err != nil {
		return nil, err
	}
	protoToBeRendered.Service = allServices

	sourceCodeInfo, err := buildSourceCodeInfo(renderer.Model.Types)
	if err != nil {
		return nil, err
	}
	protoToBeRendered.SourceCodeInfo = sourceCodeInfo

	symbolicReferenceDependencies, err := buildSymbolicReferences(renderer)
	if err != nil {
		return nil, err
	}
	dependencies := buildDependencies()
	dependencies = append(dependencies, symbolicReferenceDependencies...)
	dependencyNames := getNamesOfDependenciesThatWillBeImported(dependencies, renderer.Model.Methods)
	protoToBeRendered.Dependency = dependencyNames

	allFileDescriptors := append(symbolicReferenceDependencies, dependencies...)
	allFileDescriptors = append(allFileDescriptors, protoToBeRendered)
	fdSet = &dpb.FileDescriptorSet{
		File: allFileDescriptors,
	}

	return fdSet, err
}

// buildAllMessageDescriptors builds protobuf messages from the surface model types. If the type is a RPC request parameter
// the fields have to follow certain rules, and therefore have to be validated.
func buildAllMessageDescriptors(renderer *Renderer) (messageDescriptors []*dpb.DescriptorProto, err error) {
	for _, surfaceType := range renderer.Model.Types {
		message := &dpb.DescriptorProto{}
		message.Name = &surfaceType.TypeName

		for i, surfaceField := range surfaceType.Fields {
			if strings.Contains(surfaceField.NativeType, "map[string][]") {
				// Not supported for now: https://github.com/LorenzHW/gnostic-grpc-deprecated/issues/3#issuecomment-509348357
				continue
			}
			if isRequestParameter(surfaceType) {
				validateRequestParameter(surfaceField)
			}

			addFieldDescriptor(message, surfaceField, i, renderer.Package)
			addEnumDescriptorIfNecessary(message, surfaceField)
		}
		messageDescriptors = append(messageDescriptors, message)
		generatedMessages[*message.Name] = renderer.Package + "." + *message.Name
	}
	return messageDescriptors, nil
}

func addFieldDescriptor(message *dpb.DescriptorProto, surfaceField *surface_v1.Field, idx int, packageName string) {
	count := int32(idx + 1)
	fieldDescriptor := &dpb.FieldDescriptorProto{Number: &count, Name: &surfaceField.FieldName}
	fieldDescriptor.Type = getFieldDescriptorType(surfaceField.NativeType, surfaceField.EnumValues)
	fieldDescriptor.Label = getFieldDescriptorLabel(surfaceField)
	fieldDescriptor.TypeName = getFieldDescriptorTypeName(*fieldDescriptor.Type, surfaceField, packageName)

	addMapDescriptorIfNecessary(surfaceField, fieldDescriptor, message)

	message.Field = append(message.Field, fieldDescriptor)
}

func addMapDescriptorIfNecessary(f *surface_v1.Field, fieldDescriptor *dpb.FieldDescriptorProto, message *dpb.DescriptorProto) {
	if f.Kind == surface_v1.FieldKind_MAP {
		// Maps are represented as nested types inside of the descriptor.
		mapDescriptor := buildMapDescriptor(f)
		fieldDescriptor.TypeName = mapDescriptor.Name
		message.NestedType = append(message.NestedType, mapDescriptor)
	}
}

func addEnumDescriptorIfNecessary(message *dpb.DescriptorProto, f *surface_v1.Field) {
	if f.EnumValues != nil {
		message.EnumType = append(message.EnumType, buildEnumDescriptorProto(f))
	}
}

func validateRequestParameter(field *surface_v1.Field) {
	if field.Position == surface_v1.Position_PATH {
		validatePathParameter(field)
	}

	if field.Position == surface_v1.Position_QUERY {
		validateQueryParameter(field)
	}
}

// getFieldDescriptorType returns a field descriptor type for the given 'nativeType'. If it is not a scalar type
// then we have a reference to another type which will get rendered as a message.
func getFieldDescriptorType(nativeType string, enumValues []string) *dpb.FieldDescriptorProto_Type {
	protoType := dpb.FieldDescriptorProto_TYPE_MESSAGE
	if protoType, ok := protoBufScalarTypes[nativeType]; ok {
		return &protoType
	}
	if enumValues != nil {
		protoType := dpb.FieldDescriptorProto_TYPE_ENUM
		return &protoType
	}
	return &protoType
}

// getFieldDescriptorLabel returns the label for the descriptor based on the information in he surface field.
func getFieldDescriptorLabel(f *surface_v1.Field) *dpb.FieldDescriptorProto_Label {
	label := dpb.FieldDescriptorProto_LABEL_OPTIONAL
	if f.Kind == surface_v1.FieldKind_ARRAY || strings.Contains(f.NativeType, "map") {
		label = dpb.FieldDescriptorProto_LABEL_REPEATED
	}
	return &label
}

// getFieldDescriptorTypeName returns the typeName of the descriptor. A TypeName has to be set if the field is a reference to another
// descriptor or enum. Otherwise it is nil. Names are set according to the protocol buffer style guide for message names:
// https://developers.google.com/protocol-buffers/docs/style#message-and-field-names
func getFieldDescriptorTypeName(fieldDescriptorType descriptorpb.FieldDescriptorProto_Type, field *surface_v1.Field, packageName string) *string {
	// Check whether we generated this message already inside of another dependency. If so we will use that name instead.
	if n, ok := generatedMessages[field.NativeType]; ok {
		return &n
	}

	typeName := ""
	if fieldDescriptorType == dpb.FieldDescriptorProto_TYPE_MESSAGE {
		typeName = packageName + "." + field.NativeType
	}
	if fieldDescriptorType == dpb.FieldDescriptorProto_TYPE_ENUM {
		typeName = field.NativeType
	}
	return &typeName
}

// buildMapDescriptor builds the necessary descriptor to render a map. (https://developers.google.com/protocol-buffers/docs/proto3#maps)
// A map is represented as nested message with two fields: 'key', 'value' and the Options set accordingly.
func buildMapDescriptor(field *surface_v1.Field) *dpb.DescriptorProto {
	isMapEntry := true
	n := field.FieldName + "Entry"

	mapDP := &dpb.DescriptorProto{
		Name:    &n,
		Field:   buildKeyValueFields(field),
		Options: &dpb.MessageOptions{MapEntry: &isMapEntry},
	}
	return mapDP
}

// buildKeyValueFields builds the necessary 'key', 'value' fields for the map descriptor.
func buildKeyValueFields(field *surface_v1.Field) []*dpb.FieldDescriptorProto {
	k, v := "key", "value"
	var n1, n2 int32 = 1, 2
	l := dpb.FieldDescriptorProto_LABEL_OPTIONAL
	t := dpb.FieldDescriptorProto_TYPE_STRING
	keyField := &dpb.FieldDescriptorProto{
		Name:   &k,
		Number: &n1,
		Label:  &l,
		Type:   &t,
	}

	valueType := field.NativeType[11:] // This transforms a string like 'map[string]int32' to 'int32'. In other words: the type of the value from the map.
	valueField := &dpb.FieldDescriptorProto{
		Name:     &v,
		Number:   &n2,
		Label:    &l,
		Type:     getFieldDescriptorType(valueType, field.EnumValues),
		TypeName: getTypeNameForMapValueType(valueType),
	}
	return []*dpb.FieldDescriptorProto{keyField, valueField}
}

// buildEnumDescriptorProto builds the necessary descriptor to render a enum. (https://developers.google.com/protocol-buffers/docs/proto3#enum)
func buildEnumDescriptorProto(f *surface_v1.Field) *dpb.EnumDescriptorProto {
	enumDescriptor := &dpb.EnumDescriptorProto{Name: &f.NativeType}
	for enumCtr, value := range f.EnumValues {
		num := int32(enumCtr)
		name := strings.ToUpper(value)
		valueDescriptor := &dpb.EnumValueDescriptorProto{
			Name:   &name,
			Number: &num,
		}
		enumDescriptor.Value = append(enumDescriptor.Value, valueDescriptor)
	}
	return enumDescriptor
}

// buildAllServiceDescriptors builds a protobuf RPC service. For every method the corresponding gRPC-HTTP transcoding options (https://github.com/googleapis/googleapis/blob/master/google/api/http.proto)
// have to be set.
func buildAllServiceDescriptors(messages []*dpb.DescriptorProto, renderer *Renderer) (services []*dpb.ServiceDescriptorProto, err error) {
	serviceName := findValidServiceName(messages, strings.Title(renderer.Package))
	methodDescriptors, err := buildAllMethodDescriptors(renderer.Model.Methods, renderer.Model.Types)
	if err != nil {
		return nil, err
	}
	service := &dpb.ServiceDescriptorProto{
		Name:   &serviceName,
		Method: methodDescriptors,
	}
	services = append(services, service)
	return services, nil
}

func buildAllMethodDescriptors(methods []*surface_v1.Method, types []*surface_v1.Type) (allMethodDescriptors []*dpb.MethodDescriptorProto, err error) {
	for _, method := range methods {
		methodDescriptor, err := buildMethodDescriptor(method, types)
		if err != nil {
			return nil, err
		}
		allMethodDescriptors = append(allMethodDescriptors, methodDescriptor)
	}
	return allMethodDescriptors, nil
}

func buildMethodDescriptor(method *surface_v1.Method, types []*surface_v1.Type) (methodDescriptor *dpb.MethodDescriptorProto, err error) {
	options, err := buildMethodOptions(method, types)
	if err != nil {
		return nil, err
	}
	inputType, outputType := buildInputTypeAndOutputType(method.ParametersTypeName, method.ResponsesTypeName)
	methodDescriptor = &dpb.MethodDescriptorProto{
		Name:       &method.HandlerName,
		InputType:  &inputType,
		OutputType: &outputType,
		Options:    options,
	}
	return methodDescriptor, nil
}

func buildInputTypeAndOutputType(parametersTypeName string, responseTypeName string) (inputType string, outputType string) {
	inputType = parametersTypeName
	outputType = responseTypeName
	if parametersTypeName == "" {
		inputType = "google.protobuf.Empty"
	}
	if responseTypeName == "" {
		outputType = "google.protobuf.Empty"
	}
	return inputType, outputType
}

func buildMethodOptions(method *surface_v1.Method, types []*surface_v1.Type) (options *dpb.MethodOptions, err error) {
	options = &dpb.MethodOptions{}
	httpRule := getHttpRuleForMethod(method)
	httpRule.Body = getRequestBodyForRequestParameter(method.ParametersTypeName, types)
	if err := proto.SetExtension(options, annotations.E_Http, &httpRule); err != nil {
		return nil, err
	}
	return options, nil
}

// getHttpRuleForMethod constructs a HttpRule from google/api/http.proto. Enables gRPC-HTTP transcoding on 'method'.
// If not nil, body is also set.
func getHttpRuleForMethod(method *surface_v1.Method) annotations.HttpRule {
	var httpRule annotations.HttpRule
	switch method.Method {
	case "GET":
		httpRule = annotations.HttpRule{
			Pattern: &annotations.HttpRule_Get{
				Get: method.Path,
			},
		}
	case "POST":
		httpRule = annotations.HttpRule{
			Pattern: &annotations.HttpRule_Post{
				Post: method.Path,
			},
		}
	case "PUT":
		httpRule = annotations.HttpRule{
			Pattern: &annotations.HttpRule_Put{
				Put: method.Path,
			},
		}
	case "PATCH":
		httpRule = annotations.HttpRule{
			Pattern: &annotations.HttpRule_Patch{
				Patch: method.Path,
			},
		}
	case "DELETE":
		httpRule = annotations.HttpRule{
			Pattern: &annotations.HttpRule_Delete{
				Delete: method.Path,
			},
		}
	}
	return httpRule
}

// getRequestBodyForRequestParameter finds the corresponding surface model type for 'name' and returns the name of the
// field that is a request body. If no such field is found it returns nil.
func getRequestBodyForRequestParameter(name string, types []*surface_v1.Type) string {
	requestParameterType := &surface_v1.Type{}

	for _, t := range types {
		if t.TypeName == name {
			requestParameterType = t
		}
	}

	for _, f := range requestParameterType.Fields {
		if f.Position == surface_v1.Position_BODY {
			return f.FieldName
		}
	}
	return ""
}

// buildSourceCodeInfo builds the object which holds additional information, such as the description from OpenAPI
// components. This information will be rendered as a comment in the final .proto file.
func buildSourceCodeInfo(types []*surface_v1.Type) (sourceCodeInfo *dpb.SourceCodeInfo, err error) {
	allLocations := make([]*dpb.SourceCodeInfo_Location, 0)
	for idx, surfaceType := range types {
		location := &dpb.SourceCodeInfo_Location{
			Path:            []int32{4, int32(idx)},
			LeadingComments: &surfaceType.Description,
		}
		allLocations = append(allLocations, location)
	}
	sourceCodeInfo = &dpb.SourceCodeInfo{
		Location: allLocations,
	}
	return sourceCodeInfo, nil
}

// buildSymbolicReferences recursively generates all .proto definitions to external OpenAPI descriptions (URLs to other
// descriptions inside the current description).
func buildSymbolicReferences(renderer *Renderer) (symbolicFileDescriptors []*dpb.FileDescriptorProto, err error) {
	symbolicReferences := renderer.Model.SymbolicReferences
	symbolicReferences = trimAndRemoveDuplicates(symbolicReferences)

	for _, ref := range symbolicReferences {
		if _, alreadyGenerated := generatedSymbolicReferences[ref]; !alreadyGenerated {
			generatedSymbolicReferences[ref] = true

			// Lets get the standard gnostic output from the symbolic reference.
			cmd := exec.Command("gnostic", "--pb-out=-", ref)
			b, err := cmd.Output()
			if err != nil {
				return nil, err
			}

			// Construct an OpenAPI document v3.
			document, err := createOpenAPIDocFromGnosticOutput(b)
			if err != nil {
				return nil, err
			}

			// Create the surface model. Keep in mind that this resolves the references of the symbolic reference again!
			surfaceModel, err := surface_v1.NewModelFromOpenAPI3(document, ref)
			if err != nil {
				return nil, err
			}

			// Prepare surface model for recursive call. TODO: Keep discovery documents in mind.
			inputDocumentType := "openapi.v3.Document"
			if document.Openapi == "2.0.0" {
				inputDocumentType = "openapi.v2.Document"
			}
			NewProtoLanguageModel().Prepare(surfaceModel, inputDocumentType)

			// Recursively call the generator.
			recursiveRenderer := NewRenderer(surfaceModel)
			fileName := path.Base(ref)
			recursiveRenderer.Package = strings.TrimSuffix(fileName, filepath.Ext(fileName))
			newFdSet, err := recursiveRenderer.runFileDescriptorSetGenerator()
			if err != nil {
				return nil, err
			}
			renderer.SymbolicFdSets = append(renderer.SymbolicFdSets, newFdSet)

			symbolicProto := getLast(newFdSet.File)
			symbolicFileDescriptors = append(symbolicFileDescriptors, symbolicProto)
		}
	}
	return symbolicFileDescriptors, nil
}

// Protoreflect needs all the dependencies that are used inside of the FileDescriptorProto (that gets rendered)
// to work properly. Those dependencies are google/protobuf/empty.proto, google/api/annotations.proto,
// and "google/protobuf/descriptor.proto". For all those dependencies the corresponding
// FileDescriptorProto has to be added to the FileDescriptorSet. Protoreflect won't work
// if a reference is missing.
func buildDependencies() (dependencies []*dpb.FileDescriptorProto) {
	// Dependency to google/api/annotations.proto for gRPC-HTTP transcoding. Here a couple of problems arise:
	// 1. Problem: 	We cannot call descriptor.ForMessage(&annotations.E_Http), which would be our
	//				required dependency. However, we can call descriptor.ForMessage(&http) and
	//				then construct the extension manually.
	// 2. Problem: 	The name is set wrong.
	// 3. Problem: 	google/api/annotations.proto has a dependency to google/protobuf/descriptor.proto.
	http := annotations.Http{}
	fd, _ := descriptor.MessageDescriptorProto(&http)

	extensionName := "http"
	n := "google/api/annotations.proto"
	l := dpb.FieldDescriptorProto_LABEL_OPTIONAL
	t := dpb.FieldDescriptorProto_TYPE_MESSAGE
	tName := "google.api.HttpRule"
	extendee := ".google.protobuf.MethodOptions"

	httpExtension := &dpb.FieldDescriptorProto{
		Name:     &extensionName,
		Number:   &annotations.E_Http.Field,
		Label:    &l,
		Type:     &t,
		TypeName: &tName,
		Extendee: &extendee,
	}

	fd.Extension = append(fd.Extension, httpExtension)                        // 1. Problem
	fd.Name = &n                                                              // 2. Problem
	fd.Dependency = append(fd.Dependency, "google/protobuf/descriptor.proto") //3.rd Problem

	// Build other required dependencies
	e := empty.Empty{}
	fdp := dpb.DescriptorProto{}
	fd2, _ := descriptor.MessageDescriptorProto(&e)
	fd3, _ := descriptor.MessageDescriptorProto(&fdp)
	dependencies = []*dpb.FileDescriptorProto{fd, fd2, fd3}
	return dependencies
}

// getNamesOfDependenciesThatWillBeImported adds the dependencies to the FileDescriptorProto we want to render (the last one). This essentially
// makes the 'import'  statements inside the .proto definition.
func getNamesOfDependenciesThatWillBeImported(dependencies []*dpb.FileDescriptorProto, methods []*surface_v1.Method) (names []string) {
	// At last, we need to add the dependencies to the FileDescriptorProto in order to get them rendered.
	for _, fd := range dependencies {
		if isEmptyDependency(*fd.Name) && shouldAddEmptyDependency(methods) {
			// Reference: https://github.com/googleapis/gnostic-grpc/issues/8
			names = append(names, *fd.Name)
			continue
		}
		names = append(names, *fd.Name)
	}
	// Sort imports so they will be rendered in a consistent order.
	sort.Strings(names)
	return names
}

// validatePathParameter validates if the path parameter has the requested structure.
// This is necessary according to: https://github.com/googleapis/googleapis/blob/master/google/api/http.proto#L62
func validatePathParameter(field *surface_v1.Field) {
	if field.Kind != surface_v1.FieldKind_SCALAR {
		log.Println("The path parameter with the Name " + field.Name + " is invalid. " +
			"The path template may refer to one or more fields in the gRPC request message, as" +
			" long as each field is a non-repeated field with a primitive (non-message) type. " +
			"See: https://github.com/googleapis/googleapis/blob/master/google/api/http.proto#L62 for more information.")
	}
}

// validateQueryParameter validates if the query parameter has the requested structure.
// This is necessary according to: https://github.com/googleapis/googleapis/blob/master/google/api/http.proto#L118
func validateQueryParameter(field *surface_v1.Field) {
	_, isScalar := protoBufScalarTypes[field.NativeType]
	if !(field.Kind == surface_v1.FieldKind_SCALAR ||
		(field.Kind == surface_v1.FieldKind_ARRAY && isScalar) ||
		(field.Kind == surface_v1.FieldKind_REFERENCE)) {
		log.Println("The query parameter with the Name " + field.Name + " is invalid. " +
			"Note that fields which are mapped to URL query parameters must have a primitive type or" +
			" a repeated primitive type or a non-repeated message type. " +
			"See: https://github.com/googleapis/googleapis/blob/master/google/api/http.proto#L118 for more information.")
	}

}

// isEmptyDependency returns true if the 'name' of the dependency is empty.proto
func isEmptyDependency(name string) bool {
	return name == "google/protobuf/empty.proto"
}

// shouldAddEmptyDependency returns true if at least one request parameter or response parameter is empty
func shouldAddEmptyDependency(methods []*surface_v1.Method) bool {
	for _, method := range methods {
		if method.ParametersTypeName == "" || method.ResponsesTypeName == "" {
			return true
		}
	}
	return false
}

// isRequestParameter checks whether 't' is a type that will be used as a request parameter for a RPC method.
func isRequestParameter(sufaceType *surface_v1.Type) bool {
	if strings.Contains(sufaceType.Description, sufaceType.GetName()+" holds parameters to") {
		return true
	}
	return false
}

// getTypeNameForMapValueType returns the type name for the given 'valueType'.
// A type name for a field is only set if it is some kind of reference (non-scalar values) otherwise it is nil.
func getTypeNameForMapValueType(valueType string) *string {
	if _, ok := protoBufScalarTypes[valueType]; ok {
		return nil // Ok it is a scalar. For scalar values we don't set the TypeName of the field.
	}
	typeName := valueType
	return &typeName
}

// createOpenAPIDocFromGnosticOutput uses the 'binaryInput' from gnostic to create a OpenAPI document.
func createOpenAPIDocFromGnosticOutput(binaryInput []byte) (*openapiv3.Document, error) {
	document := &openapiv3.Document{}
	err := proto.Unmarshal(binaryInput, document)
	if err != nil {
		// If we execute gnostic with argument: '-pb-out=-' we get an EOF. So lets only return other errors.
		if err.Error() != "unexpected EOF" {
			return nil, err
		}
	}
	return document, nil
}

// trimAndRemoveDuplicates returns a list of URLs that are not duplicates (considering only the part until the first '#')
func trimAndRemoveDuplicates(urls []string) []string {
	result := make([]string, 0)
	for _, url := range urls {
		parts := strings.Split(url, "#")
		if !isDuplicate(result, parts[0]) {
			result = append(result, parts[0])
		}
	}
	return result
}

// isDuplicate returns true if 's' is inside 'ss'.
func isDuplicate(ss []string, s string) bool {
	for _, s2 := range ss {
		if s == s2 {
			return true
		}
	}
	return false
}

// getLast returns the last FileDescriptorProto of the array 'protos'.
func getLast(protos []*dpb.FileDescriptorProto) *dpb.FileDescriptorProto {
	return protos[len(protos)-1]
}

// getProtobufTypes maps the .proto Type (given as string) (https://developers.google.com/protocol-buffers/docs/proto3#scalar)
// to the corresponding descriptor proto type.
func getProtobufTypes() map[string]dpb.FieldDescriptorProto_Type {
	typeMapping := make(map[string]dpb.FieldDescriptorProto_Type)
	typeMapping["double"] = dpb.FieldDescriptorProto_TYPE_DOUBLE
	typeMapping["float"] = dpb.FieldDescriptorProto_TYPE_FLOAT
	typeMapping["int64"] = dpb.FieldDescriptorProto_TYPE_INT64
	typeMapping["uint64"] = dpb.FieldDescriptorProto_TYPE_UINT64
	typeMapping["int32"] = dpb.FieldDescriptorProto_TYPE_INT32
	typeMapping["fixed64"] = dpb.FieldDescriptorProto_TYPE_FIXED64

	typeMapping["fixed32"] = dpb.FieldDescriptorProto_TYPE_FIXED32
	typeMapping["bool"] = dpb.FieldDescriptorProto_TYPE_BOOL
	typeMapping["string"] = dpb.FieldDescriptorProto_TYPE_STRING
	typeMapping["bytes"] = dpb.FieldDescriptorProto_TYPE_BYTES
	typeMapping["uint32"] = dpb.FieldDescriptorProto_TYPE_UINT32
	typeMapping["sfixed32"] = dpb.FieldDescriptorProto_TYPE_SFIXED32
	typeMapping["sfixed64"] = dpb.FieldDescriptorProto_TYPE_SFIXED64
	typeMapping["sint32"] = dpb.FieldDescriptorProto_TYPE_SINT32
	typeMapping["sint64"] = dpb.FieldDescriptorProto_TYPE_SINT64
	return typeMapping
}

// findValidServiceName finds a valid service name for the gRPC service. A valid service name is not already taken by a
// message. Reference: https://github.com/googleapis/gnostic-grpc/issues/7
func findValidServiceName(messages []*dpb.DescriptorProto, serviceName string) string {
	messageNames := make(map[string]bool)

	for _, m := range messages {
		messageNames[*m.Name] = true
	}

	validServiceName := serviceName
	ctr := 0
	for {
		if nameIsAlreadyTaken, _ := messageNames[validServiceName]; !nameIsAlreadyTaken {
			return validServiceName
		}
		validServiceName = serviceName + "Service"
		if ctr > 0 {
			validServiceName += strconv.Itoa(ctr)
		}
		ctr += 1
	}
}
