package strategyrouter

import (
	"encoding/json"
	"errors"
	"io"
	"testing"
)

// a112 0.5 리뷰 유지#4 = 보안#2: 해석 거절의 오류 사슬에는 sentinel 하나만 실림.
// 둘째 %w 가 있으면 json 오류(io.ErrUnexpectedEOF · *json.SyntaxError …)가 두 번째 신원이 되어
// 호출자가 errors.Is/As 로 매니페스트 작성자가 만든 바이트의 결함 종류를 갈래 판정에 쓸 수 있게 됨.
// shadow 사본(strategyshadow decode)은 처음부터 %v 라 두 사본이 같은 규칙을 지키는지 함께 고정함.
func TestActivationDecodeFaultCarriesOnlyTheUnavailableSentinel(t *testing.T) {
	cases := []struct {
		name string
		data string
	}{
		{"truncated", "{"},       // 디코더가 io.ErrUnexpectedEOF 를 냄
		{"syntax", "{,}"},        // *json.SyntaxError
		{"type", `{"market":1}`}, // *json.UnmarshalTypeError
		{"unknown-key", `{"no_such_key":true}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := decodeProductionFamilyActivation([]byte(tc.data))
			if err == nil {
				t.Fatalf("decode accepted %q", tc.data)
			}
			if !errors.Is(err, ErrProductionFamilyActivationUnavailable) {
				t.Fatalf("decode refusal lost the Unavailable sentinel: %v", err)
			}
			// 사슬의 둘째 신원 금지 — json 쪽 오류는 문장으로만 남아야 함
			if errors.Is(err, io.ErrUnexpectedEOF) {
				t.Fatalf("decode refusal wraps io.ErrUnexpectedEOF as a second identity: %v", err)
			}
			var syntax *json.SyntaxError
			if errors.As(err, &syntax) {
				t.Fatalf("decode refusal wraps *json.SyntaxError as a second identity: %v", err)
			}
			var typeErr *json.UnmarshalTypeError
			if errors.As(err, &typeErr) {
				t.Fatalf("decode refusal wraps *json.UnmarshalTypeError as a second identity: %v", err)
			}
			// 사슬 전체가 단일 래핑인지 직접 셈 — Unwrap() []error 가 있으면 다중 %w
			if _, multi := err.(interface{ Unwrap() []error }); multi {
				t.Fatalf("decode refusal is a multi-wrap error (two %%w): %v", err)
			}
		})
	}
}
