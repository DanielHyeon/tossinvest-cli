package redact_test

// a133_adjacent_ids_test.go — 유도 식별자 둘이 구분자 한 글자로 붙어 있어도 둘 다 가려짐(a133 구현 리뷰 I-1).
//
// 규칙의 뒤 경계 그룹이 구분자를 먹으면 다음 식별자의 앞 경계가 사라져 둘째가 새던 결함 — 알림(%v 로 펼친 목록)과 로그 모두에 해당.

import (
	"fmt"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/redact"
)

func TestA133AdjacentDerivedIDsAreAllRedacted(t *testing.T) {
	// 빈 값 · id 없는 문자열은 그대로(B1 · B7 첫 판 반환).
	if redact.String("") != "" || redact.String("symbol 005930") != "symbol 005930" {
		t.Fatal("a string without account text must come back unchanged")
	}
	const (
		p1 = "pos-00112233445566778899aabb"
		p2 = "pos-ffeeddccbbaa998877665544"
		m1 = "mode-0123456789abcdef01234567"
		c1 = "corr-0123456789abcdef0123456789abcdef"
	)
	for _, in := range []string{
		p1 + " " + p2,
		p1 + "," + m1,
		p1 + "|" + c1 + "|" + p2,
		fmt.Sprintf("%v", []string{p1, p2, m1}),
		p1 + " " + p2 + " " + m1 + " " + c1,
	} {
		out := redact.String(in)
		for _, id := range []string{p1, p2, m1, c1} {
			if strings.Contains(out, id) {
				t.Fatalf("redact.String(%q) = %q — %s survived", in, out, id)
			}
		}
		if strings.Count(out, "-[redacted]") != strings.Count(in, "-") {
			t.Fatalf("redact.String(%q) = %q — every id must become <prefix>-[redacted] and nothing else", in, out)
		}
	}
}
