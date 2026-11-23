# a195-an-attribution-failed-block-can-be-released-by-the-operator Review

## 계약 작성 (2026-11-23)
- 작성: Manager(Fable). 근거 조사: 읽기 전용 조사원(Opus) — 분기 주장의 AST 는 세션 스크래치 산출물 인용이며 change 안 사본·HEAD 재확인은 tasks 1.3.
- `openspec validate a195-an-attribution-failed-block-can-be-released-by-the-operator --strict`: valid (rc 0, 2026-11-23 HEAD 1ea261100).

## 1단계 증거 (2026-11-23, 팀메이트 Opus, worktree `~/tossos-work/a195-impl`, base `064f59fd9`)
- base-commit.txt = `064f59fd99c11ea3a19adcdb804bbcdfed2470cb`. `generate_master_tracker.py --check` rc 0.
- CodeGraph 1.6.0(새 색인) · `tools/logic-map` AST 8개(`analysis/cited-ast/`) · HEAD grep · 선택 시험 커버리지(`analysis/coverage/`). 조정: `analysis/code-context/evidence-reconciliation.md`.
- 문서 정정(결정 불변): proposal 머리 주석과 코드 근거 표 4행의 「조사 인용/손으로 읽음」→ AST 인용, `Tracker.Resolve` 의 `ExpectCause` 실제 식(`firstNonEmpty(block.Cause, QUANTITY_MISMATCH)`), a192 판정 자리(`orderTerminalReleaseVerdict` B3 if@182), design D4·D5 근거 줄.
- Function Logic Map: `analysis/function-logic/cmd-tossctl--buildreconcileresolveruntime/`(pre-edit 판). 뒤집힐 기존 시험 0. `check_analysis.py --change a195-…` rc 0(required 0 — 편집 전).
- **tasks 1.4 정지 조건 충족**: `buildReconcileResolveRuntime` 본문 전체가 a166 census `internal/accountbudgetruntime/testdata/a166-wiring/shapes.json` 키 `cmd/tossctl/engine_reconcile.go||buildReconcileResolveRuntime` 로 고정 — 편집 전 Manager 결정 필요.

## proposal-freeze 리뷰 r1 (2026-11-23)
- 보이스: Codex(codex-cli 0.165.0, `codex exec --sandbox read-only --ephemeral … </dev/null`), 적대적 Eng. 작성자(Claude)와 다른 계열. 원문 `analysis/review/proposal-freeze-r1-codex.md`. 판정: **freeze 보류**.
- `openspec validate … --strict`: valid (rc 0, 이 worktree, 정정 반영 뒤).

| # | 등급 | 발견 | 처분 |
|---|---|---|---|
| R1 | P0 | 명령의 「재조회 일치」는 `diff.BlocksEntry()`(Quantities·MissingOrders)뿐이다. 닫힌 인스턴스는 `HeldBySymbol` 에서 빠지므로 닫힌 뒤 재매수·미편입 보유는 `ExternalPos` 로 통과한다. `Observe` 의 해제 규칙은 같은 상태를 분쟁으로 본다(`symbolsInDispute`, 개정 2 D8). 1단계 조정과 같은 결론 | **Manager 결정 대기** — D2·spec 문구(「차이」의 정의)와 새 가드 여부. 이번 사고 상태가 정확히 이 모양이라 가드를 세우면 사고를 못 푼다 |
| R2 | P1 | `ATTRIBUTION_FAILED` 는 거절 셋(`UNATTRIBUTED_SELL`·`SELL_ON_CLOSED`·`ENTRY_WHILE_CLOSING`, position.go:98-108 → position_projection.go:428-432)을 한 이름으로 묶는다. D1 은 셋을 같은 증거로 푼다 | **Manager 결정 대기** — 범위(거절 유형별 허용). 유형은 evidence 문자열에만 있다(position_projection.go:373) |
| R3 | P1 | 기존 하네스는 validate 를 스텁/복제한다 — production 클로저를 지나는 시험 0(커버리지 302~314 hit=0과 일치). 하네스 재사용이면 변이가 산다 | **Manager 결정 대기** — design §3 시험 전략 변경(production builder 가 돌려준 validate 를 거치는 시험 필수화) |
| R4 | P1 | 시험 전략에 자동 해제 금지(spec 시나리오 4)·혼합 원인 앞/뒤 순서·helper 결과 무시 변이·실제 비교기로 만든 외부 보유 입력이 없다 | **Manager 결정 대기** — 시험·변이 목록 추가 |
| R5 | P1 | a166 census 정지 조건 충족, 선택 시험에 census·a166 시험이 빠졌다(`-run 'A195|ReconcileResolve'` 는 `TestA166…` 를 안 잡는다) | **Manager 결정 대기** — shapes.json 갱신 승인과 선택 시험 범위. 구조 실패를 변이 CAUGHT 로 세지 말 것 |
| R6 | P2 | 도움말(:82-89)·성공 문구(:212)가 「quantity-mismatch」만 말한다 — 고치면 `newEngineReconcileResolveCmdWithDeps`·`runEngineReconcileResolve` 본문 편집(FLM 번들 2 추가) | **Manager 결정 대기** — 범위 |

- 사실 오류 정정을 요구한 발견: 0(리뷰의 「문제 없음」 목록은 1단계 정정과 일치). 문서 반영 없음.
- r2: 미실시 — r1 발견이 전부 결정 대기라 문서가 바뀌지 않았다. Manager 결정 반영 뒤 r2(1회 남음).

## Manager 결정 (2026-11-23) — r1 처분 확정
- R1 기각(D2 근거 (1)~(3)), spec 「차이」→「진입을 막는 차이(수량 불일치, 원장에만 있는 주문)」로 정정 + 외부 보유 Scenario 추가.
- R2 기각(D1 — 세 유형 모두, 문구 파싱 금지). R3·R4 수용(design §3 보강, 변이 6). R5 수용(D6 — shapes.json 키 하나, 선례 a189 `3af104af9`·a192 census-repair 절차). R6 수용(D7 — 두 함수 FLM 번들 pre-edit 추가: `analysis/function-logic/cmd-tossctl--runenginereconcileresolve`, `…--newenginereconcileresolvecmdwithdeps`; 둘 다 census 키 아님, 파일 텍스트 시험은 금지 토큰·「released」 단언만).
- 반영 문서: proposal(What Changes 1·4, Capabilities, Impact), design(D1·D2·D3 갱신, D6·D7 신설, §2·§3·§4·§5), spec delta, tasks(1.6b, 3.1·3.2·3.2b·3.3·4.1·4.2). `openspec validate --strict` valid. `check_analysis.py --change` rc 0(required 0, 번들 3).

## proposal-freeze 리뷰 r2 (2026-11-23, 마지막 회차)
- 보이스: Codex(codex-cli 0.165.0, `codex exec --sandbox read-only --ephemeral … </dev/null`), 적대적 Eng. 원문 `analysis/review/proposal-freeze-r2-codex.md`. 판정: P1 1·P2 1, P0 없음.
- r1 처분 재검토: R1 기각 취지 동의(근거 (2)만 부분 반박 → N1), R2 동의, R3 동의(단 builder 가 돌려준 validate 를 실행하는 시험은 아님 — 그 연결부는 D6 census 가 보완한다는 처분으로 이해), R4 동의(CAUGHT 미확인), R5 동의, R6 동의(N2).

| # | 등급 | 발견 | 처분 |
|---|---|---|---|
| N1 | P1 | D2 근거 (2) 「다음 대사에서 손절이 걸린다」는 조건부다 — 제외 목록·편입 설정(adoption.go:163), 가격 누락·노후(230-246), exit state 생성 실패(394-400)면 미보호, 미편입은 exit 관측 밖(exitloop.go:592). 팀메이트가 네 자리를 HEAD 로 확인 | **반영(사실 정정, 결정 불변)** — D2 근거 (2)를 「정상 편입 경로를 다시 열며, 편입 조건·보호 상태 생성이 성공하면 손절 관리 대상」으로, §4 에 편입 제외·설정·exit state 확인을 추가 |
| N2 | P2 | 「해제한 원인을 출력」은 현재 runtime 인터페이스로 불가(validate·resolve 가 `error` 만, tracker 는 builder 지역, 해제 뒤 블록 비움) — 편집 범위 「문장만」과 모순 | **Manager 결정 대기** — 원인 중립 문장만 / 원인 목록 데이터 경로(범위 확장). 문서는 모순만 지움(D7·§3.6 에 결정 대기 표시) |

- 리뷰 회차 상한 도달 — r3 없음.

## Manager 1단계 검증·freeze 2026-11-23
- proposal-freeze(커밋 `a2e469411`, Manager 가 spec delta·design 직접 확인, strict validate 재실행). N2: 원인 중립 문장만(D7 갱신). N1: 승인.

## Pre-Edit 선언 (tasks 2.1, 2026-11-23)

```text
Pre-Edit Gate:
- change id / task id: a195-an-attribution-failed-block-can-be-released-by-the-operator / 3.1·3.2·3.2b
- 대상 심볼(패키지.함수):
  편집 — main.buildReconcileResolveRuntime(validate 클로저 원인 순회 B13~B14 → 새 함수 호출),
         main.runEngineReconcileResolve(:212 성공 문장만), main.newEngineReconcileResolveCmdWithDeps(Long 문장만)
  신규 — main.reconcileResolveRefusedCause(새 파일 cmd/tossctl/reconcile_resolve_causes.go)
  데이터 — internal/accountbudgetruntime/testdata/a166-wiring/shapes.json 키 cmd/tossctl/engine_reconcile.go||buildReconcileResolveRuntime 하나
- CodeGraph definition/callers/callees/impact: analysis/code-context/codegraph-baseline.md (1.6.0, production 호출 자리:
  build 1 · run 1 · Tracker.Resolve 1 · runtime.validate() 1). impact 가 못 보는 결속 3(census·파일 텍스트 시험·OptionsGuard AST).
- CodeGraphContext 후보와 evidence reconciliation: not-applicable(analysis/code-context/codegraphcontext-context.md), 조정 evidence-reconciliation.md.
- 기존 동작 파악 근거: HEAD 064f59fd9 의 cited-ast 8개, engine_reconcile_test.go 13 시험·a166_reconcile_safety_test.go 2 시험·
  a166_wiring_census_test.go, 선택 시험 커버리지(analysis/coverage/). validate 클로저를 실행하는 기존 시험 0.
- Function Logic Map / Branch Test Map: analysis/function-logic/cmd-tossctl--{buildreconcileresolveruntime,runenginereconcileresolve,newenginereconcileresolvecmdwithdeps}/
- upstream 상속 테스트 영향: no — reconcile-resolve 는 TossOS 전용 명령. 뒤집힐 기존 단언 0, census 키 하나는 D6 절차로 갱신.
- 실패 테스트 선행 작성: yes — 3.1 RED(analysis/tdd/red.txt) 후 GREEN.
- 설정·DB·journal 변경과 rollback: 없음 — 원장 스키마·release cause·설정·토글 무변경. 해제 행은 종전 OPERATOR 형식이라 롤백 = 이전 이미지.
- 안전 불변식 §0 위반 여부 검토: 통과 — 주문 side effect 0(브로커 쓰기 핸들 없음 유지), 토글 무관, 손절·청산 즉시성 무변경,
  영구 차단은 여전히 운영자 명령만 풂, IDENTIFIER_CONFLICT·계좌 전체 AF 는 거절 유지(design D1·D5).
```

## 2단계 구현·검증 (2026-11-23, 팀메이트 Opus)
- 커밋: `5633b721f`(freeze·Pre-Edit) → `4441d8935`(3.1 RED) → `39ee1ed6c`(3.2·3.2b·3.3 GREEN) → 이 커밋(4.1·4.2).
- RED(`analysis/tdd/red.txt`): 헬퍼를 「모든 입력 거절」 빈 껍데기로 먼저 둬서 컴파일 실패가 아닌 단언 실패로 만듦. FAIL 5(원인 표, AF 해제, QM+AF, 외부 보유, 원인 중립 문구), PASS 5(기존 동작 고정 — 차이 남음·IC·계좌 전체 AF·활성 0·Observe 자동 해제 금지).
- GREEN: `reconcileResolveRefusedCause`(새 파일 cmd/tossctl/reconcile_resolve_causes.go), validate 클로저 원인 순회 8줄 → 1줄, 도움말·성공 문구 원인 중립. census 키 하나 갱신(`analysis/census-repair/shapes-key-diff.txt`: base 재현 True·인코딩 1회, 90 키 중 변경 1, sites.json 무변경).
- 시험 하네스: 통합 시험은 production `buildReconcileResolveRuntime` 이 돌려준 런타임(validate·resolve·localState·compare 그대로, 공식 스냅숏 수집만 가짜, 로컬 TLS 가짜 서버)을 production `runEngineReconcileResolve` 로 구동 — r1 R3·r2 R3 의 「builder 가 돌려준 validate 를 실행하는 시험 0」이 해소됨(커버리지 `analysis/coverage/after-engine_reconcile-blocks.txt`: 클로저 306·309 hit=1).
- 외부 보유 시험은 장부에 그 종목 인스턴스가 없는 입력으로 만듦 — CLOSED 인스턴스만 있는 장부와 `HeldBySymbol`(compare.go:155) 이 같은 비교 입력을 만든다는 근거(실제 CLOSED 인스턴스로는 재지 않음).
- 변이 원장(`analysis/mutation/ledger.md`, 하네스 `mutate.py` — git archive 사본, 시작 sha 단언, 무변이 대조군 GREEN, 원복 sha256 단언): 6/6 CAUGHT. M6(결과 무시)은 census 와 행동 시험 둘 다로 죽음. 약한 자리: M5 는 원인 표의 한 하위 시험만 죽임.
- 4.2(`analysis/tdd/selected-tests.txt`, `analysis/tdd/lint-validate-analysis.txt`): `./cmd/tossctl -run 'TestA166EngineReconcile|TestEngineReconcileResolve|A195'` ok · `./internal/accountbudgetruntime` ok · `./internal/reconcile -run 'Resolve|Restore|Observe'` ok · `make lint` rc 0(gofmt + vet 무태그·tossos_testseams) · `openspec validate --strict` valid · `check_analysis.py --change` rc 0(required 3, 증거 완전).
- 미실행 범위: `make test`(전체 스위트)·`make test-seams`·`make test-race`·`make gate`·`make sdd-check`·`make sdd-sync` — 지시대로 안 돌림. `./cmd/tossctl`·`./internal/reconcile` 는 `-run` 선택분만. `--record-landing` 안 함. 실계좌·운영 장부·tossctl 실행 없음.
- 4.3(구현 적대 리뷰·gstack review): 미실시 — Manager 지시로 보류.
