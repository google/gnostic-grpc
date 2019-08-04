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
	"github.com/golang/protobuf/descriptor"
	"github.com/golang/protobuf/proto"
	dpb "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/golang/protobuf/ptypes/empty"
	openapiv3 "github.com/googleapis/gnostic/OpenAPIv3"
	surface_v1 "github.com/googleapis/gnostic/surface"
	"google.golang.org/genproto/googleapis/api/annotations"
	"log"
	nethttp "net/http"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

var protoBufScalarTypes = getProtobufTypes()
var openAPITypesToProtoBuf = getOpenAPITypesToProtoBufTypes()
var openAPIScalarTypes = getOpenAPIScalarTypes()

// Gathers all symbolic references we generated in recursive calls.
var generatedSymbolicReferences = make(map[string]bool, 0)

// Gathers all messages that have been generated from symbolic references in recursive calls.
var generatedMessages = make(map[string]string, 0)

// Uses the output of gnostic to return a dpb.FileDescriptorSet (in bytes). 'renderer' contains
// the 'model' (surface model) which has all the relevant data to create the dpb.FileDescriptorSet.
// There are four main steps:
// 		1. buildDependencies to build all static FileDescriptorProto we need.
// 		2. buildSymbolicReferences 	recursively executes this plugin to generate all FileDescriptorSet based on symbolic
// 									references. A symbolic reference is an URL to another OpenAPI description inside of
//									current description.
//		3. buildMessagesFromTypes is called to create all messages which will be rendered in .proto
//		4. buildServiceFromMethods is called to create a RPC service which will be rendered in .proto
func (renderer *Renderer) runFileDescriptorSetGenerator() (fdSet *dpb.FileDescriptorSet, err error) {
	syntax := "proto3"
	n := renderer.Package + ".proto"

	// mainProto is the proto we ultimately want to render.
	mainProto := &dpb.FileDescriptorProto{
		Name:    &n,
		Package: &renderer.Package,
		Syntax:  &syntax,
	}
	fdSet = &dpb.FileDescriptorSet{
		File: []*dpb.FileDescriptorProto{mainProto},
	}

	buildDependencies(fdSet)
	err = buildSymbolicReferences(fdSet, renderer)
	if err != nil {
		return nil, err
	}

	addDependencies(fdSet)

	err = buildMessagesFromTypes(mainProto, renderer)
	if err != nil {
		return nil, err
	}

	err = buildServiceFromMethods(mainProto, renderer)
	if err != nil {
		return nil, err
	}

	return fdSet, err
}

// Adds the dependencies to the FileDescriptor we want to render. This essentially makes the 'import' statements
// inside the .proto definition.
func addDependencies(fdSet *dpb.FileDescriptorSet) {
	// At last, we need to add the dependencies to the FileDescriptorProto in order to get them rendered.
	lastFdProto := getLast(fdSet.File)
	for _, fd := range fdSet.File {
		if fd != lastFdProto {
			lastFdProto.Dependency = append(lastFdProto.Dependency, *fd.Name)
		}
	}
}

// buildSymbolicReferences recursively generates all .proto definitions to external OpenAPI descriptions (URLs to other
// descriptions inside the current description).
func buildSymbolicReferences(fdSet *dpb.FileDescriptorSet, renderer *Renderer) (err error) {
	symbolicReferences := renderer.Model.SymbolicReferences
	symbolicReferences = trimAndRemoveDuplicates(symbolicReferences)

	symbolicFileDescriptorProtos := make([]*dpb.FileDescriptorProto, 0)
	for _, ref := range symbolicReferences {
		if _, alreadyGenerated := generatedSymbolicReferences[ref]; !alreadyGenerated {
			generatedSymbolicReferences[ref] = true

			// Lets get the standard gnostic output from the symbolic reference.
			cmd := exec.Command("gnostic", "--pb-out=-", ref)
			b, err := cmd.Output()
			if err != nil {
				return err
			}

			// Construct an OpenAPI document v3.
			document, err := createOpenAPIDocFromGnosticOutput(b)
			if err != nil {
				return err
			}

			// Create the surface model. Keep in mind that this resolves the references of the symbolic reference again!
			surfaceModel, err := surface_v1.NewModelFromOpenAPI3(document, ref)
			if err != nil {
				return err
			}

			// Recursively call the generator.
			recursiveRenderer := NewRenderer(surfaceModel)
			fileName := path.Base(ref)
			recursiveRenderer.Package = strings.TrimSuffix(fileName, filepath.Ext(fileName))
			newFdSet, err := recursiveRenderer.runFileDescriptorSetGenerator()
			if err != nil {
				return err
			}
			renderer.SymbolicFdSets = append(renderer.SymbolicFdSets, newFdSet)

			symbolicProto := getLast(newFdSet.File)
			symbolicFileDescriptorProtos = append(symbolicFileDescriptorProtos, symbolicProto)
		}
	}

	fdSet.File = append(symbolicFileDescriptorProtos, fdSet.File...)
	return nil
}

// Protoreflect needs all the dependencies that are used inside of the FileDescriptorProto (that gets rendered)
// to work properly. Those dependencies are google/protobuf/empty.proto, google/api/annotations.proto,
// and "google/protobuf/descriptor.proto". For all those dependencies the corresponding
// FileDescriptorProto has to be added to the FileDescriptorSet. Protoreflect won't work
// if a reference is missing.
func buildDependencies(fdSet *dpb.FileDescriptorSet) {
	// Dependency to google/api/annotations.proto for gRPC-HTTP transcoding. Here a couple of problems arise:
	// 1. Problem: 	We cannot call descriptor.ForMessage(&annotations.E_Http), which would be our
	//				required dependency. However, we can call descriptor.ForMessage(&http) and
	//				then construct the extension manually.
	// 2. Problem: 	The name is set wrong.
	// 3. Problem: 	google/api/annotations.proto has a dependency to google/protobuf/descriptor.proto.
	http := annotations.Http{}
	fd, _ := descriptor.ForMessage(&http)

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
	fd2, _ := descriptor.ForMessage(&e)
	fd3, _ := descriptor.ForMessage(&fdp)
	dependencies := []*dpb.FileDescriptorProto{fd, fd2, fd3}

	// According to the documentation of protoReflect.CreateFileDescriptorFromSet the file I want to print
	// needs to be at the end of the array. All other FileDescriptorProto are dependencies.
	fdSet.File = append(dependencies, fdSet.File...)
}

// Builds protobuf messages from the surface model types. If the type is a RPC request parameter
// the fields have to follow certain rules, and therefore have to be validated.
func buildMessagesFromTypes(descr *dpb.FileDescriptorProto, renderer *Renderer) (err error) {
	types := renderer.Model.Types

	for _, t := range types {
		message := &dpb.DescriptorProto{}
		setMessageDescriptorName(message, t.Name)

		for i, f := range t.Fields {
			if isRequestParameter(t) {
				if f.Position == surface_v1.Position_PATH {
					validatePathParameter(f)
				}

				if f.Position == surface_v1.Position_QUERY {
					validateQueryParameter(f)
				}
			}
			ctr := int32(i + 1)
			fieldDescriptor := &dpb.FieldDescriptorProto{Number: &ctr}
			setFieldDescriptorLabel(fieldDescriptor, f)
			setFieldDescriptorName(fieldDescriptor, f)
			setFieldDescriptorType(fieldDescriptor, f)
			setFieldDescriptorTypeName(fieldDescriptor, f, renderer.Package)

			// Maps are represented as nested types inside of the descriptor.
			if f.Kind == surface_v1.FieldKind_MAP {
				if strings.Contains(f.Type, "map[string][]") {
					// Not supported for now: https://github.com/LorenzHW/gnostic-grpc/issues/3#issuecomment-509348357
					continue
				}
				mapDescriptorProto := buildMapDescriptorProto(f)
				fieldDescriptor.TypeName = mapDescriptorProto.Name
				message.NestedType = append(message.NestedType, mapDescriptorProto)
			}
			message.Field = append(message.Field, fieldDescriptor)
		}
		descr.MessageType = append(descr.MessageType, message)
		generatedMessages[*message.Name] = renderer.Package + "." + *message.Name
	}
	return nil
}

// Builds a protobuf RPC service. For every method the corresponding gRPC-HTTP transcoding options (https://github.com/googleapis/googleapis/blob/master/google/api/http.proto)
// have to be set.
func buildServiceFromMethods(descr *dpb.FileDescriptorProto, renderer *Renderer) (err error) {
	methods := renderer.Model.Methods
	serviceName := strings.Title(renderer.Package)

	service := &dpb.ServiceDescriptorProto{
		Name: &serviceName,
	}
	descr.Service = []*dpb.ServiceDescriptorProto{service}

	for _, method := range methods {
		mOptionsDescr := &dpb.MethodOptions{}
		requestBody := getRequestBodyForRequestParameters(method.ParametersTypeName, renderer.Model.Types)
		httpRule := getHttpRuleForMethod(method, requestBody)
		if err := proto.SetExtension(mOptionsDescr, annotations.E_Http, &httpRule); err != nil {
			return err
		}

		method.ParametersTypeName = cleanTypeName(method.ParametersTypeName)
		method.ResponsesTypeName = cleanTypeName(method.ResponsesTypeName)
		method.ParametersTypeName = strings.Title(method.ParametersTypeName)
		method.ResponsesTypeName = strings.Title(method.ResponsesTypeName)

		if method.ParametersTypeName == "" {
			method.ParametersTypeName = "google.protobuf.Empty"
		}
		if method.ResponsesTypeName == "" {
			method.ResponsesTypeName = "google.protobuf.Empty"
		}

		mDescr := &dpb.MethodDescriptorProto{
			Name:       &method.Name,
			InputType:  &method.ParametersTypeName,
			OutputType: &method.ResponsesTypeName,
			Options:    mOptionsDescr,
		}

		service.Method = append(service.Method, mDescr)
	}
	return nil
}

// Builds the necessary descriptor to render a map. (https://developers.google.com/protocol-buffers/docs/proto3#maps)
// A map is represented as nested message with two fields: 'key', 'value' and the Options set accordingly.
func buildMapDescriptorProto(field *surface_v1.Field) *dpb.DescriptorProto {
	isMapEntry := true
	n := field.Name + "Entry"

	mapDP := &dpb.DescriptorProto{
		Name:    &n,
		Field:   buildKeyValueFields(field),
		Options: &dpb.MessageOptions{MapEntry: &isMapEntry},
	}
	return mapDP
}

// Builds the necessary 'key', 'value' fields for the map descriptor.
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

	valueType := field.Type[11:] // This transforms a string like 'map[string]int32' to 'int32'. In other words: the type of the value from the map.
	valueField := &dpb.FieldDescriptorProto{
		Name:     &v,
		Number:   &n2,
		Label:    &l,
		Type:     getProtoTypeForMapValueType(valueType),
		TypeName: getTypeNameForMapValueType(valueType),
	}
	return []*dpb.FieldDescriptorProto{keyField, valueField}
}

// Validates if the path parameter has the requested structure.
// This is necessary according to: https://github.com/googleapis/googleapis/blob/master/google/api/http.proto#L62
func validatePathParameter(field *surface_v1.Field) {
	if field.Kind != surface_v1.FieldKind_SCALAR {
		log.Println("The path parameter with the Name " + field.Name + " is invalid. " +
			"The path template may refer to one or more fields in the gRPC request message, as" +
			" long as each field is a non-repeated field with a primitive (non-message) type. " +
			"See: https://github.com/googleapis/googleapis/blob/master/google/api/http.proto#L62 for more information.")
	}
}

// Validates if the query parameter has the requested structure.
// This is necessary according to: https://github.com/googleapis/googleapis/blob/master/google/api/http.proto#L118
func validateQueryParameter(field *surface_v1.Field) {
	if !(field.Kind == surface_v1.FieldKind_SCALAR ||
		(field.Kind == surface_v1.FieldKind_ARRAY && openAPIScalarTypes[field.Type]) ||
		(field.Kind == surface_v1.FieldKind_REFERENCE)) {
		log.Println("The query parameter with the Name " + field.Name + " is invalid. " +
			"Note that fields which are mapped to URL query parameters must have a primitive type or" +
			" a repeated primitive type or a non-repeated message type. " +
			"See: https://github.com/googleapis/googleapis/blob/master/google/api/http.proto#L118 for more information.")
	}

}

// Checks whether 't' is a type that will be used as a request parameter for a RPC method.
func isRequestParameter(t *surface_v1.Type) bool {
	if strings.Contains(t.Description, t.GetName()+" holds parameters to") {
		return true
	}
	return false
}

// Sets the Type of 'fd' according to the information from the surface field 'f'.
func setFieldDescriptorType(fd *dpb.FieldDescriptorProto, f *surface_v1.Field) {
	var protoType dpb.FieldDescriptorProto_Type
	if t, ok := protoBufScalarTypes[f.Format]; ok { // Let's see if we can get the type from f.format
		protoType = t
	} else if t, ok := protoBufScalarTypes[f.Type]; ok { // Maybe this works.
		protoType = t
	} else if t, ok := openAPITypesToProtoBuf[f.Type]; ok { // Safety check
		protoType = t
	} else {
		// TODO: What about Enums?
		// Ok, is it either a reference or an array of non scalar-types or a map. All of those get represented as message
		// inside the descriptor.
		protoType = dpb.FieldDescriptorProto_TYPE_MESSAGE
	}
	fd.Type = &protoType

}

// Sets the Name of 'fd'. The convention inside .proto is, that all field names are
// lowercase and all messages and types are capitalized if they are not scalar types (int64, string, ...).
func setFieldDescriptorName(fd *dpb.FieldDescriptorProto, f *surface_v1.Field) {
	name := cleanName(f.Name)
	name = strings.ToLower(name)
	fd.Name = &name
}

// Sets a Label for 'fd'. If it is an array we need the 'repeated' label.
func setFieldDescriptorLabel(fd *dpb.FieldDescriptorProto, f *surface_v1.Field) {
	label := dpb.FieldDescriptorProto_LABEL_OPTIONAL
	if f.Kind == surface_v1.FieldKind_ARRAY || strings.Contains(f.Type, "map") {
		label = dpb.FieldDescriptorProto_LABEL_REPEATED
	}
	fd.Label = &label
}

// Sets the TypeName of 'fd'. A TypeName has to be set if the field is a reference to another message. Otherwise it is nil.
// The convention inside .proto is, that all field names are lowercase and all messages and types are capitalized if
// they are not scalar types (int64, string, ...).
func setFieldDescriptorTypeName(fd *dpb.FieldDescriptorProto, f *surface_v1.Field, packageName string) {
	// A field with a type of Message always has a typeName associated with it (the name of the Message).
	if *fd.Type == dpb.FieldDescriptorProto_TYPE_MESSAGE {
		typeName := packageName + "." + cleanTypeName(f.Type)

		// Check whether we generated this message already inside of another dependency. If so we will use that name instead.
		if n, ok := generatedMessages[f.Type]; ok {
			typeName = n
		}
		fd.TypeName = &typeName
	}
}

// Finds the corresponding surface model type for 'name' and returns the name of the field
// that is a request body. If no such field is found it returns nil.
func getRequestBodyForRequestParameters(name string, types []*surface_v1.Type) *string {
	requestParameterType := &surface_v1.Type{}

	for _, t := range types {
		if t.Name == name {
			requestParameterType = t
		}
	}

	for _, f := range requestParameterType.Fields {
		if f.Position == surface_v1.Position_BODY {
			return &f.Name
		}
	}
	return nil
}

// Constructs a HttpRule from google/api/http.proto. Enables gRPC-HTTP transcoding on 'method'.
// If not nil, body is also set.
func getHttpRuleForMethod(method *surface_v1.Method, body *string) annotations.HttpRule {
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

	if body != nil {
		httpRule.Body = *body
	}

	return httpRule
}

// Returns the type name for the given 'valueType'. A type name for a field is only set if it is some kind of
// reference (non-scalar values) otherwise it is nil.
func getTypeNameForMapValueType(valueType string) *string {
	if _, ok := protoBufScalarTypes[valueType]; ok {
		return nil // Ok it is a scalar. For scalar values we don't set the TypeName of the field.
	}
	if _, ok := openAPIScalarTypes[valueType]; ok {
		return nil // Ok it is a scalar. For scalar values we don't set the TypeName of the field.
	}
	typeName := cleanTypeName(valueType)
	return &typeName
}

// Returns the 'protoType' for the given 'valueType'. If we don't have a scalar 'protoType', we have some kind of
// reference to another object and therefore return the 'Message' type. d
func getProtoTypeForMapValueType(valueType string) *dpb.FieldDescriptorProto_Type {
	protoType := dpb.FieldDescriptorProto_TYPE_MESSAGE
	if protoType, ok := protoBufScalarTypes[valueType]; ok {
		return &protoType
	}
	if _, ok := openAPIScalarTypes[valueType]; ok {
		if protoType, ok := openAPITypesToProtoBuf[valueType]; ok {
			return &protoType
		}
	}
	return &protoType
}

// Uses the 'binaryInput' from gnostic to create a OpenAPI document.
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

// 'url' is a list of URLs to other OpenAPI descriptions. We need the base of all URLs and no duplicates.
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

// Returns true if 's' is inside 'ss'.
func isDuplicate(ss []string, s string) bool {
	for _, s2 := range ss {
		if s == s2 {
			return true
		}
	}
	return false
}

// returns the last FileDescriptorProto of the array 'protos'.
func getLast(protos []*dpb.FileDescriptorProto) *dpb.FileDescriptorProto {
	return protos[len(protos)-1]
}

// A map for this: https://developers.google.com/protocol-buffers/docs/proto3#scalar
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

// Maps OpenAPI data types (https://swagger.io/docs/specification/data-models/data-types/)
// to protobuf data types.
func getOpenAPITypesToProtoBufTypes() map[string]dpb.FieldDescriptorProto_Type {
	return map[string]dpb.FieldDescriptorProto_Type{
		"string":  dpb.FieldDescriptorProto_TYPE_STRING,
		"integer": dpb.FieldDescriptorProto_TYPE_INT32,
		"number":  dpb.FieldDescriptorProto_TYPE_FLOAT,
		"boolean": dpb.FieldDescriptorProto_TYPE_BOOL,
		"object":  dpb.FieldDescriptorProto_TYPE_MESSAGE,
		// Array not set: could be either scalar or non-scalar value.
	}
}

// All scalar types from OpenAPI.
func getOpenAPIScalarTypes() map[string]bool {
	return map[string]bool{
		"string":  true,
		"integer": true,
		"number":  true,
		"boolean": true,
	}
}

// Sets the name of the 'messageDescriptorProto'
func setMessageDescriptorName(messageDescriptorProto *dpb.DescriptorProto, name string) {
	name = cleanTypeName(name)
	messageDescriptorProto.Name = &name
}

// Removes characters which are not allowed for message names or field names inside .proto files.
func cleanName(name string) string {
	name = convertStatusCodes(name)
	name = strings.Replace(name, "-", "_", -1)
	name = strings.Replace(name, " ", "", -1)
	name = strings.Replace(name, "(", "", -1)
	name = strings.Replace(name, ")", "", -1)
	name = strings.Replace(name, "{", "", -1)
	name = strings.Replace(name, "}", "", -1)
	name = strings.Replace(name, "/", "_", -1)
	name = strings.Replace(name, "$", "", -1)
	return name
}

// Since our convention is that all messages inside .proto are capitalized, we set the typeName accordingly.
func cleanTypeName(name string) string {
	name = cleanName(name)
	// Make camelCase
	name = strings.Replace(name, "_", " ", -1)
	name = strings.Title(name)
	name = strings.Replace(name, " ", "", -1)
	return name
}

// Converts a string status code like: "504" into the corresponding text ("Gateway Timeout")
func convertStatusCodes(name string) string {
	code, err := strconv.Atoi(name)
	if err == nil {
		statusText := nethttp.StatusText(code)
		if statusText == "" {
			log.Println("It seems like you have an status code that is currently not known to net.http.StatusText. This might cause the plugin to fail.")
			statusText = "unknownStatusCode"
		}
		name = strings.Replace(statusText, " ", "_", -1)
	}
	return name
}
