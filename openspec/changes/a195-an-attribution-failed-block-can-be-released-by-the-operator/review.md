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
