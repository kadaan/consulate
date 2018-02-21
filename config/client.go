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

import "time"

const (
	// DefaultQueryIdleConnectionTimeout is the default maximum amount
	// of time an idle (keep-alive) connection will remain idle before
	// closing itself.
	DefaultQueryIdleConnectionTimeout = 90 * time.Second

	// DefaultQueryMaxIdleConnectionCount is the default maximum number
	// of idle (keep-alive) connections across all hosts.
	DefaultQueryMaxIdleConnectionCount = 100

	// DefaultQueryTimeout is the default time limit for requests.
	DefaultQueryTimeout = 5 * time.Second
)

// ClientConfig represents the configuration for the http.Client.
type ClientConfig struct {
	QueryTimeout                time.Duration
	QueryMaxIdleConnectionCount int
	QueryIdleConnectionTimeout  time.Duration
}

// DefaultClientConfig gets a default ClientConfig.
func DefaultClientConfig() *ClientConfig {
	return &ClientConfig{
		QueryIdleConnectionTimeout:  DefaultQueryIdleConnectionTimeout,
		QueryMaxIdleConnectionCount: DefaultQueryMaxIdleConnectionCount,
		QueryTimeout:                DefaultQueryTimeout,
	}
}
