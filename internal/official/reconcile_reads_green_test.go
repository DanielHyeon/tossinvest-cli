package official

// reconcile_reads_green_test.go — a121 GREEN: 변이 원장이 찾은 빈칸 — 스키마 거절이 **어느 가드에서** 났는지 문구로
// 가른다(뒤 가드가 앞 가드를 대신 막는 우연을 막음), 종목 조회가 응답의 심볼을 되돌리는지 잰다.

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestDecodeReconcilePageNamesTheGuardThatRefused(t *testing.T) {
	for body, want := range map[string]string{
		`null`:                                "not an object",
		`[]`:                                  "not an object",
		`{"nextCursor":null,"hasNext":false}`: "conditionalOrders is missing",
		`{"conditionalOrders":null,"nextCursor":null,"hasNext":false}`:   "conditionalOrders is not an array",
		`{"conditionalOrders":[null],"nextCursor":null,"hasNext":false}`: "conditionalOrders[0] is not an object",
		`{"conditionalOrders":[],"nextCursor":null}`:                     "hasNext is missing",
		`{"conditionalOrders":[],"nextCursor":null,"hasNext":null}`:      "hasNext is not a boolean",
		`{"conditionalOrders":[],"hasNext":false}`:                       "nextCursor is missing",
		`{"conditionalOrders":[],"nextCursor":5,"hasNext":false}`:        "nextCursor is neither null nor a string",
		`{"conditionalOrders":[],"nextCursor":null,"hasNext":true}`:      "hasNext is true but nextCursor is empty",
	} {
		_, err := DecodeReconcileConditionalPage([]byte(body))
		if !errors.Is(err, ErrReconcileSchema) || !strings.Contains(err.Error(), want) {
			t.Fatalf("body %s: want ErrReconcileSchema naming %q, got %v", body, want, err)
		}
	}
}

func TestReconcileInstrumentReturnsTheBrokersEchoNotItsInput(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/oauth2/token":
			fmt.Fprint(w, `{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`)
		case "/api/v1/stocks":
			// 브로커가 다른 종목을 돌려준 경우 — 대사의 양성 대조가 이 차이를 봐야 한다.
			fmt.Fprint(w, `{"result":[{"symbol":"000660","name":"x","market":"KOSPI"}]}`)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)
	c := New(Credentials{APIKey: "k", SecretKey: "s"}, filepath.Join(t.TempDir(), "t.json"),
		WithBaseURL(srv.URL), WithHTTPClient(srv.Client()), WithAccountSeq(7))
	echo, err := c.ReconcileInstrument(context.Background(), "005930")
	if err != nil {
		t.Fatal(err)
	}
	if echo != "000660" {
		t.Fatalf("instrument read returned %q — it must return what the broker echoed (000660), not its input", echo)
	}
}
