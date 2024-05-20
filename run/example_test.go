//nolint:errcheck
package run_test

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/mutility/cli/run"
)

func runMain(app *run.Application, args ...string) {
	err := app.Main(context.TODO(), run.DefaultEnviron().WithArgs(append(os.Args[:1:1], args...)))
	if err != nil {
		app.Ferror(os.Stdout, err)
	}
}

func ExampleApp_empty() {
	app, _ := run.App("noarg", "")
	app.Debug()

	// output:
	// []
	//   cmd: noarg
}

func ExampleMustApp_empty() {
	app := run.MustApp("nah", "")
	app.SetCommands()
	app.Debug("hello")

	// output:
	// [hello] err: unexpected argument: "hello"
	//   cmd: nah
}

func Example_nil() {
	app := run.MustApp("noway", "")
	app.Debug("hello")

	// output:
	// [hello] err: unexpected argument: "hello"
	//   cmd: noway
}

func ExampleString() {
	try := func(args ...string) {
		app := run.MustApp("str", "")
		argVal := run.String("arg", "")
		app.SetArgs(argVal.Arg("arg"))
		app.Debug(args...)
	}

	try("hello")
	try("-")
	try("-n")
	try("--no")
	try("--", "--okay")

	// output:
	// [hello]
	//   cmd: str
	//   arg: arg=hello
	// [-]
	//   cmd: str
	//   arg: arg=-
	// [-n] err: unexpected flag: -n
	//   cmd: str
	//   arg: arg=
	// [--no] err: unexpected flag: --no
	//   cmd: str
	//   arg: arg=
	// [-- --okay]
	//   cmd: str
	//   arg: arg=--okay
}

func ExampleFileLike() {
	try := func(args ...string) {
		app, _ := run.App("file", "", run.FileLike[quotedstring]("arg", "").Arg("arg"))
		app.Debug(args...)
	}

	try("hello")
	try("-")
	try("-n")
	try("--no")
	try("--", "--okay")

	// output:
	// [hello]
	//   cmd: file
	//   arg: arg="hello"
	// [-]
	//   cmd: file
	//   arg: arg="-"
	// [-n] err: unexpected flag: -n
	//   cmd: file
	//   arg: arg=""
	// [--no] err: unexpected flag: --no
	//   cmd: file
	//   arg: arg=""
	// [-- --okay]
	//   cmd: file
	//   arg: arg="--okay"
}

func ExampleFileLikeSlice_many() {
	try := func(args ...string) {
		app := run.MustApp("files", "", run.FileLikeSlice[quotedstring]("args", "").Args("arg"))
		app.Debug(args...)
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
	// [hello world]
	//   cmd: files
	//   arg: args=["hello" "world"]
	// [-- head world]
	//   cmd: files
	//   arg: args=["head" "world"]
	// [mid -- world]
	//   cmd: files
	//   arg: args=["mid" "world"]
	// [trail world --]
	//   cmd: files
	//   arg: args=["trail" "world"]
	// [dash -- -- world]
	//   cmd: files
	//   arg: args=["dash" "--" "world"]
	// [unquoted -n] err: unexpected flag: -n
	//   cmd: files
	//   arg: args=["unquoted"]
	// [quoted -- -n]
	//   cmd: files
	//   arg: args=["quoted" "-n"]
	// [- -n] err: unexpected flag: -n
	//   cmd: files
	//   arg: args=["-"]
	// [- -- -n]
	//   cmd: files
	//   arg: args=["-" "-n"]
}

func ExampleIntSlice_many() {
	try := func(args ...string) {
		app := run.MustApp("ints", "", run.IntSlice("arg", "", 0).Args("arg"))
		app.Debug(args...)
	}

	try()
	try("5")
	try("6", "-2", "17", "+-3")

	// output:
	// [] err: expected "<arg> ..."
	//   cmd: ints
	//   arg: arg=[]
	// [5]
	//   cmd: ints
	//   arg: arg=[5]
	// [6 -2 17 +-3] err: arg: parsing "+-3" as int: invalid syntax
	//   cmd: ints
	//   arg: arg=[6 -2 17]
}

func ExampleUintSlice_many() {
	try := func(args ...string) {
		app := run.MustApp("uints", "", run.UintSlice("arg", "", 0).Args("arg"))
		app.Debug(args...)
	}

	try()
	try("5", "2")
	try("6", "-2", "17") // -2 blocks the parse, then has no home

	// output:
	// [] err: expected "<arg> ..."
	//   cmd: uints
	//   arg: arg=[]
	// [5 2]
	//   cmd: uints
	//   arg: arg=[5 2]
	// [6 -2 17] err: unexpected flag: -2
	//   cmd: uints
	//   arg: arg=[6]
}

type quotedstring string

func (qs quotedstring) String() string {
	return strconv.Quote(string(qs))
}

func ExampleStringLike() {
	app := run.MustApp("quoted", "", run.StringLike[quotedstring]("files", "").Arg("file"))
	app.Debug("hello")

	// output:
	// [hello]
	//   cmd: quoted
	//   arg: files="hello"
}

func ExampleIntLike() {
	type myint int8
	type myunt uint8
	try := func(args ...string) {
		app := run.MustApp("intish", "", run.IntLike[myint]("i8", "", 10).Arg("i"), run.UintLike[myunt]("u8", "", 0).Arg("u"))
		app.Debug(args...)
	}

	try("-100", "200") // fine
	try("100", "-100") // -100 invalid
	try("200", "100")  // 200 out of range

	// output:
	// [-100 200]
	//   cmd: intish
	//   arg: i8=-100
	//   arg: u8=200
	// [100 -100] err: u: parsing "-100" as run_test.myunt: invalid syntax
	//   cmd: intish
	//   arg: i8=100
	//   arg: u8=0
	// [200 100] err: i: parsing "200" as run_test.myint: value out of range
	//   cmd: intish
	//   arg: i8=0
	//   arg: u8=0
}

// Allow short flags to be grouped can alter error messages.
// Negative numbers might be sequences of short flags, so are only offered to relevant option types.
// (Normally only single-digit negative numbers look like a short flag; others are treated as positional.)
func ExampleApplication_AllowGroupShortFlags() {
	app := run.MustApp("shorts", "")
	u := run.UintLike[uint8]("u8", "", 10)
	app.SetArgs(u.Arg("u"))
	app.AllowGroupShortFlags(false) // default
	app.Debug("-1")
	app.Debug("-100")
	app.AllowGroupShortFlags(true)
	app.Debug("-1")
	app.Debug("-100")

	// output:
	// [-1] err: unexpected flag: -1
	//   cmd: shorts
	//   arg: u8=0
	// [-100] err: u: parsing "-100" as uint8: invalid syntax
	//   cmd: shorts
	//   arg: u8=0
	// [-1] err: unexpected flag: -1
	//   cmd: shorts
	//   arg: u8=0
	// [-100] err: unexpected flag: -100
	//   cmd: shorts
	//   arg: u8=0
}

func ExampleStringOf_enum() {
	app := run.MustApp("enum", "", run.StringOf[quotedstring]("letter", "", "alpha", "bravo", "charlie").Arg("abbrev"))
	app.Debug("delta")
	app.Debug("bravo")

	// output:
	// [delta] err: abbrev: "delta" not one of "alpha", "bravo", "charlie"
	//   cmd: enum
	//   arg: letter=""
	// [bravo]
	//   cmd: enum
	//   arg: letter="bravo"
}

func ExampleFloatLike() {
	app := run.MustApp("floaty", "", run.FloatLike[float64]("pct", "").Arg("pct"))
	app.Debug("+-12.34")
	app.Debug("12.34")
	app.Debug("12")
	app.Debug(".34")

	// output:
	// [+-12.34] err: pct: parsing "+-12.34" as float64: invalid syntax
	//   cmd: floaty
	//   arg: pct=0
	// [12.34]
	//   cmd: floaty
	//   arg: pct=12.34
	// [12]
	//   cmd: floaty
	//   arg: pct=12
	// [.34]
	//   cmd: floaty
	//   arg: pct=0.34
}

func ExampleFloatLikeSlice() {
	app := run.MustApp("floats", "", run.FloatLikeSlice[float64]("pcts", "").Args("pct"))
	app.Debug("12.34", "+12", "-.34", "+-12.34")

	// output:
	// [12.34 +12 -.34 +-12.34] err: pct: parsing "+-12.34" as float64: invalid syntax
	//   cmd: floats
	//   arg: pcts=[12.34 12 -0.34]
}

func ExampleNamedOf() {
	digit := run.NamedOf("digit", "", []run.NamedValue[int]{
		{Name: "one", Value: 1},
		{Name: "two", Value: 2},
		{Name: "three", Value: 3},
	})
	app := run.MustApp("named", "", digit.Arg("d"))
	app.Debug("four")
	app.Debug("two")
	app.Debug("four")

	// output:
	// [four] err: d: "four" not one of "one", "two", "three"
	//   cmd: named
	//   arg: digit=0
	// [two]
	//   cmd: named
	//   arg: digit=2
	// [four] err: d: "four" not one of "one", "two", "three"
	//   cmd: named
	//   arg: digit=2
}

func ExampleNamedSliceOf() {
	digit := run.NamedSliceOf("digits", "", []run.NamedValue[int]{
		{Name: "one", Value: 1},
		{Name: "two", Value: 2},
		{Name: "three", Value: 3},
	})
	app := run.MustApp("nameds", "", digit.Args("digit"))
	app.Debug("two", "three")
	app.Debug("two", "four")

	// output:
	// [two three]
	//   cmd: nameds
	//   arg: digits=[2 3]
	// [two four] err: digit: "four" not one of "one", "two", "three"
	//   cmd: nameds
	//   arg: digits=[2]
}

func ExampleApplication_SetFlags() {
	try := func(args ...string) {
		app := run.MustApp("flags", "")
		app.SetFlags(run.String("a", "").Flag(), run.String("b", "").Flag())
		app.Debug(args...)
	}
	try("--b", "beta")
	try("--b=gamma", "--a")

	// output:
	// [--b beta]
	//   cmd: flags
	//   flag: a=
	//   flag: b=beta
	// [--b=gamma --a] err: --a: expected <value>
	//   cmd: flags
	//   flag: a=
	//   flag: b=gamma
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
		app.Debug(args...)
	}
	try()
	try("--digit", "three")
	try("--digit", "four")

	// output:
	// []
	//   cmd: default
	//   flag: digit=2
	// [--digit three]
	//   cmd: default
	//   flag: digit=3
	// [--digit four] err: --digit: "four" not one of "one", "two", "three"
	//   cmd: default
	//   flag: digit=0
}

func ExampleMustCmd() {
	app := run.MustApp("commands", "",
		run.MustCmd("foo", "does foo", run.Handler(func(context.Context, run.Environ) error { _, err := fmt.Println("running foo"); return err })),
		run.MustCmd("bar", "does bar", run.Handler(func(context.Context, run.Environ) error { _, err := fmt.Println("running bar"); return err })),
	)

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
		app := run.MustApp("enable", "", run.Enabler("en", "", false, true).Flag())
		app.Debug(args...)
	}

	try()
	try("--en")
	try("--en", "--en")
	try("--en=true")

	// output:
	// []
	//   cmd: enable
	//   flag: en=false
	// [--en]
	//   cmd: enable
	//   flag: en=true
	// [--en --en] err: --en: repeated
	//   cmd: enable
	//   flag: en=true
	// [--en=true] err: unexpected flag value: --en=true
	//   cmd: enable
	//   flag: en=false
}

func ExampleToggler() {
	try := func(args ...string) {
		app := run.MustApp("toggle", "", run.Toggler("en", "", false, true).Flag())
		app.Debug(args...)
	}

	try()
	try("--en")
	try("--en", "--en")
	try("--en", "--en", "--en")
	try("--en=false")

	// output:
	// []
	//   cmd: toggle
	//   flag: en=false
	// [--en]
	//   cmd: toggle
	//   flag: en=true
	// [--en --en]
	//   cmd: toggle
	//   flag: en=false
	// [--en --en --en]
	//   cmd: toggle
	//   flag: en=true
	// [--en=false] err: unexpected flag value: --en=false
	//   cmd: toggle
	//   flag: en=false
}

func ExampleAccumulator() {
	try := func(args ...string) {
		run.MustApp("accum", "",
			run.Accumulator[quotedstring]("ha", "", "", "ha").Flag(),
			run.Accumulator("no", "", 0, -2).Flag(),
		).Debug(args...)
	}

	try()
	try("--ha")
	try("--no")
	try("--ha", "--no", "--ha")
	try("--ha", "--no", "--ha", "--ha", "--no", "--no", "--no")
	try("--ha=boop")

	// output:
	// []
	//   cmd: accum
	//   flag: ha=""
	//   flag: no=0
	// [--ha]
	//   cmd: accum
	//   flag: ha="ha"
	//   flag: no=0
	// [--no]
	//   cmd: accum
	//   flag: ha=""
	//   flag: no=-2
	// [--ha --no --ha]
	//   cmd: accum
	//   flag: ha="haha"
	//   flag: no=-2
	// [--ha --no --ha --ha --no --no --no]
	//   cmd: accum
	//   flag: ha="hahaha"
	//   flag: no=-8
	// [--ha=boop] err: unexpected flag value: --ha=boop
	//   cmd: accum
	//   flag: ha=""
	//   flag: no=0
}

func ExampleParser() {
	try := func(args ...string) {
		run.MustApp("parser", "",
			run.Parser("url", "", url.ParseRequestURI).Arg("url"),
		).Debug(args...)
	}

	try("rel")
	try("schema:relative")
	try("schema:/rooted")
	try("https://example.com/")

	// output:
	// [rel] err: url: parse "rel": invalid URI for request
	//   cmd: parser
	//   arg: url=<nil>
	// [schema:relative]
	//   cmd: parser
	//   arg: url=schema:relative
	// [schema:/rooted]
	//   cmd: parser
	//   arg: url=schema:/rooted
	// [https://example.com/]
	//   cmd: parser
	//   arg: url=https://example.com/
}

func ExampleParserSlice() {
	try := func(args ...string) {
		run.MustApp("parser", "",
			run.ParserSlice("urls", "", url.ParseRequestURI).Args("url"),
		).Debug(args...)
	}

	try("rel")
	try("schema:relative", "schema:/rooted", "https://example.com/")

	// output:
	// [rel] err: url: parse "rel": invalid URI for request
	//   cmd: parser
	//   arg: urls=[]
	// [schema:relative schema:/rooted https://example.com/]
	//   cmd: parser
	//   arg: urls=[schema:relative schema:/rooted https://example.com/]
}
