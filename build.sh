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


BUILD_DIR="$(cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd)"
BINARY_DIR="$BUILD_DIR/.bin"
VERSION=$(cat $BUILD_DIR/.version)

function verbose() { echo -e "$*"; }
function error() { echo -e "ERROR: $*" 1>&2; }
function fatal() { echo -e "ERROR: $*" 1>&2; exit 1; }
function pushd () { command pushd "$@" > /dev/null; }
function popd () { command popd > /dev/null; }

function trap_add() {
  localtrap_add_cmd=$1; shift || fatal "${FUNCNAME} usage error"
  for trap_add_name in "$@"; do
    trap -- "$(
      extract_trap_cmd() { printf '%s\n' "$3"; }
      eval "extract_trap_cmd $(trap -p "${trap_add_name}")"
      printf '%s\n' "${trap_add_cmd}"
    )" "${trap_add_name}" || fatal "unable to add to trap ${trap_add_name}"
  done
}
declare -f -t trap_add

function get_platform() {
  unameOut="$(uname -s)"
  case "${unameOut}" in
    Linux*)
      echo "linux"
    ;;
    Darwin*)
      echo "darwin"
    ;;
    *)
      echo "Unsupported machine type :${unameOut}"
      exit 1
    ;;
  esac
}

PLATFORM=$(get_platform)
GOX="gox"
GOMETALINTER=$BINARY_DIR/gometalinter
GOMETALINTER_URL="https://github.com/alecthomas/gometalinter/releases/download/v2.0.4/gometalinter-2.0.4-$PLATFORM-amd64.tar.gz"
CONSUL=$BINARY_DIR/consul
CONSUL_URL="https://releases.hashicorp.com/consul/1.0.6/consul_1.0.6_${PLATFORM}_amd64.zip"

function download_consul() {
  if [ ! -f "$CONSUL" ]; then
    verbose "   --> $CONSUL"
    local tmpdir=`mktemp -d`
    trap_add "rm -rf $tmpdir" EXIT
    pushd $tmpdir
    curl -L -s -O $CONSUL_URL || fatal "failed to download '$CONSUL_URL': $?"
    for i in *.zip; do
      [ "$i" = "*.zip" ] && continue
      unzip -q "$i" && rm -r "$i"
    done
    popd
    mkdir -p $BINARY_DIR
    cp $tmpdir/* $BINARY_DIR/
  fi
}

function download_gometalinter() {
  if [ ! -f "$GOMETALINTER" ]; then
    verbose "   --> $GOMETALINTER"
    local tmpdir=`mktemp -d`
    trap_add "rm -rf $tmpdir" EXIT
    pushd $tmpdir
    curl -L -s -O $GOMETALINTER_URL || fatal "failed to download '$GOMETALINTER_URL': $?"
    for i in *.tar.gz; do
      [ "$i" = "*.tar.gz" ] && continue
      tar xzf "$i" -C $tmpdir --strip-components 1 && rm -r "$i"
    done
    popd
    mkdir -p $BINARY_DIR
    cp $tmpdir/* $BINARY_DIR/
  fi
}

function download_gox() {
  if [ ! -x "$(command -v $GOX)" ]; then
    echo "   --> $GOX"
    go get github.com/mitchellh/gox || fatal "go get 'github.com/mitchellh/gox' failed: $?"
  fi
}

function download_binaries() {
  download_gox || fatal "failed to download 'gox': $?"
  download_gometalinter || fatal "failed to download 'gometalinter': $?"
  download_consul || fatal "failed to download 'consul': $?"
  export PATH=$PATH:$BINARY_DIR
}

function run() {
  local revision=`git rev-parse HEAD`
  local branch=`git rev-parse --abbrev-ref HEAD`
  local host=`hostname`
  local buildDate=`date -u +"%Y-%m-%dT%H:%M:%SZ"`
  go version | grep -q 'go version go1.14.1 ' || fatal "go version is not 1.14.1"

  if [ -z "$TRAVIS" ]; then
    verbose "Cleanup dist..."
    rm -rf dist/*
  fi

  verbose "Fetching binaries..."
  download_binaries

  verbose "Getting dependencies..."

  local gofiles=$(find . -path ./vendor -prune -o -print | grep '\.go$')

  verbose "Installing dependencies..."
  go install ./... || fatal "go install failed: $?"
  go test -i ./... || fatal "go test install failed: $?"

  verbose "Formatting source..."
  if [[ ${#gofiles[@]} -gt 0 ]]; then
    while read -r gofile; do
      gofmt -s -w $PWD/$gofile
    done <<< "$gofiles"
  fi

  if [ -n "$TRAVIS" ] && [ -n "$(git status --porcelain)" ]; then
    fatal "Source not formatted"
  fi

  verbose "Linting source..."
  $GOMETALINTER --disable-all --enable=vet --enable=gocyclo --cyclo-over=15 --enable=golint --min-confidence=.85 --enable=ineffassign --skip=Godeps --skip=vendor --skip=third_party --skip=testdata --vendor ./... || fatal "gometalinter failed: $?"

  verbose "Checking licenses..."
  licRes=$(
  for file in $(find . -type f -iname '*.go' ! -path './vendor/*'); do
    head -n3 "${file}" | grep -Eq "(Copyright|generated|GENERATED)" || error "  Missing license in: ${file}"
  done;)
  if [ -n "${licRes}" ]; then
  	fatal "license header checking failed:\n${licRes}"
  fi

  verbose "Running tests..."
  if [ -n "$TRAVIS" ]; then
    if [ ! -x "$(command -v goveralls)" ]; then
      echo "Getting goveralls..."
      go get github.com/mattn/goveralls || fatal "go get 'github.com/mattn/goveralls' failed: $?"
    fi
    goveralls -v -service=travis-ci -ignore=main.go,testutil/server.go,testutil/golden.go || fatal "goveralls: $?"
  else
    go test -v ./... || fatal "$gopackage tests failed: $?"
  fi

  XC_ARCH=${XC_ARCH:-"386 amd64 arm arm64"}
  XC_OS=${XC_OS:-"solaris darwin freebsd linux windows"}
  if [ -z "$TRAVIS" ]; then
    XC_OS=$(go env GOOS)
    XC_ARCH=$(go env GOARCH)
  fi

  verbose "Building binaries..."
  $GOX -os="${XC_OS}" -arch="${XC_ARCH}" -osarch="!darwin/arm !darwin/arm64" -ldflags "-X github.com/kadaan/consulate/version.Version=$VERSION -X github.com/kadaan/consulate/version.Revision=$revision -X github.com/kadaan/consulate/version.Branch=$branch -X github.com/kadaan/consulate/version.BuildUser=$USER@$host -X github.com/kadaan/consulate/version.BuildDate=$buildDate" -output="dist/{{.Dir}}_{{.OS}}_{{.Arch}}"  || fatal "gox failed: $?"

  if [ -n "$TRAVIS" ]; then
    verbose "Creating archives..."
    cd dist
    set -x
    for f in *; do
      filename=$(basename "$f")
      extension="${filename##*.}"
      filename="${filename%.*}"
      if [[ "$filename" != "$extension" ]] && [[ -n "$extension" ]]; then
        extension=".$extension"
      else
        extension=""
      fi
      archivename="$filename.tar.gz"
      verbose "   --> $archivename"
      genericname="consulate$extension"
      mv -f "$f" "$genericname"
      tar -czf $archivename "$genericname"
      rm -rf "$genericname"
    done
  fi
}

run "$@"
