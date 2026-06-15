// Package terminal menyediakan abstraksi pseudo-terminal (PTY) lintas platform.
//
// Interface PTY sengaja dipisah dari implementasinya agar dukungan Windows
// (ConPTY) bisa ditambahkan tanpa merombak SessionService (lihat PRD NFR-6 &
// §13.6). Implementasi Unix memakai github.com/creack/pty; implementasi
// Windows menyusul di fase lanjutan.
package terminal

import "io"

// Options adalah parameter untuk men-spawn sebuah PTY baru.
type Options struct {
	Shell string   // path shell, mis. /bin/bash
	Args  []string // argumen tambahan untuk shell (opsional)
	Cwd   string   // direktori kerja awal
	Env   []string // environment lengkap ("KEY=VALUE"); bila nil pakai os.Environ()
	Cols  uint16   // lebar awal terminal
	Rows  uint16   // tinggi awal terminal
}

// PTY adalah pseudo-terminal yang menjalankan satu proses shell.
//
// Read/Write mengalirkan I/O mentah ke proses. Resize menyesuaikan ukuran
// terminal. Close mematikan proses dan melepaskan file descriptor. Wait
// memblokir sampai proses selesai dan mengembalikan exit code lewat error.
type PTY interface {
	io.ReadWriteCloser

	// Resize menyetel ukuran terminal (kolom x baris).
	Resize(cols, rows uint16) error

	// Wait memblokir sampai proses berakhir, mengembalikan error proses
	// (mis. *exec.ExitError) bila keluar dengan kode non-nol.
	Wait() error

	// Pid mengembalikan process id dari shell yang berjalan.
	Pid() int
}
