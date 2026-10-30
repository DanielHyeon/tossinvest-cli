package prototype

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

func TestProcessFlock(t *testing.T) {
	if os.Getenv("A166_CHILD") == "1" {
		cert, e := os.ReadFile(os.Getenv("A166_CERT"))
		if e != nil {
			t.Fatal(e)
		}
		roots := x509.NewCertPool()
		if !roots.AppendCertsFromPEM(cert) {
			t.Fatal("cert")
		}
		l := &limiter{tls: &tls.Config{RootCAs: roots}, lockPath: os.Getenv("A166_LOCK"), evidence: os.Getenv("A166_LOG")}
		c := &http.Client{Transport: l, Timeout: 5 * time.Second}
		for i := 0; i < 2; i++ {
			r, e := c.Get(os.Getenv("A166_URL"))
			if e != nil {
				t.Fatal(e)
			}
			io.Copy(io.Discard, r.Body)
			r.Body.Close()
		}
		return
	}
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "ok") }))
	defer s.Close()
	d := t.TempDir()
	cert := filepath.Join(d, "cert.pem")
	os.WriteFile(cert, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: s.Certificate().Raw}), 0600)
	log := filepath.Join(d, "attempts")
	lock := filepath.Join(d, "lock")
	run := func() *exec.Cmd {
		c := exec.Command(os.Args[0], "-test.run=^TestProcessFlock$", "-test.timeout=10s")
		c.Env = append(os.Environ(), "A166_CHILD=1", "A166_CERT="+cert, "A166_LOG="+log, "A166_LOCK="+lock, "A166_URL="+s.URL)
		return c
	}
	a, b := run(), run()
	if e := a.Start(); e != nil {
		t.Fatal(e)
	}
	if e := b.Start(); e != nil {
		t.Fatal(e)
	}
	if e := a.Wait(); e != nil {
		t.Fatal(e)
	}
	if e := b.Wait(); e != nil {
		t.Fatal(e)
	}
	raw, e := os.ReadFile(log)
	if e != nil {
		t.Fatal(e)
	}
	var times []int64
	for _, line := range strings.Fields(string(raw)) {
		n, e := strconv.ParseInt(line, 10, 64)
		if e != nil {
			t.Fatal(e)
		}
		times = append(times, n)
	}
	sort.Slice(times, func(i, j int) bool { return times[i] < times[j] })
	if len(times) != 4 {
		t.Fatalf("writes=%d", len(times))
	}
	for i := 1; i < len(times); i++ {
		if time.Duration(times[i]-times[i-1]) < 350*time.Millisecond {
			t.Fatalf("spacing %v", time.Duration(times[i]-times[i-1]))
		}
	}
	if time.Duration(times[3]-times[0]) < time.Second {
		t.Fatal("four in one second")
	}
}

func TestKilledLockOwnerReleases(t *testing.T) {
	if os.Getenv("A166_KILL_CHILD") == "1" {
		f, e := os.OpenFile(os.Getenv("A166_LOCK"), os.O_CREATE|os.O_RDWR, 0600)
		if e != nil {
			t.Fatal(e)
		}
		if e = syscall.Flock(int(f.Fd()), syscall.LOCK_EX); e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(os.Getenv("A166_READY"), []byte("ready"), 0600); e != nil {
			t.Fatal(e)
		}
		select {}
	}
	d := t.TempDir()
	path := filepath.Join(d, "lock")
	ready := filepath.Join(d, "ready")
	c := exec.Command(os.Args[0], "-test.run=^TestKilledLockOwnerReleases$")
	c.Env = append(os.Environ(), "A166_KILL_CHILD=1", "A166_LOCK="+path, "A166_READY="+ready)
	if e := c.Start(); e != nil {
		t.Fatal(e)
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		if _, e := os.Stat(ready); e == nil {
			break
		}
		if time.Now().After(deadline) {
			c.Process.Kill()
			c.Wait()
			t.Fatal("child did not lock")
		}
		time.Sleep(10 * time.Millisecond)
	}
	if e := c.Process.Kill(); e != nil {
		t.Fatal(e)
	}
	c.Wait()
	f, e := os.OpenFile(path, os.O_RDWR, 0600)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	start := time.Now()
	if e = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); e != nil {
		t.Fatal(e)
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	time.Sleep(350 * time.Millisecond)
	if time.Since(start) < 350*time.Millisecond {
		t.Fatal("cooldown")
	}
}

func TestKillAfterTLSFirstWriteAndRestart(t *testing.T) {
	var calls atomic.Int32
	entered := make(chan struct{})
	release := make(chan struct{})
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			close(entered)
			<-release
		}
		fmt.Fprint(w, "ok")
	}))
	defer s.Close()
	d := t.TempDir()
	cert := filepath.Join(d, "cert.pem")
	if err := os.WriteFile(cert, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: s.Certificate().Raw}), 0600); err != nil {
		t.Fatal(err)
	}
	log := filepath.Join(d, "attempts")
	lock := filepath.Join(d, "lock")
	run := func() *exec.Cmd {
		c := exec.Command(os.Args[0], "-test.run=^TestProcessFlock$", "-test.timeout=10s")
		c.Env = append(os.Environ(), "A166_CHILD=1", "A166_CERT="+cert, "A166_LOG="+log, "A166_LOCK="+lock, "A166_URL="+s.URL)
		return c
	}
	first := run()
	if err := first.Start(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		first.Process.Kill()
		first.Wait()
		close(release)
		t.Fatal("first request missing")
	}
	if err := first.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	first.Wait()
	close(release)
	second := run()
	if err := second.Run(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Fields(string(data))
	if len(lines) != 3 {
		t.Fatalf("writes=%d", len(lines))
	}
	a, _ := strconv.ParseInt(lines[0], 10, 64)
	b, _ := strconv.ParseInt(lines[1], 10, 64)
	if time.Duration(b-a) < 350*time.Millisecond {
		t.Fatalf("restart spacing=%v", time.Duration(b-a))
	}
}
