package positionpolicyrpc

// risk_relaxation_client.go 는 a066 5.5 운영자 해제 두 route 의 tossctl 쪽임.
//
// 같은 Dial 된 Client(같은 endpoint · 토큰 · descriptor 검증)를 타고, 자기 요청 도우미를 둠 — 이유는 a079 격리 해제와
// 같음: 오류 어휘. Client.call 은 원격 코드를 positionpolicy 오류로 옮기는데 "stale 해제"에는 정직한 positionpolicy
// 대응이 없음. 경로 문자열은 engine 패키지를 import 하지 않기 위해 여기 다시 적고, engine 쪽 시험이 두 값이 같음을
// 고정함.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskrelaxation"
)

const (
	RiskRelaxationEntryLockPath = "/v1/risk-relaxation/entry-lock-release"
	RiskRelaxationLatchPath     = "/v1/risk-relaxation/risk-latch-release"
)

func (c *Client) ReleaseEntryLossLock(ctx context.Context,
	req riskrelaxation.EntryLockReleaseRequest) (riskrelaxation.Result, error) {
	var result riskrelaxation.Result
	return result, c.callRiskRelaxation(ctx, RiskRelaxationEntryLockPath, req, &result)
}

func (c *Client) ReleaseRiskOverageLatch(ctx context.Context,
	req riskrelaxation.LatchReleaseRequest) (riskrelaxation.Result, error) {
	var result riskrelaxation.Result
	return result, c.callRiskRelaxation(ctx, RiskRelaxationLatchPath, req, &result)
}

func (c *Client) callRiskRelaxation(ctx context.Context, path string, input, output any) error {
	encoded, err := json.Marshal(input)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(encoded))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	response, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	limited := io.LimitReader(response.Body, 1<<20)
	if response.StatusCode/100 != 2 {
		var remote rpcError
		if err := json.NewDecoder(limited).Decode(&remote); err != nil {
			// 5.5 이전 엔진에는 이 route 가 없어 mux 의 기본 404 가 옴 — "제공하지 않음"으로 말함.
			if response.StatusCode == http.StatusNotFound {
				return riskrelaxation.ErrUnwired
			}
			return fmt.Errorf("risk relaxation control: HTTP %d", response.StatusCode)
		}
		return decodeRemoteRiskRelaxationError(remote)
	}
	if err := json.NewDecoder(limited).Decode(output); err != nil {
		// 응답을 못 읽었다고 해제가 안 된 것은 아님 — 호출자가 "결과 불명"으로 말해야 함.
		return fmt.Errorf("risk relaxation control: the release may have been recorded but its response could not be read: %w", err)
	}
	return nil
}

func decodeRemoteRiskRelaxationError(remote rpcError) error {
	base := map[string]error{
		"invalid":           riskrelaxation.ErrInvalidRequest,
		"stale":             riskrelaxation.ErrStale,
		"audit_unavailable": riskrelaxation.ErrAuditUnavailable,
		"unwired":           riskrelaxation.ErrUnwired,
	}[remote.Code]
	if base == nil {
		base = errors.New("risk relaxation control: remote failure")
	}
	return fmt.Errorf("%w: %s", base, strings.TrimSpace(remote.Message))
}
