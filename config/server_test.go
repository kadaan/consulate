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

func TestDefaultServerConfig(t *testing.T) {
	c := DefaultServerConfig()
	if c.ListenAddress != DefaultListenAddress {
		t.Errorf("ListenAddress: want %v, got %v", DefaultListenAddress, c.ListenAddress)
	}
	if c.ConsulAddress != DefaultConsulAddress {
		t.Errorf("ConsulAddress: want %v, got %v", DefaultConsulAddress, c.ConsulAddress)
	}
	if c.ReadTimeout != DefaultReadTimeout {
		t.Errorf("ReadTimeout: want %v, got %v", DefaultReadTimeout, c.ReadTimeout)
	}
	if c.WriteTimeout != DefaultWriteTimeout {
		t.Errorf("WriteTimeout: want %v, got %v", DefaultWriteTimeout, c.WriteTimeout)
	}
	if c.ShutdownTimeout != DefaultShutdownTimeout {
		t.Errorf("ShutdownTimeout: want %v, got %v", DefaultShutdownTimeout, c.ShutdownTimeout)
	}
}
