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
	// DefaultConsulCacheDuration is the default duration that Consul results will be cached.
	DefaultConsulCacheDuration = 1 * time.Second
)

// CacheConfig represents the configuration of caching the Consulate server.
type CacheConfig struct {
	ConsulCacheDuration time.Duration
}

// DefaultCacheConfig gets a default CacheConfig.
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		ConsulCacheDuration: DefaultConsulCacheDuration,
	}
}
