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
	"bytes"
	"github.com/kadaan/consulate/checks"
	"github.com/kadaan/consulate/spi"
	"github.com/kadaan/consulate/testutil"
	"io/ioutil"
	"testing"
	"text/template"
)

var apiTests = []apiTestData{
	{"/about", spi.StatusOK, `{"Version":"","Revision":"","Branch":"","BuildUser":"","BuildDate":"","GoVersion":"go1.9.3"}`},
	{"/health", spi.StatusOK, `{"Status":"Ok"}`},
	{"/verify/checks", spi.StatusCheckError, `{"Status":"Failed","Checks":{"check1b":{"Node":"{{.ConsulNodeName}}","CheckID":"check1b","Name":"check 1","Status":"warning","Notes":"","Output":"","ServiceID":"service1","ServiceName":"service1","ServiceTags":[],"Definition":{"HTTP":"","Header":null,"Method":"","TLSSkipVerify":false,"TCP":"","Interval":0,"Timeout":0,"DeregisterCriticalServiceAfter":0},"CreateIndex":0,"ModifyIndex":0},"check1c":{"Node":"{{.ConsulNodeName}}","CheckID":"check1c","Name":"check 1","Status":"critical","Notes":"","Output":"","ServiceID":"service1","ServiceName":"service1","ServiceTags":[],"Definition":{"HTTP":"","Header":null,"Method":"","TLSSkipVerify":false,"TCP":"","Interval":0,"Timeout":0,"DeregisterCriticalServiceAfter":0},"CreateIndex":0,"ModifyIndex":0}}}`},
	{"/verify/checks?status=passing", spi.StatusCheckError, `{"Status":"Failed","Checks":{"check1b":{"Node":"{{.ConsulNodeName}}","CheckID":"check1b","Name":"check 1","Status":"warning","Notes":"","Output":"","ServiceID":"service1","ServiceName":"service1","ServiceTags":[],"Definition":{"HTTP":"","Header":null,"Method":"","TLSSkipVerify":false,"TCP":"","Interval":0,"Timeout":0,"DeregisterCriticalServiceAfter":0},"CreateIndex":0,"ModifyIndex":0},"check1c":{"Node":"{{.ConsulNodeName}}","CheckID":"check1c","Name":"check 1","Status":"critical","Notes":"","Output":"","ServiceID":"service1","ServiceName":"service1","ServiceTags":[],"Definition":{"HTTP":"","Header":null,"Method":"","TLSSkipVerify":false,"TCP":"","Interval":0,"Timeout":0,"DeregisterCriticalServiceAfter":0},"CreateIndex":0,"ModifyIndex":0}}}`},
	{"/verify/checks?status=warning", spi.StatusCheckError, `{"Status":"Failed","Checks":{"check1c":{"Node":"{{.ConsulNodeName}}","CheckID":"check1c","Name":"check 1","Status":"critical","Notes":"","Output":"","ServiceID":"service1","ServiceName":"service1","ServiceTags":[],"Definition":{"HTTP":"","Header":null,"Method":"","TLSSkipVerify":false,"TCP":"","Interval":0,"Timeout":0,"DeregisterCriticalServiceAfter":0},"CreateIndex":0,"ModifyIndex":0}}}`},
	{"/verify/checks?status=critical", spi.StatusOK, `{"Status":"Ok"}`},
	{"/verify/checks/check1b?status=warning", spi.StatusOK, `{"Status":"Ok"}`},
	{"/verify/checks/check%202?verbose", spi.StatusOK, `{"Status":"Ok","Checks":{"check2a":{"Node":"{{.ConsulNodeName}}","CheckID":"check2a","Name":"check 2","Status":"passing","Notes":"","Output":"","ServiceID":"service2","ServiceName":"service2","ServiceTags":[],"Definition":{"HTTP":"","Header":null,"Method":"","TLSSkipVerify":false,"TCP":"","Interval":0,"Timeout":0,"DeregisterCriticalServiceAfter":0},"CreateIndex":0,"ModifyIndex":0}}}`},
	{"/verify/service/service1?status=critical", spi.StatusOK, `{"Status":"Ok"}`},
	{"/verify/service/service2?verbose", spi.StatusOK, `{"Status":"Ok","Checks":{"check2a":{"Node":"{{.ConsulNodeName}}","CheckID":"check2a","Name":"check 2","Status":"passing","Notes":"","Output":"","ServiceID":"service2","ServiceName":"service2","ServiceTags":[],"Definition":{"HTTP":"","Header":null,"Method":"","TLSSkipVerify":false,"TCP":"","Interval":0,"Timeout":0,"DeregisterCriticalServiceAfter":0},"CreateIndex":0,"ModifyIndex":0}}}`},
	{"/verify/service/service2?pretty", spi.StatusOK, `{
    "Status": "Ok"
}`},

}

type apiTestData struct {
	path       string
	statusCode int
	body       string
}

type bodyTemplateData struct {
	ConsulNodeName string
}

func TestApi(t *testing.T) {
	t.Log("Starting API tests...")

	server := newServer(t)
	defer server.Stop()

	server.AddService("service1", []string{})
	server.AddCheck("check1a", "check 1", "service1", checks.HealthPassing)
	server.AddCheck("check1b", "check 1", "service1", checks.HealthWarning)
	server.AddCheck("check1c", "check 1", "service1", checks.HealthCritical)

	server.AddService("service2", []string{})
	server.AddCheck("check2a", "check 2", "service2", checks.HealthPassing)

	for _, d := range apiTests {
		t.Logf("  --> %s", d.path)
		verifyApiCall(t, server, d)
	}
	t.Log("Finished API tests")
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

	data := bodyTemplateData{
		ConsulNodeName: s.GetConsulNodeName(),
	}
	tmpl, err := template.New(d.path).Parse(d.body)
	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, data); err != nil {
		t.Errorf("Failed to execute body template: %s", err)
	}
	expectedBody := tpl.String()

	if body != expectedBody {
		t.Errorf("%q => Body: %q, want %q", d.path, body, expectedBody)
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
