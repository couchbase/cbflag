package cbflag

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/couchbase/backup/man"
)

type Command struct {
	Name     string
	Usage    string
	Run      func()
	help     bool
	commands []*Command
	flags    []*Flag
	parent   *Command
	cml      *CommandLine
}

func NewCommand(name, usage string, cb func()) *Command {
	rv := &Command{
		Name:     name,
		Usage:    usage,
		Run:      cb,
		commands: make([]*Command, 0),
		flags:    make([]*Flag, 0),
	}

	rv.AddFlag(helpFlag(&rv.help))
	return rv
}

func (c *Command) AddCommand(cmd *Command) {
	cmd.cml = c.cml
	cmd.parent = c
	c.commands = append(c.commands, cmd)
}

func (c *Command) AddFlag(flag *Flag) {
	c.flags = append(c.flags, flag)
}

func (c *Command) parse(args []string) {
	if len(args) == 0 {
		c.usage()
		return
	}

	if !strings.HasPrefix(args[0], "-") {
		c.parseCommands(args)
	} else {
		c.parseFlags(args)
	}
}

func (c *Command) parseCommands(args []string) {
	if len(args) == 0 {
		c.usage()
		return
	}

	for _, cmd := range c.commands {
		if cmd.Name == args[0] {
			cmd.parse(args[1:])
			return
		}
	}
	fmt.Fprintf(c.cml.out, "Invalid subcommand `%s`\n\n", args[0])
	c.usage()
}

func (c *Command) parseFlags(args []string) {
	if len(args) == 0 {
		c.usage()
		return
	}

	for i := 0; i < len(args); i++ {
		if !(strings.HasPrefix(args[i], "-") || strings.HasPrefix(args[i], "--")) {
			fmt.Fprintf(c.cml.out, "Expected flag: %s\n\n", args[i])
			c.usage()
			return
		}

		flag := c.findFlagByName(args[i])
		if flag == nil {
			fmt.Fprintf(c.cml.out, "Unknown flag: %s\n\n", args[i])
			c.usage()
			return
		}

		if flag.found() {
			fmt.Fprintf(c.cml.out, "Argument for -%/--% already specified\n\n", flag.short, flag.long)
			c.usage()
			return
		}

		flag.markFound(args[i])

		if !flag.isFlag {
			if (i + 1) > len(args) {
				fmt.Fprintf(c.cml.out, "Expected argument for flag: %s\n\n", args[i])
				c.usage()
				return
			}

			i++
			if err := flag.value.Set(args[i]); err != nil {
				fmt.Fprintf(c.cml.out, "Unable to process value for flag: %s\n\n", args[i])
				c.usage()
				return
			}
		} else {
			flag.value.Set("true")
		}

		if err := flag.validate(); err != nil {
			fmt.Fprintf(c.cml.out, "%s\n\n", err.Error())
			c.usage()
			return
		}

	}

	// Check to see if the help flag was specified
	if c.help {
		flag := c.findFlagByName("-h")
		if flag.foundLong {
			c.showManual(c.Name, "default")
		} else {
			c.usage()
		}
		return
	}

	// Check that all required flags have been specified
	allRequired := true
	for _, flag := range c.flags {
		if flag.required && !flag.found() {
			fmt.Fprintf(c.cml.out, "Flag required, but not specified: -%s/--%s\n", flag.short, flag.long)
			allRequired = false
		}
	}

	if !allRequired {
		fmt.Fprintf(c.cml.out, "\n")
		c.usage()
		return
	}

	c.Run()
}

func (c *Command) findFlagByName(f string) *Flag {
	if strings.HasPrefix(f, "--") {
		f = f[2:]
	} else if strings.HasPrefix(f, "-") {
		f = f[1:]
	}

	for _, flag := range c.flags {
		if flag.short == f || flag.long == f {
			return flag
		}
	}

	return nil
}

func (c *Command) showManual(page, installType string) {
	abspath, err := filepath.Abs(os.Args[0])
	if err != nil {
		fmt.Fprintf(c.cml.out, "Unable to get path to man files due to `%s`\n", err.Error())
		return
	}

	exedir := filepath.Dir(abspath)
	loc := man.CouchbaseInstallPath(exedir)
	if installType == "default" {
		loc = man.StandaloneInstallPath(exedir)
	}

	if err := man.ShowManual(loc, page); err != nil {
		fmt.Printf("%s\n", err.Error())
		return
	}
}

func (c *Command) usage() {
	s := c.getFullName()

	if c.hasCommands() {
		s += " [<command>]"
	}

	if c.hasFlags() {
		s += " [<args>]"
	}

	s += "\n\n"

	if c.hasCommands() {
		maxLen := 0
		for _, cmd := range c.commands {
			if len(cmd.Name) > maxLen {
				maxLen = len(cmd.Name)
			}
		}

		for _, cmd := range c.commands {
			spaces := strings.Repeat(" ", maxLen-len(cmd.Name))
			s += fmt.Sprintf("  %s%s   %s\n", cmd.Name, spaces, cmd.Usage)
		}
		s += "\n"
	}

	if c.hasRequiredFlags() {
		s += "Required Flags:\n\n"
		for _, flag := range c.flags {
			if flag.required {
				s += flag.usageString()
			}
		}
		s += "\n"
	}

	if c.hasOptionalFlags() {
		s += "Optional Flags:\n\n"
		for _, flag := range c.flags {
			if !flag.required {
				s += flag.usageString()
			}
		}
		s += "\n"
	}

	fmt.Fprint(c.cml.out, s)
}

func (c *Command) getFullName() string {
	if c.parent != nil {
		return c.parent.getFullName() + " " + c.Name
	}

	return c.Name
}

func (c *Command) hasCommands() bool {
	return len(c.commands) > 0
}

func (c *Command) hasFlags() bool {
	return len(c.flags) > 0
}

func (c *Command) hasOptionalFlags() bool {
	for _, flag := range c.flags {
		if !flag.required {
			return true
		}
	}

	return false
}

func (c *Command) hasRequiredFlags() bool {
	for _, flag := range c.flags {
		if flag.required {
			return true
		}
	}

	return false
}
