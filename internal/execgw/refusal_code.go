package execgw

// refusal_code.go 는 a094 R1 의 브로커 거절 code 분류기임.
//
// # 왜 있나
//
// 브로커는 계약(openapi 422)과 다른 status(409)로 `opposite-pending-order-exists` 를 돌려줬고, 상태 코드 표는 409 를
// 모호로 분류해 확정 거절이 무기한 IN_DOUBT 동결로 뒤집혔음(2026-08-07 · 세 건). 계약과 실물이 일치한 필드는 본문의
// `code` 하나뿐이었음 — 그래서 확정 거절 판정을 **status 가 아니라 code 필드 값**에 검.
//
// # 무엇을 지키나 (정본 order-execution 「IN_DOUBT 해소」 a094 조항)
//
//   - 본문 통짜의 부분문자열이 아니라 JSON 의 `code`(최상위)와 `error.code` 두 자리만 읽음 — message 문구로 걸지 않음.
//   - 값 비교는 대소문자 무시 + 전체 일치(`-v2` 같은 접미 값은 잡지 않음).
//   - 두 자리가 모두 있고 값이 다르면 **모호 강제** — 뒤의 분류(문구 분류 · status 422 확정 거절)를 타지 않음.
//   - JSON 아님 · code 없음 · 목록 밖 · 읽을 수 없는 code 값 → **판정 없음**(종전 경로). 모르는 것을 확정으로 바꾸지 않음.
//   - 재생 응답에는 적용하지 않음 — 이 분류기는 classifyMutation 한 곳에서만 부름(classifyReplay 는 부르지 않음, 구조 시험 고정).

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// ReasonOppositePendingOrder 는 브로커가 반대 방향 미체결 주문을 이유로 요청 자체를 거절했음을 뜻함(a094 R1).
// 그 상태에서 주문은 접수되지 않으므로 확정 거절(FAILED_CONFIRMED)임.
const ReasonOppositePendingOrder ReasonCode = "opposite_pending_order_exists"

// refusalCodeVerdict 는 code 분류기의 세 결과임(a094 D−3.5). 거절 본문에서 「확정 성공」은 나오지 않음.
type refusalCodeVerdict int

const (
	// refusalCodeNone 은 판정 없음 — 종전 경로(ClassifyBrokerRefusal → status)로 감.
	refusalCodeNone refusalCodeVerdict = iota
	// refusalCodeDefinitive 는 확정 거절 목록의 code 가 모순 없이 실렸음.
	refusalCodeDefinitive
	// refusalCodeContradictory 는 두 자리의 code 가 서로 다름 — 모호로 강제함.
	refusalCodeContradictory
)

// definitiveRefusalCodes 는 「요청 자체가 거절돼 주문이 접수되지 않음」을 뜻하는 브로커 code 목록임.
// 키는 소문자로 적고 비교는 대소문자 무시 전체 일치임. 목록을 넓히는 것은 확정 거절을 넓히는 것이므로
// code 마다 계약·실물 근거가 있어야 함(`request-in-progress` 는 원 요청이 실행됐을 수 있어 절대 넣지 않음).
var definitiveRefusalCodes = map[string]ReasonCode{
	"opposite-pending-order-exists": ReasonOppositePendingOrder,
}

// classifyRefusalCode 는 브로커 오류 응답 본문의 code 필드로 거절을 분류함.
//
// 읽는 본문은 공식 클라이언트 오류(official.APIError)의 응답 본문뿐임 — 우리가 만든 오류 문자열은 읽지 않음.
func classifyRefusalCode(err error) (ReasonCode, refusalCodeVerdict) {
	var apiErr *official.APIError
	if !errors.As(err, &apiErr) {
		return "", refusalCodeNone
	}
	top, nested, ok := refusalCodes(apiErr.Body)
	if !ok {
		return "", refusalCodeNone
	}
	// 두 자리가 모두 있고 값이 다르면 어느 쪽이 브로커의 뜻인지 말할 수 없음 — 목록 안팎과 무관하게 모호로 강제함.
	if top.present && nested.present && !strings.EqualFold(top.value, nested.value) {
		return "", refusalCodeContradictory
	}
	code := top
	if !code.present {
		code = nested
	}
	if !code.present {
		return "", refusalCodeNone
	}
	reason, listed := definitiveRefusalCodes[strings.ToLower(code.value)]
	if !listed {
		return "", refusalCodeNone
	}
	return reason, refusalCodeDefinitive
}

// codeField 는 본문 한 자리의 code 값임. present 는 비어 있지 않은 문자열이 있었다는 뜻임.
type codeField struct {
	value   string
	present bool
}

// refusalCodes 는 본문에서 최상위 `code` 와 `error.code` 를 읽음.
//
// ok=false 는 JSON 객체가 아닌 본문임 — 호출자는 판정 없음으로 다룸. null 과 빈 문자열(공백만 포함)은 「없음」임. 문자열이
// 아닌 code 값은 「있으나 어느 문자열과도 같지 않은 값」 으로 읽음 — 다른 자리의 값과 함께 있으면 모순(모호 강제)이고, 혼자면
// 목록 밖(판정 없음)임. 그래서 읽을 수 없는 값을 곁에 둔 확정 거절 code 는 확정이 되지 않음.
func refusalCodes(body string) (top, nested codeField, ok bool) {
	var root map[string]json.RawMessage
	if err := json.Unmarshal([]byte(body), &root); err != nil || root == nil {
		return codeField{}, codeField{}, false
	}
	top = readCode(root["code"])
	if raw, has := root["error"]; has && !isJSONNull(raw) {
		var inner map[string]json.RawMessage
		if err := json.Unmarshal(raw, &inner); err != nil || inner == nil {
			// `error` 가 객체가 아니면 그 아래 code 는 없음 — 최상위만으로 판정함.
			return top, codeField{}, true
		}
		nested = readCode(inner["code"])
	}
	return top, nested, true
}

// readCode 는 code 자리 하나를 읽음. 자리 없음 · null · 빈 문자열은 없음. 문자열이 아닌 값은 있음으로 읽되 값에 표식을 붙여
// 어느 문자열 code 와도 같지 않게 함(대소문자 무시 비교에서도).
func readCode(raw json.RawMessage) codeField {
	if len(raw) == 0 || isJSONNull(raw) {
		return codeField{}
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return codeField{value: unreadableCodeMark + string(raw), present: true}
	}
	if strings.TrimSpace(s) == "" {
		return codeField{}
	}
	return codeField{value: s, present: true}
}

// unreadableCodeMark 는 문자열이 아닌 code 값 앞에 붙이는 표식임 — 디코드된 JSON 문자열 code 가 이 접두로 시작할 수는 있어도
// 목록(definitiveRefusalCodes)의 값과는 같지 않고, 비교 상대가 문자열이면 원문 JSON 이 따옴표 없이 붙으므로 같아질 수 없음.
const unreadableCodeMark = "\x00non-string:"

func isJSONNull(raw json.RawMessage) bool {
	return strings.TrimSpace(string(raw)) == "null"
}
