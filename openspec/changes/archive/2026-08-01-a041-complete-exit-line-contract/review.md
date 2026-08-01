# Review: a041-complete-exit-line-contract

- Date: 2026-07-31
- Voices: Security, Test Architecture, Maintainability
- Scope: exit snapshot, policy identity, setting metadata foundation

## Findings and decisions

1. Stable policy ID만으로 mutable rung을 식별할 수 없다. ID/version/canonical digest와 deterministic snapshot/decision identity를 계약에 추가했다.
2. UI 계약이 a050에 늦게 생기면 upstream descriptor 소유권이 순환한다. transport-neutral `internal/settingmeta` 최소 계약은 a041이 소유하고 각 domain은 자기 값을 제공한다.
3. 1주 partial은 0수량 주문이 아니라 state-only promotion이어야 하고 final/breach만 정확히 1주를 청산한다.
4. 구현 전 기존 evaluator/exitloop의 Function Logic Map과 경계·race Branch Test Map이 필수다.

## Verification evidence

- OpenSpec strict validation: pass.
- Existing full Go suite: pass (proposal-freeze test review).
- LIVE/order authority change: none; defaults remain conservative.

## Verdict

계약 구현을 승인한다. default-OFF/zero-order 회귀, deterministic identity, same-snapshot consumer와 independent implementation review를 gate 조건으로 한다.

## Independent re-review · 2026-08-01

- Scope: `bb2579037f54cee20630115c1d6b802f7a933273..ace73540f431230d61630e4a321a020ce46803f0`
- Verdict: **CLEAN FOR INTEGRATION**.
- Previous finding 1 closed: quote `FetchedAt` 또는 cycle 단일 fallback에서 opaque observation ID를 만들고 DecisionID-derived intent와 typed journal/proposal/mutation provenance로 연결했다. 실제 두 observer와 SQLite journal의 concurrent test는 proposal 1건·mutation attempt 1건만 허용한다.
- Previous finding 2 closed: pre-a042 ID-only state는 fixed legacy policy identity와 정확히 일치할 때만 평가하고 common/adopted policy digest drift와 같은 ID/version 의미 변경은 fail-closed한다. a042 schema handoff는 별도로 명시했다.
- Verification: focused Go tests, `go test -race -count=1 ./internal/exitpolicy ./internal/journal ./internal/app/engine`, `make test`, `make vet`, `make validate`, Function Logic Map checker와 `make sdd-check` PASS.
- Safety: LIVE mutation을 실행하지 않았고 주문 권한·운영 토글·default-OFF 불변식을 변경하지 않았다.
