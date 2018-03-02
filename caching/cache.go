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

package caching

import (
	"github.com/kadaan/consulate/config"
	"github.com/patrickmn/go-cache"
	"time"
)

// Cache represents a simple data cache
type Cache interface {
	// Get an item from the cache. Returns the item or nil, and a bool indicating whether the key was found
	Get(key string) (interface{}, bool)

	// Add an item to the cache, replacing any existing item, using the default expiration
	Set(key string, value interface{})
}

type noOpCache struct {
}

func (c *noOpCache) Get(key string) (interface{}, bool) {
	return nil, false
}

func (c *noOpCache) Set(key string, value interface{}) {
}

type inMemoryCache struct {
	cache *cache.Cache
}

func (c *inMemoryCache) Get(key string) (interface{}, bool) {
	return c.cache.Get(key)
}

func (c *inMemoryCache) Set(key string, value interface{}) {
	c.cache.SetDefault(key, value)
}

// NewCache creates a new caching.Cache
func NewCache(config config.CacheConfig) *Cache {
	var result Cache
	if config.ConsulCacheDuration <= 0 {
		result = &noOpCache{}
	} else {
		result = &inMemoryCache{cache: cache.New(config.ConsulCacheDuration, 1*time.Minute)}
	}
	return &result
}
