//go:build unix

package engine

// a098 4.4b — 운영자 표면은 **엔진이 여는 소켓**이다.
//
// 여기서 재는 것은 두 가지다:
//
//	접근 경계  0700 디렉터리 · 0600 소켓 · 0600 descriptor · 토큰 불일치는 401
//	           — design D7.1 이 고른 접근 제어가 **파일시스템 권한**이므로
//	           그 권한이 실제로 그렇게 붙는지는 재야 하는 것이지 적어 두는 것이 아니다
//	권능 경계  이 descriptor 를 쥔 것이 **무엇을 할 수 없는지**. 소켓이 하나 더
//	           생기는 것은 토큰이 하나 더 생기는 것이고, 그 토큰이 주문에 닿으면
//	           표면이 아니라 구멍이다
//
// 그리고 승인이 **이 프로세스의 게이트**를 실제로 푸는 것을 본다 — 원장만 고치는
// 구현은 D7.1 이 없애려는 상태 그대로다.

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicyrpc"
)

func a098AlertSocketClient(socket string) *http.Client {
	transport := &http.Transport{Proxy: nil, DisableKeepAlives: true,
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, "unix", socket)
		}}
	return &http.Client{Transport: transport}
}

// a098StartAlertControl brings the endpoint up over a real journal and gate.
func a098StartAlertControl(t *testing.T, j *journal.Journal) (string, *execgw.EntryGate, *AlertControlServer) {
	t.Helper()
	dir := privateEngineTestDir(t)
	ops, gate := a098Ops(t, j)
	server, err := StartAlertControlServer(dir, ops)
	if err != nil {
		t.Fatalf("StartAlertControlServer: %v", err)
	}
	t.Cleanup(func() { _ = server.Close() })
	return dir, gate, server
}

// a098AlertToken reads the published descriptor the way a client would.
func a098AlertToken(t *testing.T, dir string) (string, string) {
	t.Helper()
	body, err := os.ReadFile(AlertControlDescriptorPath(dir))
	if err != nil {
		t.Fatalf("reading the descriptor: %v", err)
	}
	var descriptor AlertControlDescriptor
	if err := json.Unmarshal(body, &descriptor); err != nil {
		t.Fatalf("decoding the descriptor: %v", err)
	}
	if descriptor.Socket != AlertControlSocketFileName() {
		t.Fatalf("descriptor names socket %q, want %q — 이름을 안 고정하면 "+
			"다른 엔드포인트를 가리키는 descriptor 도 받아들여진다",
			descriptor.Socket, AlertControlSocketFileName())
	}
	if descriptor.PID != os.Getpid() || len(descriptor.Token) < 32 {
		t.Fatalf("descriptor = %+v", descriptor)
	}
	return descriptor.Token, filepath.Join(AlertControlDirectory(dir), descriptor.Socket)
}

// TestTheAlertSocketIsPrivateToThisUser is the access boundary, measured.
func TestTheAlertSocketIsPrivateToThisUser(t *testing.T) {
	j := openTestJournal(t)
	dir, _, server := a098StartAlertControl(t, j)

	for path, wantPerm := range map[string]os.FileMode{
		AlertControlDirectory(dir):      0o700,
		AlertControlDescriptorPath(dir): 0o600,
		AlertControlSocketPath(dir):     0o600,
	} {
		info, err := os.Lstat(path)
		if err != nil || info.Mode().Perm() != wantPerm {
			t.Fatalf("path=%s info=%v err=%v want perm=%04o", path, info, err, wantPerm)
		}
	}
	if info, _ := os.Lstat(AlertControlSocketPath(dir)); info == nil || info.Mode()&os.ModeSocket == 0 {
		t.Fatalf("the alert endpoint is not a Unix socket: %v", info)
	}
	// 그리고 position-policy 의 디렉터리가 **아니어야 한다**. 같은 디렉터리에 두면
	// 그쪽 descriptor 를 읽을 수 있는 것이 이쪽 권능까지 갖는다.
	//
	// ⛔ 비교는 **leaf 이름**이다. 전체 경로로 비교했더니 t.TempDir 이 만든 임시
	// 경로에 테스트 이름이 섞여 들어와 걸렸다 — 재려던 것은 경로가 아니라 이름이다.
	for _, shared := range []string{
		positionpolicyrpc.ControlDirectoryName, positionpolicyrpc.RuntimeControlDirectoryName,
	} {
		if filepath.Base(AlertControlDirectory(dir)) == shared {
			t.Errorf("the alert endpoint shares the %s control directory", shared)
		}
	}

	if err := server.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	for _, path := range []string{
		AlertControlDescriptorPath(dir), AlertControlSocketPath(dir), AlertControlDirectory(dir),
	} {
		if _, err := os.Lstat(path); !os.IsNotExist(err) {
			t.Errorf("%s survived Close (err=%v); descriptor 가 엔진보다 오래 살면 "+
				"안 듣는 소켓을 가리키는 토큰이 남는다", path, err)
		}
	}
}

// TestTheAlertSocketRefusesAWrongToken pins the second half of the boundary.
//
// ⛔ **두 라우트를 다 본다. 첫 판은 목록만 봤고, 뮤테이션이 그것을 잡았다** —
// 승인 라우트에서 `alertControlAuth` 를 떼어낸 판이 **네 테스트를 전부 통과했다.**
// 인증이 필요한 정도로 따지면 순서가 반대다: 목록은 읽고, 승인은 **게이트를 푼다.**
// 「토큰을 검사한다」를 라우트 하나로 재면 검사가 있는 라우트만 재는 것이다.
func TestTheAlertSocketRefusesAWrongToken(t *testing.T) {
	j := openTestJournal(t)
	dir, _, _ := a098StartAlertControl(t, j)
	token, socket := a098AlertToken(t, dir)
	client := a098AlertSocketClient(socket)

	routes := map[string]string{
		AlertControlListPath:        http.MethodGet,
		AlertControlAcknowledgePath: http.MethodPost,
	}
	headers := map[string]string{
		"no token":    "",
		"wrong token": "Bearer " + strings.Repeat("z", len(token)),
		"short token": "Bearer nope",
	}
	for path, method := range routes {
		for name, header := range headers {
			t.Run(path+" "+name, func(t *testing.T) {
				var body io.Reader
				if method == http.MethodPost {
					body = strings.NewReader(`{"operator":"intruder"}`)
				}
				req, _ := http.NewRequest(method, "http://alerts"+path, body)
				if header != "" {
					req.Header.Set("Authorization", header)
				}
				if body != nil {
					req.Header.Set("Content-Type", "application/json")
				}
				response, err := client.Do(req)
				if err != nil {
					t.Fatalf("Do: %v", err)
				}
				defer response.Body.Close()
				if response.StatusCode != http.StatusUnauthorized {
					t.Errorf("%s %s = %d, want 401", method, path, response.StatusCode)
				}
			})
		}
	}
}

// TestTheOperatorSocketListsAndAcknowledges is the whole loop over the wire.
//
// **게이트가 이 프로세스에서 풀리는 것**을 본다. 원장만 고치는 구현은 목록도 맞고
// 응답도 맞지만 `gate.Blocks()` 가 안 비고, 그것이 D7.1 이 없애려는 상태다.
func TestTheOperatorSocketListsAndAcknowledges(t *testing.T) {
	j := openTestJournal(t)
	id := a098SeedContentfulAlert(t, j)
	dir, gate, _ := a098StartAlertControl(t, j)
	gate.Block(execgw.ReasonAlertUndelivered, "critical alerts are undelivered")
	token, socket := a098AlertToken(t, dir)
	client := a098AlertSocketClient(socket)

	listBody := a098AlertCall(t, client, token, http.MethodGet, AlertControlListPath, nil, http.StatusOK)
	var rows []PendingAlert
	if err := json.Unmarshal(listBody, &rows); err != nil {
		t.Fatalf("decoding the listing: %v", err)
	}
	if len(rows) != 1 || rows[0].ID != id {
		t.Fatalf("listing = %+v, want the one pending row", rows)
	}
	// 소켓을 건너온 **바이트**에서 본다. 투영에서만 재면 전송이 원본을 덧붙여도 모른다.
	for column, sentinel := range a098AlertContentSentinels {
		if bytes.Contains(listBody, []byte(sentinel)) {
			t.Errorf("the wire response leaks the alert's %s: %s", column, listBody)
		}
	}

	ackBody := a098AlertCall(t, client, token, http.MethodPost, AlertControlAcknowledgePath,
		AcknowledgeRequest{Operator: "kim.operator"}, http.StatusOK)
	var result AcknowledgeResult
	if err := json.Unmarshal(ackBody, &result); err != nil {
		t.Fatalf("decoding the acknowledgement: %v", err)
	}
	if result.RemainingUndelivered != 0 || len(result.EntryBlocks) != 0 {
		t.Errorf("result = %+v, want nothing left", result)
	}
	if blocks := gate.Blocks(); len(blocks) != 0 {
		t.Errorf("gate.Blocks() = %v — 소켓 너머의 승인이 이 프로세스의 진입을 안 풀었다", blocks)
	}
	row, err := j.LookupAlert(context.Background(), id)
	if err != nil {
		t.Fatalf("LookupAlert: %v", err)
	}
	if row.AcknowledgedBy != "kim.operator" {
		t.Errorf("acknowledged_by = %q, want the operator's name", row.AcknowledgedBy)
	}
}

// TestTheAcknowledgeRouteRefusesMalformedCalls covers the refusals.
//
// 빈 운영자가 **400** 인 것이 중요하다: 500 이면 CLI 가 그것을 「엔진 오류」로 찍고,
// 운영자는 자기 입력이 아니라 엔진을 의심한다 — audit trail 을 지키는 유일한 거절이
// 엔진 고장처럼 보이면 그 거절은 우회 대상이 된다.
func TestTheAcknowledgeRouteRefusesMalformedCalls(t *testing.T) {
	j := openTestJournal(t)
	a098SeedContentfulAlert(t, j)
	dir, gate, _ := a098StartAlertControl(t, j)
	gate.Block(execgw.ReasonAlertUndelivered, "critical alerts are undelivered")
	token, socket := a098AlertToken(t, dir)
	client := a098AlertSocketClient(socket)

	t.Run("빈 운영자는 400", func(t *testing.T) {
		a098AlertCall(t, client, token, http.MethodPost, AlertControlAcknowledgePath,
			AcknowledgeRequest{Operator: "  "}, http.StatusBadRequest)
		if _, still := gate.Blocks()[execgw.ReasonAlertUndelivered]; !still {
			t.Error("the refused acknowledgement cleared the latch anyway")
		}
	})
	t.Run("목록 라우트에 POST 는 405", func(t *testing.T) {
		a098AlertCall(t, client, token, http.MethodPost, AlertControlListPath, nil,
			http.StatusMethodNotAllowed)
	})
	t.Run("승인 라우트에 GET 은 405", func(t *testing.T) {
		a098AlertCall(t, client, token, http.MethodGet, AlertControlAcknowledgePath, nil,
			http.StatusMethodNotAllowed)
	})
	t.Run("모르는 칸은 400", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "http://alerts"+AlertControlAcknowledgePath,
			strings.NewReader(`{"operator":"kim","surprise":1}`))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		response, err := client.Do(req)
		if err != nil {
			t.Fatalf("Do: %v", err)
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", response.StatusCode)
		}
	})
}

// TestTheAlertEndpointExposesOnlyTwoRoutes is the capability boundary.
//
// 소켓이 하나 더 생기는 것은 **토큰이 하나 더 생기는 것**이다. 그 토큰이 주문·토글·
// 엔진 기동에 닿지 않는다는 것은 적어 두는 것이 아니라 재는 것이다 — 옆 엔드포인트의
// 라우트 이름을 이 소켓에 그대로 던져서 확인한다.
func TestTheAlertEndpointExposesOnlyTwoRoutes(t *testing.T) {
	j := openTestJournal(t)
	dir, _, _ := a098StartAlertControl(t, j)
	token, socket := a098AlertToken(t, dir)
	client := a098AlertSocketClient(socket)

	for _, path := range []string{
		"/v1/apply", "/v1/preview", "/v1/positions", "/v1/runtime", "/v1/health",
		"/v1/quarantine/release", "/v1/strategy-runtime", "/",
	} {
		req, _ := http.NewRequest(http.MethodPost, "http://alerts"+path, strings.NewReader("{}"))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		response, err := client.Do(req)
		if err != nil {
			t.Fatalf("Do(%s): %v", path, err)
		}
		body, _ := io.ReadAll(io.LimitReader(response.Body, 1<<16))
		response.Body.Close()
		if response.StatusCode != http.StatusNotFound {
			t.Errorf("%s answered %d (%s); 이 소켓은 두 라우트만 연다",
				path, response.StatusCode, bytes.TrimSpace(body))
		}
	}
}

func a098AlertCall(t *testing.T, client *http.Client, token, method, path string,
	body any, wantStatus int) []byte {
	t.Helper()
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("encoding the request: %v", err)
		}
		reader = bytes.NewReader(encoded)
	}
	req, err := http.NewRequest(method, "http://alerts"+path, reader)
	if err != nil {
		t.Fatalf("building the request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	response, err := client.Do(req)
	if err != nil {
		t.Fatalf("Do(%s %s): %v", method, path, err)
	}
	defer response.Body.Close()
	out, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		t.Fatalf("reading the response: %v", err)
	}
	if response.StatusCode != wantStatus {
		t.Fatalf("%s %s = %d, want %d: %s", method, path, response.StatusCode, wantStatus,
			bytes.TrimSpace(out))
	}
	return out
}
