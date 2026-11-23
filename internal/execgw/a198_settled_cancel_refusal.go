package execgw

// a198_settled_cancel_refusal.go 는 a198 D7 — 엔진 취소가 「이미 끝난 주문」을 이유로 거절된 답을 확정 거절로 닫음.
//
// 계약(openapi 의 /api/v1/orders/{orderId}/cancel 409)은 이미 체결 · 취소 · 거부된 주문의 취소를 409 와 `error.code`
// `already-filled` · `already-canceled` · `already-rejected` 로 답함. 409 는 확정 status 목록 밖이라 그대로 두면 취소 attempt 가
// IN_DOUBT 로 남고, 같은 종목의 미종결 가드가 손절까지 재시작 때까지 막음. 그 주문은 이미 종결돼 후계 주문이 없으므로
// 「취소가 성립하지 않았다」는 확정임 — 주문의 운명은 체결 감지가 원장에 적음.
//
// `already-modified` 는 넣지 않음 — 정정의 후계 주문은 원장이 모르므로 확정으로 닫으면 치우지 못한 주문 위에 손절이 나갈 수
// 있음. 판정 규칙은 refusal_code.go 와 같음(최상위 `code`·`error.code` 두 자리, 대소문자 무시 전체 일치, 두 자리가 다르면 모호).

import (
	"errors"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// settledOrderCancelCodes 는 「그 주문은 브로커에서 이미 종결됐다」를 뜻하는 취소 거절 code 임(소문자, 전체 일치).
var settledOrderCancelCodes = map[string]bool{
	"already-filled":   true,
	"already-canceled": true,
	"already-rejected": true,
}

// settledCancelRefusal 은 모호로 분류된 취소 결과가 위 code 를 모순 없이 실었으면 확정 거절로 바꾼 결과를 돌려줌.
// 사유 코드는 status 확정 거절과 같은 journal.ReasonDispatchRejected — 새 사유 코드를 만들지 않음.
func settledCancelRefusal(out journal.DispatchOutcome) (journal.DispatchOutcome, bool) {
	if out.Class != journal.DispatchAmbiguous {
		return out, false
	}
	var apiErr *official.APIError
	if !errors.As(out.Err, &apiErr) {
		return out, false
	}
	top, nested, ok := refusalCodes(apiErr.Body)
	if !ok || (top.present && nested.present && !strings.EqualFold(top.value, nested.value)) {
		return out, false
	}
	code := top
	if !code.present {
		code = nested
	}
	if !code.present || !settledOrderCancelCodes[strings.ToLower(code.value)] {
		return out, false
	}
	out.Class = journal.DispatchRejected
	out.ReasonCode = journal.ReasonDispatchRejected
	return out, true
}
