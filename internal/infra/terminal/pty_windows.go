//go:build windows

package terminal

import "errors"

// Implementasi Windows (ConPTY) belum tersedia — placeholder agar interface
// PTY tetap terdefinisi lintas platform (PRD NFR-6, milestone M6).
func Start(opts Options) (PTY, error) {
	return nil, errors.New("terminal: dukungan Windows (ConPTY) belum diimplementasikan")
}

func defaultShell() string { return "powershell.exe" }
