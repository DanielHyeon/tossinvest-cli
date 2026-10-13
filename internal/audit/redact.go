package audit

// redact.go — a131: 감사 줄의 계좌 텍스트 치환(design D3-a). Record 와 일회성 정리 명령이 같은 이 함수를 부름.

import "github.com/JungHoonGhae/tossinvest-cli/internal/redact"

// RedactEntry 는 감사 줄의 문자열 칸에서 계좌 정보만 치환함. 행위(Action)·시각(At)은 손대지 않고, 설정 이름·이전/이후 값·
// 상세·주체는 계좌 텍스트만 `[account]` 로 바뀜 — 「누가 언제 무엇을 무엇에서 무엇으로」의 추적성은 그대로(§0.5).
// 설정 이름에 계좌가 들어가는 생산 자리가 있음(journal 의 `operating_mode:<account>` · `entry_loss_lock:<account>/…`) — 그래서 상세만이
// 아니라 문자열 칸 전부(구현 리뷰 I-2). RecordChange 의 중복 판정은 저장된(치환된) New 와 새 값을 비교하므로, 값 자체가 계좌를 담으면
// 매번 다시 씀 — 실측 값에 계좌 0 이라 오늘은 무해(잔재로 명명).
func RedactEntry(e Entry) Entry {
	e.Subject = redact.String(e.Subject)
	e.Setting = redact.String(e.Setting)
	e.Old = redact.String(e.Old)
	e.New = redact.String(e.New)
	e.Detail = redact.String(e.Detail)
	return e
}
