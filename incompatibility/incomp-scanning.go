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

package incompatibility

import (
	"errors"
	"path/filepath"
	"strings"

	openapiv3 "github.com/googleapis/gnostic/openapiv3"
	plugins "github.com/googleapis/gnostic/plugins"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

type Report int

const (
	BaseIncompatibility_Report = iota
	ID_Report
)

// Create a report of incompatibilities, write protobuf message to
// environment, bool indicates if detailed incompatibility report is
// desired.
func CreateIncompReport(env *plugins.Environment, reportType Report) {

	// Generate Base Incompatibility Report

	// If indicated by report type associate incompatibilities with line
	// references, etc. and generate an ID_REPORT
	for _, model := range env.Request.Models {
		if model.TypeUrl != "openapi.v3.Document" {
			continue
		}
		openAPIdocument := &openapiv3.Document{}
		err := proto.Unmarshal(model.Value, openAPIdocument)
		env.RespondAndExitIfError(err)
		incompatibilityReport := ScanIncompatibilities(openAPIdocument, env.Request.SourceName)
		writeProtobufMessage(incompatibilityReport, env)
		env.RespondAndExit()
	}
	env.RespondAndExitIfError(errors.New("no supported models for incompatibility reporting"))
}

func writeProtobufMessage(incompatibilityReport *IncompatibilityReport, env *plugins.Environment) {
	incompatibilityReportBytes, err :=
		prototext.MarshalOptions{Multiline: true, Indent: "    "}.
			Marshal(incompatibilityReport)
	env.RespondAndExitIfError(err)
	createdFile := &plugins.File{
		Name: trimSourceName(env.Request.SourceName) + "_compatibility.pb",
		Data: incompatibilityReportBytes,
	}
	env.Response.Files = append(env.Response.Files, createdFile)
}

func trimSourceName(pathWithExtension string) string {
	fileNameWithExtension := filepath.Base(pathWithExtension)
	if extInd := strings.IndexByte(fileNameWithExtension, '.'); extInd != -1 {
		return fileNameWithExtension[:extInd]
	}
	return pathWithExtension
}

// Scan for incompatibilities in an OpenAPI document
func ScanIncompatibilities(document *openapiv3.Document, reportIdentifier string) *IncompatibilityReport {
	return ReportOnDoc(document, reportIdentifier, IncompatibilityReporters...)
}
