package run_test

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/mutility/cli/run"
)

func parseMain(app *run.Application, args ...string) {
	_, err := app.Parse(run.DefaultEnviron().WithArgs(append(os.Args[:1:1], args...)))
	if err != nil {
		app.Ferror(os.Stdout, err)
	}
}

func runMain(app *run.Application, args ...string) {
	err := app.Main(context.TODO(), run.DefaultEnviron().WithArgs(append(os.Args[:1:1], args...)))
	if err != nil {
		app.Ferror(os.Stdout, err)
	}
}

func Example_whyrun() {
	app := run.App("runtest", "testing run")
	parseMain(app)

	// output:
}

func Example_empty_hello() {
	app := run.App("runtest", "testing run")
	app.Flags()
	app.Args()
	app.Commands()

	parseMain(app, "hello")

	// output:
	// runtest: error: unexpected argument: "hello"
}

func Example_nil() {
	app := run.App("runtest", "testing run")
	parseMain(app, "hello")

	// output:
	// runtest: error: unexpected argument: "hello"
}

func ExampleString() {
	try := func(args ...string) {
		app := run.App("runtest", "testing run")
		argVal := run.String("arg", "")
		app.Args(argVal.Pos("arg"))
		parseMain(app, args...)
		fmt.Println("argVal:", strconv.Quote(argVal.Value()))
	}

	try("hello")
	try("-")
	try("-n")
	try("--bad")
	try("--", "--okay")

	// output:
	// argVal: "hello"
	// argVal: "-"
	// runtest: error: unexpected flag: -n
	// argVal: ""
	// runtest: error: unexpected flag: --bad
	// argVal: ""
	// argVal: "--okay"
}

func ExampleFileLike() {
	try := func(args ...string) {
		app := run.App("runtest", "testing run")
		argVal := run.FileLike[quotedstring]("arg", "")
		app.Args(argVal.Pos("arg"))
		parseMain(app, args...)
		fmt.Println("argVal:", argVal.Value())
	}

	try("hello")
	try("-")
	try("-n")
	try("--bad")
	try("--", "--okay")

	// output:
	// argVal: "hello"
	// argVal: "-"
	// runtest: error: unexpected flag: -n
	// argVal: ""
	// runtest: error: unexpected flag: --bad
	// argVal: ""
	// argVal: "--okay"
}

func ExampleFileLikeSlice_many() {
	try := func(args ...string) {
		app := run.App("runtest", "testing run")
		argVal := run.FileLikeSlice[quotedstring]("arg", "")
		app.Args(argVal.Rest("arg"))
		parseMain(app, args...)
		fmt.Println("argVal:", argVal.Value())
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
	// argVal: ["hello" "world"]
	// argVal: ["head" "world"]
	// argVal: ["mid" "world"]
	// argVal: ["trail" "world"]
	// argVal: ["dash" "--" "world"]
	// runtest: error: unexpected flag: -n
	// argVal: ["unquoted"]
	// argVal: ["quoted" "-n"]
	// runtest: error: unexpected flag: -n
	// argVal: ["-"]
	// argVal: ["-" "-n"]
}

func ExampleIntSlice_many() {
	try := func(args ...string) {
		app := run.App("runtest", "testing ints")
		argVal := run.IntSlice("arg", "", 0)
		app.Args(argVal.Rest("arg"))
		parseMain(app, args...)
		fmt.Println("argVal:", argVal.Value())
	}

	try()
	try("5")
	try("6", "-2", "17")

	// output:
	// runtest: error: expected "<arg> ..."
	// argVal: []
	// argVal: [5]
	// argVal: [6 -2 17]
}

func ExampleUintSlice_many() {
	try := func(args ...string) {
		app := run.App("runtest", "testing ints")
		argVal := run.UintSlice("arg", "", 0)
		app.Args(argVal.Rest("arg"))
		parseMain(app, args...)
		fmt.Println("argVal:", argVal.Value())
	}

	try()
	try("5")
	try("6", "-2", "17") // 6 gets saved before -2 fails the parse.

	// output:
	// runtest: error: expected "<arg> ..."
	// argVal: []
	// argVal: [5]
	// runtest: error: unexpected flag: -2
	// argVal: [6]
}

type quotedstring string

func (qs quotedstring) String() string {
	return strconv.Quote(string(qs))
}

func ExampleStringLike() {
	app := run.App("runtest", "testing run")
	fileVal := run.StringLike[quotedstring]("file", "")
	app.Args(fileVal.Pos("file"))
	parseMain(app, "hello")
	fmt.Println("fileVal:", fileVal.Value())

	// output:
	// fileVal: "hello"
}

func ExampleIntLike() {
	try := func(args ...string) {
		app := run.App("runtest", "testing IntLike")
		i8 := run.IntLike[int8]("smallint", "", 10)
		u8 := run.UintLike[uint8]("smalluint", "", 0)
		app.Args(i8.Pos("i"), u8.Pos("u"))
		parseMain(app, args...)
		fmt.Println("i8:", i8.Value(), "u8:", u8.Value())
	}

	try("-100", "200") // fine
	try("100", "-100") // -100 invalid
	try("200", "100")  // 200 out of range

	// output:
	// i8: -100 u8: 200
	// runtest: error: u: parsing "-100" as uint8: invalid syntax
	// i8: 100 u8: 0
	// runtest: error: i: parsing "200" as int8: value out of range
	// i8: 0 u8: 0
}

// Allow short flags to be grouped can alter error messages.
// Negative numbers might be sequences of short flags, so are only offered to relevant option types.
// (Normally only single-digit negative numbers look like a short flag; others are treated as positional.)
func ExampleApplication_AllowGroupShortFlags() {
	app := run.App("runtest", "testing abbrev")
	u := run.UintLike[uint8]("smallint", "", 10)
	app.Args(u.Pos("u"))
	app.AllowGroupShortFlags(false) // default
	parseMain(app, "-1")
	parseMain(app, "-100")
	app.AllowGroupShortFlags(true)
	parseMain(app, "-1")
	parseMain(app, "-100")

	// output:
	// runtest: error: unexpected flag: -1
	// runtest: error: u: parsing "-100" as uint8: invalid syntax
	// runtest: error: unexpected flag: -1
	// runtest: error: unexpected flag: -100
}

func ExampleStringOf_enum() {
	app := run.App("runtest", "testing enums")
	letter := run.StringOf("letter", "", "alpha", "bravo", "charlie")
	app.Args(letter.Pos("abbrev"))
	parseMain(app, "bravo")
	fmt.Println("letter:", letter.Value())
	parseMain(app, "delta")

	// output:
	// letter: bravo
	// runtest: error: abbrev: "delta" not one of "alpha", "bravo", "charlie"
}

func ExampleNamedOf_enum() {
	app := run.App("runtest", "testing enums")
	digit := run.NamedOf("digit", "", []run.NamedValue[int]{
		{Name: "one", Value: 1},
		{Name: "two", Value: 2},
		{Name: "three", Value: 3},
	})
	app.Args(digit.Pos("name"))
	parseMain(app, "two")
	fmt.Println("digit:", digit.Value())
	parseMain(app, "four")

	// output:
	// digit: 2
	// runtest: error: name: "four" not one of "one", "two", "three"
}

func ExampleNamedSliceOf_enum() {
	app := run.App("runtest", "testing enums")
	digit := run.NamedSliceOf("digit", "", []run.NamedValue[int]{
		{Name: "one", Value: 1},
		{Name: "two", Value: 2},
		{Name: "three", Value: 3},
	})
	app.Args(digit.Rest("name"))
	parseMain(app, "two", "three")
	fmt.Println("digits:", digit.Value())
	parseMain(app, "two", "four")

	// output:
	// digits: [2 3]
	// runtest: error: name: "four" not one of "one", "two", "three"
}

func ExampleApplication_Flags() {
	try := func(args ...string) {
		app := run.App("runtest", "testing flags")
		a := run.String("a", "")
		b := run.String("b", "")
		app.Flags(a.Flag(), b.Flag())
		parseMain(app, args...)
		fmt.Println("a:", a.Value(), "b:", b.Value())
	}
	try("--b", "beta")
	try("--b=gamma", "--a")

	// output:
	// a:  b: beta
	// runtest: error: --a: no argument provided
	// a:  b: gamma
}

func ExampleFlag_Default() {
	try := func(args ...string) {
		app := run.App("runtest", "testing flag defaults")
		digit := run.NamedOf("digit", "", []run.NamedValue[int]{
			{Name: "one", Value: 1},
			{Name: "two", Value: 2},
			{Name: "three", Value: 3},
		})
		app.Flags(digit.Flag().Default("two"))
		parseMain(app, args...)
		fmt.Println("digit:", digit.Value())
	}
	try()
	try("--digit", "three")
	try("--digit", "four")

	// output:
	// digit: 2
	// digit: 3
	// runtest: error: --digit: "four" not one of "one", "two", "three"
	// digit: 0
}

func ExampleCmd() {
	app := run.App("runtest", "testing commands")
	fooCmd := run.Cmd("foo", "does foo")
	fooCmd.Runs(func(ctx context.Context, env run.Environ) error { _, err := fmt.Println("running foo"); return err })
	barCmd := run.Cmd("bar", "does bar")
	barCmd.Runs(func(ctx context.Context, env run.Environ) error { _, err := fmt.Println("running bar"); return err })
	app.Commands(fooCmd, barCmd)

	runMain(app)
	runMain(app, "foo")
	runMain(app, "bar")

	// output:
	// runtest: error: expected <command>
	// running foo
	// running bar
}