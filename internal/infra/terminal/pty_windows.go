//go:build windows

package terminal

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/UserExistsError/conpty"
)

type windowsPTY struct {
	cpty *conpty.ConPty
}

func Start(opts Options) (PTY, error) {
	shell := opts.Shell
	if shell == "" {
		shell = defaultShell()
	}

	// ConPTY menerima satu command line, bukan argv terpisah.
	parts := append([]string{shell}, opts.Args...)
	cmdLine := strings.Join(parts, " ")

	rows := opts.Rows
	cols := opts.Cols
	if rows == 0 {
		rows = 24
	}
	if cols == 0 {
		cols = 80
	}

	env := opts.Env
	if env == nil {
		env = os.Environ()
	}

	cOpts := []conpty.ConPtyOption{
		conpty.ConPtyDimensions(int(cols), int(rows)),
		conpty.ConPtyEnv(env),
	}
	if opts.Cwd != "" {
		cOpts = append(cOpts, conpty.ConPtyWorkDir(opts.Cwd))
	}

	cpty, err := conpty.Start(cmdLine, cOpts...)
	if err != nil {
		return nil, err
	}

	return &windowsPTY{cpty: cpty}, nil
}

func (p *windowsPTY) Read(b []byte) (int, error)  { return p.cpty.Read(b) }
func (p *windowsPTY) Write(b []byte) (int, error) { return p.cpty.Write(b) }

func (p *windowsPTY) Resize(cols, rows uint16) error {
	return p.cpty.Resize(int(cols), int(rows))
}

func (p *windowsPTY) Close() error { return p.cpty.Close() }

func (p *windowsPTY) Wait() error {
	code, err := p.cpty.Wait(context.Background())
	if err != nil {
		return err
	}
	if code != 0 {
		return &exitError{code: code}
	}
	return nil
}

func (p *windowsPTY) Pid() int { return p.cpty.Pid() }

func defaultShell() string { return "powershell.exe" }

// exitError membawa exit code ConPTY agar terbaca oleh exitInfo lewat
// antarmuka { ExitCode() int }, setara *exec.ExitError di Unix.
type exitError struct{ code uint32 }

func (e *exitError) Error() string { return fmt.Sprintf("exit status %d", e.code) }
func (e *exitError) ExitCode() int { return int(e.code) }
