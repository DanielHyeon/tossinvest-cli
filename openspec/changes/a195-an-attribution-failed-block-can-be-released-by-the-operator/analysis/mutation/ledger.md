# a195 변이 원장 — 시작 sha 39ee1ed6c18a64a8ea78b62d3257bb64fecc1878(단언 통과), 사본 /home/daniel/tossos-work/a195-mut-616462, GOFLAGS='-trimpath -buildvcs=false'
# 스위트: go test ./cmd/tossctl -run TestA166EngineReconcile|TestEngineReconcileResolve|A195 -count=1 ; go test ./internal/accountbudgetruntime -run TestA166WiringCensus -count=1
## 무변이 대조군: ./cmd/tossctl rc=0, ./internal/accountbudgetruntime rc=0
## M1 원인 판정 삭제(헬퍼가 항상 통과) — CAUGHT
- 파일 cmd/tossctl/reconcile_resolve_causes.go, 앵커 1회, 원복 sha256 일치
- ./cmd/tossctl: rc=1, 실패 시험 7: --- FAIL: TestA195RefusedCauseTable (0.00s); --- FAIL: TestA195RefusedCauseTable/identifier_conflict_only (0.00s); --- FAIL: TestA195RefusedCauseTable/attribution_then_identifier_conflict (0.00s); --- FAIL: TestA195RefusedCauseTable/identifier_conflict_then_attribution (0.00s); --- FAIL: TestA195RefusedCauseTable/account-wide_attribution_failure (0.00s); --- FAIL: TestA195IdentifierConflictKeepsEveryRowBlocked (0.27s); --- FAIL: TestA195AccountWideAttributionFailureIsRefused (0.27s)
- ./internal/accountbudgetruntime: rc=0, 실패 시험 0: 
## M2 허용 목록에 IDENTIFIER_CONFLICT 추가 — CAUGHT
- 파일 cmd/tossctl/reconcile_resolve_causes.go, 앵커 1회, 원복 sha256 일치
- ./cmd/tossctl: rc=1, 실패 시험 5: --- FAIL: TestA195RefusedCauseTable (0.00s); --- FAIL: TestA195RefusedCauseTable/identifier_conflict_only (0.00s); --- FAIL: TestA195RefusedCauseTable/attribution_then_identifier_conflict (0.00s); --- FAIL: TestA195RefusedCauseTable/identifier_conflict_then_attribution (0.00s); --- FAIL: TestA195IdentifierConflictKeepsEveryRowBlocked (0.33s)
- ./internal/accountbudgetruntime: rc=0, 실패 시험 0: 
## M3 종목 범위 조건 삭제(계좌 전체 AF 허용) — CAUGHT
- 파일 cmd/tossctl/reconcile_resolve_causes.go, 앵커 1회, 원복 sha256 일치
- ./cmd/tossctl: rc=1, 실패 시험 3: --- FAIL: TestA195RefusedCauseTable (0.00s); --- FAIL: TestA195RefusedCauseTable/account-wide_attribution_failure (0.00s); --- FAIL: TestA195AccountWideAttributionFailureIsRefused (0.26s)
- ./internal/accountbudgetruntime: rc=0, 실패 시험 0: 
## M4 ATTRIBUTION_FAILED 를 허용 목록에서 빼기 — CAUGHT
- 파일 cmd/tossctl/reconcile_resolve_causes.go, 앵커 1회, 원복 sha256 일치
- ./cmd/tossctl: rc=1, 실패 시험 9: --- FAIL: TestA195RefusedCauseTable (0.00s); --- FAIL: TestA195RefusedCauseTable/symbol_attribution_failure_only (0.00s); --- FAIL: TestA195RefusedCauseTable/mismatch_then_attribution (0.00s); --- FAIL: TestA195RefusedCauseTable/attribution_then_mismatch (0.00s); --- FAIL: TestA195RefusedCauseTable/attribution_then_identifier_conflict (0.00s); --- FAIL: TestA195OperatorReleasesASymbolAttributionFailure (0.29s); --- FAIL: TestA195MismatchAndAttributionReleaseTogether (0.31s); --- FAIL: TestA195ExternalHoldingDoesNotRefuseTheRelease (0.26s); --- FAIL: TestA195HelpAndSuccessOutputAreCauseNeutral (0.27s)
- ./internal/accountbudgetruntime: rc=0, 실패 시험 0: 
## M5 첫 허용 원인에서 즉시 통과 반환 — CAUGHT
- 파일 cmd/tossctl/reconcile_resolve_causes.go, 앵커 2회, 원복 sha256 일치
- ./cmd/tossctl: rc=1, 실패 시험 2: --- FAIL: TestA195RefusedCauseTable (0.00s); --- FAIL: TestA195RefusedCauseTable/attribution_then_identifier_conflict (0.00s)
- ./internal/accountbudgetruntime: rc=0, 실패 시험 0: 
## M6 validate 클로저가 헬퍼 결과를 무시 — CAUGHT
- 파일 cmd/tossctl/engine_reconcile.go, 앵커 1회, 원복 sha256 일치
- ./cmd/tossctl: rc=1, 실패 시험 2: --- FAIL: TestA195IdentifierConflictKeepsEveryRowBlocked (0.26s); --- FAIL: TestA195AccountWideAttributionFailureIsRefused (0.30s)
- ./internal/accountbudgetruntime: rc=1, 실패 시험 4: --- FAIL: TestA166WiringCensusBaseline (0.02s); --- FAIL: TestA166WiringCensusRejectsBypasses (0.09s); --- FAIL: TestA166WiringCensusRejectsBypasses/new-unclassified-http (0.01s); --- FAIL: TestA166WiringCensusRejectsBypasses/escaped-http-method (0.01s)
## 원복 뒤 대조군: ./cmd/tossctl rc=0, ./internal/accountbudgetruntime rc=0
## 닿음(변이가 앉은 줄이 선택 시험에서 실행되는가) — `../coverage/after-engine_reconcile-blocks.txt`(무변이 작업 트리, covermode=set)
- M1: reconcile_resolve_causes.go 14.67-15.31 hit=1(함수 진입 블록 — 앵커 바로 뒤)
- M2: 15.31-16.10 hit=1(QM case 조건 평가 블록), M5 의 QM `continue` 도 같은 블록
- M3·M4: 17.62-18.12 hit=1(AF case 조건·continue), M5 의 AF `continue` 도 같은 블록
- M6: engine_reconcile.go 309.4-309.47 hit=1(클로저의 `return reconcileResolveRefusedCause(blocks)`)
- 판정: 6/6 CAUGHT. M6 은 census(구조 — `TestA166WiringCensusBaseline`)와 행동(production builder 의 validate 를 쓰는 `TestA195IdentifierConflictKeepsEveryRowBlocked`·`TestA195AccountWideAttributionFailureIsRefused`) 둘 다로 죽음.
- 약한 자리: M5 를 죽인 것은 `TestA195RefusedCauseTable/attribution_then_identifier_conflict` 하나뿐(통합 시험은 Blocks() 정렬 순서가 IC 를 먼저 둬서 못 잡음).
# a195 변이 원장 — 시작 sha b7aeb035ca554022da69243aac2bce34ad0d092f(단언 통과), 사본 /home/daniel/tossos-work/a195-mut-756090, GOFLAGS='-trimpath -buildvcs=false'
# 스위트: go test ./cmd/tossctl -run TestA166EngineReconcile|TestEngineReconcileResolve|A195 -count=1 ; go test ./internal/accountbudgetruntime -run TestA166WiringCensus -count=1 ; go test ./internal/reconcile -run Resolve|Restore|Observe -count=1
## 무변이 대조군: ./cmd/tossctl rc=0, ./internal/accountbudgetruntime rc=0, ./internal/reconcile rc=0
## M7 Observe 원인 거름 완화 — AF 도 자동 해제 후보(원인·해제 조건 둘 다) — CAUGHT
- 파일 internal/reconcile/mismatch.go, 앵커 1회, 원복 sha256 일치
- ./cmd/tossctl: rc=1, 실패 시험 1: --- FAIL: TestA195ObserveNeverReleasesAnAttributionFailure (0.30s)
- ./internal/accountbudgetruntime: rc=0, 실패 시험 0: 
- ./internal/reconcile: rc=0, 실패 시험 0: 
## M7b Observe 원인 거름만 완화(해제 조건 Release 검사는 남김) — SURVIVED
- 파일 internal/reconcile/mismatch.go, 앵커 1회, 원복 sha256 일치
- ./cmd/tossctl: rc=0, 실패 시험 0: 
- ./internal/accountbudgetruntime: rc=0, 실패 시험 0: 
- ./internal/reconcile: rc=0, 실패 시험 0: 
## M8 원인을 보지 않고 종목 범위면 통과(R11) — CAUGHT
- 파일 cmd/tossctl/reconcile_resolve_causes.go, 앵커 1회, 원복 sha256 일치
- ./cmd/tossctl: rc=1, 실패 시험 3: --- FAIL: TestA195RefusedCauseTable (0.00s); --- FAIL: TestA195RefusedCauseTable/symbol-scoped_identifier_conflict (0.00s); --- FAIL: TestA195RefusedCauseTable/symbol-scoped_empty_cause (0.00s)
- ./internal/accountbudgetruntime: rc=0, 실패 시험 0: 
- ./internal/reconcile: rc=0, 실패 시험 0: 
## M9 빈 원인을 수량 불일치로 취급(R9) — CAUGHT
- 파일 cmd/tossctl/reconcile_resolve_causes.go, 앵커 1회, 원복 sha256 일치
- ./cmd/tossctl: rc=1, 실패 시험 2: --- FAIL: TestA195RefusedCauseTable (0.00s); --- FAIL: TestA195RefusedCauseTable/symbol-scoped_empty_cause (0.00s)
- ./internal/accountbudgetruntime: rc=0, 실패 시험 0: 
- ./internal/reconcile: rc=0, 실패 시험 0: 
## 원복 뒤 대조군: ./cmd/tossctl rc=0, ./internal/accountbudgetruntime rc=0, ./internal/reconcile rc=0
## 닿음 — M7~M9 (구현 적대 리뷰 r1 반영 뒤)
- M7·M7b: `go test ./cmd/tossctl -run TestA195ObserveNeverReleasesAnAttributionFailure -covermode=set -coverpkg=./internal/reconcile` — mismatch.go 666.34-668.49 hit=1(원인·해제 조건 평가), 668.49-672.13 hit=1(AF `continue`), 701.4-701.44 hit=1(QM 양성 대조군의 `out.Cleared` 추가).
- M8·M9: reconcile_resolve_causes.go 15.31-16.10·17.62-18.12 hit=1(`../coverage/after-engine_reconcile-blocks.txt`).
- 판정: M7·M8·M9 CAUGHT. **M7b SURVIVED** — 원인 거름만 풀어도 바로 뒤 `block.Release != ReleaseOnAdjustedReconcile` 가 AF 를 걸러 행동이 같음. 복원된 AF 블록의 Release 는 항상 `ReleaseOperatorOnly`(blockFromReconcileState mismatch.go:1408, QM 만 :1412 에서 바뀜)이고 Observe 가 새로 만드는 블록은 QM 뿐이라, 패키지 밖 시험으로 AF 에 `ReleaseOnAdjustedReconcile` 을 줄 길이 없음 — 도달 가능한 상태에서 동등 변이. 두 겹 거름 중 한 겹만 지우는 변이는 남은 겹이 막음(새 가드·production 편집 없이 기록만).
