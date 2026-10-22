# a162 추가 통합시험 조사

- 기준: test-tagged `internal/weeklyvaluelane/strategyflow_testseam.go`와 `internal/officialfx/authority_testseam.go`.
- `strategyflowEvaluation`은 ordinal 1..7을 입력으로 받지만 현재는 ordinal>1에서 leg 1만 native reserve/fill로 전이한 뒤 해당 ordinal을 reserve한다. native reservation은 `PositiveLegCount+1`만 허용하므로 3..7은 현재 fixture로 구성되지 않는다.
- 최소 변경: build-tag fixture 안에서 prior ordinal `1..k-1`을 각기 existing `ApplyReservation`과 `ApplyPositiveFillAtomic`로 전이한다. 이전 state/risk를 누적 사용하고 loop 안에서 초기화하지 않는다. 각 leg는 unique reservation/fill/idempotency identities와 advance된 official MarketWeek를 사용한다.
- 시간: current fixture는 week만 이동하고 evaluation evidence timestamp는 고정한다. 추가 fixture는 every synthetic week and evaluated/observed/fresh timestamps together so native `ValidateMarketWeek`/freshness remains consistent. This remains test-only synthetic chronology, not a market-calendar or account-currentness claim.
- FX: tagged `officialfx.EvidenceForAuthorityTest` supplies existing opaque test-only evidence. The fixture passes its public observation into the native tagged `mintFrozenFX` path, matching lane frozen-FX construction without a production scalar constructor or runtime authority import.
- Stale: change only a freshness boundary after a fresh control has passed; assert native `EVIDENCE_STALE` is returned before capture and no quantity is exposed. Every positive/refused result remains `LedgerCurrentness=unverified`, `ExecutionAuthority=false`.
- Scope: no production source, no live service, no full package suite/gate. Existing a162 implementation evidence/logs are not replaced.

## CodeGraph availability

`tools/codegraph` is absent in this checkout (`stat .../tools/codegraph: directory not found`), so no new graph index/query was run. Existing ASTs above are the function-branch source of truth for this test-only seam change.

## 2026-10-20 CodeGraph correction

`codegraph` is available at `/home/daniel/.local/bin/codegraph`. Narrow `query` results for `strategyflowEvaluation` and `strategyflowAdvanceFirstLeg` are stored under `codegraph/`; no index rebuild or sync was run. The earlier "tools/codegraph absent" statement referred only to the repository-local path and is superseded.

## 구현 결과 정정

- 현재 tag-only fixture의 `strategyflowAdvanceLeg`가 native reservation/atomic fill prefix를 만든다. 이전 `strategyflowAdvanceFirstLeg` query는 초기 조사 기록이며 현재 구현 근거는 [strategyflowAdvanceLeg-final.json](codegraph/strategyflowAdvanceLeg-final.json)과 `analysis/function-logic/internal-weeklyvaluelane--strategyflowadvanceleg/`이다.
- adapter에는 test-only FX mutate 및 active-reservation removal 반례를 추가했다. 두 반례 모두 native refusal과 adapter 수량 부재를 검증하며 production package에는 포함되지 않는다.
