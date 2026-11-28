# a063 post-implementation gstack review

Review basis: session-start `e65e394b` to current tracked changes plus all seven untracked source/test/unit files. This scoped comparator does not replace immutable `da80ce31b6a1ab5d443016768f970a82bab102db` or its completion checks. Only a063 implementation and directly relevant callers were reviewed. The separate Terra adversarial review completed code-clear before this pass began.

## Findings

1. **INFORMATIONAL, confidence 10/10 — `internal/console/templates.go:596`.** `{{if .RenewalAttemptedAt}}` tests a `time.Time` struct, which remains truthy when zero. Missing, invalid, or expiry-mismatched diagnostics can consequently print an invented year-0001 last attempt. Require `IsZero` guard and HTML regression for those unknown states. Fixed by Terra and independently re-read/tested.
2. **INFORMATIONAL, confidence 10/10 — `internal/console/data.go:267-270`.** `v := attestView{Path: c.opts.Attestation}` followed by the blank-path return leaves the new diagnostic state empty. `Console.New` accepts an empty Attestation option. Require unknown initialization and blank-path regression. Fixed by Terra and independently re-read/tested.

**Final verdict: code clear for this scoped review.** Two informational findings were fixed; zero unresolved findings and zero critical findings. This is not final acceptance.

## Review coverage

- Read the complete `/review` skill (shared preamble reconciled with previously fully read autoplan), mandatory checklist, testing/maintainability/security/performance/API-contract/red-team and design-lite checklists. Applied critical then informational passes, followed by the specialist perspectives in this reviewer context. No source edits made.
- SQL, database migration and LLM boundaries: not present in this diff. Service execution has literal argv and no shell/fallback; template is source-controlled and explicitly profile-bound. Test stub proves direct argv exit 23 propagation, not systemd runtime acceptance.
- Qualification: compared every changed predicate/message with the base diff. `evaluateIssues` is the single policy evaluation; public `Evaluate` preserves successful nil reasons, and `IncompleteError.Unwrap` preserves `errors.Is(ErrIncomplete)`. Refusal codes are copied, deduplicated, sorted, and never inferred by parsing raw text. No broker/order/engine control path added.
- Persistence/security: full filename suffix avoids same-directory basename collision; recording uses one resolved path snapshot. Default OFF bypasses status work. Unix reader opens with no-follow/nonblocking flags, verifies regular file/mode/current UID on that descriptor, and limits reads to 4097 bytes. Closed enum/schema and time checks fail to unknown. Writer uses 0600 temporary file, propagates failures, syncs/closes before rename, then syncs directory. Existing attestation remains independent of status-write failure.
- UI/API: current attestation supplies expiry; issued-status mismatch becomes unknown. Stale is strictly over 12h; warning boundary is inclusive 72h. Renewal values do not enter `Usable` or gate reasons. Fixed malformed-attestation text and normal escaped template interpolation avoid raw renewal error disclosure. New `role=status` text uses the existing surface; no new action or approval controls.
- Platform/distribution: secure reader has Linux/macOS implementation and explicit unsupported/unknown fallback elsewhere. Linux user-systemd deployment remains approval-bound; no installation or reload performed. Existing release-build blocker is recorded separately.
- Test review: inspected new and changed tests for issued/refused/OFF behavior, actual custom path, single resolver invocation, override rejection, status-write failure preserving issuance, temporary-file failure preserving prior status, future/duplicate/oversize/offset/empty-code rejection, mode/symlink/FIFO, stale/mismatch, exact boundaries, fixed error text and advisory usability. Existing qualification tests plus nil-result test cover policy compatibility; exact message preservation additionally verified in source comparison.
- Documentation: operations instructions describe the opt-in flag, current resolver precedence, binary/unit pairing and rollback approval. No related root TODO identified. a063 test-plan and core review dispositions were read.

## Limits and completion boundaries

This is an actual independent Codex reviewer applying gstack checklists; specialist perspectives were performed in one context, not separately dispatched models. Separate Terra adversarial work is identified as such. Claude adversarial, external Codex CLI structured review, Greptile/PR integrations, browser visual QA and cross-model consensus were not run. No telemetry, installs/upgrades, routing edits, memories or external messages were performed. No synthetic quality score or external gate result is claimed.

Tests and compilation demonstrate only their exercised cases; no claim of exhaustive fault injection, concurrent writer stress, native Windows/macOS runtime, real unit execution or operational renewal acceptance. Controlled expiry-warning regression proof in `analysis/regression-proof.md` establishes mutation sensitivity, not retroactive original RED-first work.

Immutable-base coverage, unrelated all-change validation (`a119`), existing full release cross-build `productionFileUID`, and operational approval/three actual qualifying days/fresh attestation remain independent acceptance concerns. This review cannot approve deployment, archive, or final completion.

## Independent final verification

- `go test ./internal/soak ./cmd/tossctl ./internal/console -run 'Test(Renewal|LoadRenewal|SaveRenewal|SoakAttest|Evaluate|BuildAttestation|Dashboard.*Renewal|DashboardRedacts)' -count=1` passed, exit 0 (soak 0.020s; CLI 0.269s; console 0.073s).
- After the UI corrections, `go test ./internal/console -run 'Test(Dashboard.*Renewal|DashboardRedacts|RenewalWarning)' -count=1` passed, exit 0 (0.058s). `TestDashboardMismatchedRenewalStatusAndBlankPathAreUnknownWithoutZeroTimestamp` proves rendered mismatch unknown/no year 0001 plus blank-path unknown/not usable; `TestDashboardRedactsMalformedAttestationContents` now also checks no zero timestamp.
- `git diff --check` passed for scoped tracked implementation files. See `implementation-reviewed-digests.json` for exact final source/test/unit/operations bytes, including untracked files.
- Source fixes verified: `attestView{Path: c.opts.Attestation, RenewalState: "unknown"}` and `{{if not .RenewalAttemptedAt.IsZero}}`. The retained known stale timestamp behavior and attestation usability were not altered by those fixes.

## GSTACK REVIEW REPORT

Code/security/QA/design-lite checklist pass complete in this context. Findings: 2 fixed, 0 unresolved. Recommendation: continue Manager acceptance work because the concrete unknown-state rendering defects are corrected and the scoped tests pass. Preserve original-base, broader validation/platform and operational blockers; do not archive or deploy based on this report.
