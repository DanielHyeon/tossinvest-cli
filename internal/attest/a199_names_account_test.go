package attest_test

// a199_names_account_test.go — a199 D1a: 만료된 attestation 의 계좌 확인은 Verify 가 쓰는 **같은 비교**여야 함.
//
// Verify 는 첫 실패만 돌려주고 만료(B5)가 계좌(B6)보다 먼저라, 만료된 파일은 계좌가 달라도 ErrExpired 로만 보임.
// 그래서 인터록은 만료일 때 계좌를 따로 물음. 그 물음이 Verify 와 다른 규칙이면(예: 가림 꼴을 같다고 보는
// SameAccountMasked) 두 정본이 되므로, 새 메서드의 답이 Verify 의 계좌 판정과 언제나 같음을 표로 고정함.

import (
	"errors"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/attest"
)

func TestA199NamesAccountIsVerifysAccountComparison(t *testing.T) {
	now := time.Date(2026, 11, 24, 3, 0, 0, 0, time.UTC)
	cases := []struct {
		name     string
		attested string
		resolved string
		want     bool
	}{
		{"같은 숫자열", "12345678", "12345678", true},
		{"하이픈 표기만 다름", "123-45-678", "12345678", true},
		{"다른 계좌", "123-45-678", "999-99-999", false},
		{"가림 꼴은 같은 계좌가 아님(Verify 와 같음)", "****5678", "12345678", false},
		{"숫자가 없는 표기", "abc", "abc", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a := attest.Attestation{
				FormatVersion: attest.FormatVersion,
				AccountRef:    tc.attested,
				IssuedAt:      now.Add(-time.Hour),
				ExpiresAt:     now.Add(time.Hour),
				Endpoints:     []string{"GET /api/v1/accounts"},
			}
			if got := a.NamesAccount(tc.resolved); got != tc.want {
				t.Fatalf("NamesAccount(%q) on %q = %v, want %v", tc.resolved, tc.attested, got, tc.want)
			}
			// 같은 입력으로 Verify 의 계좌 판정(만료 전 · endpoint 충족)과 답이 같아야 함 — 두 정본 금지.
			verifyMismatch := errors.Is(a.Verify(now, tc.resolved, nil), attest.ErrAccountMismatch)
			if verifyMismatch == tc.want {
				t.Fatalf("Verify account verdict (mismatch=%v) disagrees with NamesAccount=%v", verifyMismatch, tc.want)
			}
		})
	}
}
