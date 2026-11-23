package soak_test

// a196 — 엔진 감시는 감독 증거의 기한을 attestation 파일 값만으로 「실린 증거 시각 + (만료 − 발급)」 으로 계산함.
// 그 식이 맞으려면 발급자가 감독 증거를 받아들이는 기간과 attestation 의 유효 기간이 같은 값이어야 함 — 여기서 그 전제를 고정함.
// 발급자가 두 값을 갈라 쓰기 시작하거나 경계(나이 == 유효 기간이면 제외)를 바꾸면 이 시험이 먼저 깨짐.

import (
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/soak"
)

// TestA196SupervisedEvidenceLeavesExactlyAtTheDeadline: 나이 V−1ns 는 실리고, 정확히 V 와 V+1ns 는 빠짐(V = 유효 기간).
func TestA196SupervisedEvidenceLeavesExactlyAtTheDeadline(t *testing.T) {
	now := soakStart.AddDate(0, 0, 3)
	validity := criteria().Validity
	cases := []struct {
		name string
		age  time.Duration
		kept bool
	}{
		{"나이 V−1ns 는 실림", validity - time.Nanosecond, true},
		{"정확히 V 는 빠짐(기한 정각은 이미 지난 것)", validity, false},
		{"V+1ns 는 빠짐", validity + time.Nanosecond, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a, err := soak.BuildAttestation(soak.Summarize(threeCleanDays()), criteria(), now, "tester", "",
				supervisedProofs(now.Add(-tc.age)))
			if err != nil {
				t.Fatalf("BuildAttestation: %v", err)
			}
			if got := a.ExpiresAt.Sub(a.IssuedAt); got != validity {
				t.Fatalf("ExpiresAt−IssuedAt = %s, want the acceptance window %s", got, validity)
			}
			kept := len(a.SupervisedBy) == len(soak.LiveOnlyEndpoints())
			if kept != tc.kept || (!kept && len(a.SupervisedBy) != 0) {
				t.Fatalf("SupervisedBy=%+v, want kept=%v", a.SupervisedBy, tc.kept)
			}
		})
	}
}

// TestA196TheFileDeadlineIsWhenTheIssuerStopsCarryingTheProof: 실린 파일로 계산한 기한 = 그 증거가 빠지기 시작하는 시각.
// 기한 1ns 전의 발급은 증거를 싣고, 기한 정각의 발급은 싣지 않음.
func TestA196TheFileDeadlineIsWhenTheIssuerStopsCarryingTheProof(t *testing.T) {
	now := soakStart.AddDate(0, 0, 3)
	// 기한이 발급 1시간 뒤가 되게 — soak 기록의 최대 나이(48시간) 안에서 기한 앞뒤의 발급을 둘 다 재현하려고.
	proofAt := now.Add(-criteria().Validity).Add(time.Hour)
	first, err := soak.BuildAttestation(soak.Summarize(threeCleanDays()), criteria(), now, "tester", "", supervisedProofs(proofAt))
	if err != nil || len(first.SupervisedBy) == 0 {
		t.Fatalf("first issue: err=%v supervised=%+v", err, first.SupervisedBy)
	}
	// 엔진 감시가 쓰는 식(파일 값만): 실린 증거 시각 + (만료 − 발급).
	deadline := first.SupervisedBy[0].At.Add(first.ExpiresAt.Sub(first.IssuedAt))

	for _, tc := range []struct {
		name string
		at   time.Time
		kept bool
	}{
		{"기한 1ns 전 발급", deadline.Add(-time.Nanosecond), true},
		{"기한 정각 발급", deadline, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a, err := soak.BuildAttestation(soak.Summarize(threeCleanDays()), criteria(), tc.at, "tester", "", supervisedProofs(proofAt))
			if err != nil {
				t.Fatalf("BuildAttestation at %s: %v", tc.at, err)
			}
			if kept := len(a.SupervisedBy) > 0; kept != tc.kept {
				t.Fatalf("issued at %s: supervised kept=%v, want %v", tc.at, kept, tc.kept)
			}
		})
	}
}

// TestA196ANonDefaultValidityMovesBothValues: `--validity` 로 기본값이 아닌 유효 기간을 줘도 만료와 증거 수용이 같은 값을 씀
// (구현 리뷰 P2-4 — 수용 기간을 30일로 박는 변이가 죽어야 함).
func TestA196ANonDefaultValidityMovesBothValues(t *testing.T) {
	now := soakStart.AddDate(0, 0, 3)
	c := criteria()
	c.Validity = 10 * 24 * time.Hour
	for _, tc := range []struct {
		name string
		age  time.Duration
		kept bool
	}{
		{"10일 − 1ns 는 실림", c.Validity - time.Nanosecond, true},
		{"정확히 10일은 빠짐", c.Validity, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a, err := soak.BuildAttestation(soak.Summarize(threeCleanDays()), c, now, "tester", "", supervisedProofs(now.Add(-tc.age)))
			if err != nil {
				t.Fatalf("BuildAttestation: %v", err)
			}
			if got := a.ExpiresAt.Sub(a.IssuedAt); got != c.Validity {
				t.Fatalf("ExpiresAt−IssuedAt = %s, want %s", got, c.Validity)
			}
			if kept := len(a.SupervisedBy) > 0; kept != tc.kept {
				t.Fatalf("supervised kept=%v, want %v", kept, tc.kept)
			}
		})
	}
}
