/*
Copyright 2016-Present Couchbase, Inc.

Use of this software is governed by the Business Source License included in
the file licenses/BSL-Couchbase.txt.  As of the Change Date specified in that
file, in accordance with the Business Source License, use of this software will
be governed by the Apache License, Version 2.0, included in the file
licenses/APL2.txt.
*/

package cbflag

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExitCodeParseCommandsNoArgs tests that the parseCommands() function exits with Success (0) exit code if no
// arguments were specified (not an explicit error case, only help is printed).
func TestExitCodeParseCommandsNoArgs(t *testing.T) {
	command := NewCommand("", "", "", func() {})
	contextPtr := &Context{NewCLI("", ""), []string{}}
	exitCode := command.parseCommands(contextPtr, []string{})
	require.Equal(t, ExitCodeSuccess, exitCode)
}

// TestExitCodeParseCommandsWrongSubcommand tests that the parseCommands() function exits with CLIUsageError (64) exit
// code if a wrong subcommand was specified.
func TestExitCodeParseCommandsWrongSubcommand(t *testing.T) {
	command := NewCommand("actualName", "", "", func() {})
	contextPtr := &Context{NewCLI("", ""), []string{}}
	exitCode := command.parseCommands(contextPtr, []string{"wrongName"})
	require.Equal(t, ExitCodeCLIUsageError, exitCode)
}

// TestExitCodeParseFlagsNoArgs tests that the parseFlags() function exits with Success (0) exit code if no arguments
// were specified (not an explicit error case, only help is printed).
func TestExitCodeParseFlagsNoArgs(t *testing.T) {
	command := NewCommand("", "", "", func() {})
	contextPtr := &Context{NewCLI("", ""), []string{}}
	exitCode := command.parseFlags(contextPtr, []string{})
	require.Equal(t, ExitCodeSuccess, exitCode)
}

// TestExitCodeParseFlagsInvalidEnvFlag tests that the parseFlags() function exits with CLIUsageError (64) exit code if
// an unparsable (eg a string for an int flag) environment flag was specified.
func TestExitCodeParseFlagsUnparsableEnvFlag(t *testing.T) {
	command := NewCommand("", "", "", func() {})
	var result int
	flag := IntFlag(&result, 0, "", "", "FOO", "", []string{},
		func(Value) error { return assert.AnError }, true, true)
	command.AddFlag(flag)
	contextPtr := &Context{NewCLI("", ""), []string{}}
	os.Setenv("FOO", "thisIsNotAnInt")
	defer os.Unsetenv("FOO")
	exitCode := command.parseFlags(contextPtr, []string{})
	require.Equal(t, ExitCodeCLIUsageError, exitCode)
}

// TestExitCodeParseFlagsInvalidEnvFlag tests that the parseFlags() function exits with CLIUsageError (64) exit code if
// an environment flag with a value that couldn't be validated was specified.
func TestExitCodeParseFlagsInvalidEnvFlag(t *testing.T) {
	command := NewCommand("", "", "", func() {})
	result := ""
	clusterFlag := StringFlag(&result, "", "", "", "CB_CLUSTER", "", []string{},
		func(Value) error { return assert.AnError }, true, true)
	command.AddFlag(clusterFlag)
	contextPtr := &Context{NewCLI("", ""), []string{}}
	os.Setenv("CB_CLUSTER", "nonEmptyAddress")
	defer os.Unsetenv("CB_CLUSTER")
	exitCode := command.parseFlags(contextPtr, []string{})
	require.Equal(t, ExitCodeCLIUsageError, exitCode)
}

// TestExitCodeParseFlagsNotFlag tests that the parseFlags() function exits with CLIUsageError (64) exit code if not a
// flag (does not start with "-" or "--") was specified.
func TestExitCodeParseFlagsNotFlag(t *testing.T) {
	command := NewCommand("", "", "", func() {})
	contextPtr := &Context{NewCLI("", ""), []string{}}
	exitCode := command.parseFlags(contextPtr, []string{"notFlag"})
	require.Equal(t, ExitCodeCLIUsageError, exitCode)
}

// TestExitCodeParseFlagsUnknownFlag tests that the parseFlags() function exits with CLIUsageError (64) exit code if an
// unknown flag was specified.
func TestExitCodeParseFlagsUnknownFlag(t *testing.T) {
	command := NewCommand("", "", "", func() {})
	contextPtr := &Context{NewCLI("", ""), []string{}}
	exitCode := command.parseFlags(contextPtr, []string{"--unknownFlag"})
	require.Equal(t, ExitCodeCLIUsageError, exitCode)
}

// TestExitCodeParseFlagsFlagAlreadySpecified1 tests that the parseFlags() function exits with CLIUsageError (64) exit
// code if a flag is specified repeatedly.
func TestExitCodeParseFlagsFlagAlreadySpecified1(t *testing.T) {
	command := NewCommand("", "", "", func() {})
	result := ""
	clusterFlag := StringFlag(&result, "", "c", "", "", "", []string{},
		func(Value) error { return nil }, true, true)
	// In this case we simulate a flag being specified repeatedly by explicitly setting the variable that tracks if
	// this flag has already been found before we execute the parsing function.
	clusterFlag.foundShort = true
	command.AddFlag(clusterFlag)
	contextPtr := &Context{NewCLI("", ""), []string{}}
	exitCode := command.parseFlags(contextPtr, []string{"-c"})
	require.Equal(t, ExitCodeCLIUsageError, exitCode)
}

// TestExitCodeParseFlagsFlagAlreadySpecified2 tests that the parseFlags() function exits with CLIUsageError (64) exit
// code if a flag is specified repeatedly (more natural setup than the first test).
func TestExitCodeParseFlagsFlagAlreadySpecified2(t *testing.T) {
	command := NewCommand("", "", "", func() {})
	result := ""
	clusterFlag := StringFlag(&result, "", "c", "cluster", "", "", []string{},
		func(Value) error { return nil }, true, true)
	command.AddFlag(clusterFlag)
	contextPtr := &Context{NewCLI("", ""), []string{}}
	exitCode := command.parseFlags(contextPtr, []string{"-c", "address", "--cluster"})
	require.Equal(t, ExitCodeCLIUsageError, exitCode)
}

// TestExitCodeParseFlagsOptHandlerError tests that the parseFlags() function exits with CLIUsageError (64) exit code
// if the optHandler() function fails.
func TestExitCodeParseFlagsOptHandlerError(t *testing.T) {
	command := NewCommand("", "", "", func() {})
	result := ""
	clusterFlag := StringFlag(&result, "", "c", "", "", "", []string{},
		func(Value) error { return nil }, true, true)
	command.AddFlag(clusterFlag)
	contextPtr := &Context{NewCLI("", ""), []string{}}
	exitCode := command.parseFlags(contextPtr, []string{"-c"})
	require.Equal(t, ExitCodeCLIUsageError, exitCode)
}

// TestExitCodeParseFlagsFlagValueError tests that the parseFlags() function exits with CLIUsageError (64) exit code if
// an innappropriate value for a flag is provided.
func TestExitCodeParseFlagsFlagValueError(t *testing.T) {
	command := NewCommand("", "", "", func() {})
	result := 0
	clusterFlag := IntFlag(&result, 0, "c", "", "", "", []string{},
		func(Value) error { return nil }, true, true)
	command.AddFlag(clusterFlag)
	contextPtr := &Context{NewCLI("", ""), []string{}}
	exitCode := command.parseFlags(contextPtr, []string{"-c", "stringNotInt"})
	require.Equal(t, ExitCodeCLIUsageError, exitCode)
}

// TestExitCodeParseFlagsValidationError tests that the parseFlags() function exits with CLIUsageError (64) exit code if
// the value for a flag fails to get validated.
func TestExitCodeParseFlagsValidationError(t *testing.T) {
	command := NewCommand("", "", "", func() {})
	result := ""
	// In this case we simulate a validation fail by suplying a ValidatorFn function that fails on all inputs.
	clusterFlag := StringFlag(&result, "", "c", "", "", "", []string{},
		func(Value) error { return assert.AnError }, true, true)
	command.AddFlag(clusterFlag)
	contextPtr := &Context{NewCLI("", ""), []string{}}
	exitCode := command.parseFlags(contextPtr, []string{"-c", "address"})
	require.Equal(t, ExitCodeCLIUsageError, exitCode)
}

// TestExitCodeParseFlagsHelp tests that the parseFlags() function exits with Success (0) exit code if the help flag
// was specified (not an explicit error case, only help is printed).
func TestExitCodeParseFlagsHelp(t *testing.T) {
	command := NewCommand("", "", "", func() {})
	command.help = true
	result := true
	helpFlag := helpFlag(&result)
	command.AddFlag(helpFlag)
	contextPtr := &Context{NewCLI("", ""), []string{}}
	exitCode := command.parseFlags(contextPtr, []string{"-h"})
	require.Equal(t, ExitCodeSuccess, exitCode)
}

// TestExitCodeParseFlagsNotAllRequired tests that the parseFlags() function exits with CLIUsageError (64) exit code if
// not all required flags were specified.
func TestExitCodeParseFlagsNotAllRequired(t *testing.T) {
	command := NewCommand("", "", "", func() {})
	result := ""
	clusterFlag := StringFlag(&result, "", "c", "", "", "", []string{},
		func(Value) error { return nil }, true, true)
	missingRequiredFlag := StringFlag(&result, "", "", "", "", "", []string{},
		func(Value) error { return nil }, true, true)
	command.AddFlag(clusterFlag)
	command.AddFlag(missingRequiredFlag)
	contextPtr := &Context{NewCLI("", ""), []string{}}
	exitCode := command.parseFlags(contextPtr, []string{"-c", "address"})
	require.Equal(t, ExitCodeCLIUsageError, exitCode)
}

// TestExitCodeParseFlagsSuccess tests that the parseFlags() function exits with Success (0) exit code if its execution
// was successful.
func TestExitCodeParseFlagsSuccess(t *testing.T) {
	command := NewCommand("", "", "", func() {})
	result := ""
	clusterFlag := StringFlag(&result, "", "c", "", "", "", []string{},
		func(Value) error { return nil }, true, true)
	command.AddFlag(clusterFlag)
	contextPtr := &Context{NewCLI("", ""), []string{}}
	exitCode := command.parseFlags(contextPtr, []string{"-c", "address"})
	require.Equal(t, ExitCodeSuccess, exitCode)
}
