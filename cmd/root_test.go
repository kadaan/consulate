package cmd

import (
	"bytes"
	"github.com/kadaan/consulate/testutil"
	"io"
	"os"
	"testing"
)

func TestRootCommand(t *testing.T) {
	rootCmd.SetArgs([]string{"version"})
	exitCode, _, actualStderr := executeCli(Execute)

	if exitCode != 0 {
		t.Errorf("ExitCode => got: %d, want: 0", exitCode)
	}

	expectedStderr := testutil.Get(t, actualStderr)
	if !bytes.Equal(expectedStderr, actualStderr) {
		t.Errorf("StdErr =>\n  want: %s\n  got: %s", string(expectedStderr), string(actualStderr))
	}
}

func TestRootCommandWithConfig(t *testing.T) {
	rootCmd.SetArgs([]string{"version"})
	cfgFile = "testdata/TestRootCommandWithConfig.config"
	exitCode, actualStdout, actualStderr := executeCli(Execute)

	t.Log(string(actualStdout))

	if exitCode != 0 {
		t.Errorf("ExitCode => got: %d, want: 0", exitCode)
	}

	expectedStderr := testutil.Get(t, actualStderr)
	if !bytes.Equal(expectedStderr, actualStderr) {
		t.Errorf("StdErr =>\n  want: %s\n  got: %s", string(expectedStderr), string(actualStderr))
	}
}

func executeCli(body func()) (int, []byte, []byte) {
	originalStdout := os.Stdout
	originalStderr := os.Stderr
	originalOsExit := osExit
	defer func() {
		os.Stdout = originalStdout
		os.Stderr = originalStderr
		osExit = originalOsExit
	}()

	stdoutR, stdoutW, _ := os.Pipe()
	os.Stdout = stdoutW

	stderrR, stderrW, _ := os.Pipe()
	os.Stderr = stderrW

	var exitCode int
	osExit = func(code int) {
		exitCode = code
	}

	stdoutC := make(chan []byte)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, stdoutR)
		stdoutC <- buf.Bytes()
	}()

	stderrC := make(chan []byte)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, stderrR)
		stderrC <- buf.Bytes()
	}()

	body()

	stdoutW.Close()
	stderrW.Close()
	stdout := <-stdoutC
	stderr := <-stderrC

	return exitCode, stdout, stderr
}
