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
	"reflect"
	"testing"
)

var emptyFunc = func() {}
var intFunc = func(num int) int { return 2 * num }
var stringFunc = func(name string) string { return name }
var tripleArgsFunc = func(_ int, _ int, _ int) string { return "" }
var tripleOutFunc = func() (int, int, int) { return 0, 0, 0 }

func TestGenericGeneration(t *testing.T) {

	var funcCreationTest = []struct {
		function          interface{}
		expOutT           reflect.Type
		expInT            reflect.Type
		expConversionFail bool
	}{
		{emptyFunc, nil, nil, true},
		{intFunc, reflect.TypeOf(0), reflect.TypeOf(0), false},
		{stringFunc, reflect.TypeOf(""), reflect.TypeOf(""), false},
		{tripleArgsFunc, nil, nil, true},
		{tripleOutFunc, nil, nil, true},
	}
	for ind, trial := range funcCreationTest {
		genGeneric, err := makeGenericOperation(trial.function)
		if !trial.expConversionFail {
			typeFlag := trial.expInT != genGeneric.inputType &&
				trial.expOutT != genGeneric.outputType
			errorFlag := err == 1
			existenceFlag := genGeneric.op == nil
			if typeFlag || errorFlag || existenceFlag {
				t.Errorf("Incorrect Generic Function Generation at trial %d", ind)
			}
		} else if err == 0 {
			t.Errorf("Error not reported for Bad Input at trial %d", ind)
		}
	}
}

func TestGenericFunctionInvocation(t *testing.T) {
	var funcCreationTest = []struct {
		function  interface{}
		input     interface{}
		expOutput interface{}
	}{
		{intFunc, 5, 10},
		{stringFunc, "hello", "hello"},
	}
	for ind, trial := range funcCreationTest {
		genGeneric, genErr := makeGenericOperation(trial.function)
		if genErr == 1 {
			t.Errorf("Error reported in Generic Generation at trial %d", ind)
		} else {
			res, errInvoc := genGeneric.op(trial.input)
			if errInvoc == 1 {
				t.Errorf("Error reported in Generic Invocation at trial %d", ind)
			} else if res != trial.expOutput {
				t.Errorf("Error reported in Generic Output comparison at trial %d, expected %v, got %v", ind, trial.expOutput, res)
			}
		}

	}
}
