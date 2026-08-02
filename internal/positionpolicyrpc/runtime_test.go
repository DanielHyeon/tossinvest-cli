//go:build unix

package positionpolicyrpc

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestDialRuntimeHonorsAlreadyCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := DialRuntime(ctx, RuntimeDescriptorPath(t.TempDir())); !errors.Is(err, context.Canceled) {
		t.Fatalf("DialRuntime() error=%v, want context.Canceled", err)
	}
}

func privateRuntimeTree(t *testing.T) (string, string, net.Listener) {
	t.Helper()
	engineDir := privateTestDir(t)
	controlDir := RuntimeControlDirectory(engineDir)
	if err := os.Mkdir(controlDir, 0o700); err != nil {
		t.Fatal(err)
	}
	socketPath := RuntimeSocketPath(engineDir)
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	if err := os.Chmod(socketPath, 0o600); err != nil {
		t.Fatal(err)
	}
	descriptor := RuntimeDescriptor{Socket: RuntimeSocketFileName,
		Token: strings.Repeat("x", 32), PID: os.Getpid()}
	body, err := json.Marshal(descriptor)
	if err != nil {
		t.Fatal(err)
	}
	writePrivateTestDescriptor(t, RuntimeDescriptorPath(engineDir), string(body))
	return engineDir, socketPath, listener
}

func TestRuntimeClientMethodSetIsReadOnly(t *testing.T) {
	typ := reflect.TypeOf(&RuntimeClient{})
	if typ.NumMethod() != 1 || typ.Method(0).Name != "Runtime" {
		t.Fatalf("RuntimeClient method set exposes more than Runtime: %v", typ.NumMethod())
	}
}

func TestDialRuntimeRejectsDescriptorSocketPathInjection(t *testing.T) {
	engineDir, _, _ := privateRuntimeTree(t)
	malicious := RuntimeDescriptor{Socket: "../command.sock", Token: strings.Repeat("x", 32), PID: os.Getpid()}
	body, _ := json.Marshal(malicious)
	writePrivateTestDescriptor(t, RuntimeDescriptorPath(engineDir), string(body))
	if _, err := DialRuntime(context.Background(), RuntimeDescriptorPath(engineDir)); err == nil {
		t.Fatal("DialRuntime accepted descriptor socket traversal")
	}
}

func TestPrepareRuntimeSocketNeverDeletesNonSocket(t *testing.T) {
	engineDir := privateTestDir(t)
	if err := os.Mkdir(RuntimeControlDirectory(engineDir), 0o700); err != nil {
		t.Fatal(err)
	}
	path := RuntimeSocketPath(engineDir)
	if err := os.WriteFile(path, []byte("not a socket"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := PrepareRuntimeSocket(path); err == nil {
		t.Fatal("PrepareRuntimeSocket accepted a regular file")
	}
	if _, err := os.Lstat(path); err != nil {
		t.Fatalf("PrepareRuntimeSocket deleted non-socket: %v", err)
	}
}

func TestDecodeRuntimeDescriptorRejectsUnknownAndTrailingContent(t *testing.T) {
	for _, body := range []string{
		`{"socket":"runtime.sock","token":"` + strings.Repeat("x", 32) + `","pid":1,"extra":true}`,
		`{"socket":"runtime.sock","token":"` + strings.Repeat("x", 32) + `","pid":1} {}`,
	} {
		if _, err := decodeRuntimeDescriptor(strings.NewReader(body)); err == nil {
			t.Fatalf("accepted malformed descriptor: %s", body)
		}
	}
}
