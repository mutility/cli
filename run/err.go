package run

import (
	"cmp"
	"strconv"
	"strings"
)

type errCmd Command

func ec(cmd *Command) *errCmd { return (*errCmd)(cmd) }

func (e *errCmd) Command() *Command { return (*Command)(e) }
func (e *errCmd) msg(s ...string) string {
	msg := strings.Join(s, ": ")
	if e.parent == nil {
		return msg
	}
	return e.Command().CommandName() + ": " + msg
}

type HelpDisabledError struct {
	*errCmd
}

func (e HelpDisabledError) Error() string {
	return e.msg("help requested")
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
	*errCmd
	flag  *Flag
	after string
}

func (e missingFlagValueError) Error() string {
	hint := cmp.Or(e.flag.hint, "<value>")
	return e.msg(e.after, "expected "+hint)
}

type missingCmdError struct {
	*errCmd
}

func (e missingCmdError) Error() string {
	return e.msg("expected <command>")
}

type extraFlagValueError struct {
	*errCmd
	arg string
}

func (e extraFlagValueError) Error() string {
	return e.msg("unexpected flag value", e.arg)
}

type flagParseError struct {
	*errCmd
	flag *Flag
	val  string
	err  error
}

func (e flagParseError) Error() string {
	return e.msg(e.val, e.err.Error())
}

type argParseError struct {
	*errCmd
	arg *Arg
	val string
	err error
}

func (e argParseError) Error() string {
	return e.msg(e.arg.name, e.err.Error())
}

type extraFlagError struct {
	*errCmd
	flag string
}

func (e extraFlagError) Error() string {
	return e.msg("unexpected flag", e.flag)
}

type extraArgsError struct {
	*errCmd
	args []string
}

func (e extraArgsError) Error() string {
	if len(e.args) == 1 {
		return e.msg("unexpected argument", strconv.Quote(e.args[0]))
	}
	return e.msg("unexpected arguments", strings.Join(e.args, " "))
}

type missingArgsError struct {
	*errCmd
	args []Arg
}

func (e missingArgsError) Error() string {
	a := make([]string, len(e.args))
	for i, arg := range e.args {
		a[i] = "<" + arg.name + ">"
		if _, ok := arg.option.(valuesParser); ok {
			a[i] += " ..."
		}
	}
	return e.msg("expected " + strconv.Quote(strings.Join(a, " ")))
}

type badFlagError struct {
	*errCmd
	flag *Flag
	at   string
}

func (e badFlagError) Error() string {
	return e.msg("broken flag", e.at)
}

type badArgError struct {
	*errCmd
	arg *Arg
	at  string
}

func (e badArgError) Error() string {
	return e.msg("broken argument", e.at)
}
