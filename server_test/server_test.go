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

package server_test

import (
	"github.com/kadaan/consulate/checks"
	"github.com/kadaan/consulate/spi"
	"github.com/kadaan/consulate/testutil"
	"io/ioutil"
	"testing"
)

var apiTests = []apiTestData{
	{"/about", spi.StatusOK, `{"Version":"","Revision":"","Branch":"","BuildUser":"","BuildDate":"","GoVersion":"go1.9.3"}`},
	{"/health", spi.StatusOK, `{"Status":"Ok"}`},
	{"/verify/checks?status=passing", spi.StatusOK, `{"Status":"Ok"}`},
}

type apiTestData struct {
	path       string
	statusCode int
	body       string
}

func TestApi(t *testing.T) {
	//if testing.Verbose() {
	t.Log("Starting API tests...")
	//}
	server := newServer(t)
	defer server.Stop()

	server.AddService("service1", []string{})
	server.AddCheck("check1a", "check 1", "service1", checks.HealthPassing)

	for _, d := range apiTests {
		t.Logf("  --> %s", d.path)
		verifyApiCall(t, server, d)
	}
	//if testing.Verbose() {
	t.Log("Finished API tests")
	//}
}

func verifyApiCall(t *testing.T, s *testutil.WrappedTestServer, d apiTestData) {
	r, err := s.Client().Get(s.Url(d.path))
	if r != nil && r.Body != nil {
		defer r.Body.Close()
	}
	if err != nil {
		t.Error(err)
	}
	if r.StatusCode != d.statusCode {
		t.Errorf("%q => StatusCode: %v, want %v", d.path, r.StatusCode, d.statusCode)
	}
	bb, err := ioutil.ReadAll(r.Body)
	if err != nil {
		t.Error(err)
	}
	body := string(bb)
	if body != d.body {
		t.Errorf("%q => Body: %q, want %q", d.path, body, d.body)
	}
}

func newServer(t *testing.T) *testutil.WrappedTestServer {
	server, err := testutil.NewTestServer(t)
	if err != nil {
		if server != nil {
			defer server.Stop()
		}
		t.Fatalf("Failed to start test consulate server: %v", err)
	}
	return server.Wrap(t)
}
