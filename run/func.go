package run

import "context"

type Context struct {
	context.Context
	Environ
	Command *Command
}

// Handler can be passed to (*Command).Runs, or used applied as an option in CmdOpt.
type Handler func(Context) error

func (h Handler) applyCommand(cmd *Command) error {
	return cmd.SetHandler(h)
}

type Param[T any] interface{ Value() T }

// Handler1 adapts a func(Context, T1) for (*command).Runs.
func Handler1[
	T1 any, V1 Param[T1],
](
	handler func(Context, T1) error,
	v1 V1,
) Handler {
	return func(ctx Context) error {
		return handler(ctx, v1.Value())
	}
}

// Handler2 adapts a func(Context, T1, T2) for (*command).Runs.
func Handler2[
	T1 any, V1 Param[T1],
	T2 any, V2 Param[T2],
](
	handler func(Context, T1, T2) error,
	v1 V1, v2 V2,
) Handler {
	return func(ctx Context) error {
		return handler(ctx, v1.Value(), v2.Value())
	}
}

// Handler3 adapts a func(Context, T1, T2, T3) for (*command).Runs.
func Handler3[
	T1 any, V1 Param[T1],
	T2 any, V2 Param[T2],
	T3 any, V3 Param[T3],
](
	handler func(Context, T1, T2, T3) error,
	v1 V1, v2 V2, v3 V3,
) Handler {
	return func(ctx Context) error {
		return handler(ctx, v1.Value(), v2.Value(), v3.Value())
	}
}

// Handler4 adapts a func(Context, T1...T4) for (*command).Runs.
func Handler4[
	T1 any, V1 Param[T1],
	T2 any, V2 Param[T2],
	T3 any, V3 Param[T3],
	T4 any, V4 Param[T4],
](
	handler func(Context, T1, T2, T3, T4) error,
	v1 V1, v2 V2, v3 V3, v4 V4,
) Handler {
	return func(ctx Context) error {
		return handler(ctx, v1.Value(), v2.Value(), v3.Value(), v4.Value())
	}
}

// Handler5 adapts a func(Context, T1...T5) for (*command).Runs.
func Handler5[
	T1 any, V1 Param[T1],
	T2 any, V2 Param[T2],
	T3 any, V3 Param[T3],
	T4 any, V4 Param[T4],
	T5 any, V5 Param[T5],
](
	handler func(Context, T1, T2, T3, T4, T5) error,
	v1 V1, v2 V2, v3 V3, v4 V4, v5 V5,
) Handler {
	return func(ctx Context) error {
		return handler(ctx, v1.Value(), v2.Value(), v3.Value(), v4.Value(), v5.Value())
	}
}

// Handler6 adapts a func(Context, T1...T6) for (*command).Runs.
func Handler6[
	T1 any, V1 Param[T1],
	T2 any, V2 Param[T2],
	T3 any, V3 Param[T3],
	T4 any, V4 Param[T4],
	T5 any, V5 Param[T5],
	T6 any, V6 Param[T6],
](
	handler func(Context, T1, T2, T3, T4, T5, T6) error,
	v1 V1, v2 V2, v3 V3, v4 V4, v5 V5, v6 V6,
) Handler {
	return func(ctx Context) error {
		return handler(ctx, v1.Value(), v2.Value(), v3.Value(), v4.Value(), v5.Value(), v6.Value())
	}
}

// Handler7 adapts a func(Context, T1...T7) for (*command).Runs.
func Handler7[
	T1 any, V1 Param[T1],
	T2 any, V2 Param[T2],
	T3 any, V3 Param[T3],
	T4 any, V4 Param[T4],
	T5 any, V5 Param[T5],
	T6 any, V6 Param[T6],
	T7 any, V7 Param[T7],
](
	handler func(Context, T1, T2, T3, T4, T5, T6, T7) error,
	v1 V1, v2 V2, v3 V3, v4 V4, v5 V5, v6 V6, v7 V7,
) Handler {
	return func(ctx Context) error {
		return handler(ctx, v1.Value(), v2.Value(), v3.Value(), v4.Value(), v5.Value(), v6.Value(), v7.Value())
	}
}

type pass[T any] struct{ param T }

func (v pass[T]) Value() T    { return v.param }
func Pass[T any](v T) pass[T] { return pass[T]{v} }
