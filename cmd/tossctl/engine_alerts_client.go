package main

// engine_alerts_client.go 는 엔진이 연 알림 소켓을 부르는 쪽이다.
//
// # 왜 여기에 있고 새 패키지가 아닌가
//
// 선례(`exitquarantine` + `positionpolicyrpc`)가 도메인 타입을 별도 패키지에 두는
// 이유는 소비자가 둘이기 때문이다 — 콘솔과 CLI. 여기 소비자는 `tossctl` 하나이고
// 그것은 이미 `internal/app/engine` 을 import 한다. 두 번째 소비자가 생기면 그때
// 뽑는다 (task 4.4 · YAGNI).
//
// 전송(dial)만 플랫폼에 따라 갈리고 — Unix 소켓이 없는 곳에서는 부를 수단 자체가
// 없다 — 요청 두 개는 갈리지 않는다. 그래서 dial 만 build tag 파일에 있다.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
)

// alertControlClient 는 이 호스트의 엔진 프로세스에 붙은 한 개의 연결이다.
//
// 토큰을 들고 있지만 **절대 출력하지 않는다** (안전 불변식 8). 아래 두 메서드가
// 만드는 오류 문자열에도 토큰은 안 들어간다 — 넣으면 운영자가 화면을 붙여 넣는
// 순간 그 토큰이 승인 권한과 함께 밖으로 나간다.
type alertControlClient struct {
	baseURL string
	token   string
	http    *http.Client
}

// Pending 은 밀린 critical 알림을 그대로 받아 온다.
func (c *alertControlClient) Pending(ctx context.Context) ([]engine.PendingAlert, error) {
	var rows []engine.PendingAlert
	if err := c.call(ctx, http.MethodGet, engine.AlertControlListPath, nil, &rows); err != nil {
		return nil, err
	}
	return rows, nil
}

// Acknowledge 는 운영자가 본 것을 기록한다.
//
// 운영자 이름은 **인자로만** 온다. 이 파일 어디에도 기본값이 없다: 원장의 「공백
// 아닌 이름」 검사가 audit trail 의 유일한 보호이고, 여기서 이름을 만들어 주면
// 그 검사는 통과하면서 누가 봤는지는 사라진다 (task 4.4 R5).
func (c *alertControlClient) Acknowledge(ctx context.Context,
	req engine.AcknowledgeRequest) (engine.AcknowledgeResult, error) {
	var result engine.AcknowledgeResult
	if err := c.call(ctx, http.MethodPost, engine.AlertControlAcknowledgePath, req, &result); err != nil {
		return engine.AcknowledgeResult{}, err
	}
	return result, nil
}

// alertControlError 는 엔진이 돌려준 거절이다.
//
// 엔진은 운영자 입력 실수(빈 이름)를 400 으로, 자기 고장을 500 으로 구분해서 준다.
// 그 구분을 여기서 뭉개면 운영자가 자기 오타를 엔진 고장으로 읽는다.
type alertControlError struct {
	status  int
	code    string
	message string
}

func (e *alertControlError) Error() string {
	if strings.TrimSpace(e.message) == "" {
		return fmt.Sprintf("engine alerts: HTTP %d", e.status)
	}
	return fmt.Sprintf("engine alerts: %s: %s", e.code, e.message)
}

func (c *alertControlClient) call(ctx context.Context, method, path string, input, output any) error {
	var body io.Reader
	if input != nil {
		encoded, err := json.Marshal(input)
		if err != nil {
			return fmt.Errorf("engine alerts: encoding the request: %w", err)
		}
		body = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	if input != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	response, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("engine alerts: calling the engine: %w", err)
	}
	defer response.Body.Close()
	limited := io.LimitReader(response.Body, 1<<20)
	if response.StatusCode/100 != 2 {
		var remote struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}
		_ = json.NewDecoder(limited).Decode(&remote)
		return &alertControlError{status: response.StatusCode, code: remote.Code,
			message: strings.TrimSpace(remote.Message)}
	}
	if output == nil {
		return nil
	}
	if err := json.NewDecoder(limited).Decode(output); err != nil {
		return fmt.Errorf("engine alerts: decoding the response: %w", err)
	}
	return nil
}
