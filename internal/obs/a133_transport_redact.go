package obs

// a133_transport_redact.go — 시스템 밖으로 나가는 알림에서 계좌 정보를 지우는 전송 경계(a133, 안전 불변식 8).
//
// 사용자 결정(2026-10-13): 계좌 정보는 로그에 남기지 않고, 외부 알림(ntfy)으로도 내보내지 않음.
// 이 함수는 유일한 생산 전송 수단 Ntfy.Publish 의 첫 문장에서만 불림 — 실시간 critical · 일반 등급 · 이관 · outbox 재전송 ·
// 콘솔 알림 시험이 모두 그 한 자리를 지나므로 경로마다 사본을 둘 필요가 없음(design D1 · D2).
// 규칙 자체는 a131 의 internal/redact 하나 — 여기서는 그 규칙을 부르기만 함(사본 금지).

import "github.com/JungHoonGhae/tossinvest-cli/internal/redact"

// redactForTransport 는 알림 하나를 계좌 규칙에 통과시킨 사본을 돌려줌.
//
// 바꾸는 칸: 사건 유형(Tags 헤더의 첫 원소) · 제목(Title 헤더) · 본문 · 태그 각 원소 — 요청 바이트가 되는 문자열 전부.
// 바꾸지 않는 칸: 등급 · 우선순위(숫자 판정 입력). 순수 함수라 오류 · 패닉 경로가 없어 Publish 의 반환(전달 판정 입력)을 바꾸지 않음.
// Tags 는 새 슬라이스에 씀 — 호출자가 넘긴 슬라이스를 고치지 않게.
func redactForTransport(msg Notification) Notification {
	msg.Type = EventType(redact.String(string(msg.Type)))
	msg.Title = redact.String(msg.Title)
	msg.Body = redact.String(msg.Body)
	if len(msg.Tags) > 0 {
		tags := make([]string, len(msg.Tags))
		for i, tag := range msg.Tags {
			tags[i] = redact.String(tag)
		}
		msg.Tags = tags
	}
	return msg
}
