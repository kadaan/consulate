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

import "testing"

func TestDefaultClientConfig(t *testing.T) {
	c := DefaultClientConfig()
	if c.QueryIdleConnectionTimeout != DefaultQueryIdleConnectionTimeout {
		t.Errorf("QueryIdleConnectionTimeout: want %v, got %v", DefaultQueryIdleConnectionTimeout, c.QueryIdleConnectionTimeout)
	}
	if c.QueryMaxIdleConnectionCount != DefaultQueryMaxIdleConnectionCount {
		t.Errorf("QueryMaxIdleConnectionCount: want %v, got %v", DefaultQueryMaxIdleConnectionCount, c.QueryMaxIdleConnectionCount)
	}
	if c.QueryTimeout != DefaultQueryTimeout {
		t.Errorf("QueryTimeout: want %v, got %v", DefaultQueryTimeout, c.QueryTimeout)
	}
}
