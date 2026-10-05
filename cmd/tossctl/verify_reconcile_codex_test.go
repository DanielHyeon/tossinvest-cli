package main

// verify_reconcile_codex_test.go — a121 4.2 외부 모델(codex, clean) 발견 CG-4(승인 읽기)·CG-6(읽기용 클라이언트의 신원
// 재확인)의 시험.

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

func setReconcileTerminalIO(t *testing.T, in io.Reader, out io.Writer) {
	t.Helper()
	previous := verifyReconcileTerminal
	verifyReconcileTerminal = func() (io.Reader, io.Writer, bool) { return in, out, true }
	t.Cleanup(func() { verifyReconcileTerminal = previous })
}

type errAfterReader struct {
	data string
	err  error
	done bool
}

func (r *errAfterReader) Read(p []byte) (int, error) {
	if r.done {
		return 0, r.err
	}
	r.done = true
	return copy(p, r.data), nil
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("terminal gone") }

// TestReconcileApprovalNeedsACompleteLine 는 codex CG-4 다 — "y" 뒤에 개행 없이 EOF·읽기 오류가 오면 승인이 아니다.
func TestReconcileApprovalNeedsACompleteLine(t *testing.T) {
	for name, reader := range map[string]io.Reader{
		"y-then-eof":   strings.NewReader("y"),
		"y-then-error": &errAfterReader{data: "y", err: io.ErrUnexpectedEOF},
	} {
		t.Run(name, func(t *testing.T) {
			setReconcileTerminalIO(t, reader, &strings.Builder{})
			if err := verifyReconcileApprove(context.Background(), reconcileApprovalFixture()); err == nil {
				t.Fatal("an answer without a complete line approved the reconciliation")
			}
		})
	}
}

// TestReconcileApprovalHonoursCancellation — 답을 기다리는 중 ctx 가 취소되면 즉시 거절한다(잠금을 쥔 채 막히지 않음).
func TestReconcileApprovalHonoursCancellation(t *testing.T) {
	blocked, _ := io.Pipe()
	setReconcileTerminalIO(t, blocked, &strings.Builder{})
	ctx, cancel := context.WithCancel(context.Background())
	go func() { time.Sleep(50 * time.Millisecond); cancel() }()
	done := make(chan error, 1)
	go func() { done <- verifyReconcileApprove(ctx, reconcileApprovalFixture()) }()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("a cancelled approval must say so; got %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("the approval kept blocking after its context was cancelled")
	}
}

// TestReconcileApprovalRefusesWhenTheQuestionCannotBeShown — 질문을 단말에 못 쓰면 승인이 아니다.
func TestReconcileApprovalRefusesWhenTheQuestionCannotBeShown(t *testing.T) {
	setReconcileTerminalIO(t, strings.NewReader("y\n"), failingWriter{})
	if err := verifyReconcileApprove(context.Background(), reconcileApprovalFixture()); err == nil {
		t.Fatal("an approval whose question could not be shown went through")
	}
}

// TestReconcileReaderRechecksIdentityBeforeItsFirstScopedRead 는 codex CG-6 이다 — 계좌 판정과 읽기가 토큰 캐시를 공유하므로,
// 승인 대기 중 캐시가 다른 자격으로 바뀌면 읽기용 클라이언트가 보는 계좌가 달라진다. 첫 계좌 범위 읽기 전에 읽기용
// 클라이언트로 Accounts() 를 다시 불러, 다르면 읽지 않는다.
func TestReconcileReaderRechecksIdentityBeforeItsFirstScopedRead(t *testing.T) {
	for name, second := range map[string]string{
		"other-account-same-seq": accountsBody(accountRow("999-99-998901", 5)), // 같은 seq·같은 끝 4자리, 다른 계좌
		"same-account-other-seq": accountsBody(accountRow("123-45-678901", 9)),
		"two-accounts":           accountsBody(accountRow("123-45-678901", 5), accountRow("555-55-550000", 6)),
	} {
		t.Run(name, func(t *testing.T) {
			var accounts, lists atomic.Int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch r.URL.Path {
				case "/oauth2/token":
					fmt.Fprint(w, `{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`)
				case "/api/v1/accounts":
					if accounts.Add(1) == 1 {
						fmt.Fprint(w, accountsBody(accountRow("123-45-678901", 5)))
					} else {
						fmt.Fprint(w, second)
					}
				case "/api/v1/conditional-orders", "/api/v1/orders":
					lists.Add(1)
					fmt.Fprint(w, `{"result":{"conditionalOrders":[],"orders":[],"nextCursor":null,"hasNext":false}}`)
				default:
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			t.Cleanup(srv.Close)
			reader, _, err := newVerifyReconcileReader(context.Background(), official.Credentials{APIKey: "k", SecretKey: "s"},
				filepath.Join(t.TempDir(), "token.json"), official.WithBaseURL(srv.URL), official.WithHTTPClient(srv.Client()))
			if err != nil {
				t.Fatalf("construction: %v", err)
			}
			_, err = reader.ReconcileConditionalOrdersPage(context.Background(), "OPEN", "005930", "", 100)
			if !errors.Is(err, errReconcileIdentityChanged) {
				t.Fatalf("a changed reading identity must stop the reads; got %v", err)
			}
			_, err = reader.ReconcileOpenOrdersPage(context.Background(), "005930", "", 100)
			if !errors.Is(err, errReconcileIdentityChanged) {
				t.Fatalf("the plain-order read must stop too; got %v", err)
			}
			if n := lists.Load(); n != 0 {
				t.Fatalf("%d list read(s) went out under a changed identity", n)
			}
		})
	}
}
