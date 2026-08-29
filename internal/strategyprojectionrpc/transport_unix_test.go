//go:build unix

package strategyprojectionrpc

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojection"
)

type projectionReaderStub struct{ snapshot strategyprojection.Snapshot }

func (s projectionReaderStub) Read(context.Context) (strategyprojection.Snapshot, error) {
	return s.snapshot, nil
}

func TestAuthenticatedUnixProjectionCrossNamespaceAndReadOnlyAuthority(t *testing.T) {
	dir := shortRuntimeDir(t)
	snapshot := strategyprojection.DormantSnapshot(time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC))
	server, err := Start(dir, projectionReaderStub{snapshot: snapshot})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = server.Close() })

	descriptorPath := DescriptorPath(dir)
	info, err := os.Lstat(descriptorPath)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("descriptor mode=%v err=%v", info, err)
	}
	client, err := Dial(context.Background(), descriptorPath)
	if err != nil {
		t.Fatal(err)
	}
	got, err := client.Read(context.Background())
	if err != nil || !reflect.DeepEqual(got, snapshot) {
		t.Fatalf("snapshot=%+v err=%v", got, err)
	}

	typeName := reflect.TypeOf(client)
	for _, forbidden := range []string{"Preview", "Apply", "Order", "Activate", "Protect", "Mutate"} {
		if _, exists := typeName.MethodByName(forbidden); exists {
			t.Fatalf("runtime read client exposes %s", forbidden)
		}
	}
}

func TestStartReclaimsExactDeadProjectionEndpoint(t *testing.T) {
	dir := shortRuntimeDir(t)
	if err := os.Mkdir(ControlDirectory(dir), 0o700); err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("unix", SocketPath(dir))
	if err != nil {
		t.Fatal(err)
	}
	listener.(*net.UnixListener).SetUnlinkOnClose(false)
	if err := os.Chmod(SocketPath(dir), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	if err := writeDescriptor(DescriptorPath(dir), Descriptor{SchemaVersion: descriptorSchema, Socket: SocketFileName,
		Token: strings.Repeat("x", 43), PID: 1 << 30}); err != nil {
		t.Fatal(err)
	}
	server, err := Start(dir, projectionReaderStub{snapshot: strategyprojection.DormantSnapshot(time.Now().UTC())})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = server.Close() })
	client, err := Dial(context.Background(), DescriptorPath(dir))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Read(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestStartRefusesLiveProjectionOwnerWithoutRemovingIt(t *testing.T) {
	dir := shortRuntimeDir(t)
	first, err := Start(dir, projectionReaderStub{snapshot: strategyprojection.DormantSnapshot(time.Now().UTC())})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = first.Close() })
	if _, err := Start(dir, projectionReaderStub{snapshot: strategyprojection.DormantSnapshot(time.Now().UTC())}); err == nil {
		t.Fatal("second live projection owner was accepted")
	}
	client, err := Dial(context.Background(), DescriptorPath(dir))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Read(context.Background()); err != nil {
		t.Fatalf("live owner was damaged: %v", err)
	}
}

func TestUnixEndpointStrictMethodBodyQueryAuthAndRouteGuards(t *testing.T) {
	dir := shortRuntimeDir(t)
	server, err := Start(dir, projectionReaderStub{snapshot: strategyprojection.DormantSnapshot(time.Now().UTC())})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = server.Close() })
	descriptor, err := readDescriptor(DescriptorPath(dir))
	if err != nil {
		t.Fatal(err)
	}
	client := unixHTTPClient(SocketPath(dir))
	for _, test := range []struct {
		method, path, token, body string
		want                      int
	}{
		{http.MethodGet, ProjectionPath, "bad", "", http.StatusUnauthorized},
		{http.MethodPost, ProjectionPath, descriptor.Token, "", http.StatusMethodNotAllowed},
		{http.MethodGet, ProjectionPath + "?market=KR", descriptor.Token, "", http.StatusBadRequest},
		{http.MethodGet, ProjectionPath, descriptor.Token, `{}`, http.StatusBadRequest},
		{http.MethodGet, "/v1/preview", descriptor.Token, "", http.StatusNotFound},
		{http.MethodPost, "/v1/apply", descriptor.Token, `{}`, http.StatusNotFound},
		{http.MethodPost, "/v1/orders", descriptor.Token, `{}`, http.StatusNotFound},
		{http.MethodPost, "/v1/activation", descriptor.Token, `{}`, http.StatusNotFound},
		{http.MethodPost, "/v1/protection", descriptor.Token, `{}`, http.StatusNotFound},
	} {
		req, _ := http.NewRequest(test.method, "http://unix"+test.path, strings.NewReader(test.body))
		req.Header.Set("Authorization", "Bearer "+test.token)
		response, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		_ = response.Body.Close()
		if response.StatusCode != test.want {
			t.Errorf("%s %s=%d want=%d", test.method, test.path, response.StatusCode, test.want)
		}
	}
}

func shortRuntimeDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("/tmp", "sprpc-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	return dir
}

// TestUnixClientRejectsOversizedResponse 는 a108 의 두 팔 중 **크기 상한**만 남긴 것이다.
//
// 사라진 팔은 "unknown field" 였다. a112 의 승인된 spec 이 그 반대를 요구한다 —
// "older readers가 additive unknown fields를 무시할 수 있어야 한다 (SHALL)" 와
// `legacy projection reader` 시나리오. 권위 순서상 승인된 OpenSpec 이 현재 테스트보다
// 위이고, 엄격함이 실제로 하던 일은 엔진이 필드를 하나 더 실을 때 구버전 콘솔 화면을
// 죽이는 것뿐이었다(토큰·0600 소켓·Validate 가 이미 각각의 일을 한다).
//
// 새 계약은 `a112_additive_fields_test.go` 가 양쪽으로 잰다: 모르는 필드는 무시하고,
// 의미가 틀린 스냅샷은 여전히 거절한다.
func TestUnixClientRejectsOversizedResponse(t *testing.T) {
	for _, test := range []struct {
		name string
		body []byte
	}{
		{"oversized", []byte(strings.Repeat("x", MaxProjectionBytes+1))},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write(test.body) }))
			defer server.Close()
			client := &Client{baseURL: server.URL, token: strings.Repeat("x", 43), http: server.Client()}
			if _, err := client.Read(context.Background()); err == nil {
				t.Fatal("strict client accepted invalid response")
			}
		})
	}
}
