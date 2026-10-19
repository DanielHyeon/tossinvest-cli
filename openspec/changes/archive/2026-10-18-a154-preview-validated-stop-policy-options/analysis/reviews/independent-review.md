# a154 independent implementation review

Date: 2026-10-18
Verdict: PASS for reviewed implementation. No outstanding actionable source defect found. This is not a full-gate, rendered-browser acceptance, lifecycle or runtime approval.

## Scope and evidence

Reviewed canonical design/spec, new calculator/HTTP handlers/tests, console route, template/CSS and static guard diffs. No production files changed by reviewer. No operating service, account, broker or live API used.

Independent execution: `go test ./internal/console -run 'TestA154|TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate' -count=1` passed on final reviewed files (package receipt: `ok .../internal/console 0.577s`). Earlier run also passed before the real database test was added. These focused runs do not substitute for full gate.

## Findings and disposition

- Initial evidence gap: fake counters alone did not satisfy approved actual storage invariance evidence. Resolved by `TestA154ActualOptimizationDatabaseUnchanged`: real temporary SQLite store wired to HTTP harness; snapshot/audit values and control/snapshot/candidate/application/audit row counts unchanged after calculation and rejected forged key. Independent final run includes this test. Source additionally has no store or operational-service call in calculation path; row counts alone would not establish arbitrary row-content immutability.
- Manager UI corrections verified: market error linked to select; computed zero explicitly states no purchasable quantity; initial required-input instructions and manual ATR/low limitations present.

## Acceptance reasoning

- Exact rational stop formulas, mathematical grid floor, geometry-before-sizing validation, independent low/ATR failure, missing/invalid sizing retention and minimum of three exact quantity limits match design. Amounts have at most eight fractional places because validated prices/grid have that precision and quantity is integral; eight-place rendering therefore loses no price/amount precision. Percent alone rounds; positive tiny percent has explicit threshold.
- KR 70000 oracle produces stops 69500/68900/69200, quantities 20/9/12, risks 10000/9900/9600. US 100 oracle produces stops 99.3/97.99/97, quantities 142/49/33, risks 99.4/98.49/99. Half-up 0.0078125% becomes 0.007813%; tiny positive ratio remains visibly nonzero. Tests assert these independently specified values.
- URL-query input, unknown/duplicate body fields, duplicate CSRF, missing authorization, invalid method and oversized requests are blocked. No-store encloses session/method checks; existing trusted-network/session semantics remain unchanged. Native escaped templates preserve valid values; no arbitrary policy coefficient, stored capability or apply action exists in calculator.
- New compute-only static exemption retains existing state-changing CSRF assertions. Session wrapper recognition is bounded to the named console no-store wrapper and `c.session0`, with negative tests.
- Initial stop, strategic exit and common protection remain distinguished; KR/US arithmetic does not claim US Parker support. Costs/target/FX/execution guarantees explicitly excluded near result. Existing operating registry, selector and store behavior unchanged.

## Remaining process verification

Actual isolated mobile/desktop/keyboard browser evidence and full repository gate remain separate Manager/writer obligations; not executed or claimed by this reviewer. No performance or trading recommendation follows from these synthetic numbers.

## Reviewed SHA-256

- `internal/console/a154_hypothetical.go`: `f72b2087e483d4f986bb85d764a3a0e742e52860a120e8cb9da6adc363676667`
- `internal/console/a154_hypothetical_http.go`: `d6e02de880e0173a3c104db8f6919a59920b29c533295d86fd59696b7e029f6c`
- `internal/console/a154_hypothetical_test.go`: `d80bfdc7c2059576093545c52b8a4fc18024feb7266afc9898c424384da92541`
- `internal/console/a154_hypothetical_http_test.go`: `ab4c6f63a104e575e2910e0f3e7af3ff18ca7790f4d663cc5110a4ab943f5178`
- `internal/console/a154_route_guard_test.go`: `6834ca3f28e5a6cefc239aafbeb94edb754b82ffe8361843fdaae28e5ebb963b`
- `internal/console/console.go`: `0239fcd72e977f7f8b702d559c4a91105185baa15de307660f9bc778114aaf5e`
- `internal/console/templates_optimization.go`: `c3f3344afd7c5e0de09b6812f697a6bee3409868603632657cbce6eaf59ab69c`
- `internal/console/templates.go`: `eb4c5ae2eb16af0f5f95e6101c6a1e1d9f970150b6d0caac09a188de785b777d`
- `internal/console/static_test.go`: `1eda5647fa47d9cef7dd129859056e12ab026e94c92b44507a1d708b52f7f3df`

## Final landmark correction rereview

Manager's rendered inspection found a nested main landmark after the initial review. Initial PASS did not cover rendered-browser acceptance and missed this template defect. Resolved by removing only the result template's redundant main pair; shared head/foot still supplies one pair. HTTP rendering regression asserts exactly one opening and closing main. Preserved `analysis/verification/main-landmark-red.log` shows the regression failing before correction; `focused-final6.log` passes afterward (0.606s). Reviewed both logs and current source; no calculator/security changes. PASS remains valid for corrected source; two affected hashes above refreshed.

Read writer's `analysis/verification/render-observations.md`: isolated KR/US oracles, narrow viewport scroll, keyboard focus and linked market error observed, no browser warnings/errors. These are writer observations rather than independent browser execution. Manager separately confirmed US rendering and empty console errors. Full gate remains separate.

## First full-gate repair review

First aggregate gate failed at step 5; original `/tmp/tossos-a154-final-gate.log` remains evidence of failure, not superseded by focused success. Reviewed the repair: four existing UI tests first validate and remove exactly one hypothetical section, then run their original operational-policy assertions unchanged. `stripHypothetical` requires one compute POST form, exact named controls once, input types and containment; rejects override form attributes, event handlers, unexpected controls, scripts/links/iframes and extra forms/buttons. Negative mutations add a forged setting key, an operating apply button and an operating form; each mutation is asserted non-no-op and rejected. This is a bounded new calculator exemption, not a general relaxation of operating form checks.

Result scroll wrapper restores exact existing `table-scroll` class marker, moving the calculator styling class onto its table; descendant cell styling and scroll-region keyboard accessibility remain intact. No calculation, request authority or storage code changed. Repair source review PASS. Full-console run and aggregate rerun are writer/Manager-owned and not claimed complete here.

Additional reviewed files:
- `internal/console/optimization_ui_contract_test.go`: `e806f159ccd115d62e550d42f321aa422057389bbf60eaab531524d996264b8f`
- `internal/console/a154_contract_scope_test.go`: `3dc48c10f104504299fdfb1e5608a482a460f523e32dfe0f29a979dcb4a804bc`

Receipt follow-up: final helper hash remains `3dc48c10f104504299fdfb1e5608a482a460f523e32dfe0f29a979dcb4a804bc`, matching the reviewed source. Read preserved `gate-repair-focused.log`: final helper focused checks PASS, exit 0 (0.054s). `console-repair.log`: full console package PASS, exit 0 (117.773s), compiled before the final helper-only strengthening. This distinction is preserved; final aggregate rerun remains pending.
