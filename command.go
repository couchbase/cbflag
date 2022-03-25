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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/couchbase/cbflag/man"
)

type Command struct {
	Name        string
	Desc        string
	ManPage     string
	Run         func()
	Hidden      bool
	IsManualCmd bool
	help        bool
	initialized bool
	Commands    []*Command
	Flags       []*Flag
}

// ExitCode defines a type of exit codes that can be used by functions that use the cbflag library. The exit codes that
// we are using follow the recommendation for restricting user-defined exit codes to the range 65-113 (in addition to 0,
// for success) described, for example, in https://tldp.org/LDP/abs/html/exitcodes.html or on Linux distributions in
// sysexits.h that can be found in the /usr/include directory.
type ExitCode uint8

const (
	ExitCodeSuccess       ExitCode = 0  // Successful termination
	ExitCodeCLIUsageError ExitCode = 64 // Command line usage error
)

func NewCommand(name, usage, manPage string, cb func()) *Command {
	rv := &Command{
		Name:        name,
		Desc:        usage,
		ManPage:     manPage,
		Run:         cb,
		Hidden:      false,
		IsManualCmd: false,
		Commands:    make([]*Command, 0),
		Flags:       make([]*Flag, 0),
	}

	return rv
}

func (c *Command) AddCommand(cmd *Command) {
	c.Commands = append(c.Commands, cmd)
}

func (c *Command) AddFlag(flag *Flag) {
	c.Flags = append(c.Flags, flag)
}

func (c *Command) initialize() {
	if c.initialized {
		return
	}

	c.initialized = true
	c.AddFlag(helpFlag(&c.help))

	for curIdx, curFlag := range c.Flags {
		for cmpIdx, cmpFlag := range c.Flags {
			if curIdx == cmpIdx {
				continue
			}

			if curFlag.short != "" && curFlag.short == cmpFlag.short {
				panic(fmt.Sprintf("Found multiple flags defined for `%s`", curFlag.short))
			}

			if curFlag.long != "" && curFlag.long == cmpFlag.long {
				panic(fmt.Sprintf("Found multiple flags defined for `%s`", curFlag.long))
			}
		}
	}
}

func (c *Command) parse(ctx *Context, args []string) ExitCode {
	c.initialize()

	if c.IsManualCmd {
		return c.showManual(ctx)
	}

	ctx.prevCmds = append(ctx.prevCmds, c.Name)
	if len(args) == 0 {
		// Check if there are flags set via environment variables
		return c.parseFlags(ctx, args)
	}

	if !strings.HasPrefix(args[0], "-") {
		return c.parseCommands(ctx, args)
	} else {
		return c.parseFlags(ctx, args)
	}
}

func (c *Command) parseCommands(ctx *Context, args []string) ExitCode {
	if len(args) == 0 {
		fmt.Fprint(ctx.cli.Writer, c.usageTitle(ctx)+c.Usage())
		// No commands are specified, print help and exit with 0 exit code
		return ExitCodeSuccess
	}

	for _, cmd := range c.Commands {
		if cmd.Name == args[0] {
			return cmd.parse(ctx, args[1:])
		}
	}

	fmt.Fprintf(ctx.cli.Writer, "Invalid subcommand `%s`\n\n", args[0])
	fmt.Fprint(ctx.cli.Writer, c.usageTitle(ctx)+c.Usage())
	// Got an invalid subcommand, exit with a non-zero exit code
	return ExitCodeCLIUsageError
}

func (c *Command) parseFlags(ctx *Context, args []string) ExitCode {
	// Process environment variables first
	var hasEnvironmentVar bool
	for i := 0; i < len(c.Flags); i++ {
		value := os.Getenv(c.Flags[i].env)
		if value != "" {
			err := c.Flags[i].value.Set(value)
			if err != nil {
				fmt.Fprintf(ctx.cli.Writer, "value of '%s' is not valid\n", c.Flags[i].env)
				fmt.Fprint(ctx.cli.Writer, c.usageTitle(ctx)+c.Usage())
				// Failed to parse flag, exit with a non-zero exit code
				return ExitCodeCLIUsageError
			}

			c.Flags[i].markFound(value, true, false)
			hasEnvironmentVar = true
			err = c.Flags[i].validate()

			if err != nil {
				fmt.Fprintf(ctx.cli.Writer, "%s\n\n", err.Error())
				fmt.Fprint(ctx.cli.Writer, c.usageTitle(ctx)+c.Usage())
				// Failed to validate flag, exit with a non-zero exit code
				return ExitCodeCLIUsageError
			}
		}
	}
	// If there are no Flags or Environment variables print the help
	if len(args) == 0 && !hasEnvironmentVar {
		fmt.Fprint(ctx.cli.Writer, c.usageTitle(ctx)+c.Usage())
		// Print help and exit with 0 exit code
		return ExitCodeSuccess
	}

	for i := 0; i < len(args); i++ {
		if !(strings.HasPrefix(args[i], "-") || strings.HasPrefix(args[i], "--")) {
			fmt.Fprintf(ctx.cli.Writer, "Expected flag: %s\n\n", args[i])
			fmt.Fprint(ctx.cli.Writer, c.usageTitle(ctx)+c.Usage())
			// Flag parser expects "-" or "--" prefix for a flag, exit with a non-zero exit code
			return ExitCodeCLIUsageError
		}

		flag, isDeprecated := c.findFlagByName(args[i])
		if flag == nil {
			fmt.Fprintf(ctx.cli.Writer, "Unknown flag: %s\n\n", args[i])
			fmt.Fprint(ctx.cli.Writer, c.usageTitle(ctx)+c.Usage())
			// Unknown flag specified, exit with a non-zero exit code
			return ExitCodeCLIUsageError
		}

		if isDeprecated {
			fmt.Fprintf(ctx.cli.Writer, "Warning: %s is deprecated, use -%s/--%s\n", args[i],
				flag.short, flag.long)
		}

		if flag.foundNonEnv() {
			extra := ""
			if flag.deprecatedFlagSpecified() {
				extra = fmt.Sprintf(" by a deprecated flag (%s)", flag.deprecatedFlagsString())
			}

			fmt.Fprintf(ctx.cli.Writer, "Argument for -%s/--%s already specified%s\n\n",
				flag.short, flag.long, extra)
			fmt.Fprint(ctx.cli.Writer, c.usageTitle(ctx)+c.Usage())
			// Argument for a flag is specified repeatedly, exit with a non-zero exit code
			return ExitCodeCLIUsageError
		}

		flag.markFound(args[i], false, isDeprecated)

		if !flag.isFlag {
			opt := args[i]
			value := ""
			if (i + 1) < len(args) {
				value = args[i+1]
			}

			value, hadOption, err := flag.optHandler(opt, value)
			if err != nil {
				fmt.Fprint(ctx.cli.Writer, err.Error())
				fmt.Fprint(ctx.cli.Writer, "\n\n"+c.usageTitle(ctx)+c.Usage())
				// Error in optHandler, exit with a non-zero exit code
				return ExitCodeCLIUsageError
			}

			if err := flag.value.Set(value); err != nil {
				fmt.Fprintf(ctx.cli.Writer, "Unable to process value for flag: %s. %s\n\n", args[i], err.Error())
				fmt.Fprint(ctx.cli.Writer, c.usageTitle(ctx)+c.Usage())
				// Failed to process value for flag, exit with a non-zero exit code
				return ExitCodeCLIUsageError
			}

			if hadOption {
				i++
			}
		} else {
			// We know we have a bool flag here and "true" will parse so no need to check the error
			flag.value.Set("true") //nolint:errcheck
		}

		if err := flag.validate(); err != nil {
			fmt.Fprintf(ctx.cli.Writer, "%s\n\n", err.Error())
			fmt.Fprint(ctx.cli.Writer, c.usageTitle(ctx)+c.Usage())
			// Failed to validate the flag, exit with a non-zero exit code
			return ExitCodeCLIUsageError
		}

	}

	// Check to see if the help flag was specified
	if c.help {
		flag, _ := c.findFlagByName("-h")
		if flag.foundLong && c.ManPage != "" {
			c.showManual(ctx)
		} else {
			fmt.Fprint(ctx.cli.Writer, c.usageTitle(ctx)+c.Usage())
		}

		return ExitCodeSuccess
	}

	// Check that all required flags have been specified
	allRequired := true
	for _, flag := range c.Flags {
		if flag.required && !flag.found() {
			fmt.Fprint(ctx.cli.Writer, "Flag required, but not specified: ")
			if flag.short != "" {
				fmt.Fprintf(ctx.cli.Writer, "-%s", flag.short)
			}

			if flag.short != "" && flag.long != "" {
				fmt.Fprint(ctx.cli.Writer, "/")
			}

			if flag.long != "" {
				fmt.Fprintf(ctx.cli.Writer, "--%s", flag.long)
			}

			fmt.Fprint(ctx.cli.Writer, "\n")
			allRequired = false
		}
	}

	if !allRequired {
		fmt.Fprintf(ctx.cli.Writer, "\n")
		fmt.Fprint(ctx.cli.Writer, c.usageTitle(ctx)+c.Usage())
		// Not all required flags have been specified, exit with a non-zero exit code
		return ExitCodeCLIUsageError
	}

	c.Run()
	return ExitCodeSuccess
}

func (c *Command) findFlagByName(f string) (*Flag, bool) {
	if strings.HasPrefix(f, "--") {
		f = f[2:]
	} else {
		f = strings.TrimPrefix(f, "-")
	}

	for _, flag := range c.Flags {
		if flag.short == f || flag.long == f {
			return flag, false
		}

		for _, dep := range flag.deprecated {
			if dep == f {
				return flag, true
			}
		}
	}

	return nil, false
}

func (c *Command) showManual(ctx *Context) ExitCode {
	mcmd := exec.Command("man", filepath.Join(ctx.cli.ManPath, c.ManPage))
	mcmd.Stdout = os.Stdout

	if err := man.ShowManual(ctx.cli.ManPath, c.ManPage); err != nil {
		fmt.Fprint(ctx.cli.Writer, err.Error()+"\n")
		return ExitCodeCLIUsageError
	}

	return ExitCodeSuccess
}

func (c *Command) usageTitle(ctx *Context) string {
	s := strings.Join(ctx.prevCmds, " ")

	if c.hasCommands() {
		s += " [<command>]"
	}

	if c.hasFlags() {
		s += " [<args>]"
	}

	s += "\n\n"
	return s
}

func (c *Command) Usage() string {
	s := ""
	if c.hasCommands() {
		maxLen := 0
		for _, cmd := range c.Commands {
			if len(cmd.Name) > maxLen {
				maxLen = len(cmd.Name)
			}
		}

		for _, cmd := range c.Commands {
			if !cmd.Hidden {
				spaces := strings.Repeat(" ", maxLen-len(cmd.Name))
				s += fmt.Sprintf("  %s%s   %s\n", cmd.Name, spaces, cmd.Desc)
			}
		}
		s += "\n"
	}

	if c.hasRequiredFlags() {
		s += "Required Flags:\n\n"
		for _, flag := range c.Flags {
			if flag.required {
				s += flag.usageString()
			}
		}
		s += "\n"
	}

	if c.hasOptionalFlags() {
		s += "Optional Flags:\n\n"
		for _, flag := range c.Flags {
			if !flag.required {
				s += flag.usageString()
			}
		}
		s += "\n"
	}

	return s
}

func (c *Command) hasCommands() bool {
	return len(c.Commands) > 0
}

func (c *Command) hasFlags() bool {
	return len(c.Flags) > 0
}

func (c *Command) hasOptionalFlags() bool {
	for _, flag := range c.Flags {
		if !flag.required {
			return true
		}
	}

	return false
}

func (c *Command) hasRequiredFlags() bool {
	for _, flag := range c.Flags {
		if flag.required {
			return true
		}
	}

	return false
}
