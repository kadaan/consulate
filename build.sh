#!/usr/bin/env bash

# Copyright © 2018 Joel Baranick <jbaranick@gmail.com>
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

VERSION=0.0.1

run() {
  if [ ! -f $GOPATH/bin/dep ]; then
      unameOut="$(uname -s)"
      case "${unameOut}" in
        Linux*)
          echo "Getting dep..."
          curl -L -s https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 -o $GOPATH/bin/dep
        ;;
        Darwin*)
          echo "Getting dep..."
          curl -L -s https://github.com/golang/dep/releases/download/v0.4.1/dep-darwin-amd64 -o $GOPATH/bin/dep
        ;;
        *)
          echo "Unsupported machine type :${unameOut}"
          exit 1
        ;;
      esac
      chmod +x $GOPATH/bin/dep
  fi
  echo "Retrieving dependencies..."
  dep ensure

  echo "Formatting source..."
  local gofiles=$(find . -path ./vendor -prune -o -print | grep '\.go$')
  if [[ ${#gofiles[@]} -gt 0 ]]; then
    while read -r gofile; do
        gofmt -w $PWD/$gofile
    done <<< "$gofiles"
  fi

  echo "Checking licenses..."
  licRes=$(
  for file in $(find . -type f -iname '*.go' ! -path './vendor/*'); do
    head -n3 "${file}" | grep -Eq "(Copyright|generated|GENERATED)" || echo -e "  ${file}"
  done;)
  if [ -n "${licRes}" ]; then
  	echo -e "license header checking failed:\n${licRes}"
  	exit 1
  fi

  echo "Building binaries..."
  local revision=`git rev-parse HEAD`
  local branch=`git rev-parse --abbrev-ref HEAD`
  local host=`hostname`
  local buildDate=`date -u +"%Y-%m-%dT%H:%M:%SZ"`
  if [ ! -x "$(command -v gox)" ]; then
    echo "Getting gox..."
    go get github.com/mitchellh/gox
  fi
  gox -ldflags "-X github.com/kadaan/consulate/version.Version=$VERSION -X github.com/kadaan/consulate/version.Revision=$revision -X github.com/kadaan/consulate/version.Branch=$branch -X github.com/kadaan/consulate/version.BuildUser=$USER@$host -X github.com/kadaan/consulate/version.BuildDate=$buildDate" -output="dist/{{.Dir}}_{{.OS}}_{{.Arch}}"
  echo ""
  echo "done"
}

run "$@"
