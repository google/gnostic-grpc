// Copyright 2021 Google Inc. All Rights Reserved.
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

package utils

import (
	"os/exec"

	openapiv3 "github.com/googleapis/gnostic/openapiv3"
	"google.golang.org/protobuf/proto"
)

func ParseOpenAPIDoc(input string) (*openapiv3.Document, error) {
	cmd := exec.Command("gnostic", "--pb-out=-", input)
	b, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	documentv3, err := CreateOpenAPIDocFromGnosticOutput(b)
	return documentv3, err
}

func CreateOpenAPIDocFromGnosticOutput(binaryInput []byte) (*openapiv3.Document, error) {
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
