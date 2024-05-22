package run

import (
	"slices"
)

type options[T any] struct {
	name     string
	desc     string
	value    *[]T
	parse    func(string) (T, error)
	prefixOK string   // set to - to allow -[^-]+, or -- to allow --.+ in arg context
	strOK    []string // include unusual values such as - to allow them in arg context
	see      []*Command
}

func (o *options[T]) description() string                    { return o.desc }
func (o *options[T]) seeAlso() []*Command                    { return o.see }
func (o *options[T]) setSeeAlso(cmds ...*Command)            { o.see = cmds }
func (o *options[T]) parseDefault(arg string) error          { _, err := o.got([]string{arg}); return err }
func (o *options[T]) parseValues(args []string) (int, error) { return o.got(args) }
func (o *options[T]) okValues() []string                     { return o.strOK }
func (o *options[T]) okPrefix() string                       { return o.prefixOK }
func (o *options[T]) withPrefixOK(ok string) *options[T]     { o.prefixOK = ok; return o }
func (o *options[T]) withStrOK(ok []string) *options[T]      { o.strOK = ok; return o }

func (o *options[T]) got(args []string) (int, error) {
	*o.value = make([]T, 0, len(args))
	for i, arg := range args {
		v, err := o.parse(arg)
		if err != nil {
			return i, err
		}
		*o.value = append(*o.value, v)
	}
	return len(args), nil
}

func (o *options[T]) Value() []T { return *o.value }

// TODO: add FlagOn / FlagsOn, implemented as flags that split the string on a substring?

// Args returns an multi-Arg definition for this option with a custom alias.
func (o *options[T]) Args(name string) Arg {
	return Arg{option: o, name: name}
}

// StringSlice creates and option that stores a slice of string values.
// This differs from String by supporting Rest().
func StringSlice(name, desc string) *options[string] {
	return StringLikeSlice[string](name, desc)
}

// StringSliceVar creates and option that stores a slice of string values.
// This differs from String by supporting Rest().
func StringSliceVar(p *[]string, name, desc string) *options[string] {
	return StringLikeSliceVar(p, name, desc)
}

// StringLikeSlice creates and option that stores a slice of string-like values.
// This differs from StringLike by supporting Rest().
func StringLikeSlice[T ~string](name, desc string) *options[T] {
	parse := func(s string) (T, error) { return T(s), nil }
	return ParserSlice(name, desc, parse)
}

// StringLikeSlice creates and option that stores a slice of string-like values.
// This differs from StringLike by supporting Rest().
func StringLikeSliceVar[T ~string](p *[]T, name, desc string) *options[T] {
	parse := func(s string) (T, error) { return T(s), nil }
	return ParserSliceVar(p, name, desc, parse)
}

// StringSliceOf creates an option that stores a string-like slice of value from the provided list.
// This is suitable for small to medium sets of string-like names.
func StringSliceOf[T ~string](name, desc string, names ...T) *options[T] {
	var v []T
	return StringSliceVarOf[T](&v, name, desc, names...)
}

// StringSliceVarOf creates an option that stores a string-like slice of value from the provided list.
// This is suitable for small to medium sets of string-like names.
func StringSliceVarOf[T ~string](p *[]T, name, desc string, names ...T) *options[T] {
	nvs := make([]NamedValue[T], len(names))
	for i, nam := range names {
		nvs[i] = NamedValue[T]{Name: string(nam), Value: nam}
	}
	return NamedSliceVarOf(p, name, desc, nvs)
}

// NamedSliceOf creates an option that stores a slice of any type of value, looked up from the provided mapping.
// This is suitable for small to medium sets of names.
func NamedSliceOf[T any](name, desc string, mapping []NamedValue[T]) *options[T] {
	var v []T
	return NamedSliceVarOf(&v, name, desc, mapping)
}

// NamedSliceVarOf creates an option that stores a slice of any type of value, looked up from the provided mapping.
// This is suitable for small to medium sets of names.
func NamedSliceVarOf[T any](p *[]T, name, desc string, mapping []NamedValue[T]) *options[T] {
	mapping = slices.Clone(mapping)
	return ParserSliceVar(p, name, desc, (namedValues[T])(mapping).parse)
}

// FileSlice creates an option that stores a string slice of filenames.
// This differs from StringSlice by accepting "-" as a positional argument,
// and from File by supporting Rest().
func FileSlice(name, desc string) *options[string] {
	var v []string
	return FileSliceVar(&v, name, desc)
}

// FileSliceVar creates an option that stores a string slice of filenames.
// This differs from StringSlice by accepting "-" as a positional argument,
// and from File by supporting Rest().
func FileSliceVar(p *[]string, name, desc string) *options[string] {
	return FileLikeSliceVar(p, name, desc)
}

// FileLikeSlice creates an option that stores a string-like slice of filenames.
// This differs from StringLikeSlice by accepting "-" as a positional argument,
// and from FileLike by supporting Rest().
func FileLikeSlice[T ~string](name, desc string) *options[T] {
	var v []T
	return FileLikeSliceVar(&v, name, desc)
}

// FileLikeSliceVar creates an option that stores a string-like slice of filenames.
// This differs from StringLikeSlice by accepting "-" as a positional argument,
// and from FileLike by supporting Rest().
func FileLikeSliceVar[T ~string](p *[]T, name, desc string) *options[T] {
	return StringLikeSliceVar[T](p, name, desc).withStrOK(dashOK)
}

// IntSlice creates and option that stores a slice of int values.
// It converts strings like [strconv.ParseInt].
// This differs from Int by supporting Rest().
func IntSlice(name, desc string, base int) *options[int] {
	return IntLikeSlice[int](name, desc, base)
}

// IntSliceVar creates and option that stores a slice of int values.
// It converts strings like [strconv.ParseInt].
// This differs from Int by supporting Rest().
func IntSliceVar(p *[]int, name, desc string, base int) *options[int] {
	return IntLikeSliceVar(p, name, desc, base)
}

// IntLikeSlice creates and option that stores a slice of int-like values.
// It converts strings like [strconv.ParseInt].
// This differs from IntLike by supporting Rest().
func IntLikeSlice[T ~int | ~int8 | ~int16 | ~int32 | ~int64](name, desc string, base int) *options[T] {
	return ParserSlice(name, desc, parseIntLike[T](base)).withPrefixOK("-")
}

// IntLikeSliceVar creates and option that stores a slice of int-like values.
// It converts strings like [strconv.ParseInt].
// This differs from IntLike by supporting Rest().
func IntLikeSliceVar[T ~int | ~int8 | ~int16 | ~int32 | ~int64](p *[]T, name, desc string, base int) *options[T] {
	return ParserSliceVar(p, name, desc, parseIntLike[T](base)).withPrefixOK("-")
}

// UintSlice creates and option that stores a slice of uint values.
// It converts strings like [strconv.ParseUint].
// This differs from Uint by supporting Rest().
func UintSlice(name, desc string, base int) *options[uint] {
	return UintLikeSlice[uint](name, desc, base)
}

// UintSliceVar creates and option that stores a slice of uint values.
// It converts strings like [strconv.ParseUint].
// This differs from Uint by supporting Rest().
func UintSliceVar(p *[]uint, name, desc string, base int) *options[uint] {
	return UintLikeSliceVar(p, name, desc, base)
}

// UintLikeSlice creates and option that stores a slice of uint-like values.
// It converts strings like [strconv.ParseUint].
// This differs from UintLike by supporting Rest().
func UintLikeSlice[T ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64](name, desc string, base int) *options[T] {
	return ParserSlice(name, desc, parseUintLike[T](base))
}

// UintLikeSliceVar creates and option that stores a slice of uint-like values.
// It converts strings like [strconv.ParseUint].
// This differs from UintLike by supporting Rest().
func UintLikeSliceVar[T ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64](p *[]T, name, desc string, base int) *options[T] {
	return ParserSliceVar(p, name, desc, parseUintLike[T](base))
}

// FloatLikeSlice creates and option that stores a slice of float-like values.
// It converts strings like [strconv.ParseFloat].
// This differs from FloatLike by supporting Rest().
func FloatLikeSlice[T ~float32 | ~float64](name, desc string) *options[T] {
	return ParserSlice(name, desc, parseFloatLike[T]).withPrefixOK("-")
}

// FloatLikeSliceVar creates and option that stores a slice of float-like values.
// It converts strings like [strconv.ParseFloat].
// This differs from FloatLike by supporting Rest().
func FloatLikeSliceVar[T ~float32 | ~float64](p *[]T, name, desc string) *options[T] {
	return ParserSliceVar(p, name, desc, parseFloatLike[T]).withPrefixOK("-")
}

// ParserSlice creates an option that stores a slice of T values.
// It converts strings by calling parse.
func ParserSlice[T any](name, desc string, parse func(string) (T, error)) *options[T] {
	var v []T
	return ParserSliceVar(&v, name, desc, parse)
}

// ParserSliceVar creates an option that stores a slice of T values.
// It converts strings by calling parse.
func ParserSliceVar[T any](p *[]T, name, desc string, parse func(string) (T, error)) *options[T] {
	return &options[T]{
		name:  name,
		desc:  desc,
		value: p,
		parse: parse,
	}
}
