package cbflag

import (
	"os"
)

type Context struct {
	cli      *CommandLine
	prevCmds []string
}

type CommandLine struct {
	Name     string
	Desc     string
	ManPath  string
	ManName  string
	Commands []*Command
	Flags    []*Flag
	manpath  string
	Writer   *os.File
}

func NewCommandLine(progName, progUsage string) *CommandLine {
	return &CommandLine{
		Name:     progName,
		Desc:     progUsage,
		ManPath:  "",
		ManName:  "",
		Commands: make([]*Command, 0),
		Flags:    make([]*Flag, 0),
		Writer:   os.Stdout,
	}
}

func (c *CommandLine) AddCommand(cmd *Command) {
	c.Commands = append(c.Commands, cmd)
}

func (c *CommandLine) AddFlag(flag *Flag) {
	c.Flags = append(c.Flags, flag)
}

func (c *CommandLine) Parse(args []string) {
	cmd := &Command{
		Name:     c.Name,
		Desc:     c.Desc,
		Run:      nil,
		Commands: c.Commands,
		Flags:    c.Flags,
	}

	context := &Context{
		cli:      c,
		prevCmds: make([]string, 0),
	}
	cmd.parse(context, args[1:])
}

func (c *CommandLine) Usage() string {
	cmd := &Command{
		Name:     c.Name,
		Desc:     c.Desc,
		Run:      nil,
		Commands: c.Commands,
		Flags:    c.Flags,
	}

	return cmd.Usage()
}
