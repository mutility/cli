package run

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
)

var (
	ErrRedefined = errors.New("already set")
	ErrMissing   = errors.New("missing")
)

func App(name, desc string) *Application {
	return &Application{
		Command: Command{
			name: name,
			desc: desc,
		},
	}
}

func AppOpt(name, desc string, opts ...CmdOption) (*Application, error) {
	app := App(name, desc)
	return app, applyOpts(&app.Command, opts)
}

type Application struct {
	Command

	allowGroupShortFlags bool
}

func (a *Application) AllowGroupShortFlags(f bool) {
	a.allowGroupShortFlags = f
}

func (a *Application) Ferror(w io.Writer, err error) {
	fmt.Fprintf(w, "%s: error: %v\n", a.Name(), err)
}

// Main parses arguments and attemps to run the specified command handler.
// If the command-line is invalid, it prints help for the selected command.
func (a *Application) Main(ctx context.Context, env Environ) error {
	cmd, err := a.Parse(env)
	switch err := err.(type) {
	case nil:
	case extraArgsError:
		return errors.Join(err, err.cmd.PrintHelp(ctx, env, a))
	case missingArgsError:
		return errors.Join(err, err.cmd.PrintHelp(ctx, env, a))
	default:
		return err
	}

	// non-leaf commands may have or omit a handler.
	if cmd.handler == nil && len(cmd.cmds) > 0 {
		// it's the user's mistake error to select a command with an omitted handler.
		return missingCmdError{cmd}
	}
	return cmd.run(ctx, env)
}

// Parse attemps to parse arguments and returns the selected command.
func (a *Application) Parse(env Environ) (*Command, error) {
	// options should implement one or more of the following to indicate what they accept.
	//   - flagParser is invoked for --name: parseFlag(); it accepts no arguments, and should not also implement valueParser
	//   - inlineParser is invoked for --name=val: parseInline("val")
	//   - valueParser is invoked for --name val: parseValue("val"), or positional ... val => parseValue("val")
	//   - valuesParser is invoked for positional ... a b c => parseValue(["a", "b", "c"]), and returns the count it parsed.
	//
	// Common combos include inlineParser+valueParser, valuesParser with or without valueParser, and flagParser with or without inlineParser.
	type (
		flagParser   interface{ parseFlag() error }
		inlineParser interface{ parseInline(string) error }
		valueParser  interface{ parseValue(string) error }
		valuesParser interface{ parseValues([]string) (int, error) }
	)

	if len(env.Args) < 1 {
		return nil, wrap(ErrMissing, "program name")
	}

	arg0 := env.Args[0]
	_ = arg0

	cur := &a.Command
	canFlag := true
	carg := 0
	showHelp := false

	maybeFlag := func(arg string) bool { return strings.HasPrefix(arg, "--") || (len(arg) == 2 && arg[0] == '-') }
	if a.allowGroupShortFlags {
		maybeFlag = func(arg string) bool { return strings.HasPrefix(arg, "-") }
	}

	for i := 1; i < len(env.Args); {
		arg := env.Args[i]
		if canFlag {
			if arg == "--" {
				canFlag = false
				i++
				continue
			}

			if idx, rem := cur.lookupFlag(arg); idx >= 0 && canFlag {
				opt := &cur.flags[idx]
				switch rem {
				case 0: // --arg possibly with following val
					switch parse := opt.option.(type) {
					case flagParser: // --arg <ignored>
						if err := parse.parseFlag(); err != nil {
							return nil, flagParseError{cur, opt, arg, err}
						}
						i += 1
					case valueParser: // --arg val
						if i+1 >= len(env.Args) {
							return nil, missingFlagValueError{cur, opt, arg}
						}
						if err := parse.parseValue(env.Args[i+1]); err != nil {
							return nil, flagParseError{cur, opt, arg, err}
						}
						i += 2
					default:
						return nil, badFlagError{cur, opt, arg}
					}
				default: // --arg=val; rem points to v
					switch parse := opt.option.(type) {
					case inlineParser:
						if err := parse.parseInline(arg[rem:]); err != nil {
							return nil, flagParseError{cur, opt, arg, err}
						}
						i += 1
					default:
						return nil, extraFlagValueError{cur, arg}
					}
				}
				opt.valueSet = true
				continue
			}

			if !cur.noHelp && (arg == "-h" || arg == "--help") {
				showHelp = true
				i++
				continue
			}

			if maybeFlag(arg) && (carg >= len(cur.args) || !cur.args[carg].can(arg)) {
				return nil, extraFlagError{cur, arg}
			}
		}

		if carg < len(cur.args) {
			opt := &cur.args[carg]
			switch parser := opt.option.(type) {
			case valuesParser:
				args := env.Args[i:]
				if !canFlag {
					took, err := parser.parseValues(args)
					if err != nil {
						return nil, argParseError{cur, opt, args[took], err}
					}
					i += took
				} else {
					uncan := len(args) // track end of canFlag, to handle where processing ends
					for i, arg := range args {
						if arg == "--" {
							uncan = i
							immutArgs := args
							args = make([]string, len(args)-1)
							copy(args[:i], immutArgs[:i])
							copy(args[i:], immutArgs[i+1:])
							break
						} else if !opt.can(arg) {
							args = args[:i]
							break
						}
					}
					took, err := parser.parseValues(args)
					if err != nil {
						return nil, argParseError{cur, opt, args[took], err}
					}
					i += took
					if took > uncan {
						i++
						canFlag = false
					}
				}

			case valueParser:
				if err := parser.parseValue(arg); err != nil {
					return nil, argParseError{cur, opt, arg, err}
				}
				i += 1
			default:
				return nil, badArgError{cur, opt, arg}
			}
			carg++
			continue
		}

		if idx := cur.lookupCmd(arg); idx >= 0 {
			cmd := cur.cmds[idx]
			cmd.parent = cur
			cur = cmd
			carg = 0
			i++
			continue
		}

		return nil, extraArgsError{cur, env.Args[i:]}
	}

	if showHelp {
		if cur.noHelp {
			return nil, HelpDisabledError{cur}
		}
		return helpCommand(a, cur), nil
	}

	if carg < len(cur.args) {
		return nil, missingArgsError{cur, cur.args[carg:]}
	}

	for cmd := cur; cmd != nil; cmd = cmd.parent {
		for f := range cmd.flags {
			flag := &cmd.flags[f]
			if flag.defaultSet && !flag.valueSet {
				err := flag.option.parseDefault(flag.defaultString)
				if err != nil {
					return cur, flagParseError{cur, flag, flag.defaultString, err}
				}
			}
		}
	}

	return cur, nil
}

func wrap(e error, m string) error {
	if e == nil {
		return e
	}
	return fmt.Errorf("%s: %w", m, e)
}
