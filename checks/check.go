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

import (
	"github.com/pkg/errors"
	"time"
)

// HealthStatus represents the status of a Consul health check.
type HealthStatus int

const (
	// HealthPassing represents Consul health check in the Passing state.
	HealthPassing HealthStatus = iota

	// HealthMaintenance represents Consul health check in the Maintenance state.
	HealthMaintenance

	// HealthWarning represents Consul health check in the Warning state
	HealthWarning

	// HealthCritical represents Consul health check in the Critical state
	HealthCritical
)

var healthNames = []string{"passing", "maintenance", "warning", "critical"}

func (s HealthStatus) String() string {
	return healthNames[s]
}

// ParseHealthStatus parses a string into a HealthStatus
func ParseHealthStatus(s string) (HealthStatus, bool) {
	for i, n := range healthNames {
		if n == s {
			return HealthStatus(i), true
		}
	}
	return 0, false
}

// ResultStatus represents the status of a Consulate call.
type ResultStatus string

const (
	// Ok represents a successful call to Consulate.
	Ok ResultStatus = "Ok"

	// Failed represents a failed call to Consulate.
	Failed ResultStatus = "Failed"

	// NoChecks represents verify check call to Consulate which had no checks to verify.
	NoChecks ResultStatus = "No Checks"
)

// Result represents the result of a Consulate call.
type Result struct {
	Status ResultStatus
	Detail string            `json:",omitempty"`
	Checks map[string]*Check `json:",omitempty"`
}

// Check the result of a Consul check.
type Check struct {
	Node        string
	CheckID     string
	Name        string
	Status      string
	Notes       string
	Output      string
	ServiceID   string
	ServiceName string
	ServiceTags []string
	Definition  CheckDefinition
	CreateIndex uint64
	ModifyIndex uint64
}

// MatchesStatus returns True if the Check has a Status the same or worse than the specified status.
func (c *Check) MatchesStatus(s HealthStatus) (bool, error) {
	parsedStatus, parsed := ParseHealthStatus(c.Status)
	if !parsed {
		return false, errors.Errorf("Unsupported status: %s", c.Status)
	}
	return parsedStatus > s, nil
}

// IsServiceId returns True if the Check ServiceId matches the specified serviceId.
func (c *Check) IsServiceId(serviceId string) bool {
	return serviceId == c.ServiceID
}

// IsServiceName returns True if the Check ServiceName matches the specified serviceName.
func (c *Check) IsServiceName(serviceName string) bool {
	return serviceName == c.ServiceName
}

// IsCheckId returns True if the Check CheckID/Name matches the specified check.
func (c *Check) IsCheckId(checkId string) bool {
	return checkId == c.CheckID
}

// IsCheckName returns True if the Check CheckName matches the specified checkName.
func (c *Check) IsCheckName(checkName string) bool {
	return checkName == c.Name
}

// CheckDefinition represents the configuration of a Consul check.
type CheckDefinition struct {
	HTTP                           string
	Header                         map[string][]string
	Method                         string
	TLSSkipVerify                  bool
	TCP                            string
	Interval                       time.Duration
	Timeout                        time.Duration
	DeregisterCriticalServiceAfter time.Duration
}
