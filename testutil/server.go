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

package testutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/consul/lib/freeport"
	consulTestUtil "github.com/hashicorp/consul/testutil"
	"github.com/hashicorp/consul/testutil/retry"
	"github.com/kadaan/consulate/checks"
	"github.com/kadaan/consulate/client"
	"github.com/kadaan/consulate/config"
	"github.com/kadaan/consulate/server"
	"github.com/kadaan/consulate/spi"
	"io"
	"net/http"
	"testing"
)

type TestServer struct {
	spi.RunningServer
	httpAddr   string
	httpClient *http.Client
	svr        spi.RunningServer
	consulSvr  *consulTestUtil.TestServer
}

type WrappableTestServer struct {
	TestServer
}

type WrappedTestServer struct {
	s *TestServer
	t *testing.T
}

func NewTestServer(t *testing.T) (*WrappableTestServer, error) {
	consulServer := newConsulServer(t)
	ports := freeport.GetT(t, 1)
	httpAddr := fmt.Sprintf(":%v", ports[0])
	svrconfig := config.DefaultServerConfig()
	svrconfig.ListenAddress = httpAddr
	svrconfig.ConsulAddress = consulServer.HTTPAddr
	svr, err := server.NewServer(svrconfig).Start()
	if err != nil {
		if svr != nil {
			defer svr.Stop()
		}
		return nil, err
	}
	testsvr := &WrappableTestServer{
		TestServer{
			httpAddr:   httpAddr,
			httpClient: client.CreateClient(config.DefaultClientConfig()),
			svr:        svr,
			consulSvr:  consulServer,
		},
	}
	err = testsvr.waitForAPI(t)
	if err != nil {
		if svr != nil {
			defer svr.Stop()
		}
		return nil, err
	}
	return testsvr, nil
}

func (s *TestServer) Stop() {
	if s.svr != nil {
		defer s.svr.Stop()
	}
	if s.consulSvr != nil {
		defer s.consulSvr.Stop()
	}
}

type failer struct {
	failed bool
}

func (f *failer) Log(args ...interface{}) { fmt.Println(args) }
func (f *failer) FailNow()                { f.failed = true }

func newConsulServer(t *testing.T) *consulTestUtil.TestServer {
	svr, err := consulTestUtil.NewTestServerConfigT(t, func(c *consulTestUtil.TestServerConfig) {
		c.LogLevel = "ERR"
	})
	if err != nil {
		if svr != nil {
			defer svr.Stop()
		}
		t.Fatalf("Failed to start test consul server: %v", err)
	}
	return svr
}

func (s *TestServer) waitForAPI(t *testing.T) error {
	f := &failer{}
	retry.Run(f, func(r *retry.R) {
		resp, err := s.httpClient.Get(s.Url(t, "/about"))
		if err != nil {
			r.Error(err)
		}
		defer resp.Body.Close()
		if err := s.RequireOK(t, resp); err != nil {
			r.Error("Response returned from API was not 'OK'", err)
		}
	})
	if f.failed {
		t.Fatalf("Failed waiting for Consulate API to start")
	}
	return nil
}

func (s *TestServer) Client() *http.Client {
	return s.httpClient
}

func (s *TestServer) Url(t *testing.T, path string) string {
	if s == nil {
		t.Fatal("s is nil")
	}
	if path == "" {
		t.Fatal("path is empty")
	}
	return fmt.Sprintf("http://127.0.0.1%s%s", s.httpAddr, path)
}

func (s *TestServer) consulUrl(t *testing.T, path string) string {
	if s == nil {
		t.Fatal("s is nil")
	}
	if path == "" {
		t.Fatal("path is empty")
	}
	return fmt.Sprintf("http://%s%s", s.consulSvr.HTTPAddr, path)
}

func (s *TestServer) RequireOK(t *testing.T, resp *http.Response) error {
	if resp.StatusCode != 200 {
		t.Fatalf("bad status code: %d", resp.StatusCode)
	}
	return nil
}

func (s *TestServer) AddService(t *testing.T, name string, tags []string) {
	svc := &consulTestUtil.TestService{
		Name:    name,
		Tags:    tags,
		Address: "",
		Port:    0,
	}
	payload := s.encodePayload(t, svc)
	s.put(t, "/v1/agent/service/register", payload)
}

func (s *TestServer) AddCheck(t *testing.T, id string, name string, serviceID string, status checks.HealthStatus) {
	chk := &consulTestUtil.TestCheck{
		ID:   id,
		Name: name,
		TTL:  "10m",
	}
	if serviceID != "" {
		chk.ServiceID = serviceID
	}

	payload := s.encodePayload(t, chk)
	s.put(t, "/v1/agent/check/register", payload)

	switch status {
	case checks.HealthPassing:
		s.put(t, "/v1/agent/check/pass/"+id, nil)
	case checks.HealthWarning:
		s.put(t, "/v1/agent/check/warn/"+id, nil)
	case checks.HealthCritical:
		s.put(t, "/v1/agent/check/fail/"+id, nil)
	default:
		t.Fatalf("Unrecognized status: %s", status)
	}
}

func (s *TestServer) put(t *testing.T, path string, body io.Reader) *http.Response {
	req, err := http.NewRequest("PUT", s.consulUrl(t, path), body)
	if err != nil {
		t.Fatalf("failed to create PUT request: %s", err)
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		t.Fatalf("failed to make PUT request: %s", err)
	}
	if err := s.RequireOK(t, resp); err != nil {
		defer resp.Body.Close()
		t.Fatalf("not OK PUT: %s", err)
	}
	return resp
}

func (s *TestServer) encodePayload(t *testing.T, payload interface{}) io.Reader {
	var encoded bytes.Buffer
	enc := json.NewEncoder(&encoded)
	if err := enc.Encode(payload); err != nil {
		t.Fatalf("failed to encode payload: %v", err)
	}
	return &encoded
}

func (s *TestServer) GetConsulNodeName() string {
	return s.consulSvr.Config.NodeName
}

func (s *TestServer) Wrap(t *testing.T) *WrappedTestServer {
	return &WrappedTestServer{s, t}
}

func (w *WrappedTestServer) Stop() {
	defer w.s.Stop()
}

func (w *WrappedTestServer) Client() *http.Client {
	return w.s.Client()
}

func (w *WrappedTestServer) Url(path string) string {
	return w.s.Url(w.t, path)
}

func (w *WrappedTestServer) RequireOK(resp *http.Response) error {
	return w.s.RequireOK(w.t, resp)
}

func (w *WrappedTestServer) AddService(name string, tags []string) {
	w.s.AddService(w.t, name, tags)
}

func (w *WrappedTestServer) AddCheck(id string, name string, serviceID string, status checks.HealthStatus) {
	w.s.AddCheck(w.t, id, name, serviceID, status)
}

func (w *WrappedTestServer) GetConsulNodeName() string {
	return w.s.GetConsulNodeName()
}
