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
)

type Context struct {
	cli      *CLI
	prevCmds []string
}

type CLI struct {
	Name     string
	Desc     string
	ManPath  string
	ManPage  string
	Run      func()
	Commands []*Command
	Flags    []*Flag
	Writer   *os.File
}

func NewCLI(progName, progUsage string) *CLI {
	return &CLI{
		Name:     progName,
		Desc:     progUsage,
		ManPath:  "",
		ManPage:  "",
		Run:      nil,
		Commands: make([]*Command, 0),
		Flags:    make([]*Flag, 0),
		Writer:   os.Stdout,
	}
}

func (c *CLI) AddCommand(cmd *Command) {
	c.Commands = append(c.Commands, cmd)
}

func (c *CLI) AddFlag(flag *Flag) {
	c.Flags = append(c.Flags, flag)
}

func (c *CLI) Parse(args []string) {
	cmd := &Command{
		Name:     c.Name,
		Desc:     c.Desc,
		Run:      c.Run,
		ManPage:  c.ManPage,
		Commands: c.Commands,
		Flags:    c.Flags,
	}

	context := &Context{
		cli:      c,
		prevCmds: make([]string, 0),
	}

	// Parse context and arguments, exit if parsing returns a non-zero exit code
	if exitCode := cmd.parse(context, args[1:]); exitCode != 0 {
		os.Exit(int(exitCode))
	}
}

func (c *CLI) Usage() string {
	cmd := &Command{
		Name:     c.Name,
		Desc:     c.Desc,
		Run:      nil,
		ManPage:  c.ManPage,
		Commands: c.Commands,
		Flags:    c.Flags,
	}

	return cmd.Usage()
}
