# Review: a043-show-exit-lines-in-trading-views

- Date: 2026-07-31
- Voices: Security, Test Architecture, Maintainability, UI/UX

## Findings and decisions

1. `ExitLineView`는 console 소유가 아니라 `internal/operatorview`의 transport-neutral DTO다.
2. 화면은 a041/a042 snapshot을 재계산하지 않고 decision ID로만 order를 연결한다. symbol/time 근사 join은 금지한다.
3. StockOS의 contextual navigation을 참고하되 거래 화면은 read-only이며 입력 control을 0개로 유지한다.
4. 360px, keyboard/ARIA, CSP, stale/unknown/1주/broker-only order를 증거로 남긴다.
5. 구현 중 schema를 재확인한 결과 `mutation_attempts.decision_id`는 Guardian 결정 FK이고 exit-line
   decision이 아니었다. 따라서 deterministic join은 `broker_order_id → attempt.intent_id →
   exit_event.proposed_intent_id`로 고정하고, 연결된 event의 exit decision/snapshot을 표시한다.

## Verification evidence

> Correction (2026-08-01): the earlier GREEN statement and approval below were
> recorded before exact-commit independent review. Review of `6bebbe2` blocked on
> missing market identity, overly broad attempt-state attribution, final-only SQL
> bounding, and pre-filter evidence scopes. They are superseded by this correction;
> task 3.2 remains unchecked until the next exact-commit review.

> Correction (2026-08-01, second exact-commit review): review of `21d5a3b`
> remained BLOCKED because broker order ids were trimmed, OPEN/CLOSED overlap used
> a trimmed bare id, delimiter-encoded recursive paths could corrupt opaque ids,
> and cycle/depth coverage did not isolate those branches. The hardening GREEN
> entry below remains historical command evidence for `21d5a3b`, not approval.
> Task 3.2 and the gate remain unchecked pending review of the next exact commit.

- OpenSpec strict validation: pass.
- Mutation capability: none by contract.
- Dependency baseline: implementation starts from `70aabdc`, after a041/a042 were
  integrated and gated. `base-commit.txt` was advanced from the portfolio-planning
  commit so a043's Function Logic Map and diff gate measure this change rather than
  attributing its prerequisite snapshots to this UI change.
- RED: complete/stale/unknown/1-share positions, exact/unlinked same-symbol order,
  forbidden controls, and POST 405 fixtures failed before the view wiring.
- The pre-review GREEN list is historical evidence for the first implementation,
  not approval of the current exact commit. Current hardening verification is
  recorded only after all focused/full/race/vet/strict/SDD commands complete.
- Hardening GREEN (2026-08-01): `make test`, `make vet`, focused
  `go test -race ./internal/journal ./internal/operatorview ./internal/console ./cmd/tossctl`,
  strict OpenSpec validation, Function Logic Map validation, `make sdd-sync`, and
  `make sdd-check` pass. The 1,000-parent adversarial lineage fixture completes
  within its two-second context and returns one fail-closed result without recursion fan-out.
- Second-review RED (2026-08-01): focused tests reproduced `" O-1 "` collapsing
  into `"O-1"`, market/day/invalid-time reuse being hidden by the OPEN set, and
  a valid `PREFIX|ROOT → ROOT` lineage being misclassified as a cycle. The prior
  combined branch fixture did not reach a pure single-parent cycle; no pure
  depth-over-32 fixture existed.
- Second-review GREEN (2026-08-01): broker ids remain byte-exact (only `len==0`
  is empty); OPEN/CLOSED overlap uses account + canonical market + market-local
  day + opaque id with a tagged raw fallback; recursive visited state is a JSON
  array with exact `json_each` equality. Exact depth 32 stays linked and depth 33
  fails closed. Focused tests, `make test`, `make vet`, Function Logic Map
  validation, and serial focused race for all broker-order lineage plus changed
  console identity/filter tests pass. Strict OpenSpec validation, `make sdd-sync`,
  and `make sdd-check` also pass. A broad journal+console race run hit the
  repository test binary's 10-minute timeout under concurrent race load and is
  explicitly not counted as PASS evidence.

## Verdict

독립 재리뷰 대기. canonical DTO, account+market+market-local-day identity,
opaque broker id, CONFIRMED PLACE/AMEND evidence, collision-free bounded lineage,
no-recompute/no-fuzzy-link와 input-free 렌더 검증이 승인 조건이다.
