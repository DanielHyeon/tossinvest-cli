# Branch Test Map: `buildReconcileResolveRuntime` (pre-edit 판)

- Source: `cmd/tossctl/engine_reconcile.go` (237-319), base `064f59fd9`.
- 커버 측정: `../../coverage/pre-edit-buildReconcileResolveRuntime-blocks.txt` — `go test ./cmd/tossctl -run 'TestA166EngineReconcile|TestEngineReconcileResolve' -count=1 -covermode=set`(2026-11-23, PASS). hit=1 은 그 선택 시험에서 실행됨, 0 은 그 범위에서 미실행. 패키지 전체 시험으로는 재지 않았다.
- 시험 이름은 grep 으로 확인한 것만 적는다. production validate 클로저를 부르는 시험은 grep 으로 0(`rt.validate`·`runtime.validate()` 호출 없음; engine_reconcile_test.go:251·:365 는 validate 를 **손으로 복제/스텁**한다).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | loadConfig 실패 | 없음(grep) | 해당 없음(기존 분기) | 미실행 — 블록 239.16 hit=0 |
| B2 | OpenAPI 경로 실패 | 없음(grep) | 해당 없음(기존 분기) | 미실행 — 243.16 hit=0 |
| B3 | 자격 읽기 실패 | 없음(grep) | 해당 없음(기존 분기) | 미실행 — 247.16 hit=0 |
| B4 | 자격 빈 값 | 없음(grep) | 해당 없음(기존 분기) | 미실행 — 250.103 hit=0 |
| B5 | 예산 required 실패·취소 컨텍스트 | `TestA166EngineReconcileLocalTLS`(:87·:105) | 해당 없음(기존 분기) | yes — 255.16 hit=1 |
| B6 | 계좌 조회 실패 | 없음(grep) | 해당 없음(기존 분기) | 미실행 — 259.16 hit=0 |
| B7 | 계좌 신원 실패 | 없음(grep) | 해당 없음(기존 분기) | 미실행 — 263.16 hit=0 |
| B8 | configDir 있음 | `TestA166EngineReconcileLocalTLS`(:91) | 해당 없음(기존 분기) | yes — 269.60 hit=1 (거짓 쪽 미측정) |
| B9 | 장부 열기 실패 | 없음(grep) | 해당 없음(기존 분기) | 미실행 — 273.16 hit=0 |
| B10 | Restore 실패 | 없음(grep) | 해당 없음(기존 분기) | 미실행 — 278.45 hit=0 |
| B11 | 통화 빈 값 → 기본값 | `TestA166EngineReconcileLocalTLS`(기본 설정) | 해당 없음(기존 분기) | yes — 283.20 hit=1 |
| B12 | 활성 블록 0 → 거절 | 없음 — production 클로저를 부르는 시험 0 | 미실행 | 미실행 — 304.24 hit=0. design §3 의 7 이 새로 덮어야 함 |
| B13 | 블록 순회 | 없음 — 위와 같음 | 미실행 | 미실행 — 307.4·307.33 hit=0. design §3 의 1~6 |
| B14 | QM 아닌 원인 → 거절 | 없음 — `TestEngineReconcileResolveRefusesNonQuantityCauses` 는 validate 를 스텁으로 바꿔 이 분기를 지나지 않음 | 미실행 | 미실행 — 308.62 hit=0. design §3 의 3·4·6 |

뒤집힐 기존 시험: 0 — `ATTRIBUTION_FAILED` 거절을 단언하는 시험 함수는 없다(grep `ATTRIBUTION_FAILED`·`ReconcileCauseAttributionFailed`·`cause-specific recovery evidence` in `cmd/tossctl/*_test.go`: 스텁 문자열 1건, engine_reconcile_test.go:252). 단 `TestA166WiringCensusBaseline`(internal/accountbudgetruntime)은 본문 편집 시 **반드시** 실패한다 — 그 시험의 데이터(shapes.json)를 같이 갱신해야 한다.
