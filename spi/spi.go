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

package spi

import "net/http"

const (
	// StatusOK (200) represents a Consulate call which succeeded.
	StatusOK = http.StatusOK

	// StatusBadRequestError (400) represents a request to Consulate which could not be understood.
	StatusBadRequestError = http.StatusBadRequest

	// StatusNoChecksError (404) represents an error when no Consul checks exist.
	StatusNoChecksError = http.StatusNotFound

	// StatusUnprocessableResponseError (502) represents an error indicating that Consulate could not parse the response from Consul.
	StatusUnprocessableResponseError = http.StatusBadGateway

	// StatusCheckError (503) represents an error indicating one or more Consul checks have failed.
	StatusCheckError = http.StatusServiceUnavailable

	// StatusConsulUnavailableError (504) represents an error indicating the Consul did not respond promptly.
	StatusConsulUnavailableError = http.StatusGatewayTimeout
)

// Server represents a server that can be started.
type Server interface {
	// Start begins the Server.
	Start() (RunningServer, error)
}

// RunningServer represents a server is running and can be stopped.
type RunningServer interface {
	// Stop terminates the current RunningServer.
	Stop()
}
