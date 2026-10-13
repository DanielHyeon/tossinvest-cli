# Branch Test Map: `Ntfy.Publish`

- Source: `internal/obs/ntfy.go` (85-139, base) — 편집 전. 편집 뒤 좌표는 1.3 에서 갱신.

| Branch | Anchor | Scenario | Test | a133 |
|---|---|---|---|---|
| B1 | if at 87:2 | 토픽 없음 → ErrNtfyNotConfigured | `TestNtfyWithoutATopicIsNotConfigured`(기존) | 무편집 |
| B2 | if at 91:2 | 서버 비면 ntfy.sh | 기존 패키지 시험 | 무편집 |
| B3 | if at 96:2 | 시한 기본값 | 기존 패키지 시험 | 무편집 |
| B4 | if at 104:2 | 요청 구성 실패 | 기존 패키지 시험 | 무편집 |
| B5 | if at 110:2 | Title 헤더 | `TestA133NtfyBoundaryScrubsEveryByte`(신설) · `TestNtfyPublishesToTheTopic`(기존) | 편집 영향 |
| B6 | if at 117:2 | 토큰 헤더 | `TestNtfySendsTheBearerToken`(기존) | 무편집 |
| B7 | if at 122:2 | 기본 클라이언트 | 기존 패키지 시험 | 무편집 |
| B8 | if at 126:2 | 전송 실패 | 기존 패키지 시험 | 무편집 |
| B9 | if at 134:2 | 2xx 아님 → 오류 | `TestNtfyReportsARefusal`(기존) · `TestA133DeliveryJudgementIsUnchanged`(신설) | 무편집 |
