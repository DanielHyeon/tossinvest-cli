# Review — a063-align-attestation-renewal-profile

- Date: 2026-08-03
- Stage: proposal freeze; implementation not started
- Voices: maintainability/release review + adversarial operations/security perspective
- Evidence: a060 `issues.md` I7, installed user-systemd unit/timer read-only inspection, current path
  resolution in `cmd/tossctl/soak.go`

## Findings and disposition

- **Accepted**: bind the timer to one explicit config profile instead of copying or relabeling legacy
  evidence. One profile authority removes the input/output asymmetry without weakening attestation.
- **Accepted**: make the service definition repository-backed and test its argv/failure behavior. The
  retired untracked soak autostart artifact demonstrated the maintenance cost of external-only logic.
- **Accepted with gate**: preserve non-zero renewal status and surface impending expiry on an existing
  operator surface. The warning is advisory and cannot stop a running engine or relax startup checks.
- **Required**: human approval before installing/reloading the external service or starting the
  multi-day production survey. No operating toggle or live-order command is authorized by this plan.
- **Required deadline**: collect at least three qualifying survey days and verify a fresh attestation
  before the current one expires on 2026-08-29.

## Scope decision

The plan is ready for implementation after the independent Manager verifies the artifacts. No runtime
code, external service, survey process or operating setting was changed while creating this change.

Function Logic Map: not-applicable — this commit creates planning artifacts and amends comments/contracts
only; implementation tasks will create any required analysis before editing existing functions.

## 2026-09-05 continuation — proposal freeze reopened

- Manager: Astra; implementation/evidence teammate: `a063_impl` (Terra); separate adversarial reviewer:
  `a063_adversary` (Terra). Post-implementation gstack review and Manager acceptance remain pending.
- The preceding 2026-08-03 review is historical. Its deadline and readiness statement do not establish
  current operational state or authorize deployment.
- Accepted adversarial findings: expired deadline, incomplete current path/AST evidence, unresolved
  operator warning surface, missing concrete approval record, and final gate ordered before its own
  prerequisites.
- Disposition: proposal/design now distinguish historical observations from current evidence; tasks
  preserve three actual consecutive qualifying survey days, record deployment approval scope and unit
  digest, and place the final gate after prerequisite engineering and operational evidence.
- Current freeze status: pending current CodeGraph/AST evidence and selection of the existing operator
  warning surface. No production implementation has been released by Manager at this stage.
- The earlier Function Logic Map exemption applies only to that historical planning commit. Any current
  claim about existing function branches or implementation must use generated AST evidence first.
- No archive, live survey activation, service installation/reload, operating-toggle change or live order
  is authorized by this review entry. Explicit operational approval is still required by the spec.

### Independent adversarial design disposition

- `a063_adversary` reviewed the revised D5 contract and generated pre-edit evidence, then approved the
  bounded non-live design after correction of inaccurate AST branch labels.
- Accepted: full-attestation-path status filename to prevent same-directory profile collisions;
  bounded closed diagnostics; read integrity checks; explicit 72-hour inclusive warning and strictly
  over-12-hour stale boundary; default-OFF preservation; separate advisory and startup-denial state.
- Required RED/GREEN cases include malformed attestation error redaction, default/custom profile
  bindings, same-directory different-basename isolation, future/stale/invalid diagnostics, exact
  threshold boundaries, successful renewal replacing failure, and status-write failure remaining
  nonzero without corrupting attestation.
- Current evidence: `analysis/codegraph-hard-evidence.md`, `analysis/path-resolution/`,
  `analysis/function-logic/`, and `analysis/sanitized-operational-evidence-2026-09-05.md`.
- gstack proposal review remains pending its final report. Post-implementation adversarial and gstack
  reviews have not run yet. `issues.md` I1/I2 remain final-acceptance blockers.

### gstack proposal review and Manager implementation release

- Completed sequential CEO, Design, Eng and DX perspectives using the gstack `autoplan` and phase
  skills. Evidence: `analysis/gstack/proposal-review-report.md` and its linked phase/test-plan artifacts.
- All four scoped design verdicts are clear after accepted corrections. This was one independent
  Codex review context, not four different models; outside Claude/Codex CLI voices and browser/mockup
  verification were not performed. No cross-model consensus or runtime/UI proof is claimed.
- Manager reviewed the revised contract, current generated maps and review dispositions. Release:
  Terra `a063_impl` may implement and test this bounded non-live design, preserving the original base.
  Every additional existing function, including qualification evaluator/issuer refactoring, requires
  its generated pre-edit map before editing, and refreshed maps plus real tests after editing.
- Coding/test ownership: CLI renewal diagnostics, shared typed evaluation/status persistence, console
  read-only warning, source-controlled service template, focused tests, operations documentation and
  a063 PM synchronization. Core OpenSpec decisions remain Manager-owned.
- Separate implementation adversarial review must finish before post-implementation gstack review;
  Manager independently assesses resulting diff/test evidence afterward.
- This releases engineering work only. Final gate, operational acceptance and archive remain blocked
  by I1/I2 until legitimately resolved; no service, binary or console deployment is authorized here.

### Manager interim implementation findings — unresolved until tested

1. `SaveRenewalStatus` scoped `err` inside the chmod/write block, losing those failures before rename.
   Require propagated failure, no replacement after write failure, and meaningful failure-path proof.
2. `LoadRenewalStatus` checked a pathname with Lstat and then used unbounded `os.ReadFile`. Require a
   bounded read from a verified descriptor and prevention of symlink substitution; a size precheck
   alone is not a bounded-read guarantee.
3. `Summary.Evaluate` changed a successful nil reason slice into a non-nil empty slice. Preserve the
   established result shape as part of default-OFF compatibility.
4. The initial template did not yet display attempt age/expiry horizon, and stale handling discarded
   a known attempt timestamp. Verify visible age/horizon and stale/unknown semantics against D5.

These findings were sent to the Terra implementation owner. They do not replace the separate
post-implementation adversarial review or gstack review.

- Terra subsequently reports fixes for findings 1–4 and focused command
  `go test ./internal/soak ./cmd/tossctl ./internal/console` exit 0. Verification includes an injected
  temporary-file failure preserving prior status, descriptor-based bounded reading, nil-result
  compatibility, visible age/horizon, and an isolated service stub whose exit 23 is preserved.
  Manager/independent final verification remains pending.
- Additional Manager finding: untagged `syscall.Stat_t`/`O_NOFOLLOW` code conflicts with the Windows
  release target declared in `.github/workflows/release.yml`. Platform separation and supported-target
  cross-compilation are required before acceptance; default-OFF must remain buildable on all targets.

### Separate adversarial implementation review — open

- The reviewer independently ran focused soak/CLI/console tests, the renewal-status race test,
  strict a063 validation and diff whitespace checks successfully. No live-order/toggle/engine side
  effect or raw renewal diagnostic rendering path was found in this pass.
- Blocking defect: failed/refused diagnostics with zero reason codes were accepted as valid. Require
  nonempty closed codes and regression coverage; UTC representation must also be consistent with D5.
- Required missing tests: actual custom configured attestation paths and same-directory different
  basenames; exact/just-beyond 72h and 12h boundaries; CLI issuance followed by status-write failure;
  default-OFF no-file and exact legacy compatibility; future/offset/oversize/duplicate/empty-code
  hostile schema cases; malformed-attestation HTML error redaction.
- Terra must address these findings and receive the adversary's re-review before post-implementation
  gstack review. A passing subset does not close omitted contract cases.
- Earlier fixture/compilation failures are not pre-change behavioral RED evidence. That proof remains
  unestablished until an isolated pre-change or controlled regression run actually demonstrates it.

- Terra added the requested boundary/custom-path/hostile-input/CLI failure tests. Controlled regression
  proof is now in `analysis/regression-proof.md`: suppressing the expiry warning only in a temporary
  Go overlay made the exact-72h test fail; the same test passed without the overlay. This establishes
  regression sensitivity, not retroactive original RED-first development.
- Adversarial re-review found a further blocking race in configuration binding: recording mode selected
  the diagnostic path, then reloaded/resolved the attestation path again before publication. Reuse the
  first resolved path throughout the recording attempt; keep legacy OFF resolution behavior unchanged.
  A single-snapshot regression and another reviewer verification are required.

### Adversarial implementation re-review — code clear

- `a063_adversary` independently verified the single-snapshot fix and re-ran the targeted CLI, soak
  and console suites successfully. Covered fixes include path/default-OFF behavior, status schema and
  file safety, exact 72h/12h boundaries and malformed-attestation redaction.
- Verdict: no remaining concrete code defect found in this scoped pass. This is code-level clearance,
  not final gate, operational acceptance or archive approval.
- Implementation-scope inventory against session-start `e65e394b` identifies five changed existing
  production functions and four existing test functions across six files. This inventory is advisory
  scope evidence only; it does not replace the immutable `da80...` completion comparator.
- Post-adversarial gstack code review has now been requested on the same scoped implementation,
  including untracked source/template/test files. Its final verdict is pending.

### gstack post-implementation review — correction pending

- The structured/specialist review found that testing a `time.Time` value directly in the template
  renders its zero value as a present time. Missing/invalid/mismatched diagnostics can consequently
  display year 0001. Require an explicit `IsZero` guard and rendered-HTML regression.
- Also verify an empty configured attestation path produces the specified unknown diagnostic state.
- The implementation owner must correct this finding after its current test sessions finish, then
  rerun affected tests and obtain gstack re-verification. Earlier code-clear applies to the reviewed
  earlier snapshot; it does not close newly identified findings.

### 2026-09-06 Manager continuation disposition

- The pending UI findings above were corrected by the implementation teammate and independently
  retested by gstack. `analysis/gstack/implementation-review.md` records code clear and the exact
  fifteen-file digest manifest. Manager recomputed all fifteen values with zero mismatches and
  inspected the refused mobile and unknown desktop synthetic browser screenshots; these are not
  production deployment evidence.
- Terra continuation verifier completed eleven directly validated function-analysis bundles. The
  frozen-base checker requires six scoped test bundles, distinct from the earlier session-only
  inventory of four. Three base ASTs were generated retrospectively; no pre-edit provenance is
  claimed. Global analysis still exits 1 with 327 missing rows outside this scope.
- Separate Terra continuation adversary reviewed the deployment proposal and found three
  procedural defects. All were corrected and re-reviewed: validated backup before timer stop,
  repeated mutable-state checks, and fail-closed service-state inspection. A subsequent independent
  gstack continuation review verified the corrections. See `analysis/adversarial-continuation-review.md`
  and `analysis/gstack/continuation-review.md`; specialist perspectives are not separate model voices.
- Manager checked the candidate binary and both unit-template hashes against the deployment
  proposal: all match. No installation, timer activation, survey restart, engine restart or trading
  mutation was performed. Explicit operational approval and same-profile acceptance remain pending.
- Final acceptance is withheld: global function evidence, all-change validation and the required
  operational proof remain unresolved. Current administrative command outcomes are recorded in
  `analysis/verification.md`. Tasks 4.1–4.5 stay open, and neither archive nor the next a0xx may proceed.
