//go:build linux || darwin

package terminal

import (
	"os"
	"os/exec"

	"github.com/creack/pty"
)

// unixPTY adalah implementasi PTY untuk Unix (Linux/macOS) memakai creack/pty.
type unixPTY struct {
	cmd  *exec.Cmd
	ptmx *os.File
}

// Start men-spawn shell di dalam pseudo-terminal baru dengan ukuran awal.
func Start(opts Options) (PTY, error) {
	shell := opts.Shell
	if shell == "" {
		shell = defaultShell()
	}

	cmd := exec.Command(shell, opts.Args...)
	cmd.Dir = opts.Cwd
	if opts.Env != nil {
		cmd.Env = opts.Env
	} else {
		cmd.Env = os.Environ()
	}

	rows := opts.Rows
	cols := opts.Cols
	if rows == 0 {
		rows = 24
	}
	if cols == 0 {
		cols = 80
	}

	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{Rows: rows, Cols: cols})
	if err != nil {
		return nil, err
	}

	return &unixPTY{cmd: cmd, ptmx: ptmx}, nil
}

func (p *unixPTY) Read(b []byte) (int, error)  { return p.ptmx.Read(b) }
func (p *unixPTY) Write(b []byte) (int, error) { return p.ptmx.Write(b) }

func (p *unixPTY) Resize(cols, rows uint16) error {
	return pty.Setsize(p.ptmx, &pty.Winsize{Rows: rows, Cols: cols})
}

// Close mematikan proses shell lalu menutup file descriptor PTY.
func (p *unixPTY) Close() error {
	if p.cmd.Process != nil {
		// Best-effort: minta keluar baik-baik dulu, lalu paksa.
		_ = p.cmd.Process.Signal(os.Interrupt)
		_ = p.cmd.Process.Kill()
	}
	return p.ptmx.Close()
}

func (p *unixPTY) Wait() error { return p.cmd.Wait() }

func (p *unixPTY) Pid() int {
	if p.cmd.Process == nil {
		return 0
	}
	return p.cmd.Process.Pid
}

// defaultShell menebak shell default pengguna dari $SHELL, fallback ke /bin/bash.
func defaultShell() string {
	if s := os.Getenv("SHELL"); s != "" {
		return s
	}
	return "/bin/bash"
}
