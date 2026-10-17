# Branch Test Map: `Notifier.claimAndDeliver`

- Source: `internal/obs/notifier.go` (:288-361); **편집 뒤** 측정 — `analysis/harness/coverage-post-r25fix.json`(연결 워크트리 `55963f29`, `./internal/obs` 시험 99개를 하나씩). 편집 전 번들은 `analysis/pre-edit/unit3/`에 보존.
- 재번호: 없음(분기 일곱 그대로, 줄만 이동).
- 25라운드 수리(`55963f29`) 재번호: 새 B24(반납 행 없음 갈래) · B25(그 조건부 차단), 옛 B24(반납 선점) → B26, 옛 B25(소진 로그) → B27.

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 301:2 | claim 실패 → 잠금 안 무조건 래치 | `TestAClaimThatFailsAttemptsTheDurableBlock`, `TestAClaimThatFailsBlocksNewEntries` | 해당 없음(분기 불변) | 블록 298.16-311.19을 시험 4개가 실행, PASS |
| B2 | if at 314:3 | 로그 | `TestAClaimThatFailsAttemptsTheDurableBlock`, `TestAClaimThatFailsBlocksNewEntries` | 해당 없음(분기 불변) | 블록 311.19-313.4을 시험 3개가 실행, PASS |
| B3 | if at 317:3 | 래치 | `TestAClaimThatFailsAttemptsTheDurableBlock`, `TestAClaimThatFailsBlocksNewEntries` | 해당 없음(분기 불변) | 블록 314.20-317.4을 시험 3개가 실행, PASS |
| B4 | switch at 324:2 | claim 결과 분기 | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092ARecordDoesNotWaitForAnotherSendersTransport` | 해당 없음(분기 불변) | 블록 321.2-321.27을 시험 51개가 실행, PASS |
| B5 | case at 325:2 | 이미 정착 | `TestNotifierIsConcurrencySafe`, `TestOneConditionIsOneSend` | 해당 없음(분기 불변) | 블록 322.28-331.43을 시험 5개가 실행, PASS |
| B6 | case at 335:2 | 남의 임차 → logClaimHeld(INFO) | `TestAHeldRowIsNotWhispered`, `TestALeaseLineCarriesItsOwnName` | 변이 L11 · L12 | 블록 332.34-344.43을 시험 7개가 실행, PASS |
| B7 | if at 354:2 | lost → 판정 없음 | `TestA092AReleaseAfterTheEpochReadIsHonoured`, `TestA092AReleaseBeforeTheEpochReadRelatches` | 해당 없음(분기 불변) | 블록 351.10-356.3을 시험 7개가 실행, PASS |
