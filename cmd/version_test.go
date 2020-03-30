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

package cmd

import (
	"bytes"
	"strconv"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	output := new(bytes.Buffer)
	rootCmd.SetArgs([]string{"version"})
	rootCmd.SetOutput(output)
	e := rootCmd.Execute()
	if e != nil {
		t.Errorf("Version command failed with: %v", e)
	}
	expected := strconv.Quote(`Consulate, version  (branch: , revision: )
  build user:       
  build date:       
  go version:       go1.14.1`)
	result := strconv.Quote(output.String())
	if result != expected {
		t.Errorf("Version command: want '%v', got '%s'", expected, result)
	}
}
