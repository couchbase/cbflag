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
	ManName  string
	Commands []*Command
	Flags    []*Flag
	manpath  string
	Writer   *os.File
}

func NewCLI(progName, progUsage string) *CLI {
	return &CLI{
		Name:     progName,
		Desc:     progUsage,
		ManPath:  "",
		ManName:  "",
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

func (c *CLI) Usage() string {
	cmd := &Command{
		Name:     c.Name,
		Desc:     c.Desc,
		Run:      nil,
		Commands: c.Commands,
		Flags:    c.Flags,
	}

	return cmd.Usage()
}
