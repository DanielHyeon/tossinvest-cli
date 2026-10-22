package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// stocksServer 는 공식 /api/v1/stocks 를 흉내 냄. 요청한 symbols 를 적고, status 가 0 이 아니면 그 상태로 답함.
type stocksServer struct {
	mu      sync.Mutex
	asked   []string
	status  int
	records map[string]string // symbol → StockInfo JSON 한 개(응답에 없을 종목은 넣지 않음)
}

func (s *stocksServer) handler(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/oauth2/token":
			_, _ = w.Write([]byte(`{"access_token":"t","expires_in":86400,"token_type":"Bearer"}`))
		case "/api/v1/stocks":
			if r.Method != http.MethodGet {
				t.Errorf("non-GET %s", r.Method)
			}
			s.mu.Lock()
			s.asked = append(s.asked, r.URL.Query().Get("symbols"))
			status := s.status
			s.mu.Unlock()
			if status != 0 {
				w.WriteHeader(status)
				_, _ = w.Write([]byte(`{"error":{"code":"too-many-requests"}}`))
				return
			}
			var rows []string
			for _, sym := range strings.Split(r.URL.Query().Get("symbols"), ",") {
				if rec, ok := s.records[sym]; ok {
					rows = append(rows, rec)
				}
			}
			_, _ = w.Write([]byte(`{"result":[` + strings.Join(rows, ",") + `]}`))
		default:
			t.Errorf("unexpected call %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

func stockInfo(symbol, securityType string) string {
	st := ""
	if securityType != "" {
		st = `"securityType":"` + securityType + `",`
	}
	return `{"symbol":"` + symbol + `","name":"n","englishName":"e","isinCode":"KR0","market":"KOSPI",` + st +
		`"isCommonShare":true,"status":"ACTIVE","currency":"KRW","sharesOutstanding":"1"}`
}

func newStocksClient(t *testing.T, s *stocksServer) *official.Client {
	server := httptest.NewServer(s.handler(t))
	t.Cleanup(server.Close)
	return official.New(official.Credentials{APIKey: "k", SecretKey: "s"}, filepath.Join(t.TempDir(), "token.json"),
		official.WithBaseURL(server.URL))
}

// STOCK·ETF·ETN 은 원문 그대로, securityType 키 누락·응답 누락은 null(미분류). 다음 실행은 새 종목만 묻고 앞 값은 그대로.
func TestUpdateTypesRecordsRawEnumAndMissing(t *testing.T) {
	s := &stocksServer{records: map[string]string{
		"005930": stockInfo("005930", "STOCK"), "069500": stockInfo("069500", "ETF"),
		"Q50001": stockInfo("Q50001", "ETN"), "000001": stockInfo("000001", ""), // 키 누락
		"000777": stockInfo("000777", "REIT"),
	}}
	client := newStocksClient(t, s)
	path := filepath.Join(t.TempDir(), "d2", "security-types.json")
	run1, err := updateTypes(context.Background(), client, path, "KR", "2026-10-23",
		[]string{"000001", "005930", "069500", "123456", "Q50001"}, time.Now) // 123456 = 응답에서 빠짐
	if err != nil {
		t.Fatal(err)
	}
	if run1.Calls != 1 || !reflect.DeepEqual(run1.Unclassified, []string{"000001", "123456"}) ||
		!reflect.DeepEqual(run1.EmptyType, []string{"000001"}) || !reflect.DeepEqual(run1.NotInResponse, []string{"123456"}) {
		t.Fatalf("run1 %+v", run1)
	}
	rc, err := readTypes(path)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"005930": "STOCK", "069500": "ETF", "Q50001": "ETN"}
	for sym, v := range rc.Types {
		if w, ok := want[sym]; ok {
			if v == nil || *v != w {
				t.Errorf("%s = %v want %s", sym, v, w)
			}
		} else if v != nil {
			t.Errorf("%s = %v want null", sym, *v)
		}
	}
	if len(rc.Types) != 5 {
		t.Fatalf("types %v", rc.Types)
	}
	// 둘째 실행: 이미 있는 종목(null 포함)은 다시 묻지 않음 — 새 000777 만
	run2, err := updateTypes(context.Background(), client, path, "KR", "2026-10-26",
		[]string{"000001", "000777", "005930"}, time.Now)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(run2.Queried, []string{"000777"}) || s.asked[len(s.asked)-1] != "000777" {
		t.Fatalf("run2 %+v asked %v", run2, s.asked)
	}
	rc, _ = readTypes(path)
	if *rc.Types["000777"] != "REIT" || *rc.Types["005930"] != "STOCK" || len(rc.Runs) != 2 || rc.Runs[1].Date != "2026-10-26" {
		t.Fatalf("merged %+v", rc)
	}
	// 평가기(d2_load_receipts)가 읽는 모양: {"types": {sym: 문자열|null}}
	raw, _ := os.ReadFile(path)
	var shape struct {
		Types map[string]*string `json:"types"`
	}
	if err := json.Unmarshal(raw, &shape); err != nil || shape.Types["069500"] == nil || shape.Types["123456"] != nil {
		t.Fatalf("shape %v %s", err, raw)
	}
	// 모두 아는 종목이면 호출 0
	run3, err := updateTypes(context.Background(), client, path, "KR", "2026-10-27", []string{"005930"}, time.Now)
	if err != nil || run3.Calls != 0 {
		t.Fatalf("run3 %+v %v", run3, err)
	}
}

// 429 는 재시도 없이 멈추고 앞선 영수증을 바꾸지 않음.
func TestUpdateTypesStopsOn429WithoutTouchingFile(t *testing.T) {
	s := &stocksServer{records: map[string]string{"005930": stockInfo("005930", "STOCK")}}
	client := newStocksClient(t, s)
	path := filepath.Join(t.TempDir(), "security-types.json")
	if _, err := updateTypes(context.Background(), client, path, "KR", "2026-10-23", []string{"005930"}, time.Now); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(path)
	s.mu.Lock()
	s.status = http.StatusTooManyRequests
	n := len(s.asked)
	s.mu.Unlock()
	_, err := updateTypes(context.Background(), client, path, "KR", "2026-10-26", []string{"069500"}, time.Now)
	if !errors.Is(err, errRateLimited) {
		t.Fatalf("err %v", err)
	}
	after, _ := os.ReadFile(path)
	if string(before) != string(after) || len(s.asked) != n+1 {
		t.Fatalf("file changed or retried: asked %v", s.asked[n:])
	}
}

// 200 개 넘으면 묶음을 나눔.
func TestCollectTypesBatches(t *testing.T) {
	s := &stocksServer{records: map[string]string{}}
	client := newStocksClient(t, s)
	var syms []string
	for i := 0; i < 201; i++ {
		syms = append(syms, "S"+string(rune('A'+i/26%26))+string(rune('A'+i%26))+"X")
	}
	_, run, err := collectTypes(context.Background(), client, syms, map[string]*string{}, time.Now)
	if err != nil || run.Calls != 2 || len(strings.Split(s.asked[0], ",")) != 200 || len(run.Unclassified) != 201 {
		t.Fatalf("run %+v err %v asked %d", run.Calls, err, len(s.asked))
	}
}

// 기존 영수증이 깨져 있으면 묻기 전에 멈추고 파일을 그대로 둠.
func TestUpdateTypesRefusesCorruptReceipt(t *testing.T) {
	s := &stocksServer{records: map[string]string{}}
	client := newStocksClient(t, s)
	path := filepath.Join(t.TempDir(), "security-types.json")
	if err := os.WriteFile(path, []byte("{broken"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := updateTypes(context.Background(), client, path, "KR", "2026-10-23", []string{"005930"}, time.Now); err == nil {
		t.Fatal("corrupt receipt must be an error")
	}
	if raw, _ := os.ReadFile(path); string(raw) != "{broken" || len(s.asked) != 0 {
		t.Fatalf("file %q asked %v", raw, s.asked)
	}
}
