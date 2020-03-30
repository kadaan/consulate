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
	"fmt"
	"github.com/hashicorp/consul/sdk/testutil/retry"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"testing"
)

func TestServerCommand(t *testing.T) {
	output := new(bytes.Buffer)
	rootCmd.SetArgs([]string{"server"})
	rootCmd.SetOutput(output)

	log.SetOutput(output)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	go rootCmd.Execute()

	f := &failer{}
	expected := "Started Consulate server on :8080"
	retry.Run(f, func(r *retry.R) {
		result := strconv.Quote(output.String())
		if !strings.Contains(result, expected) {
			r.Error(result)
		}
	})
	if f.failed {
		t.Error("Server did not start correctly")
	}
	defer func() {
		var stopChan = make(chan os.Signal, 1)
		signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	}()
}

type failer struct {
	failed bool
}

func (f *failer) Log(args ...interface{}) { fmt.Println(args...) }
func (f *failer) FailNow()                { f.failed = true }
