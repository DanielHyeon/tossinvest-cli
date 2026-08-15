//go:build unix

package strategyprojectionrpc

// a109_the_replaced_client_lets_go_test.go — a109 §2b.3 G5 (security #1, CRITICAL).
//
// a109의 재부착은 조회 데몬의 전략 reader 자리를 **갈아끼운다**. 밀려난 값이 이 패키지의
// client 면, 그것이 쥔 유휴 연결과 그 연결을 지키는 goroutine 을 놓아 줄 방법이 오늘까지
// 없었다 — `Client` 에 Close 가 없고 `Dial` 의 transport 에는 keep-alive 를 끄는 설정도,
// 유휴 수명 상한도 없다. 엔진이 하루에 열 번 오르내리면 열 벌이 데몬 수명 내내 남는다.
//
// 이 파일은 **테스트만** 더한다. a108이 확정한 회수·발행 의례는 건드리지 않는다 —
// a109가 이 패키지에 더하는 production 변경은 `Client.Close` 하나와 transport 필드
// 하나뿐이고, 그 둘을 여기서 잰다.

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojection"
)

// a109DialedClient 는 살아 있는 endpoint 하나에 붙은 client 다.
func a109DialedClient(t *testing.T) *Client {
	t.Helper()
	dir := shortRuntimeDir(t)
	server, err := Start(dir, projectionReaderStub{
		snapshot: strategyprojection.DormantSnapshot(time.Now().UTC()),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = server.Close() })
	client, err := Dial(context.Background(), DescriptorPath(dir))
	if err != nil {
		t.Fatal(err)
	}
	return client
}

// TestTheDialedTransportKeepsNoIdleConnections 는 저장소 관례를 이 client 에도 적용한다.
//
// 같은 모양의 사설 socket client 둘이 이미 그렇게 한다:
// `internal/positionpolicyrpc/runtime_unix.go:42` 와
// `cmd/tossctl/engine_alerts_client_unix.go:85`. 이 client 만 예외였고, 그 예외가
// 재부착과 만나 「놓아 줄 수 없는 연결」이 됐다.
func TestTheDialedTransportKeepsNoIdleConnections(t *testing.T) {
	client := a109DialedClient(t)
	transport, ok := client.http.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("client transport 타입 = %T, want *http.Transport", client.http.Transport)
	}
	if !transport.DisableKeepAlives {
		t.Error("dial 한 transport 가 keep-alive 를 켜 둔다 — 자리에서 밀려난 client 마다 " +
			"유휴 연결이 남는다(형제 client 둘은 이미 DisableKeepAlives 다)")
	}
}

// TestTheClientCanBeLetGo 는 자리가 밀려난 값을 닫을 수 있는지다.
//
// 그리고 **닫는 범위**를 함께 고정한다: Close 는 유휴 연결만 놓아 주므로 같은 client 로
// 다시 읽어도 된다. 이 성질이 재부착 쪽에서 잠금 밖·경합 없이 닫아도 되는 근거다
// (`closeEvictedStrategyReader`) — 옛 자리를 아직 읽고 있는 요청이 있어도 안전하다.
func TestTheClientCanBeLetGo(t *testing.T) {
	var _ io.Closer = (*Client)(nil)

	var absent *Client
	if err := absent.Close(); err != nil {
		t.Errorf("nil client 를 닫는 것이 오류였다: %v — 자리에는 부재도 앉는다", err)
	}

	client := a109DialedClient(t)
	if _, err := client.Read(context.Background()); err != nil {
		t.Fatalf("대조군 읽기가 실패했다: %v", err)
	}
	if err := client.Close(); err != nil {
		t.Fatalf("client 를 닫는 것이 오류였다: %v", err)
	}
	if _, err := client.Read(context.Background()); err != nil {
		t.Errorf("Close 가 client 를 무력화했다: %v — Close 는 유휴 연결만 놓아 주는 "+
			"것이어야 진행 중인 읽기와 나란히 부를 수 있다", err)
	}
}

// a109SpyTransport 는 Close 가 **transport 까지** 닿는지 세는 감시자다.
type a109SpyTransport struct{ closes atomic.Int64 }

func (t *a109SpyTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("a109: 이 감시자는 요청을 보내지 않는다")
}

func (t *a109SpyTransport) CloseIdleConnections() { t.closes.Add(1) }

// TestCloseReachesTheTransport 는 Close 의 **본문**을 잰다.
//
// 왜 따로 재는가: 이 client 의 transport 는 이제 keep-alive 를 쓰지 않으므로 놓아 줄
// 유휴 연결이 애초에 없다. 그래서 Close 의 본문을 지워도 dial 한 client 로는 아무
// 차이가 관측되지 않는다 — 뮤테이션 M38 이 그렇게 살아남았다(mutation-ledger-t2.md).
// 두 방어가 겹칠 때 바깥 것이 안쪽 것의 측정을 가리는 모양이고, 그때 답은 **안쪽을
// 직접 재는 것**이다: 감시 transport 하나를 앉히고 Close 가 그것에 닿는지 본다.
func TestCloseReachesTheTransport(t *testing.T) {
	spy := &a109SpyTransport{}
	client := &Client{
		baseURL: "http://unix", token: strings.Repeat("a", 32),
		http: &http.Client{Transport: spy},
	}
	if err := client.Close(); err != nil {
		t.Fatalf("Close 가 오류를 냈다: %v", err)
	}
	if got := spy.closes.Load(); got != 1 {
		t.Errorf("Close 가 transport 에 닿은 횟수 = %d, want 1 — Close 가 아무것도 하지 "+
			"않으면 keep-alive 를 쓰는 미래의 transport 에서 연결이 그대로 남는다", got)
	}
}
