package run

import (
	"fmt"
	"go/doc/comment"
	"io"
	"strings"
)

func helpCommand(a *Application, cmd *Command) *Command {
	return &Command{name: cmd.Name() + ".--help", handler: func(ctx Context) error {
		return cmd.PrintHelp(ctx, a)
	}}
}

// PrintHelp writes usage information for this command to env.Stdout.
func (c *Command) PrintHelp(ctx Context, a *Application) error {
	return writeUsage(ctx.Stdout, a, c)
}

func writeUsage(w io.Writer, app *Application, cmd *Command) error {
	usage := []any{"Usage:", app.name}
	if cmd != &app.Command {
		usage = append(usage, cmd.CommandName())
	}
	if len(cmd.cmds) > 0 {
		usage = append(usage, "<command>")
	}
	if len(cmd.flags) > 0 || (cmd == &app.Command && len(cmd.cmds) > 0) {
		usage = append(usage, "[flags]")
	}
	for _, arg := range cmd.args {
		usage = append(usage, arg.describe())
	}
	fmt.Fprintln(w, usage...)

	if len(cmd.desc) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, cmd.desc)
	}
	if len(cmd.detail) > 0 {
		fmt.Fprintln(w)
		_, _ = w.Write((&comment.Printer{
			TextCodePrefix: "    ",
			TextWidth:      80,
		}).Text(new(comment.Parser).Parse(cmd.detail)))
	}

	if len(cmd.args) > 0 {
		args := makeTable("Arguments:")
		for _, arg := range cmd.args {
			args.Add(arg.describe(), arg.option.description())
			for _, also := range arg.option.seeAlso() {
				if cmd != also {
					args.Add("", fmt.Sprintf("(See %s %s --help)", app.name, also.name))
				}
			}

		}
		args.Write(w)
	}

	// commands have help at least help flags, unless suppressed
	if len(cmd.flags) > 0 || !cmd.noHelp {
		anyRuneString := !app.noHelp // help includes a rune+string
		for _, flag := range cmd.flags {
			if flag.rune != 0 && flag.string != "" {
				anyRuneString = true
			}
		}

		flags := makeTable("Flags:")
		flags.Max = 22
		if !cmd.noHelp {
			flags.Add("-h, --help", "Show context-sensitive help.")
		}
		for _, flag := range cmd.flags {
			var names []string
			if flag.rune != 0 {
				names = append(names, "-"+string(flag.rune))
			}
			if flag.string != "" {
				names = append(names, "--"+flag.string)
			}
			name := strings.Join(names, ", ")
			if anyRuneString && flag.rune == 0 {
				name = "    " + name
			}
			if flag.defaultSet {
				name += "=" + flag.defaultString
			} else if p := flag.hint; p != "" {
				name += "=" + p
			}

			flags.Add(name, flag.option.description())
			for _, also := range flag.option.seeAlso() {
				if cmd != also {
					flags.Add("", fmt.Sprintf("(See %s %s --help)", app.name, also.name))
				}
			}
		}
		flags.Write(w)
	}

	if len(cmd.cmds) > 0 {
		cmds := makeTable("Commands:")
		for _, cmd := range cmd.cmds {
			if !cmd.unlisted {
				cmds.Add(cmd.name, cmd.desc)
			}
		}
		cmds.Write(w)
		fmt.Fprintln(w, "\nRun \""+app.name+" <command> --help\" for more information on a command.")
	}

	_, err := w.Write(nil)
	return err
}

func makeTable(name string) table {
	return table{Name: name, Min: 6, Max: 12, Pad: 3}
}

type table struct {
	Name     string
	Items    [][2]string
	Min, Max int
	Pad      int
}

func (t *table) Add(col1, col2 string) { t.Items = append(t.Items, [2]string{col1, col2}) }

func (t *table) Write(w io.Writer) {
	longest := t.Min
	for _, it := range t.Items {
		longest = max(longest, len(it[0]))
	}
	if t.Max > 0 {
		longest = min(longest, t.Max)
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, t.Name)
	for _, it := range t.Items {
		if len(it[0]) > longest {
			fmt.Fprintf(w, "  %s\n  %-*s %s\n", it[0], longest+t.Pad, "", it[1])
		} else {
			fmt.Fprintf(w, "  %-*s %s\n", longest+t.Pad, it[0], it[1])
		}
	}
}
