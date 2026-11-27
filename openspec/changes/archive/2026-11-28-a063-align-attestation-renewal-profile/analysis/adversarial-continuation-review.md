# a063 continuation adversarial review

**Review scope:** existing implementation-review evidence, current Function Logic
Map evidence, and `analysis/deployment-plan.md` only. No product source, runtime
service, account, survey, engine, or trading control was changed.

## Verdict

**BOUNDED PLAN CORRECTIONS CLEAR; OPERATIONAL READINESS NOT ESTABLISHED.** The
three transactional findings below were corrected and independently rechecked.
The scoped code-review evidence remains a code-level result only; it neither
supplies operational proof nor closes the immutable-base completion gate. The
procedure still requires the documented same-profile evidence and explicit human
operational approval before it may be run.

## Findings

1. **RESOLVED (was HIGH) — the activation block could leave the known-good timer disabled without
   a recoverable snapshot.** It runs `systemctl --user disable --now` before it
   copies the three supported regular files to `$backup` (deployment plan lines
   88-90). A subsequent copy, install, verification, reload, or enable failure
   leaves the timer disabled and has no newly created rollback input. Capture and
   validate all rollback artifacts before changing the timer state, then stop the
   timer only after the snapshot is complete. **Resolved:** the activation block
   now copies the three bounded regular-file targets and checks each backup is a
   non-symlink regular file whose bytes match the source using `cmp -s` before
   `disable --now`. It directs the operator to the documented rollback block for
   every post-stop failure.

2. **RESOLVED (was HIGH) — activation did not revalidate the state that its preflight bounds.**
   The preflight rejects symlinks, absence, disabled/inactive timers, and drop-ins
   (lines 50-63), but the later activation block repeats only digests (lines
   77-87). State can change between the approval/evidence gate and activation.
   Repeat the regular-file, enabled/active, fragment-path, and service *and*
   timer drop-in checks immediately before mutation; fail closed on any mismatch.
   **Resolved:** immediately before snapshot/mutation it now repeats regular-file,
   exact `FragmentPath`, both-unit no-drop-in, and enabled/active timer checks.

3. **RESOLVED (was MEDIUM) — service-quiescence check treated an inspection error as inactive.**
   `systemctl --user is-active --quiet tossos-attest.service && { ...; exit 1; }`
   (line 89) continues for both an inactive service and a failed inspection. Use
   an explicit three-way result that accepts only the documented inactive exit
   status and stops on every other error. **Resolved:** it now assigns
   `ActiveState` under `set -e` and accepts the literal value `inactive` only;
   an inspection failure or any other state exits before mutation.

## Evidence boundaries confirmed

- The persisted comparison base is still
  `da80ce31b6a1ab5d443016768f970a82bab102db`; it resolves and is an ancestor of
  current `HEAD`. It was not reset for this review.
- The regenerated Function Logic Maps describe themselves as a 2026-09-06
  post-hoc alignment of current AST and test evidence. `post-edit-ast-evidence.md`
  explicitly marks its earlier snapshot as historical and non-current. These
  documents therefore do **not** invent pre-edit provenance, but they also do
  not repair the inherited immutable-base/global-checker blocker.
- Follow-up verification corrected the earlier coverage-status contradiction:
  direct validation now passes for all eleven scoped bundles (five production
  and six checker-derived test functions). The global checker still exits 1 with
  327 missing-evidence rows outside this a063 scope. This is valid scoped
  current-structure validation, not historical pre-edit provenance or a final
  gate result.
- `analysis/gstack/implementation-review.md` accurately limits its code-clear
  conclusion to the scoped implementation and states that it cannot approve
  deployment, archive, or final completion. It must continue to be read together
  with the unresolved operational and immutable-base conditions.
- Read-only checks in this review: the focused soak/CLI/console test selection
  passed; `systemd-analyze --user verify` returned 0 while warning that the
  repository templates are mode 0755; `git diff --check` passed. None proves a
  real user-systemd run, same-profile attestation, or three qualifying days.

## Required disposition

Keep final acceptance and archive blocked. Transactional review clearance does
not replace same-profile evidence, explicit human operational approval, a real
three-day qualification, a fresh attestation, the immutable-base/global-checker
conditions, or the final gate.
