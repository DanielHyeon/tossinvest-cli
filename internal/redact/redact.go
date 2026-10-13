// Package redact 는 로그·기록 파일에서 계좌 정보를 지우는 **규칙 하나**임(a131, 안전 불변식 8).
//
// 사용자 결정(2026-10-12): 계좌 정보는 로그에 남기지 않음. 모든 싱크(obs 로거 · 엔진 출력 · 마지막 오류 줄 · audit ·
// soak/검증 기록 · 일회성 정리 명령)는 이 패키지의 함수만 부름 — 규칙을 자리마다 베끼면 새 자리가 어느 사본에도 안 걸림.
//
// 무엇을 계좌 정보로 보는지(design D0):
//   - 등록된 계좌 참조 원문(공식 계좌 조회가 등록 — official.Client.accountsLocked)
//   - 등록 참조의 숫자열 꼴(숫자 8자리 이상일 때만 — 대시 위치가 달라도 같은 번호)
//   - 마스크 꼴(`*******1234` — attest.Mask, 끝 자리가 남음)
//   - 계좌와 저엔트로피 입력만으로 만든 유도 식별자(`pos-`·`rec-`·`mode-`·`adj-` + 24 hex, `corr-` + 32 hex — 대입으로 되돌릴 수 있음)
package redact

import (
	"regexp"
	"sort"
	"strings"
	"sync"
)

// Placeholder 는 계좌 텍스트를 대신하는 고정 문자열임. 따옴표·역슬래시가 없어 JSON 안에서 바꿔도 줄이 깨지지 않음.
const Placeholder = "[account]"

// minDigitPattern 은 숫자열 규칙을 쓰는 최소 자릿수임. 짧은 참조(시험의 `123-45`)로 숫자열 규칙을 만들면
// 가격·수량 같은 무관한 숫자를 지워 로그가 거짓이 됨 — 실제 계좌는 11자리.
const minDigitPattern = 8

var (
	// maskPattern 은 attest.Mask 의 모양(별 셋 이상 + 숫자, 끝 자리 사이 대시 허용)임 — 끝 4자리에 구분자가 껴도 꼬리를 남기지 않음.
	maskPattern = regexp.MustCompile(`\*{3,}[0-9](?:-?[0-9])*`)
	// derivedIDPattern 은 소금 없는 저엔트로피 유도 식별자임(journal.PositionID · reconcileStateID · operatingModeID ·
	// adjustmentID 의 24 hex, correctionID 의 32 hex).
	// 앞뒤 경계는 명시적 그룹(RE2 에 lookaround 없음) — `\b` 는 `_` 를 단어 글자로 보아 `pos-<hex>_lane` 을 놓침(구현 리뷰 I-7).
	derivedIDPattern = regexp.MustCompile(`(^|[^A-Za-z0-9])(?:(pos|rec|mode|adj)-[0-9a-f]{24}|(corr)-[0-9a-f]{32})([^0-9a-f]|$)`)
	// digitRun 은 대시로만 끊긴 숫자 덩어리임 — 덩어리의 숫자만 모은 값이 등록 번호와 같으면 그 덩어리 전체를 지움.
	digitRun = regexp.MustCompile(`[0-9]+(?:-[0-9]+)*`)
)

// registry 는 프로세스가 본 계좌 참조의 집합임. 줄이 쓰이는 **시점**의 집합으로 치환하므로, 등록보다 먼저 묶인 값도 지워짐.
type registry struct {
	mu     sync.RWMutex
	exact  map[string]bool
	digits map[string]bool
	sorted []string // 긴 것부터 — 긴 참조 안의 짧은 참조가 먼저 바뀌어 긴 것이 남는 일을 막음
}

var reg = &registry{exact: map[string]bool{}, digits: map[string]bool{}}

// Register 는 계좌 참조를 치환 대상으로 더함. 빈 값은 무시함. 메모리 쓰기 하나뿐이라 호출자의 반환·오류를 바꾸지 않음.
func Register(refs ...string) {
	reg.mu.Lock()
	defer reg.mu.Unlock()
	changed := false
	for _, ref := range refs {
		ref = strings.TrimSpace(ref)
		if ref == "" || reg.exact[ref] {
			continue
		}
		reg.exact[ref] = true
		changed = true
		if d := onlyDigits(ref); len(d) >= minDigitPattern {
			reg.digits[d] = true
		}
	}
	if changed {
		reg.sorted = reg.sorted[:0]
		for ref := range reg.exact {
			reg.sorted = append(reg.sorted, ref)
		}
		sort.Slice(reg.sorted, func(i, j int) bool {
			if len(reg.sorted[i]) != len(reg.sorted[j]) {
				return len(reg.sorted[i]) > len(reg.sorted[j])
			}
			return reg.sorted[i] < reg.sorted[j]
		})
	}
}

// String 은 문자열 안의 계좌 정보를 모두 치환함. 순서: 등록 원문 → 등록 숫자열 꼴 → 마스크 꼴 → 유도 식별자.
func String(s string) string {
	if s == "" {
		return s
	}
	reg.mu.RLock()
	for _, ref := range reg.sorted {
		if strings.Contains(s, ref) {
			s = strings.ReplaceAll(s, ref, Placeholder)
		}
	}
	if len(reg.digits) > 0 {
		s = digitRun.ReplaceAllStringFunc(s, func(run string) string {
			if reg.digits[onlyDigits(run)] {
				return Placeholder
			}
			return run
		})
	}
	reg.mu.RUnlock()
	s = maskPattern.ReplaceAllString(s, Placeholder)
	return derivedIDPattern.ReplaceAllString(s, "${1}${2}${3}-[redacted]${4}")
}

// IsAccountKey 는 구조화 필드 이름이 계좌를 싣는 필드인지 판정함. 대소문자·밑줄을 무시함.
func IsAccountKey(key string) bool {
	switch strings.ReplaceAll(strings.ToLower(strings.TrimSpace(key)), "_", "") {
	case "account", "accountref", "accountrefs", "accountno", "accountnumber", "accountseq", "maskedaccount":
		return true
	}
	return false
}

func onlyDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
