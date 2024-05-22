package run

import (
	"cmp"
	"errors"
)

type flagOnly[T any] struct {
	name  string
	desc  string
	value *T
	seen  func() (T, error)
	see   []*Command
}

func (o *flagOnly[T]) description() string           { return o.desc }
func (o *flagOnly[T]) seeAlso() []*Command           { return o.see }
func (o *flagOnly[T]) setSeeAlso(cmds ...*Command)   { o.see = cmds }
func (o *flagOnly[T]) okValues() []string            { return nil }
func (o *flagOnly[T]) okPrefix() string              { return "" }
func (o *flagOnly[T]) parseDefault(arg string) error { return o.got(false) }
func (o *flagOnly[T]) parseFlag() error              { return o.got(true) }

func (o *flagOnly[T]) got(real bool) error {
	v, err := o.seen()
	if err != nil {
		return err
	}
	*o.value = v
	return nil
}

func (o *flagOnly[T]) Value() T { return *o.value }

// Flags returns a flag definition for this option with custom aliases.
// Zero values will omit either short or long. Do not omit both.
func (o *flagOnly[T]) Flags(short rune, long string) Flag {
	return Flag{option: o, rune: short, string: long}
}

// Flag returns a flag definition for this option using its name as the long.
// Thus an option named "opt" will have a flag name "--opt".
func (o *flagOnly[T]) Flag() Flag {
	return Flag{option: o, string: o.name}
}

// Slice returns a Param that converts a T to a []T.
// This can be used to share a handler between single and slice options.
func (o *flagOnly[T]) Slice() Param[[]T] {
	return sliceOf[T]{o.value}
}

var errRepeated = errors.New("repeated")

// Enabler creates an option that defaults to unseen, gets set to seen, and errors on repeat.
func Enabler[T any](name, desc string, unseen, seen T) *flagOnly[T] {
	v := unseen
	return EnablerVar(&v, name, desc, seen)
}

// EnablerVar creates an option that defaults to unseen, gets set to seen, and errors on repeat.
func EnablerVar[T any](p *T, name, desc string, seen T) *flagOnly[T] {
	enabled := false
	return &flagOnly[T]{
		name:  name,
		desc:  desc,
		value: p,
		seen: func() (T, error) {
			if enabled {
				return seen, errRepeated
			}
			enabled = true
			return seen, nil
		},
	}
}

// Toggler creates an option that toggles between two values, defaulting to the first.
func Toggler[T any](name, desc string, unseen, seen T) *flagOnly[T] {
	v := unseen
	return TogglerVar(&v, name, desc, seen)
}

// TogglerVar creates an option that toggles between two values, defaulting to the first.
func TogglerVar[T any](p *T, name, desc string, seen T) *flagOnly[T] {
	toggle := [2]T{*p, seen}
	n := 0
	return &flagOnly[T]{
		name:  name,
		desc:  desc,
		value: p,
		seen: func() (T, error) {
			n++
			return toggle[n%2], nil
		},
	}
}

// Accumulator creates an option that starts as initial, and adds increment every time it is seen.
func Accumulator[T cmp.Ordered](name, desc string, initial, increment T) *flagOnly[T] {
	v := initial
	return AccumulatorVar(&v, name, desc, increment)
}

// AccumulatorVar creates an option that starts as initial, and adds increment every time it is seen.
func AccumulatorVar[T cmp.Ordered](p *T, name, desc string, increment T) *flagOnly[T] {
	v := *p
	return &flagOnly[T]{
		name:  name,
		desc:  desc,
		value: p,
		seen: func() (T, error) {
			v += increment
			return v, nil
		},
	}
}
