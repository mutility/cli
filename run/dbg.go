package run

import (
	"fmt"
)

// Debug parses command line arguments, printing any resulting error, selected command, and its options.
func (a *Application) Debug(args ...string) (*Command, error) {
	return a.DebugEnv(DefaultEnviron(), args...)
}

// DebugEnv parses command line arguments, printing any resulting error, selected command, and its options.
func (a *Application) DebugEnv(env Environ, args ...string) (*Command, error) {
	cmd, err := a.Parse(env.WithArgs(append(env.Args[:1:1], args...)))
	if err != nil {
		fmt.Fprintln(env.Stdout, args, "err:", err)
		if e, ok := err.(interface{ Command() *Command }); ok && e.Command() != nil {
			e.Command().debug(env)
		}
	} else {
		fmt.Fprintln(env.Stdout, args)
		cmd.debug(env)
		_, err = cmd.lookupHandler()
	}
	return cmd, err
}

func (c *Command) debug(env Environ) {
	if c == nil {
		fmt.Fprintln(env.Stdout, "  cmd: <nil>")
	} else {
		fmt.Fprintln(env.Stdout, "  cmd:", c.Name())
		for cmd, prefix := c, "  flag:"; cmd != nil; cmd, prefix = cmd.parent, "  "+prefix {
			for _, f := range cmd.Flags() {
				fmt.Fprintln(env.Stdout, prefix, f.option.debug())
			}
		}
		for cmd, prefix := c, "  arg:"; cmd != nil; cmd, prefix = cmd.parent, "  "+prefix {
			for _, a := range cmd.Args() {
				fmt.Fprintln(env.Stdout, prefix, a.option.debug())
			}
		}
	}
}

func (o *flagOnly[T]) debug() string { return o.name + "=" + fmt.Sprint(*o.value) }
func (o *option[T]) debug() string   { return o.name + "=" + fmt.Sprint(*o.value) }
func (o *options[T]) debug() string  { return o.name + "=" + fmt.Sprint(*o.value) }
