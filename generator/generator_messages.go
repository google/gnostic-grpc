package generator

import (
	"log"
	"strings"

	dpb "github.com/golang/protobuf/protoc-gen-go/descriptor"
	surface_v1 "github.com/google/gnostic/surface"
	"google.golang.org/protobuf/types/descriptorpb"
)

var protoBufScalarTypes = getProtobufTypes()

// Gathers all messages that have been generated from symbolic references in recursive calls.
var generatedMessages = make(map[string]string, 0)

// isScalarType find if the current type is a scalar type (one type and one field with format)
func getType(types []*surface_v1.Type, name string) *surface_v1.Type {
	for _, ts := range types {
		if ts.TypeName == name {
			return ts
		}
	}
	return nil
}
func isScalarType(surfaceType *surface_v1.Type) bool {
	return surfaceType != nil &&
		len(surfaceType.Fields) == 1 &&
		surfaceType.Fields[0].Name == "value" &&
		surfaceType.Fields[0].Position != surface_v1.Position_QUERY &&
		surfaceType.Fields[0].Position != surface_v1.Position_PATH &&
		!strings.Contains(strings.ToLower(surfaceType.Name), "nullable") &&
		surfaceType.Fields[0].EnumValues == nil &&
		surfaceType.Fields[0].Kind == surface_v1.FieldKind_SCALAR
}

func wrapperType(t string) string {
	switch t {
	case "string":
		return "google.protobuf.StringValue"
	case "integer":
		return "google.protobuf.Int64Value"
	case "double":
		return "google.protobuf.DoubleValue"
	case "float":
		return "google.protobuf.FloatValue"
	case "uint64":
		return "google.protobuf.UInt64Value"
	case "int32":
		return "google.protobuf.Int32Value"
	case "uint32":
		return "google.protobuf.UInt32Value"
	case "boolean":
		return "google.protobuf.BoolValue"
	case "bytes":
		return "google.protobuf.BytesValue"
	case "arrayString":
		return "ArrayString"

	default:
		return t
	}
}

// buildAllMessageDescriptors builds protobuf messages from the surface model types. If the type is a RPC request parameter
// the fields have to follow certain rules, and therefore have to be validated.
func buildAllMessageDescriptors(renderer *Renderer) (messageDescriptors []*dpb.DescriptorProto, err error) {
	for _, surfaceType := range renderer.Model.Types {
		message := &dpb.DescriptorProto{}
		message.Name = &surfaceType.TypeName

		if surfaceType.ContentType == "ANY_OF" || surfaceType.ContentType == "ONE_OF" {
			message.OneofDecl = []*dpb.OneofDescriptorProto{
				{Name: &surfaceType.Name},
			}
		}
		for i, surfaceField := range surfaceType.Fields {
			var oneOfIndex *int32
			if surfaceType.ContentType == "ANY_OF" || surfaceType.ContentType == "ONE_OF" {
				b32 := int32(0)
				oneOfIndex = &b32
			}
			format := ""
			prefix := true
			if strings.Contains(surfaceField.NativeType, "map[string][]") {
				// Not supported for now: https://github.com/LorenzHW/gnostic-grpc-deprecated/issues/3#issuecomment-509348357
				continue
			}
			if isRequestParameter(surfaceType) {
				switch surfaceField.Position {
				case surface_v1.Position_PATH:
					if surfaceField.Kind == surface_v1.FieldKind_REFERENCE {
						for _, ts := range renderer.Model.Types {
							if ts.TypeName == surfaceField.Type {
								surfaceField.Name = ts.Fields[0].Name
								surfaceField.FieldName = ts.Fields[0].Name
								surfaceField.NativeType = "string"
								//format = ts.Fields[0].Format
							}
						}
					} else {
						surfaceField.Type = "string"
						surfaceField.NativeType = "string"
						//format = surfaceField.Format
					}
				case surface_v1.Position_QUERY:
					if ts := getType(renderer.Model.Types, surfaceField.Type); ts != nil {
						if ts.Fields[0].Type == "arrayString" {
							format = surfaceField.Type
							surfaceField.Type = "string"
							surfaceField.NativeType = "string"
							surfaceField.Kind = surface_v1.FieldKind_ARRAY
							surfaceField.Name = ts.Fields[0].Name
							surfaceField.FieldName = ts.Fields[0].Name
							//format = ts.Fields[0].Format
						} else {
							surfaceField.Name = ts.Fields[0].Name
							surfaceField.FieldName = ts.Fields[0].Name
							surfaceField.NativeType = wrapperType(ts.Fields[0].Type)
							prefix = false
							//format = ts.Fields[0].Format
						}
					}
				}
			} else {
				if ts := getType(renderer.Model.Types, surfaceField.NativeType); ts != nil && isScalarType(ts) {
					surfaceField.NativeType = ts.Fields[0].NativeType
					surfaceField.Format = ts.Fields[0].Format
				}
				format = surfaceField.Format
			}

			addFieldDescriptor(message, surfaceField, i, renderer.Package, format, prefix, oneOfIndex)
			addEnumDescriptorIfNecessary(message, surfaceField)
		}

		messageDescriptors = append(messageDescriptors, message)
		generatedMessages[*message.Name] = renderer.Package + "." + *message.Name
	}
	return messageDescriptors, nil
}

// isRequestParameter checks whether 't' is a type that will be used as a request parameter for a RPC method.
func isRequestParameter(sufaceType *surface_v1.Type) bool {
	if strings.Contains(sufaceType.Description, sufaceType.GetName()+" holds parameters to") {
		return true
	}
	return false
}

func validateRequestParameter(field *surface_v1.Field) bool {
	if field.Position == surface_v1.Position_PATH {
		return validatePathParameter(field)
	}

	if field.Position == surface_v1.Position_QUERY {
		return validateQueryParameter(field)
	}
	return true
}

// validatePathParameter validates if the path parameter has the requested structure.
// This is necessary according to: https://github.com/googleapis/googleapis/blob/master/google/api/http.proto#L62
func validatePathParameter(field *surface_v1.Field) bool {
	if field.Kind != surface_v1.FieldKind_SCALAR {
		log.Println("The path parameter with the Name " + field.Name + " is invalid. " +
			"The path template may refer to one or more fields in the gRPC request message, as" +
			" long as each field is a non-repeated field with a primitive (non-message) type. " +
			"See: https://github.com/googleapis/googleapis/blob/master/google/api/http.proto#L62 for more information.")
		return false
	}
	return true
}

// validateQueryParameter validates if the query parameter has the requested structure.
// This is necessary according to: https://github.com/googleapis/googleapis/blob/master/google/api/http.proto#L118
func validateQueryParameter(field *surface_v1.Field) bool {
	_, isScalar := protoBufScalarTypes[field.NativeType]
	if !(field.Kind == surface_v1.FieldKind_SCALAR ||
		(field.Kind == surface_v1.FieldKind_ARRAY && isScalar) ||
		(field.Kind == surface_v1.FieldKind_REFERENCE)) {
		log.Println("The query parameter with the Name " + field.Name + " is invalid. " +
			"Note that fields which are mapped to URL query parameters must have a primitive type or" +
			" a repeated primitive type or a non-repeated message type. " +
			"See: https://github.com/googleapis/googleapis/blob/master/google/api/http.proto#L118 for more information.")
		return false
	}
	return true
}

func addFieldDescriptor(message *dpb.DescriptorProto, surfaceField *surface_v1.Field, idx int, packageName, format string, prefix bool, oneOfIndex *int32) {
	count := int32(idx + 1)
	fieldDescriptor := &dpb.FieldDescriptorProto{Number: &count, Name: &surfaceField.FieldName}
	fieldDescriptor.Type = getFieldDescriptorType(surfaceField.NativeType, surfaceField.EnumValues)
	fieldDescriptor.Label = getFieldDescriptorLabel(surfaceField)
	fieldDescriptor.OneofIndex = oneOfIndex
	fieldDescriptor.TypeName = getFieldDescriptorTypeName(*fieldDescriptor.Type, surfaceField, packageName, prefix)
	fieldDescriptor.Options = &dpb.FieldOptions{
		UninterpretedOption: []*dpb.UninterpretedOption{
			//{
			//	Name: []*dpb.UninterpretedOption_NamePart{
			//		{NamePart: ptr("(validator.field)")},
			//	},
			//	AggregateValue: ptr("{int_gt: 0, int_lt: 100}"),
			//},
			{
				Name: []*dpb.UninterpretedOption_NamePart{
					{NamePart: ptr("json_name")},
				},
				AggregateValue: ptr(`"` + surfaceField.Name + `"]; // @gotags: json:"` + surfaceField.Name + `" format:"` + format + `" `),
			},
		},
	}
	if format != "" {
	}
	addMapDescriptorIfNecessary(surfaceField, fieldDescriptor, message)

	message.Field = append(message.Field, fieldDescriptor)
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

// getFieldDescriptorTypeName returns the typeName of the descriptor. A TypeName has to be set if the field is a reference to another
// descriptor or enum. Otherwise it is nil. Names are set according to the protocol buffer style guide for message names:
// https://developers.google.com/protocol-buffers/docs/style#message-and-field-names
func getFieldDescriptorTypeName(fieldDescriptorType descriptorpb.FieldDescriptorProto_Type, field *surface_v1.Field, packageName string, prefix bool) *string {
	if fieldHasAReferenceToAMessageInAnotherDependency(field, fieldDescriptorType) {
		t := generatedMessages[field.NativeType]
		return &t
	}

	typeName := ""
	if fieldDescriptorType == dpb.FieldDescriptorProto_TYPE_MESSAGE {
		if prefix {
			typeName = packageName + "." + field.NativeType
		} else {
			typeName = field.NativeType
		}
	}
	if fieldDescriptorType == dpb.FieldDescriptorProto_TYPE_ENUM {
		typeName = field.NativeType
	}
	return &typeName
}

// fieldHasAReferenceToAMessageInAnotherDependency check whether we generated this message already inside of another
// dependency. If so we will use that name instead.
func fieldHasAReferenceToAMessageInAnotherDependency(field *surface_v1.Field, fieldDescriptorType descriptorpb.FieldDescriptorProto_Type) bool {
	_, messageExists := generatedMessages[field.NativeType]
	return fieldDescriptorType == dpb.FieldDescriptorProto_TYPE_MESSAGE && messageExists
}

// getFieldDescriptorLabel returns the label for the descriptor based on the information in he surface field.
func getFieldDescriptorLabel(f *surface_v1.Field) *dpb.FieldDescriptorProto_Label {
	label := dpb.FieldDescriptorProto_LABEL_OPTIONAL
	if f.Kind == surface_v1.FieldKind_ARRAY || strings.Contains(f.NativeType, "map") {
		label = dpb.FieldDescriptorProto_LABEL_REPEATED
	}
	return &label
}

func addMapDescriptorIfNecessary(f *surface_v1.Field, fieldDescriptor *dpb.FieldDescriptorProto, message *dpb.DescriptorProto) {
	if f.Kind == surface_v1.FieldKind_MAP {
		// Maps are represented as nested types inside of the descriptor.
		mapDescriptor := buildMapDescriptor(f)
		fieldDescriptor.TypeName = mapDescriptor.Name
		message.NestedType = append(message.NestedType, mapDescriptor)
	}
}

func ptr(s string) *string {
	return &s
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

// getTypeNameForMapValueType returns the type name for the given 'valueType'.
// A type name for a field is only set if it is some kind of reference (non-scalar values) otherwise it is nil.
func getTypeNameForMapValueType(valueType string) *string {
	if _, ok := protoBufScalarTypes[valueType]; ok {
		return nil // Ok it is a scalar. For scalar values we don't set the TypeName of the field.
	}
	typeName := valueType
	return &typeName
}

func addEnumDescriptorIfNecessary(message *dpb.DescriptorProto, f *surface_v1.Field) {
	if f.EnumValues != nil {
		message.EnumType = append(message.EnumType, buildEnumDescriptorProto(f))
	}
}

// buildEnumDescriptorProto builds the necessary descriptor to render a enum. (https://developers.google.com/protocol-buffers/docs/proto3#enum)
func buildEnumDescriptorProto(f *surface_v1.Field) *dpb.EnumDescriptorProto {
	enumDescriptor := &dpb.EnumDescriptorProto{Name: &f.NativeType}
	for enumCtr, value := range f.EnumValues {
		num := int32(enumCtr)
		name := value
		valueDescriptor := &dpb.EnumValueDescriptorProto{
			Name:   &name,
			Number: &num,
		}
		enumDescriptor.Value = append(enumDescriptor.Value, valueDescriptor)
	}
	return enumDescriptor
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
