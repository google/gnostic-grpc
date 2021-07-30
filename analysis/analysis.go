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

// This is an aggregation tool which scans for and aggregates incompatibilities
// in OpenAPI documents within a given directory. The only argument given to this
// tool should be the intended directory.

package main

import (
	"errors"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/googleapis/gnostic-grpc/incompatibility"
	"github.com/googleapis/gnostic-grpc/utils"
	"google.golang.org/protobuf/encoding/prototext"
)

// main function for aggreation tool
func main() {
	if len(os.Args) != 2 {
		exitIfError(errors.New("argument should be a path to a directory"))
	}
	generateAnalysis(os.Args[1])
}

// runs analysis on given directory
func generateAnalysis(dirPath string) *incompatibility.ApiSetIncompatibility {
	analysisAggregation := incompatibility.NewAnalysis()
	readingDirectoryErr := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Printf("walk error for file at %s", path)
		}
		if d.IsDir() {
			return nil
		}
		newAnalysis, analysisErr := fileHandler(path)
		if analysisErr != nil {
			log.Printf("unable to produce analysis for file %s with error <%s>", path, analysisErr.Error())
		} else {
			analysisAggregation = incompatibility.AggregateAnalysis(analysisAggregation, newAnalysis)
		}
		return nil
	})
	if readingDirectoryErr != nil {
		exitIfError(errors.New("unable to walk through directory"))
	}
	println(prototext.Format(analysisAggregation))
	return analysisAggregation
}

// fileHander attempts to parse the file at path to then create an incompatibility report
func fileHandler(path string) (*incompatibility.ApiSetIncompatibility, error) {
	openAPIDoc, err := utils.ParseOpenAPIDoc(path)
	if err != nil {
		return nil, err
	}
	incompatibilityReport := incompatibility.ScanIncompatibilities(openAPIDoc)
	log.Printf("created incompatibility report for file at %s\n", path)
	return incompatibility.FormAnalysis(incompatibilityReport, path), nil
}

func exitIfError(e error) {
	if e == nil {
		return
	}
	log.Printf("Exiting with error: %s\n", e)
	os.Exit(1)
}
