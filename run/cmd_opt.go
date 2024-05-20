package run

import "errors"

// CmdOptions specify options to apply to a command in CmdOpt.
type CmdOption interface {
	applyCommand(*Command) error
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
	var flags []Flag
	var args []Arg
	var cmds []*Command
	for _, opt := range opts {
		switch o := opt.(type) {
		case Flag:
			flags = append(flags, o)
		case Arg:
			args = append(args, o)
		case *Command:
			cmds = append(cmds, o)
		default:
			if err := opt.applyCommand(cmd); err != nil {
				errs = append(errs, err)
			}
		}
	}
	errs = appendOptSlice(errs, cmd.SetFlags, flags)
	errs = appendOptSlice(errs, cmd.SetArgs, args)
	errs = appendOptSlice(errs, cmd.SetCommands, cmds)
	return errors.Join(errs...)
}

func appendOptSlice[T any](errs []error, apply func(...T) error, s []T) []error {
	if len(s) > 0 {
		if err := apply(s...); err != nil {
			return append(errs, err)
		}
	}
	return errs
}
