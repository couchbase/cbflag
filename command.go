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
	Desc     string
	Run      func()
	help     bool
	commands []*Command
	flags    []*Flag
}

func NewCommand(name, usage string, cb func()) *Command {
	rv := &Command{
		Name:     name,
		Desc:     usage,
		Run:      cb,
		commands: make([]*Command, 0),
		flags:    make([]*Flag, 0),
	}

	rv.AddFlag(helpFlag(&rv.help))
	return rv
}

func (c *Command) AddCommand(cmd *Command) {
	c.commands = append(c.commands, cmd)
}

func (c *Command) AddFlag(flag *Flag) {
	c.flags = append(c.flags, flag)
}

func (c *Command) parse(ctx *Context, args []string) {
	ctx.prevCmds = append(ctx.prevCmds, c.Name)
	if len(args) == 0 {
		fmt.Fprint(ctx.cli.out, c.usageTitle(ctx)+c.Usage())
		return
	}

	if !strings.HasPrefix(args[0], "-") {
		c.parseCommands(ctx, args)
	} else {
		c.parseFlags(ctx, args)
	}
}

func (c *Command) parseCommands(ctx *Context, args []string) {
	if len(args) == 0 {
		fmt.Fprint(ctx.cli.out, c.usageTitle(ctx)+c.Usage())
		return
	}

	for _, cmd := range c.commands {
		if cmd.Name == args[0] {
			cmd.parse(ctx, args[1:])
			return
		}
	}
	fmt.Fprintf(ctx.cli.out, "Invalid subcommand `%s`\n\n", args[0])
	fmt.Fprint(ctx.cli.out, c.usageTitle(ctx)+c.Usage())
}

func (c *Command) parseFlags(ctx *Context, args []string) {
	if len(args) == 0 {
		fmt.Fprint(ctx.cli.out, c.usageTitle(ctx)+c.Usage())
		return
	}

	for i := 0; i < len(args); i++ {
		if !(strings.HasPrefix(args[i], "-") || strings.HasPrefix(args[i], "--")) {
			fmt.Fprintf(ctx.cli.out, "Expected flag: %s\n\n", args[i])
			fmt.Fprint(ctx.cli.out, c.usageTitle(ctx)+c.Usage())
			return
		}

		flag := c.findFlagByName(args[i])
		if flag == nil {
			fmt.Fprintf(ctx.cli.out, "Unknown flag: %s\n\n", args[i])
			fmt.Fprint(ctx.cli.out, c.usageTitle(ctx)+c.Usage())
			return
		}

		if flag.found() {
			fmt.Fprintf(ctx.cli.out, "Argument for -%/--% already specified\n\n", flag.short, flag.long)
			fmt.Fprint(ctx.cli.out, c.usageTitle(ctx)+c.Usage())
			return
		}

		flag.markFound(args[i])

		if !flag.isFlag {
			if (i + 1) >= len(args) {
				fmt.Fprintf(ctx.cli.out, "Expected argument for flag: %s\n\n", args[i])
				fmt.Fprint(ctx.cli.out, c.usageTitle(ctx)+c.Usage())
				return
			}

			i++
			if err := flag.value.Set(args[i]); err != nil {
				fmt.Fprintf(ctx.cli.out, "Unable to process value for flag: %s\n\n", args[i])
				fmt.Fprint(ctx.cli.out, c.usageTitle(ctx)+c.Usage())
				return
			}
		} else {
			flag.value.Set("true")
		}

		if err := flag.validate(); err != nil {
			fmt.Fprintf(ctx.cli.out, "%s\n\n", err.Error())
			fmt.Fprint(ctx.cli.out, c.usageTitle(ctx)+c.Usage())
			return
		}

	}

	// Check to see if the help flag was specified
	if c.help {
		flag := c.findFlagByName("-h")
		if flag.foundLong {
			c.showManual(c.Name, "default", ctx)
		} else {
			fmt.Fprint(ctx.cli.out, c.usageTitle(ctx)+c.Usage())
		}
		return
	}

	// Check that all required flags have been specified
	allRequired := true
	for _, flag := range c.flags {
		if flag.required && !flag.found() {
			fmt.Fprintf(ctx.cli.out, "Flag required, but not specified: -%s/--%s\n", flag.short, flag.long)
			allRequired = false
		}
	}

	if !allRequired {
		fmt.Fprintf(ctx.cli.out, "\n")
		fmt.Fprint(ctx.cli.out, c.usageTitle(ctx)+c.Usage())
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

func (c *Command) showManual(page, installType string, ctx *Context) {
	abspath, err := filepath.Abs(os.Args[0])
	if err != nil {
		fmt.Fprintf(ctx.cli.out, "Unable to get path to man files due to `%s`\n", err.Error())
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
		for _, cmd := range c.commands {
			if len(cmd.Name) > maxLen {
				maxLen = len(cmd.Name)
			}
		}

		for _, cmd := range c.commands {
			spaces := strings.Repeat(" ", maxLen-len(cmd.Name))
			s += fmt.Sprintf("  %s%s   %s\n", cmd.Name, spaces, cmd.Desc)
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

	return s
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
