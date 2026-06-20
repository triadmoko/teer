package service

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"teer/internal/domain"
	"teer/internal/infra/terminal"
)

type Spawner func(terminal.Options) (terminal.PTY, error)

type SessionService struct {
	emitter domain.EventEmitter
	spawn   Spawner

	mu   sync.Mutex
	live map[string]*liveSession
}

type liveSession struct {
	id   string
	pty  terminal.PTY
	done chan struct{}
}

func NewSessionService(emitter domain.EventEmitter) *SessionService {
	return &SessionService{
		emitter: emitter,
		spawn:   terminal.Start,
		live:    make(map[string]*liveSession),
	}
}

type StartOptions struct {
	ID             string            `json:"id"`
	Shell          string            `json:"shell"`
	Cwd            string            `json:"cwd"`
	Env            map[string]string `json:"env"`
	StartupCommand string            `json:"startupCommand"`
	Cols           int               `json:"cols"`
	Rows           int               `json:"rows"`
}

type ExitEvent struct {
	ID   string `json:"id"`
	Code int    `json:"code"`
	Err  string `json:"err,omitempty"`
}

func outputEvent(id string) string   { return "session:" + id + ":out" }
func exitEventName(id string) string { return "session:" + id + ":exit" }

func (s *SessionService) StartSession(opts StartOptions) error {
	if opts.ID == "" {
		return errors.New("session: id wajib diisi")
	}

	s.mu.Lock()
	if _, ok := s.live[opts.ID]; ok {
		s.mu.Unlock()
		return nil
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

	if opts.StartupCommand != "" {
		_, _ = p.Write([]byte(opts.StartupCommand + "\n"))
	}

	go s.readLoop(ls)
	return nil
}

func (s *SessionService) WriteSession(id string, data string) error {
	ls := s.get(id)
	if ls == nil {
		return errors.New("session: tidak berjalan")
	}
	_, err := ls.pty.Write([]byte(data))
	return err
}

func (s *SessionService) ResizeSession(id string, cols int, rows int) error {
	ls := s.get(id)
	if ls == nil {
		return nil
	}
	if cols <= 0 || rows <= 0 {
		return nil
	}
	return ls.pty.Resize(uint16(cols), uint16(rows))
}

func (s *SessionService) CloseSession(id string) error {
	s.mu.Lock()
	ls := s.live[id]
	s.mu.Unlock()
	if ls == nil {
		return nil
	}
	return ls.pty.Close()
}

func (s *SessionService) IsRunning(id string) bool {
	return s.get(id) != nil
}

func (s *SessionService) ListRunning() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	ids := make([]string, 0, len(s.live))
	for id := range s.live {
		ids = append(ids, id)
	}
	return ids
}

func (s *SessionService) get(id string) *liveSession {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.live[id]
}

func (s *SessionService) readLoop(ls *liveSession) {
	defer close(ls.done)

	buf := make([]byte, 32*1024)
	for {
		n, err := ls.pty.Read(buf)
		if n > 0 {
			enc := base64.StdEncoding.EncodeToString(buf[:n])
			s.emitter.Emit(outputEvent(ls.id), enc)
		}
		if err != nil {
			break
		}
	}

	code, errStr := exitInfo(ls.pty.Wait())

	s.mu.Lock()
	delete(s.live, ls.id)
	s.mu.Unlock()

	s.emitter.Emit(exitEventName(ls.id), ExitEvent{ID: ls.id, Code: code, Err: errStr})
}

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

func (s *SessionService) ServiceShutdown() error {
	s.shutdownAll()
	return nil
}

func buildEnv(override map[string]string) []string {
	env := os.Environ()
	// Selalu set TERM & COLORTERM agar warna/highlight prompt bekerja walau app
	// diluncurkan dari dock/desktop icon — proses semacam itu tidak mewarisi
	// environment terminal, sehingga TERM kosong dan bash menonaktifkan warna.
	// Renderer selalu xterm.js yang mendukung 256 color + truecolor.
	env = setEnv(env, "TERM", "xterm-256color")
	env = setEnv(env, "COLORTERM", "truecolor")
	for k, v := range override {
		env = setEnv(env, k, v)
	}
	return env
}

// setEnv menimpa nilai key bila sudah ada di env (mencegah duplikat), atau
// menambah entri baru bila belum ada. Berformat "KEY=VALUE".
func setEnv(env []string, key, val string) []string {
	prefix := key + "="
	for i, e := range env {
		if strings.HasPrefix(e, prefix) {
			env[i] = prefix + val
			return env
		}
	}
	return append(env, prefix+val)
}

// exitCoder dipenuhi *exec.ExitError (Unix) maupun *exitError ConPTY (Windows).
type exitCoder interface{ ExitCode() int }

func exitInfo(err error) (int, string) {
	if err == nil {
		return 0, ""
	}
	var ec exitCoder
	if errors.As(err, &ec) {
		return ec.ExitCode(), ""
	}
	return -1, err.Error()
}
