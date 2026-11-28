# a063 task 4.2 engine-start/interlock adversarial review

**Verdict: CLEAR — blocked evidence is internally consistent and preserves the safety boundary.**

Reviewed at 2026-09-06 KST, read-only. This review did not invoke an engine, verification, systemd action, survey, order command, or source edit.

## Bound evidence

| Artifact | SHA-256 | Result |
| --- | --- | --- |
| Repository engine-start record | `b9522cc228d0181fb74244564e5f205af6b9830da9a56b047d947c065711c506` | Matches `/tmp/a063-task42-engine-start.md` byte-for-byte. |
| Repository interlock diagnosis | `08780ebef3dcdd4fc50ed86f8b341ab5743cd021f8d6e4fb562af696f6e6ff60` | Matches `/tmp/a063-task42-engine-diagnosis.md` byte-for-byte. |
| Start event capture | `76c29f0ee927737f9a6d43b2343eb6f87479d09e6b9f399ad0be34a0964d52de` | Retains preflight, one launch submission, and subsequent process classification. |
| Diagnosis event capture | `46e555c5c52c89c9df6a9cd0d0745a0f883c4629263b0d53858a624d0e5f03c3` | Retains only local help, metadata, and supplied-log classification. |

The start task itself records the separate human direction to start an absent engine once, and restricts it to the reviewed candidate plus explicit config directory. That is distinct from the pending task-4.2 unit-installation authority; it does not widen the latter.

## Findings

- The retained start events contain one detached `engine run` launch submission. It was not retried. The launcher submission returned zero; executable-bound post-launch classification found zero surviving reviewed-candidate engine processes. Therefore same-profile identity could not be proven and the 4.2 activation block was correctly not entered.
- The event trail and record support absence of the mutable installation sequence: no target backup/install, unit disable/enable, daemon reload, user-unit start/stop/restart, or template installation. The timer remained enabled and active and the attestation service inactive. No survey start, order path, toggle change, task edit, archive, or source edit is evidenced.
- The phrase “no 4.2 mutation” is supportable when read as the specified unit-installation/activation block. The separately authorized engine launch is a potentially mutating command class, but it failed before any loop started; it should not be described as proof that no local lock or audit artifact could have been touched.
- Current `cmd/tossctl/engine.go` performs engine assembly and unmet-interlock handling before marker creation and runtime/loop construction. `cmd/tossctl/engineproc.go` likewise states that gate/interlock refusals occur before loops. The supplied sanitized error identifies missing capability coverage for required order and cancel endpoints. This supports the diagnosis that the observed refusal was the capability-attestation interlock rather than a lock-file remedy.
- Current interlock tests explicitly show a complete operator configuration can clear the startup clauses; broker-resident protection being UNWIRED gates exposure-raising mutations rather than startup. The proposed remedy is therefore not contradicted by the current source.
- The diagnosis does not invent a numeric engine exit status. It separates the shell's successful detached submission from the unavailable child exit code.
- `tossctl verify run` remains separated behind explicit human authorization. The retained diagnosis accurately says that its live, limit-only single-share order-and-cancel behavior must not be run by an agent. No verify command was executed in either evidence capture.
- Repository records omit raw process argv, credentials, account data, record contents, and the engine log. The supplied log itself was classified without copying its raw content into either repository record.

## Review commands

Read-only commands were limited to SHA-256 and byte comparison of the two records and temporary copies; inspection of retained event metadata; static source/test searches and reads; and `git diff --check` on the two evidence files. The scoped diff check exited 0.

Task 4.2 remains pending. This CLEAR approves only the accuracy and safety boundary of the blocked launch/diagnosis evidence; it does not approve a future `verify run`, engine retry, systemd activation, survey, gate, or archive.
