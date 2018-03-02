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
	"testing"
	"time"
	"reflect"
)

var cacheData = []struct {
	duration     time.Duration
	expectedType reflect.Type
	result       bool
}{
	{duration: 0, expectedType: reflect.TypeOf(&noOpCache{}), result: false},
	{duration: 1*time.Second, expectedType: reflect.TypeOf(&inMemoryCache{}), result: true},
}

func TestNewCache(t *testing.T) {
	for _, d := range cacheData {
		cache := *NewCache(config.CacheConfig{ConsulCacheDuration: d.duration})
		actualType := reflect.TypeOf(cache)
		if d.expectedType != actualType {
			t.Errorf("Type: want %v, got %v", d.expectedType, actualType)
		}
		cache.Set("foo", "bar")
		if _, ok := cache.Get("foo"); ok != d.result {
			t.Errorf("Result: want %v, got %v", d.result, ok)
		}
		time.Sleep(d.duration+100*time.Millisecond)
		if _, ok := cache.Get("foo"); ok {
			t.Error("Result: want record expired, got record")
		}
	}
}