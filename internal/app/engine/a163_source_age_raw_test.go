package engine

// a163 리뷰(2026-10-22) — 해독 불가 원문은 글자 경계에서 64 바이트 안으로 자르고 인쇄 불가 글자를 '?' 로 바꿈.
// 브로커 글자가 알림 본문에 줄을 끼워 넣거나 깨진 UTF-8 로 전송을 망치지 않게.

import (
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

func TestA163TheRawBrokerTimeIsBoundedPrintableAndValidUTF8(t *testing.T) {
	for name, raw := range map[string]string{
		"newline":        "2026-08-27\nFAKE: line",
		"control":        "2026\x00\x1b[31m",
		"multibyte edge": strings.Repeat("a", 63) + "가나다",
		"invalid utf8":   "2026\xff\xfe",
		"long ascii":     strings.Repeat("x", 100),
	} {
		got := sourceAgeRaw(raw)
		if len(got) > sourceAgeRawLimit || !utf8.ValidString(got) || strings.ContainsAny(got, "\n\r\x00\x1b") {
			t.Errorf("%s: sourceAgeRaw(%q) = %q — want ≤%d bytes, valid UTF-8, no control characters", name, raw, got, sourceAgeRawLimit)
		}
	}
	// 소수부가 긴(유효할 수도 있는) 거대 문자열은 해독하지 않고 해독 불가 — 판정과 무관한 보고 경로의 비용 상한.
	huge := "2026-03-30T10:00:00." + strings.Repeat("1", 100000) + "+09:00"
	if bucket, _, parsed := classifySourceAge(huge, time.Now()); bucket != sourceAgeUnparseable || parsed {
		t.Fatalf("a %d-byte timestamp classified %s (parsed=%v), want unparseable without parsing", len(huge), bucket, parsed)
	}
	if bucket, _, _ := classifySourceAge("2026-03-30T10:00:00.123456789+09:00", time.Date(2026, 3, 30, 1, 0, 1, 0, time.UTC)); bucket != sourceAgeLE15s {
		t.Fatalf("a normal nanosecond timestamp classified %s, want le_15s", bucket)
	}
	if got := sourceAgeRaw("garbage"); got != "garbage" {
		t.Fatalf("a short printable value must stay verbatim, got %q", got)
	}
}
