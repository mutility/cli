//nolint:errcheck
package run_test

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mutility/cli/run"
)

func printMain(app *run.Application, args ...string) {
	cmd, err := app.Parse(run.DefaultEnviron().WithArgs(append(os.Args[:1:1], args...)))
	if err != nil {
		if e, ok := err.(interface{ Command() *run.Command }); ok {
			cmd = e.Command()
			fmt.Println(args, "err:", err, "errcmd="+cmd.Debug()+debug("flags:", cmd.Flags())+debug("args:", cmd.Args())+debug("cmds:", cmd.Commands()))
		} else {
			fmt.Println(args, "err:", err)
		}
	} else {
		fmt.Println(args, "cmd:", cmd.Debug()+debug("flags:", cmd.Flags())+debug("args:", cmd.Args())+debug("cmds:", cmd.Commands()))
	}
}

func debug[T interface{ Debug() string }](name string, out []T) string {
	if len(out) == 0 {
		return ""
	}
	b := new(strings.Builder)
	fmt.Fprintf(b, " %s", name)
	for _, o := range out {
		fmt.Fprintf(b, " %s", o.Debug())
	}
	return b.String()
}

func runMain(app *run.Application, args ...string) {
	err := app.Main(context.TODO(), run.DefaultEnviron().WithArgs(append(os.Args[:1:1], args...)))
	if err != nil {
		app.Ferror(os.Stdout, err)
	}
}

func ExampleApp_empty() {
	app, _ := run.App("noarg", "")
	printMain(app)

	// output:
	// [] cmd: noarg
}

func ExampleMustApp_empty() {
	app := run.MustApp("nah", "", run.Flags(), run.Args())
	app.SetCommands()
	printMain(app, "hello")

	// output:
	// [hello] err: unexpected argument: "hello" errcmd=nah
}

func Example_nil() {
	app := run.MustApp("noway", "")
	printMain(app, "hello")

	// output:
	// [hello] err: unexpected argument: "hello" errcmd=noway
}

func ExampleString() {
	try := func(args ...string) {
		app := run.MustApp("str", "")
		argVal := run.String("arg", "")
		app.SetArgs(argVal.Pos("arg"))
		printMain(app, args...)
	}

	try("hello")
	try("-")
	try("-n")
	try("--no")
	try("--", "--okay")

	// output:
	// [hello] cmd: str args: arg=hello
	// [-] cmd: str args: arg=-
	// [-n] err: unexpected flag: -n errcmd=str args: arg=
	// [--no] err: unexpected flag: --no errcmd=str args: arg=
	// [-- --okay] cmd: str args: arg=--okay
}

func ExampleFileLike() {
	try := func(args ...string) {
		app, _ := run.App("file", "", run.Args(run.FileLike[quotedstring]("arg", "").Pos("arg")))
		printMain(app, args...)
	}

	try("hello")
	try("-")
	try("-n")
	try("--no")
	try("--", "--okay")

	// output:
	// [hello] cmd: file args: arg="hello"
	// [-] cmd: file args: arg="-"
	// [-n] err: unexpected flag: -n errcmd=file args: arg=""
	// [--no] err: unexpected flag: --no errcmd=file args: arg=""
	// [-- --okay] cmd: file args: arg="--okay"
}

func ExampleFileLikeSlice_many() {
	try := func(args ...string) {
		app := run.MustApp("files", "", run.Args(run.FileLikeSlice[quotedstring]("args", "").Rest("arg")))
		printMain(app, args...)
	}

	try("hello", "world")
	try("--", "head", "world")
	try("mid", "--", "world")
	try("trail", "world", "--")
	try("dash", "--", "--", "world")
	try("unquoted", "-n")
	try("quoted", "--", "-n")
	try("-", "-n")
	try("-", "--", "-n")

	// output:
	// [hello world] cmd: files args: args=["hello" "world"]
	// [-- head world] cmd: files args: args=["head" "world"]
	// [mid -- world] cmd: files args: args=["mid" "world"]
	// [trail world --] cmd: files args: args=["trail" "world"]
	// [dash -- -- world] cmd: files args: args=["dash" "--" "world"]
	// [unquoted -n] err: unexpected flag: -n errcmd=files args: args=["unquoted"]
	// [quoted -- -n] cmd: files args: args=["quoted" "-n"]
	// [- -n] err: unexpected flag: -n errcmd=files args: args=["-"]
	// [- -- -n] cmd: files args: args=["-" "-n"]
}

func ExampleIntSlice_many() {
	try := func(args ...string) {
		app := run.MustApp("ints", "", run.Args(run.IntSlice("arg", "", 0).Rest("arg")))
		printMain(app, args...)
	}

	try()
	try("5")
	try("6", "-2", "17", "+-3")

	// output:
	// [] err: expected "<arg> ..." errcmd=ints args: arg=[]
	// [5] cmd: ints args: arg=[5]
	// [6 -2 17 +-3] err: arg: parsing "+-3" as int: invalid syntax errcmd=ints args: arg=[6 -2 17]
}

func ExampleUintSlice_many() {
	try := func(args ...string) {
		app := run.MustApp("uints", "", run.Args(run.UintSlice("arg", "", 0).Rest("arg")))
		printMain(app, args...)
	}

	try()
	try("5", "2")
	try("6", "-2", "17") // -2 blocks the parse, then has no home

	// output:
	// [] err: expected "<arg> ..." errcmd=uints args: arg=[]
	// [5 2] cmd: uints args: arg=[5 2]
	// [6 -2 17] err: unexpected flag: -2 errcmd=uints args: arg=[6]
}

type quotedstring string

func (qs quotedstring) String() string {
	return strconv.Quote(string(qs))
}

func ExampleStringLike() {
	app := run.MustApp("quoted", "", run.Args(run.StringLike[quotedstring]("files", "").Pos("file")))
	printMain(app, "hello")

	// output:
	// [hello] cmd: quoted args: files="hello"
}

func ExampleIntLike() {
	type myint int8
	type myunt uint8
	try := func(args ...string) {
		app := run.MustApp("intish", "", run.Args(run.IntLike[myint]("i8", "", 10).Pos("i"), run.UintLike[myunt]("u8", "", 0).Pos("u")))
		printMain(app, args...)
	}

	try("-100", "200") // fine
	try("100", "-100") // -100 invalid
	try("200", "100")  // 200 out of range

	// output:
	// [-100 200] cmd: intish args: i8=-100 u8=200
	// [100 -100] err: u: parsing "-100" as run_test.myunt: invalid syntax errcmd=intish args: i8=100 u8=0
	// [200 100] err: i: parsing "200" as run_test.myint: value out of range errcmd=intish args: i8=0 u8=0
}

// Allow short flags to be grouped can alter error messages.
// Negative numbers might be sequences of short flags, so are only offered to relevant option types.
// (Normally only single-digit negative numbers look like a short flag; others are treated as positional.)
func ExampleApplication_AllowGroupShortFlags() {
	app := run.MustApp("shorts", "")
	u := run.UintLike[uint8]("u8", "", 10)
	app.SetArgs(u.Pos("u"))
	app.AllowGroupShortFlags(false) // default
	printMain(app, "-1")
	printMain(app, "-100")
	app.AllowGroupShortFlags(true)
	printMain(app, "-1")
	printMain(app, "-100")

	// output:
	// [-1] err: unexpected flag: -1 errcmd=shorts args: u8=0
	// [-100] err: u: parsing "-100" as uint8: invalid syntax errcmd=shorts args: u8=0
	// [-1] err: unexpected flag: -1 errcmd=shorts args: u8=0
	// [-100] err: unexpected flag: -100 errcmd=shorts args: u8=0
}

func ExampleStringOf_enum() {
	app := run.MustApp("enum", "", run.Args(run.StringOf[quotedstring]("letter", "", "alpha", "bravo", "charlie").Pos("abbrev")))
	printMain(app, "delta")
	printMain(app, "bravo")

	// output:
	// [delta] err: abbrev: "delta" not one of "alpha", "bravo", "charlie" errcmd=enum args: letter=""
	// [bravo] cmd: enum args: letter="bravo"
}

func ExampleFloatLike() {
	app := run.MustApp("floaty", "", run.Args(run.FloatLike[float64]("pct", "").Pos("pct")))
	printMain(app, "+-12.34")
	printMain(app, "12.34")
	printMain(app, "12")
	printMain(app, ".34")

	// output:
	// [+-12.34] err: pct: parsing "+-12.34" as float64: invalid syntax errcmd=floaty args: pct=0
	// [12.34] cmd: floaty args: pct=12.34
	// [12] cmd: floaty args: pct=12
	// [.34] cmd: floaty args: pct=0.34
}

func ExampleFloatLikeSlice() {
	app := run.MustApp("floats", "", run.Args(run.FloatLikeSlice[float64]("pcts", "").Rest("pct")))
	printMain(app, "12.34", "+12", "-.34", "+-12.34")

	// output:
	// [12.34 +12 -.34 +-12.34] err: pct: parsing "+-12.34" as float64: invalid syntax errcmd=floats args: pcts=[12.34 12 -0.34]
}

func ExampleNamedOf() {
	digit := run.NamedOf("digit", "", []run.NamedValue[int]{
		{Name: "one", Value: 1},
		{Name: "two", Value: 2},
		{Name: "three", Value: 3},
	})
	app := run.MustApp("named", "", run.Args(digit.Pos("d")))
	printMain(app, "four")
	printMain(app, "two")
	printMain(app, "four")

	// output:
	// [four] err: d: "four" not one of "one", "two", "three" errcmd=named args: digit=0
	// [two] cmd: named args: digit=2
	// [four] err: d: "four" not one of "one", "two", "three" errcmd=named args: digit=2
}

func ExampleNamedSliceOf() {
	digit := run.NamedSliceOf("digits", "", []run.NamedValue[int]{
		{Name: "one", Value: 1},
		{Name: "two", Value: 2},
		{Name: "three", Value: 3},
	})
	app := run.MustApp("nameds", "", run.Args(digit.Rest("digit")))
	printMain(app, "two", "three")
	printMain(app, "two", "four")

	// output:
	// [two three] cmd: nameds args: digits=[2 3]
	// [two four] err: digit: "four" not one of "one", "two", "three" errcmd=nameds args: digits=[2]
}

func ExampleApplication_SetFlags() {
	try := func(args ...string) {
		app := run.MustApp("flags", "")
		app.SetFlags(run.String("a", "").Flag(), run.String("b", "").Flag())
		printMain(app, args...)
	}
	try("--b", "beta")
	try("--b=gamma", "--a")

	// output:
	// [--b beta] cmd: flags flags: a= b=beta
	// [--b=gamma --a] err: --a: expected <value> errcmd=flags flags: a= b=gamma
}

func ExampleFlag_Default() {
	try := func(args ...string) {
		app := run.MustApp("default", "")
		digit := run.NamedOf("digit", "", []run.NamedValue[int]{
			{Name: "one", Value: 1},
			{Name: "two", Value: 2},
			{Name: "three", Value: 3},
		})
		app.SetFlags(digit.Flag().Default("two"))
		printMain(app, args...)
	}
	try()
	try("--digit", "three")
	try("--digit", "four")

	// output:
	// [] cmd: default flags: digit=2
	// [--digit three] cmd: default flags: digit=3
	// [--digit four] err: --digit: "four" not one of "one", "two", "three" errcmd=default flags: digit=0
}

func ExampleCommands() {
	app := run.MustApp("commands", "", run.Commands(
		run.MustCmd("foo", "does foo", run.Handler(func(context.Context, run.Environ) error { _, err := fmt.Println("running foo"); return err })),
		run.MustCmd("bar", "does bar", run.Handler(func(context.Context, run.Environ) error { _, err := fmt.Println("running bar"); return err })),
	))

	runMain(app)
	runMain(app, "foo")
	runMain(app, "bar")

	// output:
	// commands: error: expected <command>
	// running foo
	// running bar
}

func ExampleEnabler() {
	try := func(args ...string) {
		app := run.MustApp("enable", "", run.Flags(run.Enabler("en", "", false, true).Flag()))
		printMain(app, args...)
	}

	try()
	try("--en")
	try("--en", "--en")
	try("--en=true")

	// output:
	// [] cmd: enable flags: en=false
	// [--en] cmd: enable flags: en=true
	// [--en --en] err: --en: repeated errcmd=enable flags: en=true
	// [--en=true] err: unexpected flag value: --en=true errcmd=enable flags: en=false
}

func ExampleToggler() {
	try := func(args ...string) {
		app := run.MustApp("toggle", "", run.Flags(run.Toggler("en", "", false, true).Flag()))
		printMain(app, args...)
	}

	try()
	try("--en")
	try("--en", "--en")
	try("--en", "--en", "--en")
	try("--en=false")

	// output:
	// [] cmd: toggle flags: en=false
	// [--en] cmd: toggle flags: en=true
	// [--en --en] cmd: toggle flags: en=false
	// [--en --en --en] cmd: toggle flags: en=true
	// [--en=false] err: unexpected flag value: --en=false errcmd=toggle flags: en=false
}

func ExampleAccumulator() {
	try := func(args ...string) {
		app := run.MustApp("accum", "", run.Flags(
			run.Accumulator[quotedstring]("ha", "", "", "ha").Flag(),
			run.Accumulator("no", "", 0, -2).Flag()))
		printMain(app, args...)
	}

	try()
	try("--ha")
	try("--no")
	try("--ha", "--no", "--ha")
	try("--ha", "--no", "--ha", "--ha", "--no", "--no", "--no")
	try("--ha=boop")

	// output:
	// [] cmd: accum flags: ha="" no=0
	// [--ha] cmd: accum flags: ha="ha" no=0
	// [--no] cmd: accum flags: ha="" no=-2
	// [--ha --no --ha] cmd: accum flags: ha="haha" no=-2
	// [--ha --no --ha --ha --no --no --no] cmd: accum flags: ha="hahaha" no=-8
	// [--ha=boop] err: unexpected flag value: --ha=boop errcmd=accum flags: ha="" no=0
}
