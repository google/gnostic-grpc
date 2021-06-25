package main

import (
	openapiv3 "github.com/googleapis/gnostic/openapiv3"
	plugins "github.com/googleapis/gnostic/plugins"
	surface "github.com/googleapis/gnostic/surface"
	"google.golang.org/protobuf/proto"
)

// Main function for incomplatibility plugin
func main() {
	env, err := plugins.NewEnvironment()
	env.RespondAndExitIfError(err)
	println("number of models", len(env.Request.Models), "\n")
	for _, model := range env.Request.Models {
		println("model ", model.TypeUrl)
		switch model.TypeUrl {
		case "openapi.v3.Document":
			openAPIdocument := &openapiv3.Document{}
			err := proto.Unmarshal(model.Value, openAPIdocument)
			if err == nil {
				println("converted oasv3!\n")
			}
		case "surface.v1.Model":
			surfaceModel := &surface.Model{}
			err = proto.Unmarshal(model.Value, surfaceModel)
			if err == nil {
				println("converted surface model!\n")
			}
		}

	}
	env.RespondAndExit()

}
