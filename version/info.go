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

import (
	"bytes"
	"runtime"
	"strings"
	"text/template"
)

var (
	// Version is the version number of Consulate.
	Version string

	// Revision is the git revision that Consulate was built from.
	Revision string

	// Branch is the git branch that Consulate was built from.
	Branch string

	// BuildUser is the user that built Consulate.
	BuildUser string

	// BuildDate is the date that Consulate was built.
	BuildDate string
	goVersion = runtime.Version()
)

type info struct {
	Version   string
	Revision  string
	Branch    string
	BuildUser string
	BuildDate string
	GoVersion string
}

var versionInfoTmpl = `
Consulate, version {{.version}} (branch: {{.branch}}, revision: {{.revision}})
  build user:       {{.buildUser}}
  build date:       {{.buildDate}}
  go version:       {{.goVersion}}
`

// Print formats the version info as a string.
func Print() string {
	m := map[string]string{
		"version":   Version,
		"revision":  Revision,
		"branch":    Branch,
		"buildUser": BuildUser,
		"buildDate": BuildDate,
		"goVersion": goVersion,
	}
	t := template.Must(template.New("version").Parse(versionInfoTmpl))

	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "version", m); err != nil {
		panic(err)
	}
	return strings.TrimSpace(buf.String())
}

// NewInfo creates a new version info object.
func NewInfo() info {
	return info{
		Version:   Version,
		Revision:  Revision,
		Branch:    Branch,
		BuildUser: BuildUser,
		BuildDate: BuildDate,
		GoVersion: goVersion,
	}
}
