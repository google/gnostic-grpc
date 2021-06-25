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

package incompatibility

import (
	"github.com/googleapis/gnostic-grpc/incompatibility/utils"
	// openapiv3 "github.com/googleapis/gnostic/openapiv3"
	"testing"
)

// Simple check for catching server incompatibility
func ServerIncompCheck(t *testing.T, expectingServer bool, fileName string) {
	document, err := utils.ParseOpenAPIDoc(fileName)
	if err != nil {
		t.Errorf("Error while parsing input file: %s", fileName)
		return
	}
	incomp := getIncompatibilites(document).Incompatibilities
	foundServer := len(incomp) != 0 && incomp[0].Classification == "SERVERS"
	if expectingServer && !foundServer {
		t.Error("Failed to report servers incompatibility")
	} else if !expectingServer && foundServer {
		t.Error("Reported false servers incompatibility")
	}
}

func TestBasic(t *testing.T) {
	noServerPath := "../../generator/testfiles/other.yaml"
	ServerIncompCheck(t, false, noServerPath)
	serverPath := "../oas-examples/petstore.yaml"
	ServerIncompCheck(t, true, serverPath)
}
