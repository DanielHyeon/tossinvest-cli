# a063 task 4.2 — `verify run --list` independent gstack review

**Verdict: CLEAR — only for the bounded, read-only planning evidence.** This
review does not authorize a non-list `verify run`, service/systemd work, an
engine launch, a survey, a gate, or archival. Task 4.2 remains open.

## Inputs and reproducibility

Read-only inputs reviewed on 2026-09-06:

- evidence report SHA-256:
  `dea5177e3b062bf0ad47c125cf2800de5fc772125019d30a5fd31e67171cccf6`;
- adversarial re-review SHA-256:
  `c9307e175ad19016bcf411b170c59d2cf23f59c538c196648fa9a0d4a835d4b4`;
- candidate `/tmp/a063-reviewed-20260906/tossctl`: regular, non-symlink file,
  SHA-256 `d9ef7f5ecac62a9e93159ebfc5cf5cf50d1ac1ee93b68c7ff9d48704777f6a6b`;
- retained transient sanitized-output artifact: 15,658 bytes, SHA-256
  `67aaae6b5c9ae3785cb706a7a138bf9415f950a3e40c73b4bdca4bcc61fa9a2e`.

The evidence records the already-completed scoped invocation:

```bash
/tmp/a063-reviewed-20260906/tossctl --config-dir "$HOME/.config/tossctl" verify run --list
```

with exit `0`. I did **not** re-run it, and did not invoke any non-list
verification or operational command. `git diff --check --
openspec/changes/a063-align-attestation-renewal-profile/analysis/adoption`
returned `0` with no whitespace error.

## Source and test contract

Current `cmd/tossctl/verify.go` keeps `verify run` annotated
`source=official, mutating=true`, preserving the command-level warning for the
non-list path. In `runVerifyRun`, `if opts.list { verifylive.WriteSteps(...);
return nil }` occurs before context setup, execution lock, record loading,
rate-budget work, broker construction, or runner/mutation handling.

`TestVerifyRunListNeedsNoCredentialsAndSendsNothing` in
`cmd/tossctl/verify_test.go` executes `verify run --list` against a request
observer and fails if `len(srv.seen()) != 0`; it also checks that every listed
step and the expiring batch-approval boundary are rendered. Current source and
test files have no uncommitted diff. This supports the narrow no-broker-request
claim for the list branch; it does not prove behavior of any other invocation.

## Safety, provenance, and SDD assessment

The corrected evidence properly limits its affirmative conclusion to the
candidate invocation, output digest, and list-branch contract. It explicitly
quarantines the unrecorded discarded heredoc incident: there is no transcript,
event, stderr, or exit artifact, and the report draws no claim about it. That
fix resolves the prior provenance overreach.

The output documents live order, cancellation, amendment, conditional-order,
and optional trigger work, including the possible one-share market sale. It
also documents the expiring typed approval, the per-mutation
`--confirm-each` alternative, the absence of a prompt-bypass flag, and a new
approval for unlisted work. That is an accurate safety boundary and a useful
SDD trace to the source/test contract.

The explicit `--config-dir` supplies candidate-invocation provenance only:
the early list branch deliberately does not read credentials or configuration.
It therefore cannot establish that an engine uses the console profile, that an
attestation is fresh, or that any unit was installed or reloaded. The evidence
states those limits, leaves 4.2–4.5 unchecked, and makes no false final-gate or
archive claim.

## Remaining dependency

A real `verify run` is still a separate human-authorized live action, subject
to its displayed plan and confirmations. Even successful live verification
would still require same-profile engine proof before task 4.2 activation
could be reconsidered. No finding blocks retention of this limited planning
record.
