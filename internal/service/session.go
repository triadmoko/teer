package service

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sync"

	"teer/internal/domain"
	"teer/internal/infra/terminal"
)

// Spawner membangkitkan sebuah PTY baru. Diinjeksikan ke SessionService agar
// lapisan use-case tidak terikat ke implementasi PTY konkret dan bisa diuji
// dengan PTY palsu. Default-nya terminal.Start (lihat NewSessionService).
type Spawner func(terminal.Options) (terminal.PTY, error)

// SessionService mengelola siklus hidup PTY untuk setiap sesi terminal.
//
// Setiap sesi yang berjalan = 1 PTY + 1 goroutine pembaca yang meng-emit
// output ke frontend lewat domain.EventEmitter (PRD §8.2). Service ini di-bind
// ke frontend sehingga metode ber-nama-eksport bisa dipanggil via RPC.
type SessionService struct {
	emitter domain.EventEmitter
	spawn   Spawner

	mu   sync.Mutex
	live map[string]*liveSession
}

// liveSession menyimpan state runtime sebuah PTY aktif.
type liveSession struct {
	id   string
	pty  terminal.PTY
	done chan struct{}
}

// NewSessionService membuat SessionService kosong. EventEmitter dipakai untuk
// meng-emit event output/exit ke frontend.
func NewSessionService(emitter domain.EventEmitter) *SessionService {
	return &SessionService{
		emitter: emitter,
		spawn:   terminal.Start,
		live:    make(map[string]*liveSession),
	}
}

// ---- Tipe yang dilewatkan ke/dari frontend (auto-bind ke TS) ----

// StartOptions adalah parameter untuk membangkitkan sebuah sesi.
type StartOptions struct {
	ID             string            `json:"id"`
	Shell          string            `json:"shell"`
	Cwd            string            `json:"cwd"`
	Env            map[string]string `json:"env"`
	StartupCommand string            `json:"startupCommand"`
	Cols           int               `json:"cols"`
	Rows           int               `json:"rows"`
}

// ExitEvent dikirim ke frontend saat sebuah sesi berakhir.
type ExitEvent struct {
	ID   string `json:"id"`
	Code int    `json:"code"`
	Err  string `json:"err,omitempty"`
}

// outputEvent/exitEventName mengembalikan nama event untuk sebuah sesi.
// Frontend mendengarkan event ini untuk menerima byte terminal (base64).
func outputEvent(id string) string   { return "session:" + id + ":out" }
func exitEventName(id string) string { return "session:" + id + ":exit" }

// ---- Metode yang di-bind ke frontend ----

// StartSession membangkitkan PTY baru untuk sebuah sesi. Idempoten: bila sesi
// dengan id tsb sudah berjalan, langsung sukses tanpa spawn baru.
func (s *SessionService) StartSession(opts StartOptions) error {
	if opts.ID == "" {
		return errors.New("session: id wajib diisi")
	}

	s.mu.Lock()
	if _, ok := s.live[opts.ID]; ok {
		s.mu.Unlock()
		return nil // sudah berjalan
	}
	s.mu.Unlock()

	p, err := s.spawn(terminal.Options{
		Shell: opts.Shell,
		Cwd:   opts.Cwd,
		Env:   buildEnv(opts.Env),
		Cols:  uint16(opts.Cols),
		Rows:  uint16(opts.Rows),
	})
	if err != nil {
		return fmt.Errorf("gagal spawn shell: %w", err)
	}

	ls := &liveSession{id: opts.ID, pty: p, done: make(chan struct{})}

	s.mu.Lock()
	s.live[opts.ID] = ls
	s.mu.Unlock()

	// Jalankan startupCommand (opsional) — ditulis ke stdin shell.
	if opts.StartupCommand != "" {
		_, _ = p.Write([]byte(opts.StartupCommand + "\n"))
	}

	go s.readLoop(ls)
	return nil
}

// WriteSession menulis input (keystroke/paste) ke stdin shell.
func (s *SessionService) WriteSession(id string, data string) error {
	ls := s.get(id)
	if ls == nil {
		return errors.New("session: tidak berjalan")
	}
	_, err := ls.pty.Write([]byte(data))
	return err
}

// ResizeSession menyesuaikan ukuran PTY mengikuti ukuran pane di frontend.
func (s *SessionService) ResizeSession(id string, cols int, rows int) error {
	ls := s.get(id)
	if ls == nil {
		return nil // sesi tidak berjalan; abaikan resize
	}
	if cols <= 0 || rows <= 0 {
		return nil
	}
	return ls.pty.Resize(uint16(cols), uint16(rows))
}

// CloseSession mematikan PTY sebuah sesi (kill proses).
func (s *SessionService) CloseSession(id string) error {
	s.mu.Lock()
	ls := s.live[id]
	s.mu.Unlock()
	if ls == nil {
		return nil
	}
	return ls.pty.Close()
}

// IsRunning melaporkan apakah sebuah sesi sedang aktif.
func (s *SessionService) IsRunning(id string) bool {
	return s.get(id) != nil
}

// ListRunning mengembalikan daftar id sesi yang sedang aktif.
func (s *SessionService) ListRunning() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	ids := make([]string, 0, len(s.live))
	for id := range s.live {
		ids = append(ids, id)
	}
	return ids
}

// ---- Internal ----

func (s *SessionService) get(id string) *liveSession {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.live[id]
}

// readLoop membaca output PTY dan meng-emit-nya ke frontend sampai sesi
// berakhir, lalu meng-emit event exit dan membersihkan state.
func (s *SessionService) readLoop(ls *liveSession) {
	defer close(ls.done)

	buf := make([]byte, 32*1024)
	for {
		n, err := ls.pty.Read(buf)
		if n > 0 {
			// Encode base64 agar byte mentah (termasuk UTF-8 parsial &
			// sekuens kontrol) selamat melewati serialisasi JSON event.
			enc := base64.StdEncoding.EncodeToString(buf[:n])
			s.emitter.Emit(outputEvent(ls.id), enc)
		}
		if err != nil {
			break
		}
	}

	// Reap proses untuk mendapatkan exit code.
	code, errStr := exitInfo(ls.pty.Wait())

	s.mu.Lock()
	delete(s.live, ls.id)
	s.mu.Unlock()

	s.emitter.Emit(exitEventName(ls.id), ExitEvent{ID: ls.id, Code: code, Err: errStr})
}

// shutdownAll mematikan seluruh PTY aktif. Dipanggil saat aplikasi ditutup
// (PRD §13.6) untuk mencegah proses zombie.
func (s *SessionService) shutdownAll() {
	s.mu.Lock()
	sessions := make([]*liveSession, 0, len(s.live))
	for _, ls := range s.live {
		sessions = append(sessions, ls)
	}
	s.mu.Unlock()

	for _, ls := range sessions {
		_ = ls.pty.Close()
	}
}

// ServiceShutdown dipanggil oleh Wails saat aplikasi berhenti.
func (s *SessionService) ServiceShutdown() error {
	s.shutdownAll()
	return nil
}

// buildEnv menggabungkan environment proses dengan override per-sesi.
// Bila override nil, kembalikan nil agar terminal.Start memakai os.Environ().
func buildEnv(override map[string]string) []string {
	if len(override) == 0 {
		return nil
	}
	env := os.Environ()
	for k, v := range override {
		env = append(env, k+"="+v)
	}
	return env
}

// exitInfo menerjemahkan error dari Wait menjadi (kode, pesan).
func exitInfo(err error) (int, string) {
	if err == nil {
		return 0, ""
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode(), ""
	}
	return -1, err.Error()
}
