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
	DefaultListenAddress   = ":8080"
	DefaultReadTimeout     = 10 * time.Second
	DefaultWriteTimeout    = 10 * time.Second
	DefaultShutdownTimeout = 15 * time.Second
)

type ServerConfig struct {
	ListenAddress   string
	ConsulAddress   string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	ClientConfig    ClientConfig
}

func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		ListenAddress:   DefaultListenAddress,
		ConsulAddress:   DefaultListenAddress,
		ReadTimeout:     DefaultReadTimeout,
		WriteTimeout:    DefaultWriteTimeout,
		ShutdownTimeout: DefaultShutdownTimeout,
		ClientConfig:    *DefaultClientConfig(),
	}
}
