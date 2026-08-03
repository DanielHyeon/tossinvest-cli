package engine

// a063 task 5: the three quarantine routes on the engine's loopback endpoint.
//
// The claim that matters most here is the negative one: a command service
// without the capability serves exactly the route set it served before a063.
// That is how §0.2 holds at the transport layer, and it is what lets a console
// draw "no button" against an older engine instead of a 500.

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitquarantine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicyrpc"
)

func readTransportDescriptor(t *testing.T, dir string) positionpolicyrpc.Descriptor {
	t.Helper()
	body, err := os.ReadFile(positionpolicyrpc.DescriptorPath(dir))
	if err != nil {
		t.Fatal(err)
	}
	var descriptor positionpolicyrpc.Descriptor
	if err := json.Unmarshal(body, &descriptor); err != nil {
		t.Fatal(err)
	}
	return descriptor
}

type fakeQuarantineCommands struct {
	fakePolicyCommands
	mu       sync.Mutex
	rows     []exitquarantine.Row
	released int
}

func (f *fakeQuarantineCommands) Quarantines(context.Context) ([]exitquarantine.Row, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]exitquarantine.Row(nil), f.rows...), nil
}

func (f *fakeQuarantineCommands) PreviewQuarantineRelease(_ context.Context,
	req exitquarantine.Request) (exitquarantine.Preview, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, row := range f.rows {
		if row.PositionID == req.PositionID && row.Version == req.Version {
			return exitquarantine.Preview{Row: row, Capability: "opaque-quarantine-capability", WaitSeconds: 3}, nil
		}
	}
	return exitquarantine.Preview{}, exitquarantine.ErrNotQuarantined
}

func (f *fakeQuarantineCommands) ReleaseQuarantine(_ context.Context,
	req exitquarantine.ApplyRequest) (exitquarantine.Result, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if req.Capability != "opaque-quarantine-capability" {
		return exitquarantine.Result{}, exitquarantine.ErrCapabilityInvalid
	}
	if !req.Confirmed {
		return exitquarantine.Result{}, exitquarantine.ErrConfirmationRequired
	}
	f.released++
	row := f.rows[0]
	f.rows = nil
	return exitquarantine.Result{Row: row, ReleasedAt: "2026-08-04T12:00:00Z", Evidence: "server composed"}, nil
}

func quarantineTransportFixture(t *testing.T) (*fakeQuarantineCommands, string) {
	t.Helper()
	dir := privateEngineTestDir(t)
	commands := &fakeQuarantineCommands{
		fakePolicyCommands: fakePolicyCommands{state: positionpolicy.State{
			PositionID: "p-1", Symbol: "466100", AdoptionGeneration: 1, Version: 4,
			Status: positionpolicy.StatusManaged, ExitEligible: true,
		}},
		rows: []exitquarantine.Row{{
			PositionID: "p-1", Market: "kr", Symbol: "466100", Generation: 1, Version: 1,
			Reason: "ambiguous_recovery", Evidence: "identity mismatch",
			QuarantinedAt: "2026-08-03T09:03:40Z", Protection: "24929",
		}},
	}
	server, err := StartPositionPolicyCommandServer(dir, commands)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = server.Close() })
	return commands, dir
}

func TestTheQuarantineRoutesCarryAReleaseEndToEnd(t *testing.T) {
	commands, dir := quarantineTransportFixture(t)
	ctx := context.Background()

	client, err := positionpolicyrpc.Dial(ctx, positionpolicyrpc.DescriptorPath(dir))
	if err != nil {
		t.Fatal(err)
	}
	rows, err := client.Quarantines(ctx)
	if err != nil || len(rows) != 1 {
		t.Fatalf("rows=%+v err=%v", rows, err)
	}
	if rows[0].Protection != "24929" || rows[0].Reason != "ambiguous_recovery" {
		t.Fatalf("the round trip lost the row's evidence: %+v", rows[0])
	}

	preview, err := client.PreviewQuarantineRelease(ctx, exitquarantine.Request{
		PositionID: "p-1", Generation: 1, Version: 1, Actor: exitquarantine.ActorLocalOperator,
	})
	if err != nil {
		t.Fatalf("PreviewQuarantineRelease: %v", err)
	}
	if preview.Capability == "" || preview.WaitSeconds != 3 {
		t.Fatalf("preview lost its grant or delay: %+v", preview)
	}

	if _, err := client.ReleaseQuarantine(ctx, exitquarantine.ApplyRequest{
		Capability: preview.Capability, Confirmed: false,
	}); !errors.Is(err, exitquarantine.ErrConfirmationRequired) {
		t.Fatalf("unconfirmed release err = %v, want ErrConfirmationRequired", err)
	}
	result, err := client.ReleaseQuarantine(ctx, exitquarantine.ApplyRequest{
		Capability: preview.Capability, Confirmed: true,
	})
	if err != nil {
		t.Fatalf("ReleaseQuarantine: %v", err)
	}
	if result.Row.Symbol != "466100" || result.Evidence == "" {
		t.Fatalf("release result lost its record: %+v", result)
	}
	if commands.released != 1 {
		t.Fatalf("released %d times, want once", commands.released)
	}
}

func TestTheQuarantineRoutesRequireTheEngineToken(t *testing.T) {
	_, dir := quarantineTransportFixture(t)
	descriptor := readTransportDescriptor(t, dir)

	for _, route := range []struct{ method, path string }{
		{http.MethodGet, "/v1/quarantines"},
		{http.MethodPost, "/v1/quarantine/preview"},
		{http.MethodPost, "/v1/quarantine/release"},
	} {
		req, err := http.NewRequest(route.method, "http://"+descriptor.Address+route.path,
			strings.NewReader("{}"))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("%s %s without a bearer = %d, want 401", route.method, route.path, resp.StatusCode)
		}
	}
}

func TestTheQuarantineRoutesRejectTheWrongMethod(t *testing.T) {
	_, dir := quarantineTransportFixture(t)
	descriptor := readTransportDescriptor(t, dir)

	for _, route := range []struct{ method, path string }{
		{http.MethodPost, "/v1/quarantines"},
		{http.MethodGet, "/v1/quarantine/preview"},
		{http.MethodGet, "/v1/quarantine/release"},
	} {
		req, err := http.NewRequest(route.method, "http://"+descriptor.Address+route.path,
			strings.NewReader("{}"))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+descriptor.Token)
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("%s %s = %d, want 405", route.method, route.path, resp.StatusCode)
		}
	}
}

func TestAServiceWithoutQuarantineCapabilityRegistersNoNewRoutes(t *testing.T) {
	// A command service that predates a063: the mux must be exactly what it was.
	dir := privateEngineTestDir(t)
	server, err := StartPositionPolicyCommandServer(dir, &fakePolicyCommands{
		state: positionpolicy.State{PositionID: "p-1", AdoptionGeneration: 1, Version: 4},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = server.Close() })
	descriptor := readTransportDescriptor(t, dir)

	for _, path := range []string{"/v1/quarantines", "/v1/quarantine/preview", "/v1/quarantine/release"} {
		req, err := http.NewRequest(http.MethodGet, "http://"+descriptor.Address+path, nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+descriptor.Token)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("%s on a pre-a063 service = %d, want 404", path, resp.StatusCode)
		}
	}
}
