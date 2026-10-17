package official

// a131_register_test.go — 계좌 번호가 프로세스에 들어오는 유일한 입구(accountsLocked)가 받은 번호를 치환 규칙에 등록함.

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/redact"
)

func TestA131AccountsRegistersEveryAccountNumber(t *testing.T) {
	const first, second = "987-65-REGFIRST", "11223344556"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			_, _ = w.Write([]byte(`{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`))
		case "/api/v1/accounts":
			_, _ = w.Write([]byte(`{"result":[{"accountNo":"` + first + `","accountSeq":3,"accountType":"BROKERAGE"},` +
				`{"accountNo":"` + second + `","accountSeq":4,"accountType":"ISA"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	if got := redact.String(first + " " + second); !strings.Contains(got, first) {
		t.Fatalf("precondition: the sentinels must not be registered before the read: %q", got)
	}
	c := New(Credentials{APIKey: "k", SecretKey: "s"}, filepath.Join(t.TempDir(), "t.json"),
		WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	accounts, err := c.Accounts(context.Background())
	if err != nil || len(accounts) != 2 {
		t.Fatalf("Accounts = %v, %v", accounts, err)
	}
	// 반환값은 종전 그대로(원문) — 등록은 반환을 바꾸지 않음.
	if accounts[0].DisplayName != first || accounts[1].DisplayName != second {
		t.Fatalf("Accounts changed its result: %+v", accounts)
	}
	out := redact.String("first=" + first + " second=" + second + " dashed=112-23344-556")
	for _, leak := range []string{first, second, "112-23344-556"} {
		if strings.Contains(out, leak) {
			t.Fatalf("%q not registered: %q", leak, out)
		}
	}
}
