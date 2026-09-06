# a063 tasks 4.2–4.4 — independent gstack readiness review

## Verdict

**CLEAR for prepared operational readiness only.** The packet is sufficiently
specific for a human to decide whether to authorize the later operations. It
does not itself authorize, perform, or evidence installation, survey, timer
activation, attestation renewal, engine behavior, or an archive.

## Reviewed inputs

- Readiness template:
  `openspec/changes/a063-align-attestation-renewal-profile/analysis/adoption/task-4.2-4.4-operational-evidence-template.md`
  SHA-256 `9b0357a0f7e9ed2967a8cff458e77b3cb815581a1a894872f13bf985fdb6151d`.
- Deployment plan:
  `analysis/deployment-plan.md` SHA-256
  `a2bd777a5b02979d53e608d02955ed8b0a755e2df9760c521875752e17f0a0a2`.
- Current repository templates match the declared digests:
  service `0bf50972660fa14ece6953b1d43aa51fe2df407a51783c7ea0c6decb3100de96`,
  timer `0db4e0cbfada2a59de4756747f17a3d7596d80dc7a0adafb3b3460849611edde`.
- The adversarial re-review is CLEAR for prepared readiness and reports no
  operational command execution.

## Prior blocking findings verified as closed

1. **Bounded console-only survey start.** Approval must identify the existing
   console process/procedure on `$HOME/.config/tossctl`; it determines whether
   the same resolved record is already surveying, preserves it when active, and
   permits the existing console control exactly once only when absent. The
   packet forbids shell survey start, `engine run`, reset, `--record`, launch
   argument changes, and a second survey. It requires the human decision,
   actual local start time, redacted same-profile record reference, and future
   window in the completion record.
2. **Strict engine/console profile equality.** Before install or survey, the
   template requires a redacted normalized explicit running-engine config-dir
   reference that equals the normalized console target. An implicit engine
   config-dir or a mismatch stops the procedure and requires a separately
   reviewed plan. Task 4.4 repeats the equality check before accepting the
   profile's attestation/status evidence.
3. **Approval and time boundary.** Current human identity, timestamp, bounded
   scope, candidate/service/timer digests, and a future three-calendar-day
   window are required values. The expired 2026-08-29 deadline and prior survey
   records are explicitly rejected.
4. **Safe completion boundary.** Task 4.4 calls for post-window read-only
   attestation/status checks, bounded diagnostics, a non-mutating failure
   observation, and evidence of no engine restart or order mutation. The
   template holds independent review, gstack review, PM synchronization, a
   successful final `make gate`, Manager acceptance, and archive until actual
   evidence fills every pending field.

## Scope and developer-experience assessment

The packet separates preparation from mutation, names the narrow user-unit
scope, verifies candidate/template digests before installation, rejects
symlinks/drop-ins/unexpected fragments, validates backup content, and keeps
rollback bounded to the known preflight state. Its approval table and pending
completion records give the operator a concrete, auditable checklist without
requiring credentials or account identifiers. The no-duplicate survey rule and
profile-equality stop condition eliminate the earlier ambiguous decision points.

## Remaining conditions

All completion fields for 4.2–4.4 remain `pending`. No final gate has passed,
and a063 remains unaccepted and unarchivable. A human must provide the required
approval before any deployment or survey action; this review grants no such
approval.
