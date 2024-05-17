package run

import (
	"io"
	"os"
)

// DefaultEnviron returns an Environ that works like the os package.
func DefaultEnviron() (env Environ) {
	env.fillDefaults()
	return env
}

type Environ struct {
	Args      []string // Args includes the program name at index 0.
	Stdin     io.Reader
	Stdout    io.Writer
	Stderr    io.Writer
	Getenv    func(string) string
	LookupEnv func(string) (string, bool)
}

// WithArgs overrides Args.
func (e Environ) WithArgs(args []string) Environ {
	e.Args = args
	return e
}

// WithInput overrides Stdin. Pass nil to simulate a closed handle.
func (e Environ) WithInput(r io.Reader) Environ {
	e.Stdin = r
	if r == nil {
		e.Stdin = nullReader{}
	}
	return e
}

// WithOutput overrides Stdout and Stderr together.
func (e Environ) WithOutput(w io.Writer) Environ {
	e.Stdout, e.Stderr = w, w
	return e
}

// WithStdout overrides Stdout.
func (e Environ) WithStdout(w io.Writer) Environ {
	e.Stdout = w
	return e
}

// WithStderr overrides Stderr.
func (e Environ) WithStderr(w io.Writer) Environ {
	e.Stderr = w
	return e
}

// WithVariables overrides Getenv and LookupEnv to use env.
func (e Environ) WithVariables(env Variables) Environ {
	e.Getenv, e.LookupEnv = env.get, env.lookup
	return e
}

func (e *Environ) fillDefaults() {
	if e.Args == nil {
		e.Args = os.Args
	}
	if e.Stdin == nil {
		e.Stdin = os.Stdin
	}
	if e.Stdout == nil {
		e.Stdout = os.Stdout
	}
	if e.Stderr == nil {
		e.Stderr = os.Stderr
	}
	if e.Getenv == nil {
		e.Getenv = os.Getenv
	}
	if e.LookupEnv == nil {
		e.LookupEnv = os.LookupEnv
	}
}

// EntryFunc is the recommended type for your entry function.
type EntryFunc func(Environ) error

// Main runs your entry function and returns an int suitable for [os.Exit].
func Main(entry EntryFunc) int {
	err := entry(DefaultEnviron())
	if err != nil {
		return 1
	}
	return 0
}

// TestErr runs your entry function and returns its error.
func TestErr(main EntryFunc) error {
	return main(DefaultEnviron())
}

// Variables represents an environment block.
type Variables map[string]string

func (v Variables) get(env string) string            { return v[env] }
func (v Variables) lookup(env string) (string, bool) { a, b := v[env]; return a, b }

type nullReader struct{}

func (nullReader) Read([]byte) (int, error) { return 0, io.EOF }
