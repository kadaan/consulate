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
	"github.com/kadaan/consulate/config"
	"github.com/kadaan/consulate/testutil"
	"io/ioutil"
	"testing"
	"text/template"
)

var (
	OK         = config.DefaultServerConfig().SuccessStatusCode
	PartialOK  = config.DefaultServerConfig().PartialSuccessStatusCode
	NoChecks   = config.DefaultServerConfig().NoCheckStatusCode
	CheckError = config.DefaultServerConfig().ErrorStatusCode
)

var apiTests = []apiTestData{
	{"/about", OK, `{"Version":"","Revision":"","Branch":"","BuildUser":"","BuildDate":"","GoVersion":"go1.9.3"}`},
	{"/health", OK, `{"Status":"Ok"}`},
	{"/verify/checks", CheckError, `{"Status":"Failed","Counts":{"failing":1,"passing":3,"warning":2},"Checks":{"check1b":{"Node":"{{.ConsulNodeName}}","CheckID":"check1b","Name":"check 1","Status":"warning","Notes":"","Output":"","ServiceID":"service1","ServiceName":"service1","ServiceTags":[],"Definition":{"HTTP":"","Header":null,"Method":"","TLSSkipVerify":false,"TCP":"","Interval":0,"Timeout":0,"DeregisterCriticalServiceAfter":0},"CreateIndex":0,"ModifyIndex":0},"check1c":{"Node":"{{.ConsulNodeName}}","CheckID":"check1c","Name":"check 1","Status":"critical","Notes":"","Output":"","ServiceID":"service1","ServiceName":"service1","ServiceTags":[],"Definition":{"HTTP":"","Header":null,"Method":"","TLSSkipVerify":false,"TCP":"","Interval":0,"Timeout":0,"DeregisterCriticalServiceAfter":0},"CreateIndex":0,"ModifyIndex":0},"check3b":{"Node":"{{.ConsulNodeName}}","CheckID":"check3b","Name":"check 3","Status":"warning","Notes":"","Output":"","ServiceID":"service3","ServiceName":"service3","ServiceTags":[],"Definition":{"HTTP":"","Header":null,"Method":"","TLSSkipVerify":false,"TCP":"","Interval":0,"Timeout":0,"DeregisterCriticalServiceAfter":0},"CreateIndex":0,"ModifyIndex":0}}}`},
	{"/verify/checks?status=passing", CheckError, `{"Status":"Failed","Counts":{"failing":1,"passing":3,"warning":2},"Checks":{"check1b":{"Node":"{{.ConsulNodeName}}","CheckID":"check1b","Name":"check 1","Status":"warning","Notes":"","Output":"","ServiceID":"service1","ServiceName":"service1","ServiceTags":[],"Definition":{"HTTP":"","Header":null,"Method":"","TLSSkipVerify":false,"TCP":"","Interval":0,"Timeout":0,"DeregisterCriticalServiceAfter":0},"CreateIndex":0,"ModifyIndex":0},"check1c":{"Node":"{{.ConsulNodeName}}","CheckID":"check1c","Name":"check 1","Status":"critical","Notes":"","Output":"","ServiceID":"service1","ServiceName":"service1","ServiceTags":[],"Definition":{"HTTP":"","Header":null,"Method":"","TLSSkipVerify":false,"TCP":"","Interval":0,"Timeout":0,"DeregisterCriticalServiceAfter":0},"CreateIndex":0,"ModifyIndex":0},"check3b":{"Node":"{{.ConsulNodeName}}","CheckID":"check3b","Name":"check 3","Status":"warning","Notes":"","Output":"","ServiceID":"service3","ServiceName":"service3","ServiceTags":[],"Definition":{"HTTP":"","Header":null,"Method":"","TLSSkipVerify":false,"TCP":"","Interval":0,"Timeout":0,"DeregisterCriticalServiceAfter":0},"CreateIndex":0,"ModifyIndex":0}}}`},
	{"/verify/checks?status=warning", CheckError, `{"Status":"Failed","Counts":{"failing":1,"passing":5,"warning":0},"Checks":{"check1c":{"Node":"{{.ConsulNodeName}}","CheckID":"check1c","Name":"check 1","Status":"critical","Notes":"","Output":"","ServiceID":"service1","ServiceName":"service1","ServiceTags":[],"Definition":{"HTTP":"","Header":null,"Method":"","TLSSkipVerify":false,"TCP":"","Interval":0,"Timeout":0,"DeregisterCriticalServiceAfter":0},"CreateIndex":0,"ModifyIndex":0}}}`},
	{"/verify/checks?status=critical", OK, `{"Status":"Ok"}`},
	{"/verify/checks/id", NoChecks, `404 page not found`},
	{"/verify/checks/id/unknown", NoChecks, `{"Status":"No Checks","Detail":"No checks with CheckID: unknown"}`},
	{"/verify/checks/name", NoChecks, `404 page not found`},
	{"/verify/checks/name/unknown", NoChecks, `{"Status":"No Checks","Detail":"No checks with CheckName: unknown"}`},
	{"/verify/checks/id/check1b", CheckError, `{"Status":"Failed","Counts":{"failing":0,"passing":0,"warning":1},"Checks":{"check1b":{"Node":"{{.ConsulNodeName}}","CheckID":"check1b","Name":"check 1","Status":"warning","Notes":"","Output":"","ServiceID":"service1","ServiceName":"service1","ServiceTags":[],"Definition":{"HTTP":"","Header":null,"Method":"","TLSSkipVerify":false,"TCP":"","Interval":0,"Timeout":0,"DeregisterCriticalServiceAfter":0},"CreateIndex":0,"ModifyIndex":0}}}`},
	{"/verify/checks/id/check1b?status=warning", OK, `{"Status":"Ok"}`},
	{"/verify/checks/name/check%202?verbose", OK, `{"Status":"Ok","Checks":{"check2a":{"Node":"{{.ConsulNodeName}}","CheckID":"check2a","Name":"check 2","Status":"passing","Notes":"","Output":"","ServiceID":"service2","ServiceName":"service2","ServiceTags":[],"Definition":{"HTTP":"","Header":null,"Method":"","TLSSkipVerify":false,"TCP":"","Interval":0,"Timeout":0,"DeregisterCriticalServiceAfter":0},"CreateIndex":0,"ModifyIndex":0}}}`},
	{"/verify/service/id/unknown", NoChecks, `{"Status":"No Checks","Detail":"No checks for services with ServiceId: unknown"}`},
	{"/verify/service/name/unknown", NoChecks, `{"Status":"No Checks","Detail":"No checks for services with ServiceName: unknown"}`},
	{"/verify/service/id/service1?status=critical", OK, `{"Status":"Ok"}`},
	{"/verify/service/id/service2?verbose", OK, `{"Status":"Ok","Checks":{"check2a":{"Node":"{{.ConsulNodeName}}","CheckID":"check2a","Name":"check 2","Status":"passing","Notes":"","Output":"","ServiceID":"service2","ServiceName":"service2","ServiceTags":[],"Definition":{"HTTP":"","Header":null,"Method":"","TLSSkipVerify":false,"TCP":"","Interval":0,"Timeout":0,"DeregisterCriticalServiceAfter":0},"CreateIndex":0,"ModifyIndex":0}}}`},
	{"/verify/service/id/service2?pretty", OK, `{
    "Status": "Ok"
}`},
	{"/verify/service/id/service3", PartialOK, `{"Status":"Warning","Counts":{"failing":0,"passing":1,"warning":1},"Checks":{"check3b":{"Node":"{{.ConsulNodeName}}","CheckID":"check3b","Name":"check 3","Status":"warning","Notes":"","Output":"","ServiceID":"service3","ServiceName":"service3","ServiceTags":[],"Definition":{"HTTP":"","Header":null,"Method":"","TLSSkipVerify":false,"TCP":"","Interval":0,"Timeout":0,"DeregisterCriticalServiceAfter":0},"CreateIndex":0,"ModifyIndex":0}}}`},
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

	server.AddService("service3", []string{})
	server.AddCheck("check3a", "check 3", "service3", checks.HealthPassing)
	server.AddCheck("check3b", "check 3", "service3", checks.HealthWarning)

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
	tmpl, _ := template.New(d.path).Parse(d.body)
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
