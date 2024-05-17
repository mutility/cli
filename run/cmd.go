package run

import (
	"cmp"
	"context"
	"slices"
	"strings"
)

// Flag represents the named options for a Command.
type Flag struct {
	rune          rune
	string        string
	option        Option
	hint          string
	defaultString string
	defaultSet    bool
	valueSet      bool
}

// Default specifies a value that will be supplied for an unprovided flag.
func (f Flag) Default(string string) Flag {
	f.defaultString = string
	f.defaultSet = true
	return f
}

type flags []Flag

func (f flags) searchRune(index int, r rune) int     { return cmp.Compare(f[index].rune, r) }
func (f flags) searchString(index int, s string) int { return cmp.Compare(f[index].string, s) }

type commands []*Command

func (c commands) searchString(index int, s string) int { return cmp.Compare(c[index].name, s) }

// Arg represents a positional option for a Command.
type Arg struct {
	option Option
	name   string
	many   bool
}

func (a Arg) can(dashArg string) (ok bool) {
	if slices.Contains(a.option.okValues(), dashArg) {
		return true
	}

	nonDash := strings.IndexFunc(dashArg, func(r rune) bool { return r != '-' })
	return nonDash >= 0 && strings.HasPrefix(a.option.okPrefix(), dashArg[:nonDash])
}

func (a Arg) describe() string {
	desc := "<" + a.name + ">"
	if a.many {
		desc += " ..."
	}
	return desc
}

// Cmd creates a command.
func Cmd(name, desc string) *Command {
	return &Command{name: name, desc: desc}
}

// Cmd creates a command and applies options.
func CmdOpt(name, desc string, opts ...CmdOption) (*Command, error) {
	cmd := Cmd(name, desc)
	return cmd, applyOpts(cmd, opts)
}

// A command is a 'verb' for an application.
// It can be configured with Flags and Args, and a Handler.
// An application starts with an implicit root command, to which other "sub" commands can be added.
type Command struct {
	name, desc, detail string

	parent *Command
	cmds   commands
	flags  flags
	args   []Arg

	clookup func(arg string) int              // returns index in cmds of matching *Command (or -1)
	flookup func(arg string) (index, rem int) // returns index in flags of matching flag (or -1), index in arg after = (or 0)

	handler  func(context.Context, Environ) error
	noHelp   bool // don't offer -h|--help for this command
	unlisted bool // don't list this command in its parents help
}

// CommandName returns the hierarchical name for a command.
//
// For example, if this command represented git commit, it would return "commit".
func (c *Command) CommandName() string {
	parts := []string{c.name}
	for c.parent != nil {
		c = c.parent
		parts = append(parts, c.name)
	}
	slices.Reverse(parts)
	return strings.Join(parts[1:], ".")
}

// Name returns the hierarchical program name for a command.
//
// For example, if this command represented git commit, it would return "git.commit".
func (c *Command) Name() string {
	parts := []string{c.name}
	for c.parent != nil {
		c = c.parent
		parts = append(parts, c.name)
	}
	slices.Reverse(parts)
	return strings.Join(parts, ".")
}

// Details sets extra help information for a Command.
// Attempting to set details more than once causes an error.
func (c *Command) Details(detail string) error {
	if c.detail != "" {
		return wrap(ErrRedefined, c.name+" detail")
	}
	c.detail = detail
	return nil
}

// Details sets extra help information for a Command, and links it to options.
// Options reused in other commands will point to this command for further information.
//
// Attempting to set details more than once causes an error.
func (c *Command) DetailsFor(detail string, opts ...Option) error {
	if c.detail != "" {
		return wrap(ErrRedefined, c.name+" detail")
	}
	c.detail = detail
	for _, opt := range opts {
		opt.setSeeAlso(c)
	}
	return nil
}

// Runs sets the handler for a Command.
// Attempting to set more than one handler causes an error.
func (c *Command) Runs(handler func(context.Context, Environ) error) error {
	if c.handler != nil {
		return wrap(ErrRedefined, c.name+" handler")
	}
	c.handler = handler
	return nil
}

// Flags sets the named options for a command.
// Attempting to set them more than once causes an error.
func (c *Command) Flags(flags ...Flag) error {
	if c.flookup != nil {
		return wrap(ErrRedefined, c.name+" flags")
	}
	c.flags = flags

	runeFlags, stringFlags := 0, 0
	for _, f := range flags {
		if f.rune != 0 {
			runeFlags++
		}
		if f.string != "" {
			stringFlags++
		}
	}

	runeIndex := make([]int, 0, runeFlags)
	stringIndex := make([]int, 0, stringFlags)

	for i, f := range flags {
		if f.rune != 0 {
			runeIndex = append(runeIndex, i)
		}
		if f.string != "" {
			stringIndex = append(stringIndex, i)
		}
	}
	slices.SortFunc(runeIndex, func(a, b int) int { return cmp.Compare(c.flags[a].rune, flags[b].rune) })
	slices.SortFunc(stringIndex, func(a, b int) int { return cmp.Compare(c.flags[a].string, flags[b].string) })

	c.flookup = func(arg string) (index, rem int) {
		switch {
		case len(arg) == 0 || arg[0] != '-' || arg == "-":
			return -1, 0
		case arg[1] == '-':
			pos, ok := slices.BinarySearchFunc(stringIndex, arg[2:], c.flags.searchString)
			if !ok {
				eq := strings.IndexByte(arg, '=')
				if eq >= 3 {
					rem = eq + 1
					pos, ok = slices.BinarySearchFunc(stringIndex, arg[2:eq], c.flags.searchString)
				}
			}
			if ok {
				return stringIndex[pos], rem
			}
		default:
			pos, ok := slices.BinarySearchFunc(runeIndex, ([]rune(arg))[1], c.flags.searchRune)
			if ok {
				return runeIndex[pos], 0
			}
		}

		return -1, 0
	}

	return nil
}

// Args sets the positional options for a command.
// Attempting to set them more than once causes an error.
func (c *Command) Args(args ...Arg) error {
	if c.args != nil {
		return wrap(ErrRedefined, c.name+" args")
	}
	c.args = args
	return nil
}

// Commands sets the named subcommands for a command.
// Attempting to set them more than once causes an error.
func (c *Command) Commands(cmds ...*Command) error {
	if c.clookup != nil {
		return wrap(ErrRedefined, c.name+" commands")
	}
	c.cmds = cmds

	stringIndex := make([]int, 0, len(cmds))
	for i, sub := range cmds {
		if sub.name != "" {
			stringIndex = append(stringIndex, i)
		}
	}
	slices.SortFunc(stringIndex, func(a, b int) int { return cmp.Compare(cmds[a].name, cmds[b].name) })

	c.clookup = func(name string) int {
		pos, ok := slices.BinarySearchFunc(stringIndex, name, c.cmds.searchString)
		if ok {
			return stringIndex[pos]
		}
		return -1
	}

	return nil
}

// lookupCmd returns the index of the matching *Command (or -1)
func (c *Command) lookupCmd(arg string) int {
	if c.clookup == nil {
		return -1
	}
	return c.clookup(arg)
}

func (c *Command) run(ctx context.Context, env Environ) error {
	if c.handler == nil {
		panic(c.name + ": not handled")
	}
	return c.handler(ctx, env)
}

// lookupFlag returns the index of the matching flag (or -1), and the index in arg after an = (or 0)
func (c *Command) lookupFlag(arg string) (index, rem int) {
	if c.flookup == nil {
		return -1, 0
	}
	return c.flookup(arg)
}
