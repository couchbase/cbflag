package cbflag

import (
	"os"
)

type CommandLine struct {
	args    []string
	cmd     *Command
	parsed  bool
	manpath string
	out     *os.File
}

func NewCommandLine(progName, progUsage string, args []string) *CommandLine {
	rv := &CommandLine{
		args:    args[1:],
		cmd:     NewCommand(progName, progUsage, nil),
		parsed:  false,
		manpath: "",
		out:     os.Stdout,
	}
	rv.cmd.cml = rv
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

func (c *CommandLine) Parse() {
	c.parsed = true
	c.cmd.parse(c.args)
}

func (c *CommandLine) Parsed() bool {
	return c.parsed
}

func (c *CommandLine) Usage() {
	c.cmd.usage()
}
