# Branch Test Map: `ReconcileDriver.judgeHoldings`

> **편집 전 판(pre-edit) — tasks 1.5.** 커버(cover mode set, 1 = 실행됨): `../../implementation/base-cover.out`
> (`GOFLAGS=-trimpath go test ./internal/app/engine -count=1 -run 'TestAdoption|TestActiveForeignCause|TestUSIncluded|TestA095|TestInclude|TestAFailedInclude|TestAnExcluded|TestTheUnmanaged|TestACandidate|TestAStale|TestAnUnstable|TestReconcile' -coverprofile=…`,
> base `064f59fd9`, rc 0 — 원문 `../../implementation/base-cover-run.txt`, 파일은 adoption.go·reconcileloop.go 줄만 남김).
> 「Test」 칸은 grep·본문으로 고른 대표다 — set 모드는 선택 시험 전체의 합집합이라 블록을 그 시험 하나로 잰 것이 아니다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 스냅샷 신선도 기본값 | 하네스 기본(`SnapshotStaleness` 0) | 해당 없음(기존 분기) | yes — 블록 110.16 hit 1 |
| B2 | 보유 순회 | 모든 driver 시험 | 해당 없음 | yes — 119.44 hit 1 |
| B3 | market 빈 보유 → 기본 시장 | 선택 시험에 없음 | 해당 없음 | 미실행 — 121.19 hit 0 |
| B4 | 빈 종목·수량 0 건너뜀 | 선택 시험에 없음 | 해당 없음 | 미실행 — 125.71 hit 0 |
| B5 | 포지션 인스턴스 없음(fold 실패) | 선택 시험에 없음 | 해당 없음 | 미실행 — 130.17 hit 0 |
| B6 | CLOSED·수량 0 포지션 | 선택 시험에 없음 | 해당 없음 | 미실행 — 136.70 hit 0 |
| B7 | 이미 관리 중 | 수량 증가 시험(reconcileloop_test.go:844 부근) | 해당 없음 | yes — 140.23 hit 1 |
| B8 | 편입 포지션의 외부 증가 | 같은 시험 | 해당 없음 | yes — 143.19 hit 1 |
| B9 | 엔진이 연 포지션의 증가 | a095 수량 증가 시험 | 해당 없음 | yes — 145.10 hit 1 |
| B10 | RECONCILE 차단 → 무알림 continue(**a194 편집 지점**) | `TestAdoptionIsSilentUnderReconcile`(reconcileloop_test.go:358) · `TestActiveForeignCauseIsCountedAndBlocksBeforePriceRead`(:383) · `TestUSIncludedSymbolWaitsUnderAccountWidePermanentReconcile`(adoption_include_test.go:109) | 해당 없음 — a194 의 RED 는 3.1 | yes — 152.32 hit 1 |
| B11 | 스냅샷 낡음 → 무알림 | `TestA095AStaleSnapshotStaysSilent`(a095 …_internal_test.go:683) | 해당 없음 | yes — 155.13 hit 1 |
| B12 | 제외 목록 | `TestAnExcludedSymbolIsReportedNotAdopted`(reconcileloop_test.go:529) · `TestIncludeDoesNotBypassExclusion`(adoption_include_test.go:145) | 해당 없음 | yes — 163.39 hit 1 |
| B13 | 편입 꺼짐 ∧ 미지정 | `TestTheUnmanagedAlertSurvivesAdoptionBeingOff`(reconcileloop_test.go:494) · `TestA095OperatorChosenStatesStayNormal` | 해당 없음 | yes — 171.68 hit 1 |
| B14 | 후보 순회 | 편입 시험 | 해당 없음 | yes — 179.31 hit 1 |
| B15 | 편입 못 한 후보 → 무관리 | `TestACandidateWithNoQuoteIsDeferred`(reconcileloop_test.go:582) · a095 시도 실패 시험 | 해당 없음 | yes — 180.52 hit 1 |
| B16 | 무관리 보고 | 위 B12·B13·B15 시험 | 해당 없음 | yes — 188.30 hit 1 |

미실행 B3~B6 은 a194 가 바꾸지 않는 갈래다(편집 경계 밖). a194 의 새 동작(B10 안의 모으기와 순회 뒤 보고)은 3.1 의 새 시험이 RED→GREEN 으로 잰다.
