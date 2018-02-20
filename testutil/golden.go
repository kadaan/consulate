package testutil

import (
	"flag"
	"io/ioutil"
	"path/filepath"
	"testing"
)

var update = flag.Bool("test.update", false, "update .golden file")

// Get returns the golden file content. If the `test.update` is specified, it updates the
// file with the current output and returns it.
func Get(t *testing.T, actual []byte) []byte {
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
