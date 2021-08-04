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
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/googleapis/gnostic-grpc/search"
	"github.com/googleapis/gnostic-grpc/utils"
	openapiv3 "github.com/googleapis/gnostic/openapiv3"
	plugins "github.com/googleapis/gnostic/plugins"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
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
		//Parse openapidoc
		openAPIdocument := &openapiv3.Document{}
		err := proto.Unmarshal(model.Value, openAPIdocument)
		env.RespondAndExitIfError(err)

		//Generate Base Incompatibility Report
		incompatibilityReport := ScanIncompatibilities(openAPIdocument, env.Request.SourceName)

		//Write Report to File
		switch reportType {
		case BaseIncompatibility_Report:
			writeProtobufMessage(incompatibilityReport, env)
		case ID_Report:
			//TODO once branches are merged
			idReport := detailReport(incompatibilityReport)
			writeProtobufMessage(idReport, env)
		}
		env.RespondAndExit()
	}
	env.RespondAndExitIfError(errors.New("no supported models for incompatibility reporting"))
}

func writeProtobufMessage(m protoreflect.ProtoMessage, env *plugins.Environment) {
	reportBytes, err := utils.ProtoTextBytes(m)
	env.RespondAndExitIfError(err)
	createdFile := &plugins.File{
		Name: trimSourceName(env.Request.SourceName) + "_compatibility.pb",
		Data: reportBytes,
	}
	env.Response.Files = append(env.Response.Files, createdFile)
}

// creates an IDReport from a base report
func detailReport(incompatibilityReport *IncompatibilityReport) *IDReport {
	var idReport *IDReport
	var incompatibilities []*IncompatibilityDescription
	fileNode, parseErr := search.MakeNode(incompatibilityReport.ReportIdentifier)
	if parseErr != nil {
		log.Printf("FATAL: unable to parse file at %s with error %s", incompatibilityReport.ReportIdentifier, parseErr)
		return nil
	}
	for _, baseincomp := range incompatibilityReport.Incompatibilities {
		line, col, searchErr := search.FindKey(fileNode.Content[0], baseincomp.TokenPath...)
		if searchErr != nil {
			log.Printf("Warning: Unable to find incompatibilty %s", searchErr.Error())
		}
		lastTokenInPath := baseincomp.TokenPath[len(baseincomp.TokenPath)-1]
		incompatibilities = append(incompatibilities,
			newIncompatibilityDescription(line, col, baseincomp.Classification, lastTokenInPath))
	}
	idReport = newIDReport(incompatibilityReport.ReportIdentifier, incompatibilities)
	return idReport
}

// leaves base file name without any extenstions
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

func newIDReport(reportIdentifier string, incompDescriptions []*IncompatibilityDescription) *IDReport {
	return &IDReport{
		ReportIdentifier:  reportIdentifier,
		Incompatibilities: incompDescriptions,
	}
}

func newIncompatibility(classification IncompatibiltiyClassification, path ...string) *Incompatibility {
	return &Incompatibility{
		TokenPath:      path,
		Classification: classification,
		Severity:       classificationSeverity(classification),
	}
}

func newIncompatibilityDescription(line int, column int, class IncompatibiltiyClassification, lastToken string) *IncompatibilityDescription {
	return &IncompatibilityDescription{
		Line:   int32(line),
		Column: int32(column),
		Hint:   classificationHint(class),
		Class:  class,
		Token:  lastToken,
	}
}

// returns a severity level based on the given classification
func classificationSeverity(classification IncompatibiltiyClassification) Severity {
	var severityLevel Severity
	switch classification {
	case IncompatibiltiyClassification_IncompatibiltiyClassification_Default:
		severityLevel = Severity_INFO
	case IncompatibiltiyClassification_Security,
		IncompatibiltiyClassification_ParameterStyling,
		IncompatibiltiyClassification_DataValidation,
		IncompatibiltiyClassification_ExternalTranscodingSupport:
		severityLevel = Severity_WARNING
	case IncompatibiltiyClassification_InvalidOperation,
		IncompatibiltiyClassification_InvalidDataState,
		IncompatibiltiyClassification_Inheritance:
		severityLevel = Severity_FAIL
	default:
		severityLevel = Severity_Severity_Default
	}
	return severityLevel
}

// returns a hint based the given classification
func classificationHint(classification IncompatibiltiyClassification) string {
	rootHint := fmt.Sprintf("%s incompatibilities occur as a result of ",
		classification.Enum().String())
	var reason string
	switch classification {
	case IncompatibiltiyClassification_Security:
		reason = "gRPC HTTP/JSON transcoding not concerned with auth information."
	case IncompatibiltiyClassification_ParameterStyling:
		reason = "parameter styling not representable in .proto files."
	case IncompatibiltiyClassification_DataValidation:
		reason = "dataValidation (regex, array limits, etc.) not natively supported in .proto files."
	case IncompatibiltiyClassification_ExternalTranscodingSupport:
		reason = "the need for external transcoding support outside of .proto files"
	case IncompatibiltiyClassification_InvalidOperation:
		reason = "unstandard operation not fundamentally and truly supported in .proto represenation."
	case IncompatibiltiyClassification_InvalidDataState:
		reason = "data state(nullable) not representable in .proto files."
	case IncompatibiltiyClassification_Inheritance:
		reason = "Inheritance not supported in .proto files."
	default:
		return "No hint for " + classification.Enum().String()
	}
	severityRoot := fmt.Sprintf(" %s implies ",
		classificationSeverity(classification).Enum().String())
	var implication string
	switch classificationSeverity(classification) {
	case Severity_INFO:
		implication = "information not important to core api representation."
	case Severity_WARNING:
		implication = "exclusion of this feature in .proto representation removes component of api " +
			"reprsentation but lack of this feature in .proto file does not solely imply " +
			"inability of a functional transcoding environment."
	case Severity_FAIL:
		implication = "fundamental api representation feature lacks support in .proto files."
	default:
		return "No hint for " + classification.Enum().String()
	}
	return rootHint + reason + severityRoot + implication
}
