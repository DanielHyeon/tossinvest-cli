# Branch Test Map: `strategyRuntimeAttachment.Read`

인용 전용 — 기존 a109 시험 · a115 시험이 덮는다(편집 없음).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 시도 대상(failed)이면 요청이 재부착을 깨운다(비차단) | `TestTheRequestPathNeverWaitsForADial` | no — 무편집(a109 원장이 변이로 잰다) | yes |
| B2 | 부재(reader nil) → 오류 — 부재를 스냅샷으로 짓지 않음(design 대안 2 기각의 근거) | `TestTheDaemonAttachesWhenTheEngineComesUpLater` | no — 무편집(a109 원장이 변이로 잰다) | yes |
| B3 | 읽기 실패 판정 → 즉시 wake(다음 요청을 기다리지 않음) | `TestTheRequestPathNeverWaitsForADial` · `TestTheDaemonReattachesAfterTheEngineRestarts` | no — 무편집(a109 원장이 변이로 잰다) | yes |
