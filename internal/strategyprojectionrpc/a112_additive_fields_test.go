package strategyprojectionrpc

// a112 — spec "runtime lineage와 health는 lane 단위로 결정적으로 관측된다" 의
// `legacy projection reader` 시나리오:
//
//	WHEN  8-lane runtime response를 legacy market-level reader가 조회한다
//	THEN  기존 fields의 의미와 형식은 유지되고 additive lane/coordinator fields를
//	      무시해도 read가 실패하지 않는다
//
// 이 client 는 그 시나리오의 **legacy reader** 자리다. 엄격 디코딩을 켜 두면 엔진이
// 필드를 하나 더 실어 보내는 순간 화면 전체가 죽는다 — 시나리오가 금지하는 실패다.

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojection"
)

func TestClientIgnoresAdditiveFieldsFromANewerEngine(t *testing.T) {
	snapshot := strategyprojection.DormantSnapshot(time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC))
	raw, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	// 이 client 가 모르는 필드 두 개를 심는다: envelope 하나(미래의 `coordinators`)와
	// 시장 레코드 안 하나(미래의 `lanes`). 둘 다 spec 이 예고한 additive 필드다.
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatal(err)
	}
	envelope["coordinators"] = json.RawMessage(`[{"market":"KR"},{"market":"US"}]`)
	var markets map[string]map[string]json.RawMessage
	if err := json.Unmarshal(envelope["markets"], &markets); err != nil {
		t.Fatal(err)
	}
	markets["KR"]["lanes"] = json.RawMessage(`[{"family":"BREAKOUT_RETEST"}]`)
	remarshalled, err := json.Marshal(markets)
	if err != nil {
		t.Fatal(err)
	}
	envelope["markets"] = remarshalled
	body, err := json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}

	token := strings.Repeat("t", 40)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+token {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write(body)
	}))
	t.Cleanup(server.Close)

	client := &Client{baseURL: server.URL, token: token, http: server.Client()}
	got, err := client.Read(context.Background())
	if err != nil {
		t.Fatalf("구버전 reader 가 additive 필드에 죽었다: %v", err)
	}
	// 기존 필드의 의미와 형식은 그대로여야 한다.
	if got.SchemaVersion != strategyprojection.SchemaVersion || len(got.Markets) != 2 ||
		got.Markets[strategyprojection.MarketKR].Status != strategyprojection.StatusUnknown {
		t.Fatalf("기존 필드가 흔들렸다: %+v", got)
	}
}

// TestClientStillRejectsASemanticallyInvalidSnapshot 는 관용이 **판정을 버린 것이
// 아님**을 잡는다. 모르는 필드는 무시하되, 의미가 틀린 스냅샷은 여전히 거절한다.
func TestClientStillRejectsASemanticallyInvalidSnapshot(t *testing.T) {
	token := strings.Repeat("t", 40)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		// 시장이 하나뿐인 스냅샷 — Validate 가 envelope 에서 거절한다.
		_, _ = w.Write([]byte(`{"schemaVersion":"` + strategyprojection.SchemaVersion +
			`","generatedAt":"2026-08-29T12:00:00Z","runtime":{"configDigest":null,"buildDigest":null},` +
			`"markets":{"KR":{"market":"KR","status":"UNKNOWN"}},"future":1}`))
	}))
	t.Cleanup(server.Close)

	client := &Client{baseURL: server.URL, token: token, http: server.Client()}
	if _, err := client.Read(context.Background()); err == nil {
		t.Fatal("의미가 틀린 스냅샷이 통과했다 — 관용이 판정까지 삼켰다")
	}
}
