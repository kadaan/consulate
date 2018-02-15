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
	"time"
)

const (
	healthPassing = "passing"
)

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

// IsHealthy returns True if the Check has a Status of Passing.
func (c *Check) IsHealthy() bool {
	return c.Status == healthPassing
}

// IsService returns True if the Check ServiceId/Name matches the specified service.
func (c *Check) IsService(service string) bool {
	return service == c.ServiceName || service == c.ServiceID
}

// IsCheck returns True if the Check CheckID/Name matches the specified check.
func (c *Check) IsCheck(check string) bool {
	return check == c.Name || check == c.CheckID
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
