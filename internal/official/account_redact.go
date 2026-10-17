package official

// account_redact.go — a131: 공식 계좌 조회가 받은 계좌 번호를 로그 치환 규칙(internal/redact)에 등록함.

import "github.com/JungHoonGhae/tossinvest-cli/internal/redact"

// registerAccountNumbers 는 응답의 모든 계좌 번호를 등록함 — 엔진은 첫 계좌만 쓰지만 어느 계좌든 로그에 남으면 계좌 정보임.
func registerAccountNumbers(raw []apiAccount) {
	for _, a := range raw {
		redact.Register(a.AccountNo)
	}
}
