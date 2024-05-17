package run

import (
	"cmp"
	"strconv"
	"strings"
)

type HelpDisabledError struct {
	cmd *Command
}

func (e HelpDisabledError) Error() string {
	return withCommand(e.cmd, "help requested")
}

type NotOneOfError[T any] struct {
	name  string
	names []NamedValue[T]
}

func (e NotOneOfError[T]) Error() string {
	if len(e.names) < 8 {
		n := make([]string, len(e.names))
		for i, nam := range e.names {
			n[i] = strconv.Quote(nam.Name)
		}
		return strconv.Quote(e.name) + " not one of " + strings.Join(n, ", ")
	}
	return strconv.Quote(e.name) + " unsupported value"
}

type missingFlagValueError struct {
	cmd   *Command
	flag  *Flag
	after string
}

func (e missingFlagValueError) Error() string {
	hint := cmp.Or(e.flag.hint, "<value>")
	return withCommand(e.cmd, e.after, "expected "+hint)
}

type missingCmdError struct {
	cmd *Command
}

func (e missingCmdError) Error() string {
	return withCommand(e.cmd, "expected <command>")
}

type flagArgUnconsumedError struct {
	cmd *Command
	arg string
	rem int
}

func (e flagArgUnconsumedError) Error() string {
	return withCommand(e.cmd, "unused", e.arg[:e.rem-1], strconv.Quote(e.arg[e.rem:]))
}

type flagParseError struct {
	cmd  *Command
	flag *Flag
	val  string
	err  error
}

func (e flagParseError) Error() string {
	return withCommand(e.cmd, e.val, e.err.Error())
}

type argParseError struct {
	cmd *Command
	arg *Arg
	val string
	err error
}

func (e argParseError) Error() string {
	return withCommand(e.cmd, e.arg.name, e.err.Error())
}

type extraFlagError struct {
	cmd  *Command
	flag string
}

func (e extraFlagError) Error() string {
	return withCommand(e.cmd, "unexpected flag", e.flag)
}

type extraArgsError struct {
	cmd  *Command
	args []string
}

func (e extraArgsError) Error() string {
	if len(e.args) == 1 {
		return withCommand(e.cmd, "unexpected argument", strconv.Quote(e.args[0]))
	}
	return withCommand(e.cmd, "unexpected arguments", strings.Join(e.args, " "))
}

type missingArgsError struct {
	cmd  *Command
	args []Arg
}

func (e missingArgsError) Error() string {
	a := make([]string, len(e.args))
	for i, arg := range e.args {
		a[i] = "<" + arg.name + ">"
		if arg.many {
			a[i] += " ..."
		}
	}
	return withCommand(e.cmd, "expected "+strconv.Quote(strings.Join(a, " ")))
}

type badFlagError struct {
	cmd  *Command
	flag *Flag
	at   string
}

func (e badFlagError) Error() string {
	return withCommand(e.cmd, "broken flag", e.at)
}

type badArgError struct {
	cmd *Command
	arg *Arg
	at  string
}

func (e badArgError) Error() string {
	return withCommand(e.cmd, "broken argument", e.at)
}

func withCommand(cmd *Command, s ...string) string {
	msg := strings.Join(s, ": ")
	if cmd.parent == nil {
		return msg
	}
	return cmd.CommandName() + ": " + msg
}
