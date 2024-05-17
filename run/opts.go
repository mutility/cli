package run

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"unsafe"
)

type options[T any] struct {
	name     string
	desc     string
	value    *[]T
	parse    func(string) error
	prefixOK string   // set to - to allow -[^-]+, or -- to allow --.+ in arg context
	strOK    []string // include unusual values such as - to allow them in arg context
	see      []*Command
}

func (o *options[T]) description() string            { return o.desc }
func (o *options[T]) seeAlso() []*Command            { return o.see }
func (o *options[T]) setSeeAlso(cmds ...*Command)    { o.see = cmds }
func (o *options[T]) parseDefault(arg string) error  { return o.parse(arg) }
func (o *options[T]) parseMany(args argSource) error { return manyParser(o.parse)(args) }
func (o *options[T]) okValues() []string             { return o.strOK }
func (o *options[T]) okPrefix() string               { return o.prefixOK }

func (o *options[T]) Value() []T { return *o.value }

// Rest returns an multi-Arg definition for this option with a custom alias.
func (o *options[T]) Rest(name string) Arg {
	return Arg{option: o, name: name, many: true}
}

// StringSlice creates and option that stores a slice of string values.
// This differs from String by supporting Rest().
func StringSlice(name, desc string) options[string] {
	return StringLikeSlice[string](name, desc)
}

// StringLikeSlice creates and option that stores a slice of string-like values.
// This differs from StringLike by supporting Rest().
func StringLikeSlice[T ~string](name, desc string) options[T] {
	var v []T
	parse := func(s string) error {
		v = append(v, T(s))
		return nil
	}
	return options[T]{
		name:  name,
		desc:  desc,
		value: &v,
		parse: parse,
	}
}

// StringSliceOf creates an option that stores a string-like slice of value from the provided list.
// This is suitable for small to medium sets of string-like names.
func StringSliceOf[T ~string](name, desc string, names ...T) options[T] {
	nvs := make([]NamedValue[T], len(names))
	for i, nam := range names {
		nvs[i] = NamedValue[T]{Name: string(nam), Value: nam}
	}
	return NamedSliceOf(name, desc, nvs)
}

// NamedSliceOf creates an option that stores a slice of any type of value, looked up from the provided mapping.
// This is suitable for small to medium sets of names.
func NamedSliceOf[T any](name, desc string, mapping []NamedValue[T]) options[T] {
	var v []T
	mapping = slices.Clone(mapping)
	parse := func(s string) error {
		pos := slices.IndexFunc(mapping, func(v NamedValue[T]) bool {
			return v.Name == s
		})
		if pos < 0 {
			return NotOneOfError[T]{s, mapping}
		}
		v = append(v, T(mapping[pos].Value))
		return nil
	}
	return options[T]{
		name:  name,
		desc:  desc,
		value: &v,
		parse: parse,
	}
}

// FileSlice creates an option that stores a string slice of filenames.
// This differs from StringSlice by accepting "-" as a positional argument,
// and from File by supporting Rest().
func FileSlice(name, desc string) options[string] {
	return FileLikeSlice[string](name, desc)
}

// FileLikeSlice creates an option that stores a string-like slice of filenames.
// This differs from StringLikeSlice by accepting "-" as a positional argument,
// and from FileLike by supporting Rest().
func FileLikeSlice[T ~string](name, desc string) options[T] {
	o := StringLikeSlice[T](name, desc)
	o.strOK = dashOK
	return o
}

// IntSlice creates and option that stores a slice of int values.
// It converts strings like [strconv.ParseInt].
// This differs from Int by supporting Rest().
func IntSlice(name, desc string, base int) options[int] {
	return IntLikeSlice[int](name, desc, base)
}

// IntLikeSlice creates and option that stores a slice of int-like values.
// It converts strings like [strconv.ParseInt].
// This differs from IntLike by supporting Rest().
func IntLikeSlice[T ~int | ~int8 | ~int16 | ~int32 | ~int64](name, desc string, base int) options[T] {
	var v []T
	parse := func(s string) error {
		i, err := strconv.ParseInt(s, base, int(unsafe.Sizeof(v[0]))*8)
		if e, ok := err.(*strconv.NumError); ok || errors.As(err, &e) {
			return fmt.Errorf("parsing %q as %T: %v", e.Num, v, e.Err)
		}
		if err != nil {
			return err
		}
		v = append(v, T(i))
		return nil
	}
	return options[T]{
		name:     name,
		desc:     desc,
		value:    &v,
		parse:    parse,
		prefixOK: "-",
	}
}

// UintSlice creates and option that stores a slice of uint values.
// It converts strings like [strconv.ParseUint].
// This differs from Uint by supporting Rest().
func UintSlice(name, desc string, base int) options[uint] {
	return UintLikeSlice[uint](name, desc, base)
}

// UintLikeSlice creates and option that stores a slice of uint-like values.
// It converts strings like [strconv.ParseUint].
// This differs from UintLike by supporting Rest().
func UintLikeSlice[T ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64](name, desc string, base int) options[T] {
	var v []T
	parse := func(s string) error {
		i, err := strconv.ParseUint(s, base, int(unsafe.Sizeof(v[0]))*8)
		if e, ok := err.(*strconv.NumError); ok || errors.As(err, &e) {
			return fmt.Errorf("parsing %q as %T: %v", e.Num, v, e.Err)
		}
		if err != nil {
			return err
		}
		v = append(v, T(i))
		return nil
	}
	return options[T]{
		name:  name,
		desc:  desc,
		value: &v,
		parse: parse,
	}
}

func manyParser(parse func(string) error) func(argSource) error {
	return func(s argSource) error {
		for v, ok := s.PeekMany(); ok; v, ok = s.PeekMany() {
			if !ok {
				return missingArgsError{}
			}
			if err := parse(v); err != nil {
				return err
			}
			s.Next()
		}
		return nil
	}
}
