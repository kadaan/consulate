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
	// StatusOK represents a Consulate call which succeeded.
	StatusOK = http.StatusOK

	// StatusBadRequestError represents a request to Consulate which could not be understood.
	StatusBadRequestError = http.StatusBadRequest

	// StatusNoChecksError represents an error when no Consul checks exist.
	StatusNoChecksError = http.StatusNotFound

	// StatusConsulError represents an error indicating one or more Consul checks have failed.
	StatusConsulError = http.StatusInternalServerError

	// StatusConsulUnavailableError represents an error indicating the Consul did not respond promptly.
	StatusConsulUnavailableError = http.StatusServiceUnavailable

	// StatusUnprocessableResponseError represents an error indicating that Consulate could not parse the reponse from Consul.
	StatusUnprocessableResponseError = http.StatusUnprocessableEntity
)

type Server interface {
	Start() (RunningServer, error)
}

type RunningServer interface {
	Stop()
}
