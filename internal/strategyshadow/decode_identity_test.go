package strategyshadow

import (
	"encoding/json"
	"errors"
	"io"
	"testing"
)

// a112 0.5 리뷰 유지#4: 활성화 사본과 같은 규칙 — 해석 거절의 사슬에는 sentinel 하나만 실림.
// 두 사본이 한 로트 안에서 갈렸던 것(%w 대 %v)을 양쪽 시험으로 다시 못 갈리게 고정함.
func TestShadowDecodeFaultCarriesOnlyTheUnavailableSentinel(t *testing.T) {
	for _, data := range []string{"{", "{,}", `{"market":1}`, `{"no_such_key":true}`} {
		_, err := decodeProductionFamilyShadow([]byte(data))
		if err == nil {
			t.Fatalf("decode accepted %q", data)
		}
		if !errors.Is(err, ErrProductionFamilyShadowUnavailable) {
			t.Fatalf("%q: refusal lost the Unavailable sentinel: %v", data, err)
		}
		var syntax *json.SyntaxError
		var typeErr *json.UnmarshalTypeError
		if errors.Is(err, io.ErrUnexpectedEOF) || errors.As(err, &syntax) || errors.As(err, &typeErr) {
			t.Fatalf("%q: refusal carries a json error as a second identity: %v", data, err)
		}
		if _, multi := err.(interface{ Unwrap() []error }); multi {
			t.Fatalf("%q: refusal is a multi-wrap error: %v", data, err)
		}
	}
}
