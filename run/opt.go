package run

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"unsafe"
)

// Options implement the required internal interface for use as either Flags, Args, or both.
type Option interface {
	description() string
	seeAlso() []*Command
	setSeeAlso(cmds ...*Command)
	parseDefault(string) error
	okValues() []string
	okPrefix() string
}

type optionFlagArg interface {
	parseArg(flagArgSource) error
}

type optionSlice interface {
	parseMany(argSource) error
}

type flagArgSource interface {
	Peek() (string, bool)
	Next() (string, bool)
}

type argSource interface {
	PeekMany() (string, bool)
	Next() (string, bool)
}

type option[T any] struct {
	name     string
	desc     string
	value    *T
	parse    func(string) error
	prefixOK string   // set to - to allow -[^-]+, or -- to allow --.+ in arg context
	strOK    []string // include unusual values such as - to allow them in arg context
	see      []*Command
}

func (o *option[T]) description() string               { return o.desc }
func (o *option[T]) seeAlso() []*Command               { return o.see }
func (o *option[T]) setSeeAlso(cmds ...*Command)       { o.see = cmds }
func (o *option[T]) parseDefault(arg string) error     { return o.parse(arg) }
func (o *option[T]) parseArg(args flagArgSource) error { return argParser(o.parse)(args) }
func (o *option[T]) okValues() []string                { return o.strOK }
func (o *option[T]) okPrefix() string                  { return o.prefixOK }

func (o *option[T]) Value() T { return *o.value }

// Flags returns a flag definition for this option with custom aliases.
// Zero values will omit either short or long. Do not omit both.
func (o *option[T]) Flags(short rune, long string, placeholder string) Flag {
	return Flag{option: o, rune: short, string: long, hint: placeholder}
}

// Flag returns a flag definition for this option using its name as the long.
// Thus an option named "opt" will have a flag name "--opt".
func (o *option[T]) Flag() Flag {
	return Flag{option: o, string: o.name}
}

// Pos returns an Arg definition for this option with a custom alias.
func (o *option[T]) Pos(name string) Arg {
	return Arg{option: o, name: name}
}

// Slice returns a Param that converts a T to a []T.
// This can be used to share a handler between single and slice options.
func (o *option[T]) Slice() Param[[]T] {
	return sliceOf[T]{o.value}
}

// String creates an option that stores any string.
func String(name, desc string) option[string] {
	return StringLike[string](name, desc)
}

// String creates an option that stores any string-like value.
func StringLike[T ~string](name, desc string) option[T] {
	var v T
	parse := func(s string) error { v = T(s); return nil }
	return option[T]{
		name:  name,
		desc:  desc,
		value: &v,
		parse: parse,
	}
}

type NamedValue[T any] struct {
	Name  string
	Desc  string
	Value T
}

// StringOf creates an option that stores a string-like value from the provided list.
// This is suitable for small to medium sets of string-like names.
func StringOf[T ~string](name, desc string, names ...T) option[T] {
	nvs := make([]NamedValue[T], len(names))
	for i, nam := range names {
		nvs[i] = NamedValue[T]{Name: string(nam), Value: nam}
	}
	return NamedOf(name, desc, nvs)
}

// NamedOf creates an option that stores any type of value, looked up from the provided mapping.
// This is suitable for small to medium sets of names.
func NamedOf[T any](name, desc string, mapping []NamedValue[T]) option[T] {
	var v T
	mapping = slices.Clone(mapping)
	parse := func(s string) error {
		pos := slices.IndexFunc(mapping, func(v NamedValue[T]) bool {
			return v.Name == s
		})
		if pos < 0 {
			return NotOneOfError[T]{s, mapping}
		}
		v = T(mapping[pos].Value)
		return nil
	}
	return option[T]{
		name:  name,
		desc:  desc,
		value: &v,
		parse: parse,
	}
}

// File creates an option that stores a string filename.
// This differs from String by accepting "-" as a positional argument.
func File(name, desc string) option[string] {
	return FileLike[string](name, desc)
}

// FileLike creates an option that stores a string-like filename.
// This differs from StringLike by accepting "-" as a positional argument.
func FileLike[T ~string](name, desc string) option[T] {
	o := StringLike[T](name, desc)
	o.strOK = dashOK
	return o
}

var dashOK = []string{"-"}

// Int creates an option that stores any int.
// It converts strings like [strconv.ParseInt].
func Int(name, desc string, base int) option[int] {
	return IntLike[int](name, desc, base)
}

// IntLike creates an option that stores any int-like value.
// It converts strings like [strconv.ParseInt].
func IntLike[T ~int | ~int8 | ~int16 | ~int32 | ~int64](name, desc string, base int) option[T] {
	var v T
	parse := func(s string) error {
		i, err := strconv.ParseInt(s, base, int(unsafe.Sizeof(v))*8)
		if e, ok := err.(*strconv.NumError); ok || errors.As(err, &e) {
			return fmt.Errorf("parsing %q as %T: %v", e.Num, v, e.Err)
		}
		if err != nil {
			return err
		}
		v = T(i)
		return nil
	}
	return option[T]{
		name:     name,
		desc:     desc,
		value:    &v,
		parse:    parse,
		prefixOK: "-",
	}
}

// Uint creates an option that stores any uint.
// It converts strings like [strconv.ParseUint].
func Uint(name, desc string, base int) option[uint] {
	return UintLike[uint](name, desc, base)
}

// UintLike creates an option that stores any uint-like value.
// It converts strings like [strconv.ParseUint].
func UintLike[T ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64](name, desc string, base int) option[T] {
	var v T
	parse := func(s string) error {
		i, err := strconv.ParseUint(s, base, int(unsafe.Sizeof(v))*8)
		if e, ok := err.(*strconv.NumError); ok || errors.As(err, &e) {
			return fmt.Errorf("parsing %q as %T: %v", e.Num, v, e.Err)
		}
		if err != nil {
			return err
		}
		v = T(i)
		return nil
	}
	return option[T]{
		name:  name,
		desc:  desc,
		value: &v,
		parse: parse,
	}
}

// FloatLike creates an option that stores any float-like value.
// It converts strings like [strconv.ParseFloat].
func FloatLike[T ~float32 | ~float64](name, desc string) option[T] {
	var v T
	parse := func(s string) error {
		f, err := strconv.ParseFloat(s, int(unsafe.Sizeof(v))*8)
		if e, ok := err.(*strconv.NumError); ok || errors.As(err, &e) {
			return fmt.Errorf("parsing %q as %T: %v", e.Num, v, e.Err)
		}
		if err != nil {
			return err
		}
		v = T(f)
		return nil
	}
	return option[T]{
		name:  name,
		desc:  desc,
		value: &v,
		parse: parse,
	}
}

var errMissingArg = errors.New("no argument provided")

func argParser(parse func(string) error) func(flagArgSource) error {
	return func(s flagArgSource) error {
		v, ok := s.Peek()
		if !ok {
			return errMissingArg
		}
		err := parse(v)
		if err == nil {
			s.Next()
		}
		return err
	}
}

type sliceOf[T any] struct{ value *T }

func (s sliceOf[T]) Value() []T { return []T{*s.value} }
