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

package main

import (
	"github.com/googleapis/gnostic-grpc/generator"
	"github.com/googleapis/gnostic-grpc/incompatibility/incompatibility-report"
	plugins "github.com/googleapis/gnostic/plugins"
)

func main() {
	env, err := plugins.NewEnvironment()
	env.RespondAndExitIfError(err)
	switch paramLen := len(env.Request.Parameters); paramLen {
	case 0: //Run proto Generation
		generator.RunProtoGenerator(env)
	case 1: //Scan for incompatibilities
		lintIncompatibilities(env)
	default:
		exitWithMessage(env, "This plugin only supports at most one parameter during an invocation")
	}
}

func lintIncompatibilities(env *plugins.Environment) {
	if env.Request.Parameters[0].Name != "report" {
		exitWithMessage(env, "unsupported parameter name")
	}
	switch env.Request.Parameters[0].Value {
	case "incomp": // Base incompatibility scanning
		incompatibility.CreateIncompReport(env, incompatibility.BaseIncompatibility_Report)
	case "detailed-incomp": //Detailed incompatibility scanning
		incompatibility.CreateIncompReport(env, incompatibility.ID_Report)
	default:
		exitWithMessage(env, "unsupported parameter value")
	}
}

func exitWithMessage(env *plugins.Environment, errMsg string) {
	env.Response.Errors = append(env.Response.Errors, errMsg)
	env.RespondAndExit()
}
