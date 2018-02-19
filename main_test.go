// Copyright © 2018 Joel Baranick <jbaranick@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Update golden file by passing `-test.update` to program arguments
var update = flag.Bool("test.update", false, "update .golden file")

func TestMainFunction(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		main()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestMainFunction")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	actual, err := cmd.Output()
	if err != nil {
		e, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("Command error could not be parsed as ExitError: %v", err)
		}
		if !e.Success() {
			t.Fatalf("ExitError: want 0, got %v", e.String())
		}
	}
	expected := get(t, []byte(actual))
	actualString := string(actual)
	expectedString := string(expected)
	if !strings.HasPrefix(string(actual), string(expected)) {
		t.Fatalf("Output ==> \n    want:\n%v\n\n    got\n%v", expectedString, actualString)
	}
}

func get(t *testing.T, actual []byte) []byte {
	golden := filepath.Join("testdata", t.Name()+".golden")
	if *update {
		if err := ioutil.WriteFile(golden, actual, 0644); err != nil {
			t.Fatal(err)
		}
	}
	expected, err := ioutil.ReadFile(golden)
	if err != nil {
		t.Fatal(err)
	}
	return expected
}
