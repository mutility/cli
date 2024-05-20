package run

import (
	"fmt"
)

// Debug parses command line arguments, printing any resulting error, selected command, and its options.
func (a *Application) Debug(args ...string) {
	a.DebugEnv(DefaultEnviron(), args...)
}

// DebugEnv parses command line arguments, printing any resulting error, selected command, and its options.
func (a *Application) DebugEnv(env Environ, args ...string) {
	cmd, err := a.Parse(env.WithArgs(append(env.Args[:1:1], args...)))
	if err != nil {
		fmt.Fprintln(env.Stdout, args, "err:", err)
		if e, ok := err.(interface{ Command() *Command }); ok && e.Command() != nil {
			e.Command().debug(env)
		}
	} else {
		fmt.Fprintln(env.Stdout, args)
		cmd.debug(env)
	}
}

// debug returns a helpful debugging string. It is not constrained by compatibility.
func (c *Command) debug(env Environ) {
	if c == nil {
		fmt.Fprintln(env.Stdout, "  cmd: <nil>")
	} else {
		fmt.Fprintln(env.Stdout, "  cmd:", c.Name())
		for _, f := range c.Flags() {
			fmt.Fprintln(env.Stdout, "  flag:", f.option.debug())
		}
		for _, a := range c.Args() {
			fmt.Fprintln(env.Stdout, "  arg:", a.option.debug())
		}
		for _, c := range c.Commands() {
			fmt.Fprintln(env.Stdout, "  cmd:", c.name)
		}
	}
}

func (o *flagOnly[T]) debug() string { return o.name + "=" + fmt.Sprint(*o.value) }
func (o *option[T]) debug() string   { return o.name + "=" + fmt.Sprint(*o.value) }
func (o *options[T]) debug() string  { return o.name + "=" + fmt.Sprint(*o.value) }
