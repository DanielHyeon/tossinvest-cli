package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"sync"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicyrpc"
)

type fakePolicyCommands struct {
	mu           sync.Mutex
	state        positionpolicy.State
	runtime      positionpolicy.ManagementRuntime
	runtimeCalls int
	runtimeErr   error
	applies      int
	issued       bool
	used         bool
}

func (f *fakePolicyCommands) Runtime(context.Context) (positionpolicy.ManagementRuntime, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.runtimeCalls++
	return f.runtime, f.runtimeErr
}

func privateEngineTestDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("/tmp", "position-policy-engine-test-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	return dir
}

type fakeDescriptorTemp struct {
	chmodErr error
	writeErr error
	writeN   int
	writes   int
}

func (f *fakeDescriptorTemp) Chmod(os.FileMode) error { return f.chmodErr }
func (f *fakeDescriptorTemp) Write(body []byte) (int, error) {
	f.writes++
	if f.writeErr != nil || f.writeN != 0 {
		return f.writeN, f.writeErr
	}
	return len(body), nil
}
func (f *fakeDescriptorTemp) Sync() error  { return nil }
func (f *fakeDescriptorTemp) Close() error { return nil }

func TestWritePositionPolicyDescriptorPreservesChmodAndWriteErrors(t *testing.T) {
	chmodFailure := errors.New("chmod failed")
	chmod := &fakeDescriptorTemp{chmodErr: chmodFailure}
	if err := writePositionPolicyDescriptorBody(chmod, []byte("descriptor")); !errors.Is(err, chmodFailure) {
		t.Fatalf("chmod error=%v", err)
	}
	if chmod.writes != 0 {
		t.Fatalf("write called after chmod failure: %d", chmod.writes)
	}

	writeFailure := errors.New("write failed")
	write := &fakeDescriptorTemp{writeErr: writeFailure}
	if err := writePositionPolicyDescriptorBody(write, []byte("descriptor")); !errors.Is(err, writeFailure) {
		t.Fatalf("write error=%v", err)
	}
	if write.writes != 1 {
		t.Fatalf("writes=%d", write.writes)
	}

	short := &fakeDescriptorTemp{writeN: len("descriptor") - 1}
	if err := writePositionPolicyDescriptorBody(short, []byte("descriptor")); !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf("short write error=%v, want %v", err, io.ErrShortWrite)
	}
}

func (f *fakePolicyCommands) List(context.Context) ([]positionpolicy.State, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return []positionpolicy.State{f.state}, nil
}

func (f *fakePolicyCommands) Preview(_ context.Context,
	req positionpolicy.Request) (positionpolicy.Preview, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if req.ExpectedVersion != f.state.Version {
		return positionpolicy.Preview{}, positionpolicy.ErrVersionMismatch
	}
	after := f.state
	after.Version++
	f.issued = true
	return positionpolicy.Preview{Before: f.state, After: after, Action: req.Action,
		Capability: "opaque-transport-test-capability"}, nil
}

func (f *fakePolicyCommands) Apply(_ context.Context,
	req positionpolicy.ApplyRequest) (positionpolicy.State, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if !f.issued || f.used || req.Capability != "opaque-transport-test-capability" {
		return positionpolicy.State{}, positionpolicy.ErrCapabilityInvalid
	}
	f.used = true
	f.applies++
	f.state.Version++
	return f.state, nil
}

func TestEngineOwnsAuthenticatedPositionPolicyControlEndpoint(t *testing.T) {
	dir := privateEngineTestDir(t)
	commands := &fakePolicyCommands{state: positionpolicy.State{
		PositionID: "p-1", Symbol: "005930", AdoptionGeneration: 1,
		Version: 4, Status: positionpolicy.StatusManaged, ExitEligible: true,
	}}
	server, err := StartPositionPolicyCommandServer(dir, commands)
	if err != nil {
		t.Fatal(err)
	}
	descriptorPath := positionpolicyrpc.DescriptorPath(dir)
	info, err := os.Lstat(descriptorPath)
	if err != nil || info.Mode().Perm()&0o077 != 0 || !info.Mode().IsRegular() {
		t.Fatalf("descriptor info=%v err=%v", info, err)
	}
	client, err := positionpolicyrpc.Dial(context.Background(), descriptorPath)
	if err != nil {
		t.Fatal(err)
	}
	states, err := client.List(context.Background())
	if err != nil || len(states) != 1 || states[0].Version != 4 {
		t.Fatalf("states=%+v err=%v", states, err)
	}
	if err := server.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(descriptorPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("descriptor survived engine stop: %v", err)
	}
	if entries, err := os.ReadDir(dir); err != nil || len(entries) != 0 {
		t.Fatalf("control plane created non-endpoint state: entries=%v err=%v", entries, err)
	}
}

func TestPositionPolicyCommandEndpointDoesNotExposeRuntime(t *testing.T) {
	dir := privateEngineTestDir(t)
	commands := &fakePolicyCommands{}
	server, err := StartPositionPolicyCommandServer(dir, commands)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	body, err := os.ReadFile(positionpolicyrpc.DescriptorPath(dir))
	if err != nil {
		t.Fatal(err)
	}
	var descriptor positionpolicyrpc.Descriptor
	if err := json.Unmarshal(body, &descriptor); err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodGet, "http://"+descriptor.Address+"/v1/runtime", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+descriptor.Token)
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusNotFound || commands.runtimeCalls != 0 {
		t.Fatalf("runtime route status=%d calls=%d", response.StatusCode, commands.runtimeCalls)
	}
}

func TestPositionPolicyControlRejectsMissingBearerBeforeCommand(t *testing.T) {
	dir := privateEngineTestDir(t)
	commands := &fakePolicyCommands{state: positionpolicy.State{Version: 1}}
	server, err := StartPositionPolicyCommandServer(dir, commands)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	body, err := os.ReadFile(positionpolicyrpc.DescriptorPath(dir))
	if err != nil {
		t.Fatal(err)
	}
	var descriptor positionpolicyrpc.Descriptor
	if err := json.Unmarshal(body, &descriptor); err != nil {
		t.Fatal(err)
	}
	response, err := http.Get("http://" + descriptor.Address + "/v1/positions")
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unauthenticated status=%d", response.StatusCode)
	}
}

func TestPositionPolicyControlRejectsInsecureControlFilesystem(t *testing.T) {
	t.Run("group writable engine directory", func(t *testing.T) {
		dir := privateEngineTestDir(t)
		if err := os.Chmod(dir, 0o770); err != nil {
			t.Fatal(err)
		}
		if server, err := StartPositionPolicyCommandServer(dir, &fakePolicyCommands{}); err == nil {
			server.Close()
			t.Fatal("started control server in group-writable engine directory")
		}
	})

	t.Run("symlinked control directory", func(t *testing.T) {
		dir := privateEngineTestDir(t)
		if err := os.Symlink(privateEngineTestDir(t), positionpolicyrpc.ControlDirectory(dir)); err != nil {
			t.Fatal(err)
		}
		if server, err := StartPositionPolicyCommandServer(dir, &fakePolicyCommands{}); err == nil {
			server.Close()
			t.Fatal("started control server through symlinked directory")
		}
	})

	t.Run("wrong existing control directory mode", func(t *testing.T) {
		dir := privateEngineTestDir(t)
		if err := os.Mkdir(positionpolicyrpc.ControlDirectory(dir), 0o750); err != nil {
			t.Fatal(err)
		}
		if server, err := StartPositionPolicyCommandServer(dir, &fakePolicyCommands{}); err == nil {
			server.Close()
			t.Fatal("started control server in non-private control directory")
		}
	})
}

func TestCompetingControlClientsPreserveOneCASWinner(t *testing.T) {
	dir := privateEngineTestDir(t)
	commands := &fakePolicyCommands{state: positionpolicy.State{Version: 7}}
	server, err := StartPositionPolicyCommandServer(dir, commands)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	clients := make([]*positionpolicyrpc.Client, 2)
	for index := range clients {
		clients[index], err = positionpolicyrpc.Dial(context.Background(),
			positionpolicyrpc.DescriptorPath(dir))
		if err != nil {
			t.Fatal(err)
		}
	}
	errs := make([]error, 2)
	preview, err := clients[0].Preview(context.Background(), positionpolicy.Request{ExpectedVersion: 7})
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for index := range clients {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			_, errs[index] = clients[index].Apply(context.Background(), positionpolicy.ApplyRequest{
				Capability: preview.Capability,
			})
		}(index)
	}
	wg.Wait()
	wins, stale := 0, 0
	for _, err := range errs {
		if err == nil {
			wins++
		} else if errors.Is(err, positionpolicy.ErrCapabilityInvalid) {
			stale++
		}
	}
	if wins != 1 || stale != 1 || commands.applies != 1 {
		t.Fatalf("wins=%d stale=%d applies=%d errs=%v", wins, stale, commands.applies, errs)
	}
}

func TestPositionPolicyControlStrictlyBoundsAndFramesJSON(t *testing.T) {
	dir := privateEngineTestDir(t)
	commands := &fakePolicyCommands{state: positionpolicy.State{Version: 7}}
	server, err := StartPositionPolicyCommandServer(dir, commands)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	descriptorBody, err := os.ReadFile(positionpolicyrpc.DescriptorPath(dir))
	if err != nil {
		t.Fatal(err)
	}
	var descriptor positionpolicyrpc.Descriptor
	if err := json.Unmarshal(descriptorBody, &descriptor); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		contentType string
		body        []byte
		wantStatus  int
	}{
		{name: "trailing JSON value", contentType: "application/json", body: []byte(`{} {}`), wantStatus: http.StatusBadRequest},
		{name: "oversize body", contentType: "application/json", body: []byte(`{"capability":"` + string(bytes.Repeat([]byte("x"), (16<<10)+1)) + `"}`), wantStatus: http.StatusRequestEntityTooLarge},
		{name: "non JSON media type", contentType: "application/json-patch+json", body: []byte(`{}`), wantStatus: http.StatusUnsupportedMediaType},
		{name: "client supplied mutation scope", contentType: "application/json", body: []byte(`{"capability":"opaque","position_id":"p-2"}`), wantStatus: http.StatusBadRequest},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request, err := http.NewRequest(http.MethodPost, "http://"+descriptor.Address+"/v1/apply", bytes.NewReader(test.body))
			if err != nil {
				t.Fatal(err)
			}
			request.Header.Set("Authorization", "Bearer "+descriptor.Token)
			request.Header.Set("Content-Type", test.contentType)
			response, err := http.DefaultClient.Do(request)
			if err != nil {
				t.Fatal(err)
			}
			response.Body.Close()
			if response.StatusCode != test.wantStatus {
				t.Fatalf("status=%d, want %d", response.StatusCode, test.wantStatus)
			}
		})
	}
	if commands.applies != 0 {
		t.Fatalf("rejected bodies reached command service: applies=%d", commands.applies)
	}
}
