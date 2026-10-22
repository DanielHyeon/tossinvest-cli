## 1. Design and readiness

- [x] 1.1 Manager design review: approved offline scope, explicit hypothetical risk bound and non-authority boundary; independent and gstack findings resolved.
- [x] 1.2 Freeze current source/CodeGraph and selected AST evidence; map native request/result types, guards and affected tests. No production edit before READY.
- [x] 1.3 Strict OpenSpec validation and Story/Feature/registry consistency; record validation scope and outstanding blockers.

## 2. Terra implementation — implemented locally, focused verification passed

- [x] 2.1 Verify focused regressions for native proposal reuse and refusal preservation across three lanes and KR/US follow-on legs; record actual RED/mutation evidence and its limits without claiming a complete test-first sequence.
- [x] 2.2 Minimal typed adapters and bounded integer hypothetical-budget assessment; no new allocator, parser, execution seal or runtime wiring.
- [x] 2.3 Verify unknown/missing/scope/amount boundaries, budget arithmetic, input immutability and weekly reservation requirements.
- [x] 2.4 Verify dependency/type boundaries against broker, journal, gateway, engine and operational side effects.

## 3. Review and completion — local reviews passed, publication pending

- [x] 3.1 Independent review, gstack review, disposition and Manager verification of exact diff and focused test outputs.
- [x] 3.2 Preserve existing evidence and operational configuration; document untested runtime/real-account scope.
- [x] 3.3 Confirm explicit commit/push/archive authorization and lifecycle prerequisites. Full-suite/full-gate execution is prohibited by this request. Actual commit and remote publication are verified separately after execution; this checkbox does not claim an already successful push.

## Working-tree status — 2026-10-20

Code and tests are implemented locally. Completion checkboxes are included in
the lifecycle commit that carries the code and evidence together.
Design/source mapping and local implementation checks are recorded in `review.md` and
`analysis/implementation-evidence.md`.

Task 2.1 evidence is qualified: focused regressions pass and bounded-min mutation
fails, but initial fixture failures are not product RED evidence and no full
test-first sequence is claimed. Real-account currentness and live authority
remain unverified. The original non-nil FX and weekly ordinal 3–7 fixture gaps
are addressed only within the synthetic scope in section 4 below.

## 4. Additional scoped integration coverage — authorized 2026-10-20

- [x] 4.1 Extend tagged weekly fixtures through native prefix transitions; verify KR/US ordinals 3–7, risk totals, reservation rejection and input immutability.
- [x] 4.2 Exercise valid non-nil native FX inputs, before/after capture equality, independent mutation detection and invalid FX refusal.
- [x] 4.3 Verify native stale-input refusal and unverified currentness/no-authority markers on both positive and refused results; no live-account attestation claim.
- [x] 4.4 Run only affected integration tests and existing direct fixture consumer tests; preserve old evidence, append fresh logs and obtain independent/gstack/Manager review.

Local status: 4.1–4.3 implemented and focused tests passed; exact numeric assertions
and reservation-integrity wording corrected during 4.4 review. Historical
ordinal-2/non-nil-FX limitations above describe the previous validation pass;
see `analysis/additional-integration/implementation-evidence.md` for the added
weekly coverage. Actual account currentness remains unverified. These checkboxes
are committed with the code/evidence as required by WORKFLOW.
