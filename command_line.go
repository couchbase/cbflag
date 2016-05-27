package cbflag

import (
	"os"
)

type Context struct {
	cli      *CommandLine
	prevCmds []string
}

type CommandLine struct {
	cmd     *Command
	manpath string
	out     *os.File
}

func NewCommandLine(progName, progUsage string) *CommandLine {
	rv := &CommandLine{
		cmd:     NewCommand(progName, progUsage, nil),
		manpath: "",
		out:     os.Stdout,
	}

	return rv
}

func (c *CommandLine) AddCommand(cmd *Command) {
	c.cmd.AddCommand(cmd)
}

func (c *CommandLine) AddFlag(flag *Flag) {
	c.cmd.AddFlag(flag)
}

func (c *CommandLine) SetManPath(path string) {
	c.manpath = path
}

func (c *CommandLine) Parse(args []string) {
	context := &Context{
		cli:      c,
		prevCmds: make([]string, 0),
	}
	c.cmd.parse(context, args[1:])
}

func (c *CommandLine) Usage() string {
	return c.cmd.Usage()
}
