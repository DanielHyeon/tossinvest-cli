# Branch Test Map: `Ntfy.Publish`

- Source: `internal/obs/ntfy.go` (85-142) — 편집 뒤. 분기 9 개는 편집 전과 같은 ID · 같은 조건(좌표만 +3), 옛/새 ast.json 의 branches 를 순서대로 맞대어 확인.

| Branch | Anchor | Scenario | Test | a133 |
|---|---|---|---|---|
| B1 | if at 90:2 | 토픽 없음 → ErrNtfyNotConfigured | `TestNtfyWithoutATopicIsNotConfigured` | 무편집(앞에 치환 대입) |
| B2 | if at 94:2 | 서버 비면 ntfy.sh | 시험 없음 — 기본값은 실 ntfy.sh 라 자동 시험이 닿으면 안 됨(불변식: 실 endpoint 금지). `UsesPublicService` 는 `TestNtfyPublicServiceIsFlagged` | 무편집 |
| B3 | if at 99:2 | 시한 0 → 기본 | `TestNtfyPublishesToTheTopic`(Timeout 0) · `TestA133NtfyBoundaryScrubsEveryByte` | 무편집 |
| B4 | if at 107:2 | 요청 구성 실패 → 오류 | `TestA133PublishStillReportsBuildAndTransportFailures` | 무편집 — 본문은 치환된 값 |
| B5 | if at 113:2 | Title 헤더 — 치환된 제목 | `TestA133NtfyBoundaryScrubsEveryByte` · `TestA133CriticalFieldsLeaveNoAccount` · `TestNtfyPublishesToTheTopic` | 편집 영향 |
| B6 | if at 120:2 | 토큰 헤더 | `TestNtfySendsTheBearerToken` | 무편집 |
| B7 | if at 125:2 | 클라이언트 nil → 기본 클라이언트 | `TestA133TheNotificationTestCarriesOnlyItsText`(cmd/tossctl — HTTPClient 없음) · `TestA133PublishStillReportsBuildAndTransportFailures` | 무편집 |
| B8 | if at 129:2 | 전송 실패 → 오류 | `TestA133PublishStillReportsBuildAndTransportFailures` | 무편집 |
| B9 | if at 137:2 | 2xx 아님 → 오류 | `TestNtfyReportsARefusal` · `TestA133DeliveryJudgementIsUnchanged` · `TestA133TheDelivererStillLatchesOnRefusal` | 무편집 |

## a133 시험

- 경계: `TestA133NtfyBoundaryScrubsEveryByte` · `TestA133TheBoundaryCoversEveryTextField` · `TestA133TheBoundaryKeepsNonAccountNumbersAndTheCallersTags`
- 경로별(P1~P5 · P7): `TestA133BestEffortPathLeavesNoAccount` · `TestA133CriticalFieldsLeaveNoAccount` · `TestA133NormalRelayLeavesNoAccount` · `TestA133FlushRedeliveryLeavesNoAccount` ·
  `TestA133TheDelivererRedeliversWithoutTheAccount` · `TestA133TheNotificationTestCarriesOnlyItsText`
- 판정 불변: `TestA133DeliveryJudgementIsUnchanged` · `TestA133TheDelivererStillLatchesOnRefusal` · `TestA133PublishStillReportsBuildAndTransportFailures`
- 구조 핀: `TestA133TheOnlyTransportPassesTheBoundary`
