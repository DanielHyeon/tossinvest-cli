# a063 task 4.2 `verify run --list` — Manager acceptance

**Verdict: accepted only as bounded SDD planning evidence.** Task 4.2 remains
open. This acceptance neither authorizes nor records a non-list verification,
an order mutation, an engine launch, a user-unit action, a survey, a final gate,
or archive.

## Independent checks

On 2026-09-06, the Manager independently read the corrected evidence and both
separate reviews, inspected the current `--list` branch and its request-observer
test, and confirmed the following repository checks:

- `git diff --check` exited `0`;
- `openspec validate a063-align-attestation-renewal-profile --strict --no-interactive`
  passed;
- only the three new a063 evidence/review records are in scope; the pre-existing
  untracked a119 material remains excluded.

The evidence report SHA-256 is
`dea5177e3b062bf0ad47c125cf2800de5fc772125019d30a5fd31e67171cccf6`. Its
adversarial re-review is CLEAR at SHA-256
`c9307e175ad19016bcf411b170c59d2cf23f59c538c196648fa9a0d4a835d4b4`, followed
by the gstack review CLEAR at SHA-256
`3145fa3726f96c8880a704641a963aacc2c1a7618fa86e25b045a8e7ea1ed497`.

## Scope decision

The list branch returns before credential, record, network, broker, runner, and
mutation handling, and the focused request-observer test protects that branch.
The report correctly limits its affirmative result to the explicit candidate
`verify run --list` invocation and its retained digest. It expressly leaves the
unrecorded discarded heredoc incident outside its conclusion; no effect is
inferred from that incident.

A real `verify run` still performs the displayed live workflow and requires a
separate explicit human authorization. It remains the prerequisite for the
missing capability-attestation coverage, after which the same-profile engine
proof and tasks 4.2–4.5 must still be completed before the final gate or archive.
