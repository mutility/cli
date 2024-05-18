package run

import "errors"

// CmdOptions specify options to apply to a command in CmdOpt.
type CmdOption interface {
	applyCommand(*Command) error
}

// Flags sets the named options for a Command.
func Flags(flags ...Flag) CmdOption {
	return cmdOptionFunc(func(cmd *Command) error {
		return cmd.SetFlags(flags...)
	})
}

// Args sets the positional options for a Command.
func Args(args ...Arg) CmdOption {
	return cmdOptionFunc(func(cmd *Command) error {
		return cmd.SetArgs(args...)
	})
}

// Commands sets the subcommands for a Command.
func Commands(cmds ...*Command) CmdOption {
	return cmdOptionFunc(func(cmd *Command) error {
		return cmd.SetCommands(cmds...)
	})
}

// Details sets extra help information for a Command.
func Details(detail string) CmdOption {
	return cmdOptionFunc(func(cmd *Command) error {
		return cmd.SetDetails(detail)
	})
}

// Details sets extra help information for a Command, and links it to options.
func DetailsFor(detail string, opts ...Option) CmdOption {
	return cmdOptionFunc(func(cmd *Command) error {
		return cmd.SetDetailsFor(detail, opts...)
	})
}

type cmdOptionFunc func(*Command) error

func (f cmdOptionFunc) applyCommand(cmd *Command) error {
	return f(cmd)
}

func applyOpts(cmd *Command, opts []CmdOption) error {
	errs := make([]error, 0, len(opts))
	for _, opt := range opts {
		if err := opt.applyCommand(cmd); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
