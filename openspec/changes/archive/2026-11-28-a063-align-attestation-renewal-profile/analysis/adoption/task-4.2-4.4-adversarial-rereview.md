# a063 tasks 4.2–4.4 operational-readiness adversarial re-review

## Verdict: CLEAR for prepared readiness only

The revised template SHA-256 `9b0357a0f7e9ed2967a8cff458e77b3cb815581a1a894872f13bf985fdb6151d` closes both findings from the prior review. It remains a prepared, approval-gated packet; it does not authorize or record an operation.

## Closure checks

- Approval scope now explicitly includes the bounded console-profile survey procedure. It requires the approval to identify the existing console process/procedure using `$HOME/.config/tossctl`, checks whether the same resolved record is already surveying, preserves an active survey without duplication, and permits the existing console control exactly once only when absent. It forbids shell survey start, `engine run`, record reset, `--record`, changed launch arguments, and a second survey. The human-operated decision, actual start time, redacted same-profile record reference, and future window must be retained.
- The template now requires a redacted, normalized running-engine config-dir reference and strict equality with the normalized console target before install or survey. An implicit engine config-dir or any mismatch stops the procedure pending separate review. Task 4.4 repeats equality before relying on attestation/status evidence.
- The future three-day window remains explicit and cannot reuse the expired 2026-08-29 deadline or historical records. The 4.4 proof remains post-window and read-only, with a fresh attestation timestamp, bounded renewal outcome, safe failure observation, and explicit no-restart/no-order record.
- Existing restrictions still confine installation to the reviewed candidate and two user-unit templates with digest verification, regular-file/drop-in/fragment checks, backup validation, user daemon reload, and only the attest timer. They continue to prohibit engine restart, automation-toggle change, live-order mutation, unrelated units/drop-ins, and survey reset.
- Service/timer SHA-256 values still match their S snapshot values. Final independent review, gstack review, PM synchronization, successful final gate, Manager acceptance, and archive remain blocked until actual approved evidence exists.

No operational command was run during this re-review.
