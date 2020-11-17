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

package version

import "testing"

func TestPrint(t *testing.T) {
	Version = "testVersion"
	Branch = "testBranch"
	Revision = "testRevision"
	BuildUser = "testUser"
	BuildDate = "testDate"

	expected := "Consulate, version testVersion (branch: testBranch, revision: testRevision)\n  build user:       testUser\n  build date:       testDate\n  go version:       go1.15.2"
	result := Print()
	if result != expected {
		t.Errorf("Print: %q, want %q", result, expected)
	}
}

func TestNewInfo(t *testing.T) {
	Version = "testVersion"
	Branch = "testBranch"
	Revision = "testRevision"
	BuildUser = "testUser"
	BuildDate = "testDate"

	info := NewInfo()
	if info.BuildDate != BuildDate {
		t.Errorf("BuildDate: %q, want %q", info.BuildDate, BuildDate)
	}
	if info.BuildUser != BuildUser {
		t.Errorf("BuildUser: %q, want %q", info.BuildUser, BuildUser)
	}
	if info.Revision != Revision {
		t.Errorf("Revision: %q, want %q", info.Revision, Revision)
	}
	if info.Branch != Branch {
		t.Errorf("Branch: %q, want %q", info.Branch, Branch)
	}
	if info.Version != Version {
		t.Errorf("Version: %q, want %q", info.Version, Version)
	}

}
