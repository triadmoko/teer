package service

import (
	"bytes"
	"encoding/base64"
	"errors"
	"strings"
	"sync"
	"testing"

	"teer/internal/domain"
	"teer/internal/infra/terminal"
)

// --- fakePTY ---

type fakePTY struct {
	readCh   chan []byte
	written  bytes.Buffer
	closed   chan struct{}
	waitCode int
	mu       sync.Mutex
}

func newFakePTY(waitCode int) *fakePTY {
	return &fakePTY{
		readCh: make(chan []byte, 16),
		closed: make(chan struct{}),
		waitCode: waitCode,
	}
}

func (f *fakePTY) Read(p []byte) (int, error) {
	select {
	case data, ok := <-f.readCh:
		if !ok {
			return 0, errors.New("pty closed")
		}
		n := copy(p, data)
		return n, nil
	case <-f.closed:
		return 0, errors.New("pty closed")
	}
}

func (f *fakePTY) Write(p []byte) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.written.Write(p)
}

func (f *fakePTY) Close() error {
	select {
	case <-f.closed:
	default:
		close(f.closed)
	}
	return nil
}

func (f *fakePTY) Resize(cols, rows uint16) error { return nil }

func (f *fakePTY) Pid() int { return 0 }

type fakeExitErr struct{ code int }

func (e *fakeExitErr) Error() string  { return "exit status" }
func (e *fakeExitErr) ExitCode() int  { return e.code }

func (f *fakePTY) Wait() error {
	<-f.closed
	if f.waitCode == 0 {
		return nil
	}
	return &fakeExitErr{code: f.waitCode}
}

// --- fakeEmitter ---

type emittedEvent struct {
	Name string
	Data any
}

type fakeEmitter struct {
	mu     sync.Mutex
	events []emittedEvent
	notify chan emittedEvent
}

func newFakeEmitter() *fakeEmitter {
	return &fakeEmitter{notify: make(chan emittedEvent, 32)}
}

func (e *fakeEmitter) Emit(name string, data ...any) bool {
	var d any
	if len(data) > 0 {
		d = data[0]
	}
	e.mu.Lock()
	e.events = append(e.events, emittedEvent{Name: name, Data: d})
	e.mu.Unlock()
	select {
	case e.notify <- emittedEvent{Name: name, Data: d}:
	default:
	}
	return true
}

func (e *fakeEmitter) waitFor(name string) emittedEvent {
	for ev := range e.notify {
		if ev.Name == name {
			return ev
		}
	}
	return emittedEvent{}
}

// Pastikan fakeEmitter memenuhi interface domain.EventEmitter
var _ domain.EventEmitter = (*fakeEmitter)(nil)

// --- helpers ---

func newTestService(em *fakeEmitter) *SessionService {
	svc := NewSessionService(em, nil)
	return svc
}

func newTestServiceWithFakePTY(em *fakeEmitter, pty *fakePTY) *SessionService {
	svc := NewSessionService(em, nil)
	svc.spawn = func(terminal.Options) (terminal.PTY, error) { return pty, nil }
	return svc
}

// --- tests ---

func TestStartSession_IDKosong(t *testing.T) {
	svc := newTestService(newFakeEmitter())
	err := svc.StartSession(StartOptions{})
	if err == nil || !strings.Contains(err.Error(), "id wajib diisi") {
		t.Fatalf("harusnya error id wajib diisi, dapat: %v", err)
	}
}

func TestStartSession_Sukses(t *testing.T) {
	em := newFakeEmitter()
	pty := newFakePTY(0)
	svc := newTestServiceWithFakePTY(em, pty)

	err := svc.StartSession(StartOptions{ID: "s1", Cols: 80, Rows: 24})
	if err != nil {
		t.Fatalf("StartSession: %v", err)
	}
	if !svc.IsRunning("s1") {
		t.Fatal("IsRunning harus true")
	}
	ids := svc.ListRunning()
	found := false
	for _, id := range ids {
		if id == "s1" {
			found = true
		}
	}
	if !found {
		t.Fatalf("ListRunning tidak memuat s1: %v", ids)
	}
}

func TestStartSession_DuaKaliSamaID(t *testing.T) {
	spawnCount := 0
	em := newFakeEmitter()
	svc := NewSessionService(em, nil)
	pty := newFakePTY(0)
	svc.spawn = func(terminal.Options) (terminal.PTY, error) {
		spawnCount++
		return pty, nil
	}

	_ = svc.StartSession(StartOptions{ID: "s1", Cols: 80, Rows: 24})
	_ = svc.StartSession(StartOptions{ID: "s1", Cols: 80, Rows: 24})

	if spawnCount != 1 {
		t.Fatalf("spawn harus dipanggil 1x, dapat %d", spawnCount)
	}
}

func TestStartSession_SpawnerError(t *testing.T) {
	em := newFakeEmitter()
	svc := NewSessionService(em, nil)
	svc.spawn = func(terminal.Options) (terminal.PTY, error) {
		return nil, errors.New("no shell")
	}

	err := svc.StartSession(StartOptions{ID: "s1", Cols: 80, Rows: 24})
	if err == nil || !strings.Contains(err.Error(), "gagal spawn shell") {
		t.Fatalf("harusnya error gagal spawn shell, dapat: %v", err)
	}
}

func TestStartSession_StartupCommand(t *testing.T) {
	em := newFakeEmitter()
	pty := newFakePTY(0)
	svc := newTestServiceWithFakePTY(em, pty)

	err := svc.StartSession(StartOptions{ID: "s1", Cols: 80, Rows: 24, StartupCommand: "echo hello"})
	if err != nil {
		t.Fatalf("StartSession: %v", err)
	}

	pty.mu.Lock()
	written := pty.written.String()
	pty.mu.Unlock()

	if written != "echo hello\n" {
		t.Fatalf("startup command harus ditulis ke PTY, dapat: %q", written)
	}
}

func TestReadLoop_Output(t *testing.T) {
	em := newFakeEmitter()
	pty := newFakePTY(0)
	svc := newTestServiceWithFakePTY(em, pty)

	err := svc.StartSession(StartOptions{ID: "s1", Cols: 80, Rows: 24})
	if err != nil {
		t.Fatalf("StartSession: %v", err)
	}

	payload := []byte("hello terminal")
	pty.readCh <- payload

	ev := em.waitFor("session:s1:out")
	decoded, err := base64.StdEncoding.DecodeString(ev.Data.(string))
	if err != nil {
		t.Fatalf("data bukan base64 valid: %v", err)
	}
	if string(decoded) != string(payload) {
		t.Fatalf("output tidak cocok: dapat %q, harap %q", decoded, payload)
	}
}

func TestReadLoop_Exit(t *testing.T) {
	em := newFakeEmitter()
	pty := newFakePTY(42)
	svc := newTestServiceWithFakePTY(em, pty)

	err := svc.StartSession(StartOptions{ID: "s1", Cols: 80, Rows: 24})
	if err != nil {
		t.Fatalf("StartSession: %v", err)
	}

	pty.Close()

	ev := em.waitFor("session:s1:exit")
	exitEv, ok := ev.Data.(ExitEvent)
	if !ok {
		t.Fatalf("payload bukan ExitEvent: %T %v", ev.Data, ev.Data)
	}
	if exitEv.Code != 42 {
		t.Fatalf("exit code harusnya 42, dapat %d", exitEv.Code)
	}

	if svc.IsRunning("s1") {
		t.Fatal("session harus terhapus dari map setelah exit")
	}
}

func TestWriteSession_IDMati(t *testing.T) {
	em := newFakeEmitter()
	svc := newTestService(em)

	err := svc.WriteSession("tidak-ada", "data")
	if err == nil || !strings.Contains(err.Error(), "tidak berjalan") {
		t.Fatalf("harusnya error tidak berjalan, dapat: %v", err)
	}
}

func TestResizeSession_IDMati(t *testing.T) {
	em := newFakeEmitter()
	svc := newTestService(em)

	err := svc.ResizeSession("tidak-ada", 80, 24)
	if err != nil {
		t.Fatalf("ResizeSession id mati harusnya nil, dapat: %v", err)
	}
}

func TestResizeSession_DimensiNol(t *testing.T) {
	em := newFakeEmitter()
	pty := newFakePTY(0)
	svc := newTestServiceWithFakePTY(em, pty)
	_ = svc.StartSession(StartOptions{ID: "s1", Cols: 80, Rows: 24})

	err := svc.ResizeSession("s1", 0, 0)
	if err != nil {
		t.Fatalf("ResizeSession dimensi <=0 harusnya nil, dapat: %v", err)
	}
}

func TestServiceShutdown(t *testing.T) {
	em := newFakeEmitter()

	ptys := []*fakePTY{newFakePTY(0), newFakePTY(0), newFakePTY(0)}
	idx := 0
	mu := sync.Mutex{}

	svc := NewSessionService(em, nil)
	svc.spawn = func(terminal.Options) (terminal.PTY, error) {
		mu.Lock()
		p := ptys[idx]
		idx++
		mu.Unlock()
		return p, nil
	}

	for i, id := range []string{"s1", "s2", "s3"} {
		_ = svc.StartSession(StartOptions{ID: id, Cols: 80, Rows: 24})
		_ = i
	}

	_ = svc.ServiceShutdown()

	for _, p := range ptys {
		select {
		case <-p.closed:
		default:
			t.Fatal("PTY harus ter-Close setelah ServiceShutdown")
		}
	}
}

func TestBuildEnv(t *testing.T) {
	// Tanpa override: TERM & COLORTERM harus selalu di-set
	env := buildEnv(map[string]string{"MY_VAR": "hello"})

	counts := map[string]int{}
	vals := map[string]string{}
	for _, e := range env {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			counts[parts[0]]++
			vals[parts[0]] = parts[1]
		}
	}

	if vals["TERM"] != "xterm-256color" {
		t.Fatalf("TERM harus xterm-256color, dapat %q", vals["TERM"])
	}
	if vals["COLORTERM"] != "truecolor" {
		t.Fatalf("COLORTERM harus truecolor, dapat %q", vals["COLORTERM"])
	}
	if vals["MY_VAR"] != "hello" {
		t.Fatalf("MY_VAR harus hello, dapat %q", vals["MY_VAR"])
	}
	// Override workspace menimpa (tanpa duplikat)
	env2 := buildEnv(map[string]string{"TERM": "dumb", "MY_VAR": "world"})
	counts2 := map[string]int{}
	vals2 := map[string]string{}
	for _, e := range env2 {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			counts2[parts[0]]++
			vals2[parts[0]] = parts[1]
		}
	}
	if vals2["TERM"] != "dumb" {
		t.Fatalf("override TERM harus menimpa, dapat %q", vals2["TERM"])
	}
	if counts2["TERM"] != 1 {
		t.Fatalf("TERM duplikat! muncul %d kali", counts2["TERM"])
	}
	if vals2["MY_VAR"] != "world" {
		t.Fatalf("MY_VAR harus world, dapat %q", vals2["MY_VAR"])
	}
}
