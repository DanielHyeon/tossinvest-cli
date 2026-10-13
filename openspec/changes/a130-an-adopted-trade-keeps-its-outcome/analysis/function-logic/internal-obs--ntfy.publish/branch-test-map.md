# Branch Test Map: `Ntfy.Publish`

- Source: `internal/obs/ntfy.go` (85-142) — `727a841e` 기준(a133 `35936cc8` 편집 뒤, 그 뒤로 이 파일 무변경).
- 출처: a133 이 편집한 함수가 a130 창 안에 끼어서 요구됨 — 겹친 change. 시험은 a133 · 기존 시험을 인용함.
- 측정(2026-10-13, HEAD `272dba00`): 시험마다 따로
  `go test -count=1 -run '^<Test>$' -covermode=count -coverpkg=./internal/obs,./internal/redact -coverprofile=<f> ./<pkg>`
  (전부 rc 0). 칸의 `a/b` = 분기 문장을 담은 블록의 실행 횟수 / 분기 본문(같은 줄, 조건 뒤에서 시작하는 블록)의 실행 횟수.

| Branch | Anchor | Scenario | Test | 측정 |
|---|---|---|---|---|
| B1 | if at 90:2 | 토픽 없음 → `ErrNtfyNotConfigured` | `TestNtfyWithoutATopicIsNotConfigured` | 1/1 |
| B2 | if at 94:2 | 서버 비면 기본 ntfy 서버 | not-applicable — 참 갈래는 실 공개 ntfy 서버로 보내므로 자동 시험이 닿으면 안 됨(실 endpoint 금지). 거짓 갈래(지정 서버)는 `TestNtfyPublishesToTheTopic` | 참 0(어느 인용 시험도 본문 0) · 거짓 1/0 |
| B3 | if at 99:2 | 시한 0 → 기본 시한 | `TestNtfyPublishesToTheTopic` · `TestA133NtfyBoundaryScrubsEveryByte` | 1/1 · 2/2 |
| B4 | if at 107:2 | 요청 구성 실패 → 오류 | `TestA133PublishStillReportsBuildAndTransportFailures` | 2/1 |
| B5 | if at 113:2 | 치환된 제목이 Title 헤더로(참) · 제목 없음(거짓) | 참: `TestA133NtfyBoundaryScrubsEveryByte` · `TestNtfyPublishesToTheTopic`, 거짓: `TestNtfySendsTheBearerToken` | 2/2 · 1/1 · 1/0 |
| B6 | if at 120:2 | 토큰 → Authorization 헤더 | `TestNtfySendsTheBearerToken` | 1/1 |
| B7 | if at 125:2 | 클라이언트 nil → 기본 클라이언트 | `TestA133PublishStillReportsBuildAndTransportFailures` · `TestA133TheNotificationTestCarriesOnlyItsText`(`cmd/tossctl`) | 1/1 · 1/1 |
| B8 | if at 129:2 | 전송 실패 → 오류 | `TestA133PublishStillReportsBuildAndTransportFailures` · `TestA133NormalRelayLeavesNoAccount` | 1/1 · 1/1 |
| B9 | if at 137:2 | 2xx 아님 → 오류 | `TestNtfyReportsARefusal` · `TestA133DeliveryJudgementIsUnchanged` · `TestA133TheDelivererStillLatchesOnRefusal`(`internal/app/engine`) | 1/1 · 3/3 · 3/3 |

- a133 번들의 분기별 인용은 이 측정에서 전부 본문 실행 ≥ 1 로 확인됨(B2 제외 — a133 도 not-applicable).
- 미덮음 분기: B2 참 갈래 하나(위 사유).
