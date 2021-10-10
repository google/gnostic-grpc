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
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	surface "github.com/google/gnostic/surface"

	"github.com/google/gnostic-grpc/utils"
)

const (
	// When false, test behaves normally, checking output against golden test files.
	// But when changed to true, running test will actually re-generate golden test
	// files (which assumes output is correct).
	regenerateMode = false

	testFilesDirectory = "testfiles"
)

func TestFileDescriptorGeneratorParameters(t *testing.T) {
	input := "testfiles/parameters.yaml"

	protoData, err := runGeneratorWithoutPluginEnvironment(input, "parameters")
	if err != nil {
		handleError(err, t)
	}

	checkContents(t, string(protoData), "goldstandard/parameters.proto")
}

func TestFileDescriptorGeneratorRequestBodies(t *testing.T) {
	input := "testfiles/requestBodies.yaml"

	protoData, err := runGeneratorWithoutPluginEnvironment(input, "requestbodies")
	if err != nil {
		handleError(err, t)
	}

	checkContents(t, string(protoData), "goldstandard/requestbodies.proto")

}

func TestFileDescriptorGeneratorResponses(t *testing.T) {
	input := "testfiles/responses.yaml"

	protoData, err := runGeneratorWithoutPluginEnvironment(input, "responses")
	if err != nil {
		handleError(err, t)
	}
	checkContents(t, string(protoData), "goldstandard/responses.proto")
}

func TestFileDescriptorGeneratorOther(t *testing.T) {
	input := "testfiles/other.yaml"

	protoData, err := runGeneratorWithoutPluginEnvironment(input, "other")
	if err != nil {
		handleError(err, t)
	}
	checkContents(t, string(protoData), "goldstandard/other.proto")

	erroneousInput := []string{"testfiles/errors/cyclic_dependency_1.yaml"}

	for _, errorInput := range erroneousInput {
		errorMessages := map[string]bool{
			"cycle in imports: cyclic_dependency_2.proto -> cyclic_dependency_1.proto -> cyclic_dependency_2.proto": true,
		}
		protoData, err = runGeneratorWithoutPluginEnvironment(errorInput, "cyclic_dependency_1")
		if _, ok := errorMessages[err.Error()]; !ok {
			// If we don't get an error from the generator the test fails!
			handleError(err, t)
		}
	}
}

func runGeneratorWithoutPluginEnvironment(input string, packageName string) ([]byte, error) {
	surfaceModel, err := buildSurfaceModel(input)
	if err != nil {
		return nil, err
	}
	NewProtoLanguageModel().Prepare(surfaceModel, "openapi.v3.Document")
	r := NewRenderer(surfaceModel)
	r.Package = packageName

	fdSet, err := r.runFileDescriptorSetGenerator()
	r.FdSet = fdSet
	if err != nil {
		return nil, err
	}
	f, err := r.RenderProto(fdSet, "")
	if err != nil {
		return nil, err
	}
	return f.Data, err
}

func buildSurfaceModel(input string) (*surface.Model, error) {
	documentv3, err := utils.ParseOpenAPIDoc(input)
	if err != nil {
		return nil, err
	}
	surfaceModel, err := surface.NewModelFromOpenAPI3(documentv3, input)
	return surfaceModel, err
}

func writeFile(output string, protoData []byte) {
	dir := path.Dir(output)
	os.MkdirAll(dir, 0755)
	f, _ := os.Create(output)
	defer f.Close()
	f.Write(protoData)
}

func checkContents(t *testing.T, actualContents string, goldenFileName string) {
	goldenFileName = filepath.Join(testFilesDirectory, goldenFileName)

	if regenerateMode {
		writeFile(goldenFileName, []byte(actualContents))
	}

	// verify that output matches golden test files
	b, err := ioutil.ReadFile(goldenFileName)
	if err != nil {
		t.Errorf("Error while reading goldstandard file")
		t.Errorf(err.Error())
	}
	goldstandard := string(b)
	if goldstandard != actualContents {
		t.Errorf("File contents does not match.")
	}
}

func handleError(err error, t *testing.T) {
	t.Errorf("Error while executing the protoc-generator")
	if strings.Contains(err.Error(), "included an unresolvable reference") {
		t.Errorf("This could be due to the fact that 'typeName' is set wrong on a FieldDescriptorProto." +
			"For every FieldDescriptorProto where the type == 'FieldDescriptorProto_TYPE_MESSAGE' the correct typeName needs to be set.")
	}
	t.Errorf(err.Error())
}

// Sometimes I need
//func buildFdsetFromProto() {
//	b, err := ioutil.ReadFile("temp.descr")
//	if err != nil {
//		fmt.Print(err.Error())
//	}
//	fdSet := &descriptor.FileDescriptorSet{}
//	err = proto.Unmarshal(b, fdSet)
//	if err != nil {
//		fmt.Print(err.Error())
//	}
//}
