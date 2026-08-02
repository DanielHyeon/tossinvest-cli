//go:build unix

package engine

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicyrpc"
)

func runtimeSocketHTTPClient(socket string) *http.Client {
	transport := &http.Transport{Proxy: nil, DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "unix", socket)
	}}
	return &http.Client{Transport: transport}
}

func TestRuntimeOnlyUnixControlCrossNamespaceTopologyAndRouteAbsence(t *testing.T) {
	dir := privateEngineTestDir(t)
	commands := &fakePolicyCommands{runtime: positionpolicy.ManagementRuntime{
		EffectiveKnown: true,
		Effective:      positionpolicy.NewAdoptionSettings(true, .03, []string{"AAPL"}, nil, ""),
	}}
	server, err := StartPositionPolicyRuntimeServer(dir, commands)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()

	descriptorPath := positionpolicyrpc.RuntimeDescriptorPath(dir)
	socketPath := positionpolicyrpc.RuntimeSocketPath(dir)
	for path, wantPerm := range map[string]os.FileMode{
		positionpolicyrpc.RuntimeControlDirectory(dir): 0o700,
		descriptorPath: 0o600,
		socketPath:     0o600,
	} {
		info, err := os.Lstat(path)
		if err != nil || info.Mode().Perm() != wantPerm {
			t.Fatalf("path=%s info=%v err=%v want perm=%04o", path, info, err, wantPerm)
		}
	}
	if info, _ := os.Lstat(socketPath); info == nil || info.Mode()&os.ModeSocket == 0 {
		t.Fatalf("runtime endpoint is not a Unix socket: %v", info)
	}

	reader, err := positionpolicyrpc.DialRuntime(context.Background(), descriptorPath)
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := reader.Runtime(context.Background())
	if err != nil || !runtime.EffectiveKnown || !runtime.Effective.Enabled {
		t.Fatalf("runtime=%+v err=%v", runtime, err)
	}

	body, err := os.ReadFile(descriptorPath)
	if err != nil {
		t.Fatal(err)
	}
	var descriptor positionpolicyrpc.RuntimeDescriptor
	if err := json.Unmarshal(body, &descriptor); err != nil {
		t.Fatal(err)
	}
	client := runtimeSocketHTTPClient(socketPath)
	tests := []struct {
		method string
		path   string
		want   int
	}{
		{http.MethodGet, "/v1/runtime", http.StatusOK},
		{http.MethodPost, "/v1/runtime", http.StatusMethodNotAllowed},
		{http.MethodGet, "/v1/preview", http.StatusNotFound},
		{http.MethodPost, "/v1/preview", http.StatusNotFound},
		{http.MethodPost, "/v1/apply", http.StatusNotFound},
		{http.MethodGet, "/v1/reconcile", http.StatusNotFound},
		{http.MethodPost, "/v1/reconcile/resolve", http.StatusNotFound},
	}
	for _, tt := range tests {
		req, err := http.NewRequest(tt.method, "http://runtime"+tt.path, nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+descriptor.Token)
		response, err := client.Do(req)
		if err != nil {
			t.Fatalf("%s %s: %v", tt.method, tt.path, err)
		}
		_, _ = io.Copy(io.Discard, response.Body)
		response.Body.Close()
		if response.StatusCode != tt.want {
			t.Fatalf("%s %s status=%d want=%d", tt.method, tt.path, response.StatusCode, tt.want)
		}
	}

	unauthorized, err := client.Get("http://runtime/v1/runtime")
	if err != nil {
		t.Fatal(err)
	}
	unauthorized.Body.Close()
	if unauthorized.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unauthorized status=%d", unauthorized.StatusCode)
	}

	if err := server.Close(); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{descriptorPath, socketPath, positionpolicyrpc.RuntimeControlDirectory(dir)} {
		if _, err := os.Lstat(path); !os.IsNotExist(err) {
			t.Fatalf("runtime control artifact survived Close: %s err=%v", path, err)
		}
	}
}

var _ positionpolicy.RuntimeReader = (*positionpolicyrpc.RuntimeClient)(nil)

func TestCommandAndRuntimeControlsHaveIndependentLifecycles(t *testing.T) {
	dir := privateEngineTestDir(t)
	commands := &fakePolicyCommands{runtime: positionpolicy.ManagementRuntime{EffectiveKnown: true}}
	commandServer, err := StartPositionPolicyCommandServer(dir, commands)
	if err != nil {
		t.Fatal(err)
	}
	runtimeServer, err := StartPositionPolicyRuntimeServer(dir, commands)
	if err != nil {
		commandServer.Close()
		t.Fatal(err)
	}
	defer runtimeServer.Close()

	if err := commandServer.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(positionpolicyrpc.DescriptorPath(dir)); !os.IsNotExist(err) {
		t.Fatalf("command descriptor survived command Close: %v", err)
	}
	reader, err := positionpolicyrpc.DialRuntime(t.Context(), positionpolicyrpc.RuntimeDescriptorPath(dir))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := reader.Runtime(t.Context()); err != nil {
		t.Fatalf("runtime endpoint did not survive independent command Close: %v", err)
	}
}
