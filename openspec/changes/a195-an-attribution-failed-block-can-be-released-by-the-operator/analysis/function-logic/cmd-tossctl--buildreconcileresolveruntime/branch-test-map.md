# Branch Test Map: `buildReconcileResolveRuntime` (편집 뒤 판)

- Source: `cmd/tossctl/engine_reconcile.go` (239-314), 작업 트리. 재번호 `renumber.txt`(B1~B12 유지, 옛 B13·B14 삭제).
- 커버 측정: `../../coverage/after-engine_reconcile-blocks.txt` — `go test ./cmd/tossctl -run 'TestA166EngineReconcile|TestEngineReconcileResolve|A195' -count=1 -covermode=set`(PASS). 편집 전 측정 `../../coverage/pre-edit-buildReconcileResolveRuntime-blocks.txt`.
- 시험 이름은 읽어서 그 분기를 지남을 확인한 것만 적는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | loadConfig 실패 | 없음(grep) | 해당 없음(기존 분기) | 미실행 — 241.16 hit=0(편집 전도 0) |
| B2 | OpenAPI 경로 실패 | 없음(grep) | 해당 없음(기존 분기) | 미실행 — 245.16 hit=0(편집 전도 0) |
| B3 | 자격 읽기 실패 | 없음(grep) | 해당 없음(기존 분기) | 미실행 — 249.16 hit=0(편집 전도 0) |
| B4 | 자격 빈 값 | 없음(grep) | 해당 없음(기존 분기) | 미실행 — 252.103 hit=0(편집 전도 0) |
| B5 | 예산 required 실패·취소 컨텍스트 | `TestA166EngineReconcileLocalTLS` | 해당 없음(기존 분기) | yes — 257.16 hit=1 |
| B6 | 계좌 조회 실패 | 없음(grep) | 해당 없음(기존 분기) | 미실행 — 261.16 hit=0(편집 전도 0) |
| B7 | 계좌 신원 실패 | 없음(grep) | 해당 없음(기존 분기) | 미실행 — 265.16 hit=0(편집 전도 0) |
| B8 | configDir 있음 | `TestA166EngineReconcileLocalTLS` · `a195ProductionRuntime` 을 쓰는 A195 시험 | 해당 없음(기존 분기) | yes — 271.60 hit=1 |
| B9 | 장부 열기 실패 | 없음(grep) | 해당 없음(기존 분기) | 미실행 — 275.16 hit=0(편집 전도 0) |
| B10 | Restore 실패 | 없음(grep) | 해당 없음(기존 분기) | 미실행 — 280.45 hit=0(편집 전도 0) |
| B11 | 통화 빈 값 → 기본값 | `TestA166EngineReconcileLocalTLS` | 해당 없음(기존 분기) | yes — 285.20 hit=1 |
| B12 | 활성 블록 0 → 거절 | `TestA195NoActiveBlockIsRefused`(production builder 의 validate) | 편집 전에도 통과(기존 동작 고정 시험 — red.txt PASS) | yes — 306.24 hit=1 |
| (return@309) | 원인 판정 결과 전달 | `TestA195OperatorReleasesASymbolAttributionFailure` · `…MismatchAndAttributionReleaseTogether` · `…ExternalHoldingDoesNotRefuseTheRelease`(통과) · `…IdentifierConflictKeepsEveryRowBlocked` · `…AccountWideAttributionFailureIsRefused`(거절) | yes — 통과 3건이 RED 에서 실패(`analysis/tdd/red.txt`) | yes — 309.4 hit=1 |
