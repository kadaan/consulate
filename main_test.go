package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
)

// Update golden file by passing `-test.update` to program arguments
var update = flag.Bool("test.update", false, "update .golden file")

func TestMainFunction(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		main()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestMainFunction")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	actual, err := cmd.Output()
	if err != nil {
		e, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("Command error could not be parsed as ExitError: %v", err)
		}
		if !e.Success() {
			t.Fatalf("ExitError: want 0, got %v", e.String())
		}
	}
	expected := get(t, []byte(actual))
	if !bytes.Equal(actual, expected) {
		t.Fatalf("Output ==> want '%v', got '%v'", strconv.Quote(string(expected)), strconv.Quote(string(actual)))
	}
}

func get(t *testing.T, actual []byte) []byte {
	golden := filepath.Join("testdata", t.Name()+".golden")
	if *update {
		if err := ioutil.WriteFile(golden, actual, 0644); err != nil {
			t.Fatal(err)
		}
	}
	expected, err := ioutil.ReadFile(golden)
	if err != nil {
		t.Fatal(err)
	}
	return expected
}
