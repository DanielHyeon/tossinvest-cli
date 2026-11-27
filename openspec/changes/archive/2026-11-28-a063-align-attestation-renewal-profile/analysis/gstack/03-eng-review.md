# a063 Engineering proposal review

2026-09-05; follows completed CEO and Design phases. Reviewed current D5, issuer and console source, full shared resolver source, generated AST (issuer 12 branches, console 8), hard-evidence bindings, corrected maps and existing CLI/console tests. Go testing and isolated httptest/profile fixtures are the appropriate framework. No implementation test has yet run for this proposal review.

## 1. Scope and architecture

Use the existing issuer and console pipeline; add a small bounded diagnostic schema/helper, an opt-in flag, and repository service/timer templates. No new network, DB, broker adapter, scheduler or policy layer is needed. One implementer owns the code; independent review follows, with no parallel coding streams needed.

```text
Cobra --record-renewal-status OFF -> original issuer behavior
                              ON -> reject record/out override
                                 -> existing qualification/publish
                                 -> closed stage/outcome -> atomic bounded status
full resolved attestation filename + suffix ------------------------+
                                                                  |
console snapshot -> independent current attestation + status reader+
                 -> existing Usable/Reasons + separate advisory
```

The accepted path correction prevents `/shared/a.json` and `/shared/b.json` colliding. A status-write error after successful issuance must preserve the issued file but return failure. Rollout must verify binary support for the flag before human-approved installation of its unit.

## 2. Code quality and security

No free-form error parsing or duplicate qualification predicate is permitted. Add typed/stage mapping around existing results and fixed presentation text. Preserve the original issuer's branches/output when the flag is false; current console malformed-attestation error is intentionally tightened to a fixed message under D5.

The reader's 4096-byte/16-code limit, closed schema, timestamp validity, ownership/mode and regular-file checks must be enforced before trusting diagnostics. Atomic replacement is required for complete-file reads. Tests must cover symlinks and directory/nonregular input; any implementation strategy that can block on a FIFO before checking its type is a concrete review concern to verify later.

## 3. Test diagram and coverage plan

Every row below is required planned proof, not a claimed passing test. Preserve named existing tests and expand exact branch references after implementation. Frozen-clock unit tests, isolated CLI integration and real template rendering are sufficient; production execution is forbidden in these tests.

```text
ISSUER
  OFF original success/refusal/error/output -> regression comparison
  ON explicit profile -> successful same-profile issuance/status
    custom full filenames same directory -> distinct status paths
    --record/--out explicitly supplied -> refusal before issuance
    record/supervised input errors -> failed closed code/nonzero
    qualification refusal -> refused codes/nonzero/last good preserved
    attestation write error -> failed/nonzero
    status write error after success -> nonzero/good attestation intact
    later success -> prior failure replaced
READER
  valid issued/refused/failed -> fixed view + age
  missing/empty/oversize/unknown field/version/enum/duplicate codes -> unknown
  >16 reasons / future time / invalid issued expiry -> unknown
  symlink/nonregular/foreign owner/group-other access -> unknown, no blocking read
  exact 12h / >12h -> fresh / stale
  exact 72h / >72h / expired -> advisory / no horizon warning / expiry visible
  current/status expiry mismatch -> unknown; current expiry remains authoritative
CONSOLE
  attestation absent/malformed + status -> advisory still rendered
  hostile details -> never raw text or executable HTML
  warning -> unchanged Usable/Reasons except fixed malformed text
  snapshots -> zero engine, order, restart, config mutation
UNIT
  explicit config profile + boolean flag; no record/out override
  direct command exit propagated; no shell success fallback
```

Existing issuer B1-B12 and console B1-B8 remain the traceability baseline; planned diagnostic branches must not be mislabeled as existing AST branches. The earlier B9/B2 mapping issues were corrected and re-read before freeze. The companion `test-plan.md` is the artifact for implementation and isolated QA.

## 4. Performance and failure modes

The only new regular refresh work is a bounded local file read and fixed formatting. No external requests, retry loop, timer reset or broker rate-budget consumption is added by attestation renewal. Do not re-run the full soak summary solely for status classification; use existing typed results/stages.

| Failure | Handling | Planned test | Visibility |
|---|---|---|---|
| Refusal | Nonzero, preserve last good | CLI refusal | Fixed codes |
| Input/publish failure | Nonzero | CLI stage failures | Failed/unknown |
| Status failure after issue | Nonzero, retain issue | Write-failure test | Unit failure, stale/unknown |
| Hostile status | Reject | Reader matrix | Unknown |
| Stale success | Reject freshness | 12h boundary | Age + warning |
| Missing current file | No invented validity | Template matrix | Unknown expiry + diagnostics |
| Two configured basenames | Full filename suffix | Collision regression | Correct profile |

## Completion and limits

Architecture finding: one accepted path collision correction. Quality finding: one accepted malformed-attestation redaction correction. Test review: diagram produced, all proposed paths have required tests, execution pending. Performance gaps: zero in the design; implementations still require review. Critical silent plan gaps: zero after D5; no new TODOs; scope retained; one sequential coding lane; no cross-model voice ran (all six Eng consensus dimensions N/A).

Original `base-commit.txt` remains unchanged. Whole-change Function Logic Map validation is still blocked by inherited changes since that immutable base; no re-freeze or checker bypass is approved. This limits final gate/acceptance, not the scoped design's readiness for isolated implementation. Operational approval and actual qualifying-day evidence remain separate acceptance requirements.

## Final D5 refinement

Manager added a single typed internal issue evaluation that preserves every public `Summary.Evaluate` predicate, message and decision. Carry closed refusal codes from that same evaluation through a typed error/result; do not rescan or parse textual reasons. Require pre-edit AST/FLM for the existing Evaluate function, RED/GREEN tests comparing exact legacy messages and decisions, and caller regression proof. This is an in-scope compatibility-preserving refinement, not a qualification policy change.
