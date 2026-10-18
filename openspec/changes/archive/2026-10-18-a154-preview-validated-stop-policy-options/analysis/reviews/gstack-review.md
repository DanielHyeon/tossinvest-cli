# a154 gstack implementation review

Current verdict: scoped code/security/UX review PASS. No unresolved actionable blocker remains. Codex-host gstack review, not Claude/cross-model execution. Read current complete changed/new source and tests, canonical design/spec, static gate helper and writer test/render receipts. Reviewer executed no tests, app, browser or gate and made no production edits. Full frozen gate and lifecycle remain pending. Only this report edited.

## Findings and corrective rereview

Market error originally appeared separately without aria-invalid/aria-describedby on its select. Manager and this reviewer identified the same accessibility gap. Latest template connects the select to hyp-error-market, and dedicated rendered HTML assertions verify those attributes and the message. Resolved. Manager also added required-input guidance/manual-unverified low/ATR wording, explicit zero-quantity state and unknown currency text. No new apply authority or inferred defaults were introduced.

Requested numerical test strengthening rather than alleging a numerical defect: final tests independently cover 0.0078125% half-up to 0.007813%, positive below-display-unit percentage, and share cap 1. Existing behavioral RED records fractional share syntax 1.0 wrongly accepted; current integer syntax validation rejects it. Raw failure is preserved.

## Numerical and error-path assessment

Inputs use bounded unsigned decimal syntax, 32 characters and eight fractional digits; positive budget/caps and one-million integer share maximum. Common invalid geometry blocks every arm; missing/invalid low and ATR affect only their own arm. Sizing failures preserve valid stops/distances. Grid-floor handles negative intermediates mathematically before invalid stops are rejected. Entry is never silently rounded.

Fixed 0.993, low floor minus tick and ATR1.5 floor formulas match the contract. Every stop is positive, below entry and within caller bounds. RiskBasedQuantity is called only after geometry validation, then exact share/notional caps apply. The shared parser is shielded by strict local input checks. Integer quantity zero remains a valid no-position state.

Formatting via eight decimal places is exact for these outputs: entry/tick have at most eight decimal places, stop is an integer multiple of tick, distance is their difference, and risk/notional multiply by integer quantity. Thus the stop string passed into the reused helper loses no precision. Percentage alone is rounded half-up, and subunit positive values are explicitly marked. This invariant must remain true if accepted scales or quantity types change later. No targets, costs, filled risk or account-loss guarantees are calculated.

## HTTP, authority and static guards

The route is wrapped by no-store outside existing session0 and bounded CSRF/origin POST middleware. Existing trusted-network behavior is preserved, not misrepresented as unconditional authentication. Body size is capped at 8192 bytes before form parsing. Handler rejects query parameters and unknown/duplicate PostForm keys, including csrf. Numeric values are not accepted from merged URL forms. Template escaping remains standard html/template; no unsafe HTML, logging, subprocess or caller-controlled filesystem path was added.

Handler calls only the pure calculation/render path, not optimization Preview/Apply, capability generation, legacy save, holdings or account/market fetching. Existing registry and protection selection remain unchanged. No new persisted policy or default is introduced. Real temporary optimization DB test checks unchanged snapshot/audit and counts in control/snapshots/candidates/applications/audit after calculator and forged-key rejection; fake-command tests also verify preview/apply/save/holdings counters stay unchanged.

Static route detection permits only the named c.hypotheticalNoStore wrapper outside c.session0. Tests reject missing-session, arbitrary-wrapper and other-receiver variants. State-changing CSRF assertions remain, with one explicit compute-only path added. This is not a generic recursion exemption for arbitrary wrappers.

## UX and evidence limits

Title and adjacent warnings identify manual hypothetical arithmetic, non-live/unverified inputs, KR-only Parker support, excluded fees/tax/gap/slippage/FX and no total-loss cap. No apply button/token is rendered in the calculator. Required input explanation, preserved values, per-field errors, keyboard focus, unknown cells and zero-position text are present. Results use a labelled, keyboard-focusable horizontal scroll region and currency context.

Writer focused-final5.log records package PASS. Read exact KR/US numerical oracles, floor/range/malformed inputs, partial-input isolation, caps, half-up/subunit branches, unauthenticated/CSRF/query/duplicate/oversized/method refusals with no-store, real DB no-write and forged-key assertions. This reviewer did not rerun them. Final git diff whitespace check passed during review.

Read render-observations.md: writer observed actual isolated httptest templates in Chrome, KR/US expected numbers, 390x844 viewport with document width375 and table-only scroll, keyboard Tab/3px focus, linked market error and empty browser warning/error log. Screenshots were inspected in that writer's tool output but not saved; this reviewer did not inspect screenshots or perform an independent browser run. Observations precede only the final unknown-currency text clarification, covered by current template tests. Source-only impeccable scan explicitly omits parsed-page rules and reports two pre-existing side-border CSS findings; it is not evidence of a full accessibility audit. These limits are preserved rather than claiming blanket UX certification.

## Scope and remaining work

Canonical scope is the approved calculator, not validated strategy selection or CAS apply. Base c14642e979e07d16c19501fa93dd448da52ecff7 is recorded. Existing high-risk route/static functions have analysis artifacts; inspected branch map records compile/history evidence honestly rather than claiming new behavioral RED for every pre-existing branch. Full function-evidence validation, frozen sync/check and aggregate gate remain required. No live service restart, operating settings or existing research evidence was changed by reviewer.

## Corrective landmark rereview

Manager subsequently found nested main elements during isolated rendered inspection; the prior manifest is superseded. Shared head opens main and foot closes it. Latest hypothetical-result removes its redundant main pair and adds a page h1. Current HTTP regression requires one opening and one closing main, with actual RED and focused-final6 PASS preserved. Source hashes match writer: template ed2c4d2de556f5c78f71da5dbfb77c02f6166ca9f07097e81728b0808bd6a643; HTTP tests ab4c6f63a104e575e2910e0f3e7af3ff18ca7790f4d663cc5110a4ab943f5178. The only other changed manifest entry is render-observations.md documenting Manager inspection and correction. This defect was missed in initial source UX review and is now resolved. No independent browser rerun by this reviewer is claimed. Scoped PASS restored for current files; full gate pending. Earlier render observations describe their historical snapshot, not the final landmark patch.

## Final preparation metadata refresh

After the landmark correction, only tasks.md changed among prior reviewed entries. Task 3.2 initially implied sync/check execution despite preparation status; corrected to execution preparation before this refresh. Manager verification and review records were read and added. Implementation hashes and receipts remain unchanged; full gate and lifecycle still require actual completion receipts. Scoped PASS remains; no new unresolved issue.

## Gate attempt 1 repair review

Full gate attempt 1 failed at make test/stage 7 with exit 2; raw failure is preserved. Existing whole-page policy-only UI assertions conflicted with the authorized independent calculator, and the table-scroll helper expected its exact class. Current patch preserves the original four policy assertions and filters only one verified hypothetical section: exactly one compute POST form, exact allowed fields/counts, CSRF hidden control, single submit button, no additional form/action overrides, links, scripts, event handlers or contenteditable controls. Forged setting_key, apply-formaction button and extra settings form mutations explicitly fail. Thus the new scope is validated before removal rather than silently excluded from coverage. Existing operating option/action/read-only checks remain intact.

Template moves hypothetical-results class onto table, retaining table-scroll as the wrapper class and the same nowrap CSS effect; adds explicit target/RR-not-computed wording. No calculator/HTTP/security implementation changed. Focused repair receipt exits 0. Full console package receipt now records PASS in 117.773 seconds, EXIT_CODE=0. That run preceded the final helper-only strengthening; gate-repair-focused.log covers the final helper. Full gate remains attempt 1 FAIL / rerun pending. Scoped repair code/test review PASS, not aggregate validation success. Prior source manifest changed only templates_optimization.go; two repaired test files and focused/failure receipts added below. Final receipt-only refresh adds console-repair.log; all 26 prior manifest entries were independently rehashed unchanged.

Manager gate-attempt/repair metadata was subsequently reread and rehashed; no implementation or other evidence changed.

## Reviewed manifest

Manifest SHA-256: `9e5fb6a2461eab788a6db5616c1686fdd766cc019efe0b93055e465abbaba6c1`

```text
f72b2087e483d4f986bb85d764a3a0e742e52860a120e8cb9da6adc363676667  internal/console/a154_hypothetical.go
d6e02de880e0173a3c104db8f6919a59920b29c533295d86fd59696b7e029f6c  internal/console/a154_hypothetical_http.go
d80bfdc7c2059576093545c52b8a4fc18024feb7266afc9898c424384da92541  internal/console/a154_hypothetical_test.go
ab4c6f63a104e575e2910e0f3e7af3ff18ca7790f4d663cc5110a4ab943f5178  internal/console/a154_hypothetical_http_test.go
6834ca3f28e5a6cefc239aafbeb94edb754b82ffe8361843fdaae28e5ebb963b  internal/console/a154_route_guard_test.go
0239fcd72e977f7f8b702d559c4a91105185baa15de307660f9bc778114aaf5e  internal/console/console.go
eb4c5ae2eb16af0f5f95e6101c6a1e1d9f970150b6d0caac09a188de785b777d  internal/console/templates.go
c3f3344afd7c5e0de09b6812f697a6bee3409868603632657cbce6eaf59ab69c  internal/console/templates_optimization.go
1eda5647fa47d9cef7dd129859056e12ab026e94c92b44507a1d708b52f7f3df  internal/console/static_test.go
9bef3ec35a7c8702abad4f2ccd517a27b98f472044802b0b857259e200563f60  openspec/changes/a154-preview-validated-stop-policy-options/proposal.md
01e5d83b11f0e9878778193915d25cc2a245fe6b393e048ae36735f0fdc5d6d9  openspec/changes/a154-preview-validated-stop-policy-options/design.md
8a1e21db2e6dbcc70cd9cd7e922ce101ebe47513495e53ac96f24a09287a5987  openspec/changes/a154-preview-validated-stop-policy-options/specs/strategy-optimization/spec.md
3e93399264928fe5ddad5c3e78a84c042638f4af0cb6c86a34b9e7e3221e4313  openspec/changes/a154-preview-validated-stop-policy-options/tasks.md
fd33b3b6e17620efd4aca87de61b58558dad060f3013db7eb1121059df9ce8a9  openspec/changes/a154-preview-validated-stop-policy-options/base-commit.txt
02fe2945e122aae75bf3a2a41f6ff9ed6b2dfbac7bde2b35008c8a2c096eb70e  openspec/changes/a154-preview-validated-stop-policy-options/analysis/verification/behavior-red.log
ffb2292debe69ad30a8b46c93741ab6d7540974c1a958576895b34df0d029991  openspec/changes/a154-preview-validated-stop-policy-options/analysis/verification/focused-final5.log
51af2f0108ff1d79d76431841220a6c3c0e87b6fdb9a09046d9bf0307cb5d28c  openspec/changes/a154-preview-validated-stop-policy-options/analysis/verification/impeccable-source-scan.log
b7f4f91fcc56fde39fa10dbc3334590552abfe1a3af872781c3b3766fe43b69b  openspec/changes/a154-preview-validated-stop-policy-options/analysis/verification/render-observations.md
8090448e487f8021e78a33889c3f3e3cbf4227a8d710fc24e5eb6fba29dd129b  openspec/changes/a154-preview-validated-stop-policy-options/analysis/verification/main-landmark-red.log
7b51929cad153cb99985667e98270bbcae40a139251ba0d44bf2959a2d139b47  openspec/changes/a154-preview-validated-stop-policy-options/analysis/verification/focused-final6.log
5e3e7fd58c07e71f721ef9c4d6d937644ce3fb3e41f5d0f04b1960abca7d5277  openspec/changes/a154-preview-validated-stop-policy-options/analysis/reviews/manager-verification.md
de6aed42e5a2a9c80edf186a2cc07d980a398edbd0fe896c6ab685a81cdef2f7  openspec/changes/a154-preview-validated-stop-policy-options/review.md
e806f159ccd115d62e550d42f321aa422057389bbf60eaab531524d996264b8f  internal/console/optimization_ui_contract_test.go
3dc48c10f104504299fdfb1e5608a482a460f523e32dfe0f29a979dcb4a804bc  internal/console/a154_contract_scope_test.go
42dd294db16b80abbc23e5ab9ee6d8827dbdddf72406d959ba108c256f437aa1  openspec/changes/a154-preview-validated-stop-policy-options/analysis/verification/gate-attempt1-fail.log
3cf448a33c0cd0059c3f5fdbae1caf6a84a103c8108cdfaf80245843fd3db377  openspec/changes/a154-preview-validated-stop-policy-options/analysis/verification/gate-repair-focused.log
0e086a707473ab788d39ed04b9916613910f8d1b461765634adbc1963f5edb3f  openspec/changes/a154-preview-validated-stop-policy-options/analysis/verification/console-repair.log
```
