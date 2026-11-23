# Branch Test Map: `runEngineReconcileResolve` (pre-edit 판)

- Source: `cmd/tossctl/engine_reconcile.go` (107-214), base `064f59fd9`.
- 커버 측정: `../../coverage/pre-edit-runEngineReconcileResolve-and-cmd-blocks.txt`(선택 시험 `./cmd/tossctl -run 'TestA166EngineReconcile|TestEngineReconcileResolve'`, covermode=set, PASS). 시험 이름은 읽어서 그 분기를 단언함을 확인한 것만 적는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | confirm 거짓 | `TestEngineReconcileResolveRequiresExplicitConfirmationIdentityAndNote/confirmation` | 해당 없음(기존 분기) | yes — 110.34 hit=1 |
| B2 | operator 공백 | 같은 시험 `/operator` | 해당 없음(기존 분기) | yes — 114.20 hit=1 |
| B3 | note 공백 | 같은 시험 `/note` | 해당 없음(기존 분기) | yes — 118.16 hit=1 |
| B4 | deps 누락 | 없음 | 해당 없음(기존 분기) | 미실행 — 121.93 hit=0 |
| B5 | ctx nil | 없음 | 해당 없음(기존 분기) | 미실행 — 126.16 hit=0 |
| B6 | journalDir 오류 | 없음 | 해당 없음(기존 분기) | 미실행 — 130.16 hit=0 |
| B7 | 잠금 실패 | `TestEngineReconcileResolveAcquiresTheEngineLockBeforeAssembly` | 해당 없음(기존 분기) | yes — 134.16 hit=1 |
| B8 | build 오류 | 없음 | 해당 없음(기존 분기) | 미실행 — 140.16 hit=0 |
| B9 | 실패 경로 닫기 | `TestEngineReconcileResolveRefusesUnstableSnapshots`(closed 단언) | 해당 없음(기존 분기) | yes — 145.38 hit=1 |
| B10 | runtime 결선 검증 실패 | 없음 | 해당 없음(기존 분기) | 미실행 — 149.65 hit=0 |
| B11 | 1번 수집 실패 | 없음 | 해당 없음(기존 분기) | 미실행 — 158.16 hit=0 |
| B12 | 1·2번 사이 대기 실패 | 없음 | 해당 없음(기존 분기) | 미실행 — 162.80 hit=0 |
| B13 | 2번 수집 실패 | 없음 | 해당 없음(기존 분기) | 미실행 — 166.16 hit=0 |
| B14 | 2번 스냅숏 불안정 | `TestEngineReconcileResolveRefusesUnstableSnapshots` | 해당 없음(기존 분기) | yes — 170.20 hit=1 |
| B15 | 불안정 사유 빈 값 → 기본 문장 | `TestEngineReconcileResolveRefusesUnstableSnapshots`(블록 실행만 측정) | 해당 없음(기존 분기) | yes — 172.16 hit=1 |
| B16 | 2·3번 사이 대기 실패 | 없음 | 해당 없음(기존 분기) | 미실행 — 177.80 hit=0 |
| B17 | 3번 수집 실패 | 없음 | 해당 없음(기존 분기) | 미실행 — 181.16 hit=0 |
| B18 | 3번 스냅숏 변함 | `TestEngineReconcileResolveRefusesAFinalSnapshotThatChanged` | 해당 없음(기존 분기) | yes — 185.25 hit=1 |
| B19 | 변함 사유 빈 값 → 기본 문장 | 같은 시험(블록 실행만 측정) | 해당 없음(기존 분기) | yes — 187.16 hit=1 |
| B20 | 장부 재구성 실패 | 없음 | 해당 없음(기존 분기) | 미실행 — 194.16 hit=0 |
| B21 | 진입을 막는 차이 남음 | `TestEngineReconcileResolveRefusesABlockingCorrectedComparison` | 해당 없음(기존 분기) | yes — 198.24 hit=1 |
| B22 | validate 거절 | `TestEngineReconcileResolveRefusesNonQuantityCauses`(스텁 validate) | 해당 없음(기존 분기) | yes — 201.43 hit=1 |
| B23 | 해제 실패 | `TestEngineReconcileResolveDoesNotReportSuccessWhenPersistenceFails` | 해당 없음(기존 분기) | yes — 204.61 hit=1 |
| B24 | close 실패 | 없음 | 해당 없음(기존 분기) | 미실행 — 207.40 hit=0 |
| (성공 :212) | 성공 문장 | `TestEngineReconcileResolveRecordsDurableReleaseBeforeSuccess` · `TestEngineReconcileResolveIntegratesCorrectedJournalStateAndAtomicRelease`(`released` 포함 단언) | 문장 변경은 2단계 RED 대상 | yes — 210.2 hit=1 |
