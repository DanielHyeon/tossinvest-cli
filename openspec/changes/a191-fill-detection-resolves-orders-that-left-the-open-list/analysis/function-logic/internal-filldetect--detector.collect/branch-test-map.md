# Branch Test Map: `Detector.collect` (internal/filldetect)

RED 원문: `../../red/red-1-core-behavioural.txt`(base 에서 행동 실패 3), `../../red/red-2-new-api-compile.txt`(새 표면 컴파일 실패).
GREEN 원문: `../../green/green-1-filldetect.txt`(71 top-level PASS, rc 0). 변이: `../../mutation/mutation-run-1.txt`·`mutation-run-2-buildfix.txt`.
base 커버리지: `../../coverage/base-collect-blocks.txt`. B9~B12 의 시험 대응은 시험별 커버리지 실측 `../../coverage/per-test-b9-b12.txt`(추측 아님).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | OPEN 걷기 오류 | `TestRateLimitedPollIsClassifiedAsOutage` | 기존 | yes |
| B2 | 추적 집합 오류 | 없음 — base count=0 | — | — |
| B3 | 선택 계좌 빈 값 | 없음 — base count=0 | — | — |
| B4 | 추적 순회 | 전부 | 기존 | yes |
| B5 | 추적 키 오류 | 없음 — base count=0 | — | — |
| B6 | 키 중복 | `TestPollRejectsDuplicateTrackedCanonicalIdentity` | 기존 | yes |
| B7 | OPEN 행 순회 | 전부 | 기존 | yes |
| B8 | OPEN 행 파싱 실패 | `TestMalformedOrderPayloadFailsTheCycle` | 기존 | yes |
| B9 | 421 OPEN 행 정규 키 완전 → 참(블록 421.15) | `TestPollReadsTrackedOrdersThatLeftTheOpenList`, `TestPollWalksTheOpenListToTheLastPage` 외 다수(커버리지 실측) | 기존 | yes |
| B11 | 422 그 키가 추적됨 → 참: lineage·seen(블록 422.66); 거짓: 추적 안 된 OPEN 행 | 참: `TestPollSeparatesReusedOrderIDsByCanonicalScope`·`TestA191OrderStillInTheOpenListIsNeverLookedUp`; 거짓: `TestPollWalksTheOpenListToTheLastPage`(421 블록만 실행, 422 본문 0) | 기존 | yes |
| B10 | 426 else(정규 키 불완전) 진입(블록 426.9) | `TestPollRejectsIncompleteOpenScopeForATrackedOrder`, `TestA191IncompleteOpenOrderNotTrackedIsStillApplied` | 기존 | yes |
| B12 | 426 불완전 키인데 같은 id 추적 → 참: 주기 실패 return@427(블록 426.48); 거짓: 그대로 적용 | 참: `TestPollRejectsIncompleteOpenScopeForATrackedOrder`; 거짓: `TestA191IncompleteOpenOrderNotTrackedIsStillApplied`(새 — 기존 시험이 밟지 않던 갈래, 변이 m33 `> 0`→`>= 0` CAUGHT, `../../mutation/mutation-run-5-b12.txt`) | 거짓 갈래: 관측 안 함(기존 동작을 고정하는 시험 — base 에서도 통과할 시험) | yes |
| B13 | 추적 키 순회 | 전부 | 기존 | yes |
| B14 | OPEN 에서 봄 | `TestA191OrderStillInTheOpenListIsNeverLookedUp` | 컴파일 RED | yes(m3 CAUGHT) |
| B15 | 단건 오류 | 없음 — base count=0 | — | — |
| B16 | 단건 파싱 실패 | 없음 — base count=0 | — | — |
| B17 | 응답 id 빈 값 | `TestBlankPayloadIdFallsBackToTheTrackedId` | 기존 | yes |
| B18 | 응답 범위 불완전 | 없음 — base count=0 | — | — |
| B19 | 응답 범위 불일치 | `TestA191ExistingTrackedReadGuardsRunBeforeTheLookup` | 컴파일 RED | yes(m17 CAUGHT) |
| call@467 | 후보 등록 | `TestA191OrderThatLeftTheOpenListAndIsCanceledInTheClosedListIsRecordedTerminal` | **행동 RED**(OPEN_UNFILLED 로 남음) | yes(m1 CAUGHT) |
| call@470 | 조회·교체 | `TestA191TerminalSnapshotLeavesTheReconcileLocalState`, `TestA191AcceptedClosedRowKeepsTheSingleReadExecution` | **행동 RED** | yes(m19 CAUGHT) |

B2·B3·B5·B15·B16·B18 의 count=0 은 a191 이전부터의 시험 부채다. a191 은 그 분기를 바꾸지 않아 범위로 끌어오지 않음(Manager 보고).
