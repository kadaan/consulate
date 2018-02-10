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
	HealthPassing = "passing"
)

type ResultStatus string

const (
	OK       ResultStatus = "Ok"
	Failed   ResultStatus = "Failed"
	NoChecks ResultStatus = "No Checks"
)

type Result struct {
	Status ResultStatus
	Detail string `json:",omitempty"`
	Checks map[string]*Check `json:",omitempty"`
}

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

func (c *Check) IsHealthy() bool {
	return c.Status == HealthPassing
}
func (c *Check) IsService(service string) bool {
	return service == c.ServiceName || service == c.ServiceID
}

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