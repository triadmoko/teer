//go:build windows

package terminal

import "errors"

func Start(opts Options) (PTY, error) {
	return nil, errors.New("terminal: dukungan Windows (ConPTY) belum diimplementasikan")
}

func defaultShell() string { return "powershell.exe" }
