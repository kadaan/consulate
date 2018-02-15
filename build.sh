#!/usr/bin/env bash

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
  if [ ! -x "$(command -v gox)" ]; then
    echo "Getting gox..."
    go get github.com/mitchellh/gox
  fi
  echo "Formatting source..."
  local gofiles=$(find . -path ./vendor -prune -o -print | grep '\.go$')
  if [[ ${#gofiles[@]} -gt 0 ]]; then
    while read -r gofile; do
        gofmt -w $PWD/$gofile
    done <<< "$gofiles"
  fi
  echo "Building binaries..."
  local revision=`git rev-parse HEAD`
  local branch=`git rev-parse --abbrev-ref HEAD`
  local host=`hostname`
  local buildDate=`date -u +"%Y-%m-%dT%H:%M:%SZ"`
  gox -ldflags "-X github.com/kadaan/consulate/version.Version=$VERSION -X github.com/kadaan/consulate/version.Revision=$revision -X github.com/kadaan/consulate/version.Branch=$branch -X github.com/kadaan/consulate/version.BuildUser=$USER@$host -X github.com/kadaan/consulate/version.BuildDate=$buildDate" -output="dist/{{.Dir}}_{{.OS}}_{{.Arch}}"
  echo ""
  echo "done"
}

run "$@"
