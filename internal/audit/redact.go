package audit

// redact.go — a131: 감사 줄의 계좌 텍스트 치환(design D3-a). Record 와 일회성 정리 명령이 같은 이 함수를 부름.

import "github.com/JungHoonGhae/tossinvest-cli/internal/redact"

// RedactEntry 는 감사 줄의 문자열 칸에서 계좌 정보만 치환함. 행위(Action)·시각(At)은 손대지 않고, 설정 이름·이전/이후 값·
// 상세·주체는 계좌 텍스트만 `[account]` 로 바뀜 — 「누가 언제 무엇을 무엇에서 무엇으로」의 추적성은 그대로(§0.5).
// 실측(2026-10-13): 감사 줄의 계좌는 상세에만 있었고 이전/이후 값에는 0 — RecordChange 의 중복 판정은 영향 없음.
func RedactEntry(e Entry) Entry {
	e.Subject = redact.String(e.Subject)
	e.Setting = redact.String(e.Setting)
	e.Old = redact.String(e.Old)
	e.New = redact.String(e.New)
	e.Detail = redact.String(e.Detail)
	return e
}
