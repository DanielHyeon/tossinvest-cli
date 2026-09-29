# Branch Test Map: `Notifier.claimAndDeliver`

- Source: `internal/obs/notifier.go` (:282-355); **편집 뒤** 측정 — `analysis/harness/coverage-post-unit3.json`(연결 워크트리 `fbc6df5f`, `./internal/obs` 시험 99개를 하나씩). 편집 전 번들은 `analysis/pre-edit/unit3/`에 보존.
- 재번호: 없음(분기 일곱 그대로, 줄만 이동).

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 295:2 | claim 실패 → 잠금 안 무조건 래치 · 해제 후 반환 | `TestAClaimThatFailsAttemptsTheDurableBlock`, `TestAClaimThatFailsBlocksNewEntries` | 해당 없음(분기 불변) | 블록 295.16-308.19을 시험 4개가 실행, PASS |
| B2 | if at 308:3 | 로그 | `TestAClaimThatFailsAttemptsTheDurableBlock`, `TestAClaimThatFailsBlocksNewEntries` | 해당 없음(분기 불변) | 블록 308.19-310.4을 시험 3개가 실행, PASS |
| B3 | if at 311:3 | 래치 | `TestAClaimThatFailsAttemptsTheDurableBlock`, `TestAClaimThatFailsBlocksNewEntries` | 해당 없음(분기 불변) | 블록 311.20-314.4을 시험 3개가 실행, PASS |
| B4 | switch at 318:2 | claim 결과 분기 | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092ARecordDoesNotWaitForAnotherSendersTransport` | 해당 없음(분기 불변) | 블록 318.2-318.27을 시험 51개가 실행, PASS |
| B5 | case at 319:2 | 이미 정착 → 해제 후 반환 | `TestNotifierIsConcurrencySafe`, `TestOneConditionIsOneSend` | 해당 없음(분기 불변) | 블록 319.28-328.43을 시험 5개가 실행, PASS |
| B6 | case at 329:2 | 남의 임차 → 해제 후 `logClaimHeld`(INFO) | `TestAHeldRowIsNotWhispered`, `TestALeaseLineCarriesItsOwnName` | 변이 L11 · L12 | 블록 329.34-341.43을 시험 7개가 실행, PASS |
| B7 | if at 348:2 | lost → 판정 없음 | `TestA092AReleaseAfterTheEpochReadIsHonoured`, `TestA092AReleaseBeforeTheEpochReadRelatches` | 해당 없음(분기 불변) | 블록 348.10-353.3을 시험 7개가 실행, PASS |
