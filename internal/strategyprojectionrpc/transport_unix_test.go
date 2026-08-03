//go:build unix

package strategyprojectionrpc

import (
	"context"
	"encoding/json"
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

func TestUnixClientRejectsUnknownSchemaFieldsAndOversizedResponse(t *testing.T) {
	valid := strategyprojection.DormantSnapshot(time.Now().UTC())
	body, _ := json.Marshal(valid)
	unknown := append(body[:len(body)-1], []byte(`,"unknown":true}`)...)
	for _, test := range []struct {
		name string
		body []byte
	}{
		{"unknown field", unknown},
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
