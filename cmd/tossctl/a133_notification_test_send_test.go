package main

// a133_notification_test_send_test.go — P7(콘솔 「알림 테스트」, a133 design D1): 직접 만든 *obs.Ntfy 도 같은 경계를 지나고,
// 요청 URL 은 설정의 서버 + 토픽뿐이며 고정 문구는 바뀌지 않음(경계가 계좌 아닌 내용을 건드리지 않음).

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
)

func TestA133TheNotificationTestCarriesOnlyItsText(t *testing.T) {
	var path, query, body, title string
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		hits++
		path, query, body, title = r.URL.Path, r.URL.RawQuery, string(b), r.Header.Get("Title")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	if err := publishNotificationTest(context.Background(),
		config.Notifications{Enabled: true, BaseURL: srv.URL, Topic: "tossos-a133cli"}, ""); err != nil {
		t.Fatalf("publishNotificationTest: %v", err)
	}
	if hits != 1 || path != "/tossos-a133cli" || query != "" {
		t.Fatalf("hits=%d path=%q query=%q — the request must carry only the configured topic", hits, path, query)
	}
	if title != "TossOS 알림 테스트" || !strings.HasPrefix(body, "이 메시지가 보이면 critical 알림이 이 채널로 도착한다.") {
		t.Fatalf("the fixed test text changed: title=%q body=%q", title, body)
	}
}
