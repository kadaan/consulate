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

package checks

import "testing"

var matchesStatusData = []struct {
	threshold HealthStatus
	status    HealthStatus
	result    Status
}{
	{threshold: HealthPassing, status: HealthPassing, result: StatusPassing},
	{threshold: HealthPassing, status: HealthWarning, result: StatusWarning},
	{threshold: HealthPassing, status: HealthMaintenance, result: StatusFailing},
	{threshold: HealthPassing, status: HealthCritical, result: StatusFailing},
	{threshold: HealthWarning, status: HealthPassing, result: StatusPassing},
	{threshold: HealthWarning, status: HealthWarning, result: StatusPassing},
	{threshold: HealthWarning, status: HealthMaintenance, result: StatusFailing},
	{threshold: HealthWarning, status: HealthCritical, result: StatusFailing},
	{threshold: HealthMaintenance, status: HealthPassing, result: StatusPassing},
	{threshold: HealthMaintenance, status: HealthWarning, result: StatusPassing},
	{threshold: HealthMaintenance, status: HealthMaintenance, result: StatusPassing},
	{threshold: HealthMaintenance, status: HealthCritical, result: StatusFailing},
	{threshold: HealthCritical, status: HealthPassing, result: StatusPassing},
	{threshold: HealthCritical, status: HealthWarning, result: StatusPassing},
	{threshold: HealthCritical, status: HealthMaintenance, result: StatusPassing},
	{threshold: HealthCritical, status: HealthCritical, result: StatusPassing},

	//{threshold: HealthPassing, status: HealthPassing, result: false},
	//{threshold: HealthPassing, status: HealthMaintenance, result: true},
	//{threshold: HealthPassing, status: HealthWarning, result: true},
	//{threshold: HealthPassing, status: HealthCritical, result: true},
	//{threshold: HealthMaintenance, status: HealthPassing, result: false},
	//{threshold: HealthMaintenance, status: HealthMaintenance, result: false},
	//{threshold: HealthMaintenance, status: HealthWarning, result: true},
	//{threshold: HealthMaintenance, status: HealthCritical, result: true},
	//{threshold: HealthWarning, status: HealthPassing, result: false},
	//{threshold: HealthWarning, status: HealthMaintenance, result: false},
	//{threshold: HealthWarning, status: HealthWarning, result: false},
	//{threshold: HealthWarning, status: HealthCritical, result: true},
	//{threshold: HealthCritical, status: HealthPassing, result: false},
	//{threshold: HealthCritical, status: HealthMaintenance, result: false},
	//{threshold: HealthCritical, status: HealthWarning, result: false},
	//{threshold: HealthCritical, status: HealthCritical, result: false},
}

func TestMatchesStatusUnknown(t *testing.T) {
	check := Check{
		Status: "unknown",
	}
	_, e := check.MatchStatus(HealthPassing)
	if e == nil || e.Error() != "Unsupported status: unknown" {
		t.Errorf("Status: unknown, want 'Unsupported status: unknown', got '%v'", e)
	}
}

func TestMatchesStatus(t *testing.T) {
	for _, d := range matchesStatusData {
		check := Check{
			Status: d.status.String(),
		}
		if r, _ := check.MatchStatus(d.threshold); r != d.result {
			t.Errorf("Threshold %v => Status: %v, want %v, got %v", d.threshold, d.status, d.result, r)
		}
	}
}

func TestIsCheckId(t *testing.T) {
	check := Check{
		CheckID: "a",
	}
	if !check.IsCheckId("a") {
		t.Error("CheckId 'a' => 'a', want 'true', got 'false'")
	}
	if check.IsCheckId("b") {
		t.Error("CheckId 'a' => 'b', want 'false', got 'true'")
	}
}

func TestIsCheckName(t *testing.T) {
	check := Check{
		Name: "a",
	}
	if !check.IsCheckName("a") {
		t.Error("CheckName 'a' => 'a', want 'true', got 'false'")
	}
	if check.IsCheckName("b") {
		t.Error("CheckName 'a' => 'b', want 'false', got 'true'")
	}
}

func TestIsServiceId(t *testing.T) {
	check := Check{
		ServiceID: "a",
	}
	if !check.IsServiceId("a") {
		t.Error("ServiceId 'a' => 'a', want 'true', got 'false'")
	}
	if check.IsServiceId("b") {
		t.Error("ServiceId 'a' => 'b', want 'false', got 'true'")
	}
}

func TestIsServiceName(t *testing.T) {
	check := Check{
		ServiceName: "a",
	}
	if !check.IsServiceName("a") {
		t.Error("ServiceName 'a' => 'a', want 'true', got 'false'")
	}
	if check.IsServiceName("b") {
		t.Error("ServiceName 'a' => 'b', want 'false', got 'true'")
	}
}
