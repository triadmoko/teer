//go:build linux || darwin

package terminal

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"time"
)

func TestEchoRoundTrip(t *testing.T) {
	p, err := Start(Options{Shell: "/bin/sh", Cols: 80, Rows: 24})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer p.Close()

	if p.Pid() <= 0 {
		t.Fatalf("Pid tidak valid: %d", p.Pid())
	}

	if _, err := p.Write([]byte("echo teer-ok\n")); err != nil {
		t.Fatalf("Write: %v", err)
	}

	got := readUntil(t, p, "teer-ok", 3*time.Second)
	if !strings.Contains(got, "teer-ok") {
		t.Fatalf("output tidak memuat penanda; dapat: %q", got)
	}
}

func TestResize(t *testing.T) {
	p, err := Start(Options{Shell: "/bin/sh", Cols: 80, Rows: 24})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer p.Close()

	if err := p.Resize(120, 40); err != nil {
		t.Fatalf("Resize: %v", err)
	}

	if _, err := p.Write([]byte("stty size\n")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	got := readUntil(t, p, "40 120", 3*time.Second)
	if !strings.Contains(got, "40 120") {
		t.Fatalf("ukuran terminal tidak ter-set; dapat: %q", got)
	}
}

func readUntil(t *testing.T, r io.Reader, marker string, timeout time.Duration) string {
	t.Helper()
	var buf bytes.Buffer
	deadline := time.Now().Add(timeout)
	chunk := make([]byte, 4096)
	for time.Now().Before(deadline) {
		n, err := r.Read(chunk)
		if n > 0 {
			buf.Write(chunk[:n])
			if strings.Contains(buf.String(), marker) {
				return buf.String()
			}
		}
		if err != nil {
			break
		}
	}
	return buf.String()
}
