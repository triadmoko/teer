package terminal

import "io"

type Options struct {
	Shell string
	Args  []string
	Cwd   string
	Env   []string
	Cols  uint16
	Rows  uint16
}

type PTY interface {
	io.ReadWriteCloser

	Resize(cols, rows uint16) error

	Wait() error

	Pid() int
}
