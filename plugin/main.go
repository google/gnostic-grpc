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
	"errors"

	plugins "github.com/google/gnostic/plugins"

	"github.com/google/gnostic-grpc/generator"
	"github.com/google/gnostic-grpc/incompatibility"
)

func main() {
	env, err := plugins.NewEnvironment()
	env.RespondAndExitIfError(err)
	switch paramLen := len(env.Request.Parameters); paramLen {
	case 0:
		generator.RunProtoGenerator(env)
	case 1:
		resolveModeFromParameters(env)
	default:
		exitWithMessage(env, "This plugin supports at most one parameter during an invocation")
	}
}

func resolveModeFromParameters(env *plugins.Environment) {
	if env.Request.Parameters[0].Name != "report" {
		exitWithMessage(env, "unsupported parameter name")
	}
	switch env.Request.Parameters[0].Value {
	case "1": // Base incompatibility scanning
		incompatibility.GnosticIncompatibiltyScanning(env, incompatibility.BaseIncompatibility_Report)
	case "2": //Detailed incompatibility scanning
		incompatibility.GnosticIncompatibiltyScanning(env, incompatibility.FileDescriptive_Report)
	default:
		exitWithMessage(env, "unsupported parameter value")
	}
	if len(env.Response.Files) == 0 {
		env.RespondAndExitIfError(errors.New("no supported models for incompatibility reporting"))
	}
	env.RespondAndExit()
}

func exitWithMessage(env *plugins.Environment, errMsg string) {
	env.Response.Errors = append(env.Response.Errors, errMsg)
	env.RespondAndExit()
}
