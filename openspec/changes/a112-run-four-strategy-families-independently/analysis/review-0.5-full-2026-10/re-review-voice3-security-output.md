# a112 0.5 voice3 (security) 표적 재검 — P1#3 SubagentHandback 원문 (agent a5d13ac03d3f7f694, 좌표 2497427e, 저장 2026-10-05)

**P1#3(보안#1): 반영.** 원래 변이 셋과 「자리를 놓치는 모양」 탐침 넷이 모두 잡혔다. P0 는 보지 못했다.

**안전 규칙:** `~/.codex` 와 `_codex*/` 는 열지 않았다. 검색은 `internal/` 의 두 생산 파일과 시험 파일 하나, 그리고 지정된 원장 두 줄 grep 에 한정했다. openspec 전체 grep 은 하지 않았다.

**고친 자리 (2497427e 사본에서 읽음)**
- `internal/app/engine/a112_shadow_structure_test.go:491-504`: 핀 세 줄이다.
  - `{coordinate, "shadow", …3}`
  - `{collectMarket, "shadow", …3}`
  - `{collectMarket, "collected", …2}`
- `:509-528` `a112LocalObjectSites`: `info.Defs` 와 `info.Uses` 로 세고, 범위는 그 함수 선언 안에 정의된 같은 이름의 `*types.Var`(필드 제외)다.
- 대상이 되는 생산 자리:
  - `strategy_market_coordinator.go:65` `shadow := strategyShadowBatch{observed: true}` · `:93` `shadow.collect(envelope.Proposal)` · `:129` `return arbitration, refused, shadow`
  - `strategy_proposal_authority.go:400-401` `arbitration, refused, collected := …` 와 `*shadow = collected.boundTo(…)`

**재실행**
- 명령: 사본의 디스크 파일을 바꿔 가며 `go test -count=1 -tags tossos_testseams -run 'Shadow|Collect' ./internal/app/engine` 를 돌렸다. AST 시험은 디스크를 읽으므로 overlay 는 쓰지 않았다. 매 회 원본으로 되돌렸고, 끝에 `git show 2497427e:<file> | cmp` 로 두 파일 모두 원본과 같음을 확인했다.
- 대조군(무변이): `ok 31.1s`

| 변이 | 결과 | 실패한 시험 / 메시지 |
|---|---|---|
| M4 `if len(shadow.inputs) > 64 { gate = strategyFamilyGate{} }` | CAUGHT | `TestTheCoordinatorCollectsWithOneStatement…:501` — coordinate "shadow" 4 sites, want 3 |
| M2 (내 원래 철자 `refused++`) | CAUGHT | 같은 시험 1건, 4 vs 3 |
| M2g (저자 철자 `gatedInScope++`) | CAUGHT | 3건: `TestCollectMarketCarriesTheShadowBatchOnlyFromCoordination` · `TestCollectReturnsTheShadowPairBesideTheAuthorityPair` · `TestTheCoordinatorCollects…` |
| C03 collectMarket 에 `if len(collected.inputs) > 64 { refused++ }` | CAUGHT | `TestTheCoordinatorCollects…` — collectMarket "collected" 3 vs 2 |
| C03b collectMarket 에 `shadow.inputs` 읽기(포인터 자동 역참조) | CAUGHT | 같은 시험 — collectMarket "shadow" 4 vs 3 |
| 탐침 `&shadow` (`if p := &shadow; len(p.inputs) > 64 …`) | CAUGHT | 4 vs 3 |
| 탐침: 클로저 안에서 사용 (`func(){ if len(shadow.inputs) > 64 … }()`) | CAUGHT | 4 vs 3 (`ast.Inspect` 가 FuncLit 안까지 내려감) |
| 탐침: 같은 이름 가리기 (`{ shadow := len(routes); if shadow > 64 … }`) | CAUGHT | 5 vs 3 (가린 지역 변수도 이름으로 셈) |

**놓치는 모양이 있는가 — 판정 근거로만 적는다**
- 세 자리는 각각 따로 단언된다. 선언 리터럴(`:412-414`), `shadow.collect` ExprStmt 정확히 하나, 두 return 의 `Results[2]` 이다. collectMarket 쪽은 두 `*shadow` 대입의 우변 문자열까지 단언한다(`:483-486`).
- 그래서 「기존 자리 하나를 읽기로 바꿔치기」로 개수를 맞출 여지가 없다. 함수 안에서 값에 닿으려면 식별자가 필요하고, 그 식별자는 개수를 늘린다.
- 이 핀은 이 두 함수 안의 자리만 본다. 다른 함수를 통하는 길은 이번 지적의 범위 밖이라 확대하지 않는다.

**원장 대조:** 일치한다.
- `mutation-0.5-R-run1.tsv` C01(M4)·C03: CAUGHT, 실패 1건(`TestTheCoordinatorCollects…`). 재실행과 같다.
- C02(M2, 저자 철자 `gatedInScope++`): CAUGHT, 실패 3건. 내 M2g 의 시험 이름 셋과 같다. 내 원래 철자(`refused++`)로는 1건 실패, 역시 CAUGHT 다.
- `disposition.md:16` 보안#1 행: 「고침(시험) … C01 · C02 · C03 CAUGHT」. 재실행과 일치한다.

**저장소 상태:** HEAD 는 시작과 끝 모두 `2497427e00149bac92195cb38261daa225cfdcc1` 다. 그러나 **`status --short` 는 같지 않다.**
- 시작: `?? .reticle-setup-crash.log` · `?? …/analysis/harness/gate_suites.sh` · `?? w4.log`
- 끝: 위 셋에 더해 다음이 새로 생겼다.
  - `?? …/analysis/harness/btm_census.py`
  - `?? …/analysis/harness/btm_remeasure.py`
  - `?? …/analysis/measurements/gate-8.1-8.3-2026-10-05/`
- 나는 실제 저장소에 아무것도 쓰지 않았다. 쓴 곳은 스크래치패드 `…/scratchpad/re/` 뿐이다.
- 새 파일들은 병행 세션이 만든 것으로 보인다. 주인을 확인하기 바란다.
