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
	"github.com/golang/protobuf/proto"
	dpb "github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugins "github.com/google/gnostic/plugins"
	surface "github.com/google/gnostic/surface"
	prDesc "github.com/jhump/protoreflect/desc"
	prPrint "github.com/jhump/protoreflect/desc/protoprint"
)

// Renderer generates a .proto file based on the information inside Model.
type Renderer struct {
	// The model holds the necessary information from the OpenAPI description.
	Model *surface.Model
	// The FileDescriptorSet that will be printed with protoreflect
	FdSet          *dpb.FileDescriptorSet
	SymbolicFdSets []*dpb.FileDescriptorSet
	Package        string // package name
}

// NewRenderer creates a renderer.
func NewRenderer(model *surface.Model) (renderer *Renderer) {
	renderer = &Renderer{}
	renderer.Model = model
	renderer.SymbolicFdSets = make([]*dpb.FileDescriptorSet, 0)
	return renderer
}

// Render runs the renderer to generate the named files.
func (renderer *Renderer) Render(response *plugins.Response, fileName string) (err error) {
	renderer.FdSet, err = renderer.runFileDescriptorSetGenerator()

	if err != nil {
		return err
	}

	if false { //TODO: If we want to generate the descriptor file, we need an additional flag here!
		f, err := renderer.RenderDescriptor()
		if err != nil {
			return err
		}
		response.Files = append(response.Files, f)
	}

	// Render main proto definition.
	f, err := renderer.RenderProto(renderer.FdSet, fileName)
	if err != nil {
		return err
	}
	response.Files = append(response.Files, f)

	// Render external proto definitions.
	for _, externalSet := range renderer.SymbolicFdSets {
		f, err = renderer.RenderProto(externalSet, *getLast(externalSet.File).Name)
		if err != nil {
			return err
		}
		response.Files = append(response.Files, f)
	}

	return err
}

func (renderer *Renderer) RenderProto(fdSet *dpb.FileDescriptorSet, fileName string) (*plugins.File, error) {
	// Creates a protoreflect FileDescriptor, which is then used for printing.
	prFd, err := prDesc.CreateFileDescriptorFromSet(fdSet)
	if err != nil {
		return nil, err
	}

	// Print the protoreflect FileDescriptor.
	p := prPrint.Printer{}
	res, err := p.PrintProtoToString(prFd)
	if err != nil {
		return nil, err
	}

	f := NewLineWriter()
	f.WriteLine(res)

	file := &plugins.File{Name: fileName}
	file.Data = f.Bytes()

	return file, err
}

func (renderer *Renderer) RenderDescriptor() (*plugins.File, error) {
	fdSetData, err := proto.Marshal(renderer.FdSet)
	if err != nil {
		return nil, err
	}

	descriptorFile := &plugins.File{Name: renderer.Package + ".descr"}
	descriptorFile.Data = fdSetData
	return descriptorFile, nil
}
