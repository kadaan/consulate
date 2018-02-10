#!/usr/bin/env bash

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
  echo "Building binaries..."
  gox -os="linux darwin" -arch="amd64" -output="dist/{{.Dir}}_{{.OS}}_{{.Arch}}"
  echo ""
  echo "done"
}

run "$@"
