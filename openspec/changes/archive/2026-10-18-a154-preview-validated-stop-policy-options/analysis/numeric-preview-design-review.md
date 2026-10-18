# Numerical preview design review — 2026-10-18

Status: user approved hypothetical-only calculation; canonical scope reconciled by Manager. Draft review technical findings resolved. No implementation or test completion is asserted.

## Applicability and method

- Manager authored the numerical design after the user's selection of numerical preview design over an unknown-only display.
- ai-memory retrieval was attempted: scoped project unavailable, global a154 search empty. Source-first research is recorded in research-first.md; no durable memory write.
- ECC search-first: local exact-risk arithmetic, console, offline a151 formulas, existing middleware and test harness evaluated at base c14642e979e07d16c19501fa93dd448da52ecff7. Extend existing UI and RiskBasedQuantity; build only a pure calculation leaf. No dependency additions, no external ecosystem search claimed.
- grill-me/grilling: unresolved material scope is whether the user accepts manual hypothetical inputs with costs/target/FX/apply deferred. Implementation mechanics and boundary rules are resolved from source and reviewer findings.
- impeccable: applicable to the existing optimization UI; Operate mode, accessible labels/errors and actual isolated rendered review required after implementation. No rendered verification claimed at design stage.
- Independent reviewer /root/a148_independent and separate gstack reviewer /root/a148_gstack reviewed the draft against source. These are Codex-host reviews, not an assertion of a different provider or model. Terra /root/terra_a148 separately checked feasibility without editing source or running tests.

## Findings and disposition

1. Existing Store.preview persists candidates/capabilities. Excluded from the pure calculator; no new registry setting or apply binding.
2. StrategyEntryQuantity requires a production Policy and rejects zero results. Use the narrower RiskBasedQuantity after geometry validation with exact caps instead.
3. Input zero versus computed zero unspecified. Provided budget/share/notional caps must be positive; computed zero remains a valid no-position result.
4. Output precision unspecified. Prices/amounts remain exact terminating decimals; percent six-place half-up with positive values below the display unit shown as <0.000001%, never 0%.
5. Missing/invalid policy inputs could suppress unrelated results. Explicit per-arm errors; common geometry blocks all; sizing failures retain valid stops/distances.
6. Existing middleware combines URL and body forms, and static guards classify all CSRF paths as state-changing. Reject query parameters, validate PostForm duplicates including csrf, bound body to 8 KiB, add narrow compute-only static classification and no-store before guards.
7. session0 permits the configured trusted-network mode. Preserve that contract and avoid claiming unconditional session authentication.
8. Signal-bar low is not a confirmed structural low. Name it accurately. KR/US arithmetic is not US Parker runtime support. Costs and actual loss limits remain explicitly unverified.

## Verification remaining

Final draft reread: both independent and gstack reviewers confirmed all identified technical design blockers resolved after the Manager corrections. PASS applies only to the reviewed minimum draft, not scope freeze, implementation, rendered UI, tests or lifecycle. Neither reviewer edited source or ran tests in this design review.

Canonical proposal/design/delta/tasks reconciliation after scope decision; final design reread; Terra implementation and exact numeric/HTTP no-write tests; actual isolated desktop/mobile/keyboard evidence; independent and gstack implementation reviews; Manager final verification and actual full gate; formal spec synchronization/archive and PM checks before scoped commit/push. Existing evidence, live services and operating settings remain out of scope.
