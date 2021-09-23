/*
Copyright 2016-Present Couchbase, Inc.

Use of this software is governed by the Business Source License included in
the file licenses/BSL-Couchbase.txt.  As of the Change Date specified in that
file, in accordance with the Business Source License, use of this software will
be governed by the Apache License, Version 2.0, included in the file
licenses/APL2.txt.
*/

package pwd

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestGetPasswd tests the password creation and output based on a byte buffer
// as input to mock the underlying getch() methods.
func TestGetPasswd(t *testing.T) {
	type testData struct {
		input []byte

		// Due to how backspaces are written, it is easier to manually write
		// each expected output for the masked cases.
		masked    string
		password  string
		bytesLeft int
		reason    string
	}

	ds := []testData{
		testData{[]byte("abc\n"), "***", "abc", 0, "Password parsing should stop at \\n"},
		testData{[]byte("abc\r"), "***", "abc", 0, "Password parsing should stop at \\r"},
		testData{[]byte("a\nbc\n"), "*", "a", 3, "Password parsing should stop at \\n"},
		testData{[]byte("*!]|\n"), "****", "*!]|", 0, "Special characters shouldn't affect the password."},
		testData{[]byte("abc\r\n"), "***", "abc", 1,
			"Password parsing should stop at \\r; Windows LINE_MODE should be unset so \\r is not converted to \\r\\n."},
		testData{[]byte{'a', 'b', 'c', 8, '\n'}, "***\b \b", "ab", 0, "Backspace byte should remove the last read byte."},
		testData{[]byte{'a', 'b', 127, 'c', '\n'}, "**\b \b*", "ac", 0, "Delete byte should remove the last read byte."},
		testData{[]byte{'a', 'b', 127, 'c', 8, 127, '\n'}, "**\b \b*\b \b\b \b", "", 0, "Successive deletes continue to delete."},
		testData{[]byte{8, 8, 8, '\n'}, "", "", 0, "Deletes before characters are noops."},
		testData{[]byte{8, 8, 8, 'a', 'b', 'c', '\n'}, "***", "abc", 0, "Deletes before characters are noops."},
		testData{[]byte{'a', 'b', 0, 'c', '\n'}, "***", "abc", 0,
			"Nil byte should be ignored due; may get unintended nil bytes from syscalls on Windows."},
	}

	// Redirecting output for tests as they print to os.Stdout but we want to
	// capture and test the output.
	origStdOut := os.Stdout
	for _, masked := range []bool{true, false} {
		for _, d := range ds {
			pipeBytesToStdin(d.input)

			r, w, err := os.Pipe()
			require.NoError(t, err)
			os.Stdout = w

			result, err := getPasswd(masked)
			os.Stdout = origStdOut
			require.NoError(t, err)

			leftOnBuffer := flushStdin()

			// Test output (masked and unmasked).  Delete/backspace actually
			// deletes, overwrites and deletes again.  As a result, we need to
			// remove those from the pipe afterwards to mimic the console's
			// interpretation of those bytes.
			require.NoError(t, w.Close())

			output, err := ioutil.ReadAll(r)
			require.NoError(t, err)

			expectedOutput := make([]byte, 0, len(d.masked))
			if masked {
				expectedOutput = append(expectedOutput, d.masked...)
			}

			require.Equal(t, expectedOutput, output, d.reason)
			require.Equal(t, d.password, string(result), d.reason)
			require.Equal(t, d.bytesLeft, leftOnBuffer, d.reason)
		}
	}
}

// TestPipe ensures we get our expected pipe behavior.
func TestPipe(t *testing.T) {
	type testData struct {
		input    string
		password string
		expError error
	}
	ds := []testData{
		testData{"abc", "abc", io.EOF},
		testData{"abc\n", "abc", nil},
		testData{"abc\r", "abc", nil},
		testData{"abc\r\n", "abc", nil},
	}

	for _, d := range ds {
		_, err := pipeToStdin(d.input)
		if err != nil {
			t.Log("Error writing input to stdin:", err)
			t.FailNow()
		}
		pass, err := GetPasswd()
		if string(pass) != d.password {
			t.Errorf("Expected %q but got %q instead.", d.password, string(pass))
		}
		if err != d.expError {
			t.Errorf("Expected %v but got %q instead.", d.expError, err)
		}
	}
}

// flushStdin reads from stdin for .5 seconds to ensure no bytes are left on
// the buffer.  Returns the number of bytes read.
func flushStdin() int {
	ch := make(chan byte)
	go func(ch chan byte) {
		reader := bufio.NewReader(os.Stdin)
		for {
			b, err := reader.ReadByte()
			if err != nil { // Maybe log non io.EOF errors, if you want
				close(ch)
				return
			}
			ch <- b
		}
	}(ch)

	numBytes := 0
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				return numBytes
			}
			numBytes++
		case <-time.After(500 * time.Millisecond):
			return numBytes
		}
	}
}

// pipeToStdin pipes the given string onto os.Stdin by replacing it with an
// os.Pipe.  The write end of the pipe is closed so that EOF is read after the
// final byte.
func pipeToStdin(s string) (int, error) {
	pipeReader, pipeWriter, err := os.Pipe()
	if err != nil {
		fmt.Println("Error getting os pipes:", err)
		os.Exit(1)
	}
	os.Stdin = pipeReader
	w, err := pipeWriter.WriteString(s)
	pipeWriter.Close()
	return w, err
}

func pipeBytesToStdin(b []byte) (int, error) {
	return pipeToStdin(string(b))
}

// TestGetPasswd_Err tests errors are properly handled from getch()
func TestGetPasswd_Err(t *testing.T) {
	var inBuffer *bytes.Buffer
	getch = func() (byte, error) {
		b, err := inBuffer.ReadByte()
		if err != nil {
			return 13, err
		}
		if b == 'z' {
			return 'z', fmt.Errorf("Forced error; byte returned should not be considered accurate.")
		}
		return b, nil
	}
	defer func() { getch = defaultGetCh }()

	for input, expectedPassword := range map[string]string{"abc": "abc", "abzc": "ab"} {
		inBuffer = bytes.NewBufferString(input)
		p, err := GetPasswdMasked()
		if string(p) != expectedPassword {
			t.Errorf("Expected %q but got %q instead.", expectedPassword, p)
		}
		if err == nil {
			t.Errorf("Expected error to be returned.")
		}
	}
}

func TestMaxPasswordLength(t *testing.T) {
	type testData struct {
		input       []byte
		expectedErr error

		// Helper field to output in case of failure; rather than hundreds of
		// bytes.
		inputDesc string
	}

	ds := []testData{
		testData{append(bytes.Repeat([]byte{'a'}, maxLength), '\n'), nil, fmt.Sprintf("%v 'a' bytes followed by a newline", maxLength)},
		testData{append(bytes.Repeat([]byte{'a'}, maxLength+1), '\n'), ErrMaxLengthExceeded, fmt.Sprintf("%v 'a' bytes followed by a newline", maxLength+1)},
		testData{append(bytes.Repeat([]byte{0x00}, maxLength+1), '\n'), ErrMaxLengthExceeded, fmt.Sprintf("%v 0x00 bytes followed by a newline", maxLength+1)},
	}

	for _, d := range ds {
		pipeBytesToStdin(d.input)
		_, err := GetPasswd()
		if err != d.expectedErr {
			t.Errorf("Expected error to be %v; isntead got %v from %v", d.expectedErr, err, d.inputDesc)
		}
	}
}
