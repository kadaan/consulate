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

package config

import (
	"net/http"
	"time"
)

const (
	// DefaultListenAddress is the default listen address for Consulate.
	DefaultListenAddress = ":8080"

	// DefaultConsulAddress is the default address used to connect to Consul.
	DefaultConsulAddress = "localhost:8500"

	// DefaultReadTimeout is the default maximum duration for Consulate reading the entire request.
	DefaultReadTimeout = 10 * time.Second

	// DefaultWriteTimeout is the default maximum duration for Consulate writing the response.
	DefaultWriteTimeout = 10 * time.Second

	// DefaultShutdownTimeout is the default maximum duration before timing out the shutdown of the Consulate server.
	DefaultShutdownTimeout = 15 * time.Second

	// DefaultSuccessStatusCode (200) is the default status code returned when there are 1+ passing health checks, 0 warning health checks, and 0 failing health checks.
	DefaultSuccessStatusCode = http.StatusOK

	// DefaultPartialSuccessStatusCode (429) is the default status code returned when there are 1+ passing health checks and 1+ warning health checks.
	DefaultPartialSuccessStatusCode = http.StatusTooManyRequests

	// DefaultWarningStatusCode (503) is the default status code returned when there are 0 passing health checks and 1+ warning health checks.
	DefaultWarningStatusCode = http.StatusServiceUnavailable

	// DefaultErrorStatusCode (503) is the default status code returned when there are  1+ failing health checks.
	DefaultErrorStatusCode = http.StatusServiceUnavailable

	// DefaultBadRequestStatusCode (400) is the default status code returned when a request to Consulate which could not be understood.
	DefaultBadRequestStatusCode = http.StatusBadRequest

	// DefaultNoCheckStatusCode (404) is the default status code returned when no Consul checks exist.
	DefaultNoCheckStatusCode = http.StatusNotFound

	// DefaultUnprocessableStatusCode (502) is the default status code returned when Consulate could not parse the response from Consul.
	DefaultUnprocessableStatusCode = http.StatusBadGateway

	// DefaultConsulUnavailableStatusCode (504) is the default status code returned when Consul did not respond promptly.
	DefaultConsulUnavailableStatusCode = http.StatusGatewayTimeout
)

// ServerConfig represents the configuration of the Consulate server.
type ServerConfig struct {
	ListenAddress               string
	ConsulAddress               string
	ReadTimeout                 time.Duration
	WriteTimeout                time.Duration
	ShutdownTimeout             time.Duration
	SuccessStatusCode           int
	PartialSuccessStatusCode    int
	WarningStatusCode           int
	ErrorStatusCode             int
	BadRequestStatusCode        int
	NoCheckStatusCode           int
	UnprocessableStatusCode     int
	ConsulUnavailableStatusCode int
	ClientConfig                ClientConfig
	CacheConfig                 CacheConfig
}

// DefaultServerConfig gets a default ServerConfig.
func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		ListenAddress:               DefaultListenAddress,
		ConsulAddress:               DefaultConsulAddress,
		ReadTimeout:                 DefaultReadTimeout,
		WriteTimeout:                DefaultWriteTimeout,
		ShutdownTimeout:             DefaultShutdownTimeout,
		SuccessStatusCode:           DefaultSuccessStatusCode,
		PartialSuccessStatusCode:    DefaultPartialSuccessStatusCode,
		WarningStatusCode:           DefaultWarningStatusCode,
		ErrorStatusCode:             DefaultErrorStatusCode,
		BadRequestStatusCode:        DefaultBadRequestStatusCode,
		NoCheckStatusCode:           DefaultNoCheckStatusCode,
		UnprocessableStatusCode:     DefaultUnprocessableStatusCode,
		ConsulUnavailableStatusCode: DefaultConsulUnavailableStatusCode,
		ClientConfig:                *DefaultClientConfig(),
		CacheConfig:                 *DefaultCacheConfig(),
	}
}
