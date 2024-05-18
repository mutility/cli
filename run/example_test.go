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

func Example_whyrun() {
	app := run.App("runtest", "testing run")
	printMain(app)

	// output:
	// [] cmd: runtest
}

func Example_empty_hello() {
	app := run.App("runtest", "testing run")
	app.SetFlags()
	app.SetArgs()
	app.SetCommands()

	printMain(app, "hello")

	// output:
	// [hello] err: unexpected argument: "hello" errcmd=runtest
}

func Example_nil() {
	app := run.App("runtest", "testing run")
	printMain(app, "hello")

	// output:
	// [hello] err: unexpected argument: "hello" errcmd=runtest
}

func ExampleString() {
	try := func(args ...string) {
		app := run.App("runtest", "testing run")
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
	// [hello] cmd: runtest args: arg=hello
	// [-] cmd: runtest args: arg=-
	// [-n] err: unexpected flag: -n errcmd=runtest args: arg=
	// [--no] err: unexpected flag: --no errcmd=runtest args: arg=
	// [-- --okay] cmd: runtest args: arg=--okay
}

func ExampleFileLike() {
	try := func(args ...string) {
		app := run.App("runtest", "testing run")
		argVal := run.FileLike[quotedstring]("arg", "")
		app.SetArgs(argVal.Pos("arg"))
		printMain(app, args...)
	}

	try("hello")
	try("-")
	try("-n")
	try("--no")
	try("--", "--okay")

	// output:
	// [hello] cmd: runtest args: arg="hello"
	// [-] cmd: runtest args: arg="-"
	// [-n] err: unexpected flag: -n errcmd=runtest args: arg=""
	// [--no] err: unexpected flag: --no errcmd=runtest args: arg=""
	// [-- --okay] cmd: runtest args: arg="--okay"
}

func ExampleFileLikeSlice_many() {
	try := func(args ...string) {
		app := run.App("runtest", "testing run")
		argVal := run.FileLikeSlice[quotedstring]("arg", "")
		app.SetArgs(argVal.Rest("arg"))
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
	// [hello world] cmd: runtest args: arg=["hello" "world"]
	// [-- head world] cmd: runtest args: arg=["head" "world"]
	// [mid -- world] cmd: runtest args: arg=["mid" "world"]
	// [trail world --] cmd: runtest args: arg=["trail" "world"]
	// [dash -- -- world] cmd: runtest args: arg=["dash" "--" "world"]
	// [unquoted -n] err: unexpected flag: -n errcmd=runtest args: arg=["unquoted"]
	// [quoted -- -n] cmd: runtest args: arg=["quoted" "-n"]
	// [- -n] err: unexpected flag: -n errcmd=runtest args: arg=["-"]
	// [- -- -n] cmd: runtest args: arg=["-" "-n"]
}

func ExampleIntSlice_many() {
	try := func(args ...string) {
		app := run.App("runtest", "testing ints")
		argVal := run.IntSlice("arg", "", 0)
		app.SetArgs(argVal.Rest("arg"))
		printMain(app, args...)
	}

	try()
	try("5")
	try("6", "-2", "17", "+-3")

	// output:
	// [] err: expected "<arg> ..." errcmd=runtest args: arg=[]
	// [5] cmd: runtest args: arg=[5]
	// [6 -2 17 +-3] err: arg: parsing "+-3" as int: invalid syntax errcmd=runtest args: arg=[6 -2 17]
}

func ExampleUintSlice_many() {
	try := func(args ...string) {
		app := run.App("runtest", "testing ints")
		argVal := run.UintSlice("arg", "", 0)
		app.SetArgs(argVal.Rest("arg"))
		printMain(app, args...)
	}

	try()
	try("5", "2")
	try("6", "-2", "17") // -2 blocks the parse, then has no home

	// output:
	// [] err: expected "<arg> ..." errcmd=runtest args: arg=[]
	// [5 2] cmd: runtest args: arg=[5 2]
	// [6 -2 17] err: unexpected flag: -2 errcmd=runtest args: arg=[6]
}

type quotedstring string

func (qs quotedstring) String() string {
	return strconv.Quote(string(qs))
}

func ExampleStringLike() {
	app := run.App("runtest", "testing run")
	fileVal := run.StringLike[quotedstring]("file", "")
	app.SetArgs(fileVal.Pos("file"))
	printMain(app, "hello")

	// output:
	// [hello] cmd: runtest args: file="hello"
}

func ExampleIntLike() {
	type myint int8
	type myunt uint8
	try := func(args ...string) {
		app := run.App("runtest", "testing IntLike")
		i8 := run.IntLike[myint]("i8", "", 10)
		u8 := run.UintLike[myunt]("u8", "", 0)
		app.SetArgs(i8.Pos("i"), u8.Pos("u"))
		printMain(app, args...)
	}

	try("-100", "200") // fine
	try("100", "-100") // -100 invalid
	try("200", "100")  // 200 out of range

	// output:
	// [-100 200] cmd: runtest args: i8=-100 u8=200
	// [100 -100] err: u: parsing "-100" as run_test.myunt: invalid syntax errcmd=runtest args: i8=100 u8=0
	// [200 100] err: i: parsing "200" as run_test.myint: value out of range errcmd=runtest args: i8=0 u8=0
}

// Allow short flags to be grouped can alter error messages.
// Negative numbers might be sequences of short flags, so are only offered to relevant option types.
// (Normally only single-digit negative numbers look like a short flag; others are treated as positional.)
func ExampleApplication_AllowGroupShortFlags() {
	app := run.App("runtest", "testing abbrev")
	u := run.UintLike[uint8]("u8", "", 10)
	app.SetArgs(u.Pos("u"))
	app.AllowGroupShortFlags(false) // default
	printMain(app, "-1")
	printMain(app, "-100")
	app.AllowGroupShortFlags(true)
	printMain(app, "-1")
	printMain(app, "-100")

	// output:
	// [-1] err: unexpected flag: -1 errcmd=runtest args: u8=0
	// [-100] err: u: parsing "-100" as uint8: invalid syntax errcmd=runtest args: u8=0
	// [-1] err: unexpected flag: -1 errcmd=runtest args: u8=0
	// [-100] err: unexpected flag: -100 errcmd=runtest args: u8=0
}

func ExampleStringOf_enum() {
	app := run.App("runtest", "testing enums")
	letter := run.StringOf[quotedstring]("letter", "", "alpha", "bravo", "charlie")
	app.SetArgs(letter.Pos("abbrev"))
	printMain(app, "delta")
	printMain(app, "bravo")

	// output:
	// [delta] err: abbrev: "delta" not one of "alpha", "bravo", "charlie" errcmd=runtest args: letter=""
	// [bravo] cmd: runtest args: letter="bravo"
}

func ExampleFloatLike_enum() {
	app := run.App("runtest", "testing floats")
	pct := run.FloatLike[float64]("pct", "")
	app.SetArgs(pct.Pos("pct"))
	printMain(app, "+-12.34")
	printMain(app, "12.34")
	printMain(app, "12")
	printMain(app, ".34")

	// output:
	// [+-12.34] err: pct: parsing "+-12.34" as float64: invalid syntax errcmd=runtest args: pct=0
	// [12.34] cmd: runtest args: pct=12.34
	// [12] cmd: runtest args: pct=12
	// [.34] cmd: runtest args: pct=0.34
}

func ExampleFloatLikeSlice_enum() {
	app := run.App("runtest", "testing floats")
	pcts := run.FloatLikeSlice[float64]("pcts", "")
	app.SetArgs(pcts.Rest("pct"))
	printMain(app, "12.34", "+12", "-.34", "+-12.34")

	// output:
	// [12.34 +12 -.34 +-12.34] err: pct: parsing "+-12.34" as float64: invalid syntax errcmd=runtest args: pcts=[12.34 12 -0.34]
}

func ExampleNamedOf_enum() {
	app := run.App("runtest", "testing enums")
	digit := run.NamedOf("digit", "", []run.NamedValue[int]{
		{Name: "one", Value: 1},
		{Name: "two", Value: 2},
		{Name: "three", Value: 3},
	})
	app.SetArgs(digit.Pos("d"))
	printMain(app, "four")
	printMain(app, "two")
	printMain(app, "four")

	// output:
	// [four] err: d: "four" not one of "one", "two", "three" errcmd=runtest args: digit=0
	// [two] cmd: runtest args: digit=2
	// [four] err: d: "four" not one of "one", "two", "three" errcmd=runtest args: digit=2
}

func ExampleNamedSliceOf_enum() {
	app := run.App("runtest", "testing enums")
	digit := run.NamedSliceOf("digits", "", []run.NamedValue[int]{
		{Name: "one", Value: 1},
		{Name: "two", Value: 2},
		{Name: "three", Value: 3},
	})
	app.SetArgs(digit.Rest("digit"))
	printMain(app, "two", "three")
	printMain(app, "two", "four")

	// output:
	// [two three] cmd: runtest args: digits=[2 3]
	// [two four] err: digit: "four" not one of "one", "two", "three" errcmd=runtest args: digits=[2]
}

func ExampleApplication_Flags() {
	try := func(args ...string) {
		app := run.App("runtest", "testing flags")
		a := run.String("a", "")
		b := run.String("b", "")
		app.SetFlags(a.Flag(), b.Flag())
		printMain(app, args...)
	}
	try("--b", "beta")
	try("--b=gamma", "--a")

	// output:
	// [--b beta] cmd: runtest flags: a= b=beta
	// [--b=gamma --a] err: --a: expected <value> errcmd=runtest flags: a= b=gamma
}

func ExampleFlag_Default() {
	try := func(args ...string) {
		app := run.App("runtest", "testing flag defaults")
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
	// [] cmd: runtest flags: digit=2
	// [--digit three] cmd: runtest flags: digit=3
	// [--digit four] err: --digit: "four" not one of "one", "two", "three" errcmd=runtest flags: digit=0
}

func ExampleCmd() {
	app := run.App("runtest", "testing commands")
	fooCmd := run.Cmd("foo", "does foo")
	fooCmd.SetHandler(func(ctx context.Context, env run.Environ) error { _, err := fmt.Println("running foo"); return err })
	barCmd := run.Cmd("bar", "does bar")
	barCmd.SetHandler(func(ctx context.Context, env run.Environ) error { _, err := fmt.Println("running bar"); return err })
	app.SetCommands(fooCmd, barCmd)

	runMain(app)
	runMain(app, "foo")
	runMain(app, "bar")

	// output:
	// runtest: error: expected <command>
	// running foo
	// running bar
}

func ExampleEnabler() {
	try := func(args ...string) {
		app := run.App("runtest", "testing enable")
		en := run.Enabler("en", "", false, true)
		app.SetFlags(en.Flag())
		printMain(app, args...)
	}

	try()
	try("--en")
	try("--en", "--en")
	try("--en=true")

	// output:
	// [] cmd: runtest flags: en=false
	// [--en] cmd: runtest flags: en=true
	// [--en --en] err: --en: repeated errcmd=runtest flags: en=true
	// [--en=true] err: unexpected flag value: --en=true errcmd=runtest flags: en=false
}

func ExampleToggler() {
	try := func(args ...string) {
		app := run.App("runtest", "testing enable")
		en := run.Toggler("en", "", false, true)
		app.SetFlags(en.Flag())
		printMain(app, args...)
	}

	try()
	try("--en")
	try("--en", "--en")
	try("--en", "--en", "--en")
	try("--en=false")

	// output:
	// [] cmd: runtest flags: en=false
	// [--en] cmd: runtest flags: en=true
	// [--en --en] cmd: runtest flags: en=false
	// [--en --en --en] cmd: runtest flags: en=true
	// [--en=false] err: unexpected flag value: --en=false errcmd=runtest flags: en=false
}

func ExampleAccumulator() {
	try := func(args ...string) {
		app := run.App("runtest", "testing enable")
		laugh := run.Accumulator[quotedstring]("ha", "", "", "ha")
		frown := run.Accumulator("no", "", 0, -2)
		app.SetFlags(laugh.Flag(), frown.Flag())
		printMain(app, args...)
	}

	try()
	try("--ha")
	try("--no")
	try("--ha", "--no", "--ha")
	try("--ha", "--no", "--ha", "--ha", "--no", "--no", "--no")
	try("--ha=boop")

	// output:
	// [] cmd: runtest flags: ha="" no=0
	// [--ha] cmd: runtest flags: ha="ha" no=0
	// [--no] cmd: runtest flags: ha="" no=-2
	// [--ha --no --ha] cmd: runtest flags: ha="haha" no=-2
	// [--ha --no --ha --ha --no --no --no] cmd: runtest flags: ha="hahaha" no=-8
	// [--ha=boop] err: unexpected flag value: --ha=boop errcmd=runtest flags: ha="" no=0
}
