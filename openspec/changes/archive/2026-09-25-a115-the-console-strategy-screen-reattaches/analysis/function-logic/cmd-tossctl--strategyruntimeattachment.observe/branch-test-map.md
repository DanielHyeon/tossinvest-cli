# Branch Test Map: `strategyRuntimeAttachment.observe`

인용 전용 — 기존 a109 시험 · a115 시험이 덮는다(편집 없음).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 요청자 취소 — 판정 없음(F1) | `TestACancelledRequestDoesNotDetachAHealthyClient` | no — 무편집(a109 원장이 변이로 잰다) | yes |
| B2 | 옛 자리의 소식 무시(F2) — 회복 직후 탈착 방지 | `TestALateReadFailureDoesNotUnseatTheNewAttachment` | no — 무편집(a109 원장이 변이로 잰다) | yes |
| B3 | 성공 — failed=false·attached 복원(G3) | `TestARecoveredReadIsAnAttachmentAgain` · `TestTheAttachmentReportsOnlyTransitions` | no — 무편집(a109 원장이 변이로 잰다) | yes |
| B4 | (B3 안) 복원이 전이면 부착 로그 1회 | `TestTheAttachmentReportsOnlyTransitions` | no — 무편집(a109 원장이 변이로 잰다) | yes |
| B5 | 실패 전이면 탈착 로그 1회 — 문구 「데몬은 그대로 돈다」(issues R1) | `TestTheAttachmentReportsOnlyTransitions` | no — 무편집(a109 원장이 변이로 잰다) | yes |
