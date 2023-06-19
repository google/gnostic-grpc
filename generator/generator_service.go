package generator

import (
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"
	dpb "github.com/golang/protobuf/protoc-gen-go/descriptor"
	surface_v1 "github.com/google/gnostic/surface"
	"google.golang.org/genproto/googleapis/api/annotations"
)

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

func buildInputTypeAndOutputType(parametersTypeName, responseTypeName string) (inputType, outputType string) {
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

// findValidServiceName finds a valid service name for the gRPC service. A valid service name is not already taken by a
// message. Reference: https://github.com/google/gnostic-grpc/issues/7
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
