# a063 task 4.2 — read-only `verify run --list` evidence

**Result: PASS for the read-only planning check only.** This is not an
attestation renewal, engine proof, service activation, or task completion.
Task 4.2 remains open.

## Scope and command

The reviewed candidate was used with the approved console profile:

```bash
/tmp/a063-reviewed-20260906/tossctl --config-dir "$HOME/.config/tossctl" verify run --list
```

- Candidate SHA-256:
  `d9ef7f5ecac62a9e93159ebfc5cf5cf50d1ac1ee93b68c7ff9d48704777f6a6b`
- Command exit: `0`
- Sanitized output length: `15658` bytes
- Raw transient-output SHA-256 (not retained in the repository):
  `67aaae6b5c9ae3785cb706a7a138bf9415f950a3e40c73b4bdca4bcc61fa9a2e`

The scoped candidate invocation recorded above ended in `verify run --list`.
Its output was inspected only after token, session, authorization, cookie,
account, and long numeric patterns were redacted; no raw account or runtime
output is copied into this record.

## Evidence scope and correction

The affirmative conclusion here is limited to that scoped candidate command,
its retained output digest, and the current list-branch source/test contract.
`cmd/tossctl/verify.go` takes the `--list` branch before credential, record,
network, or mutation handling; the command's request-observer test covers the
same no-request behavior.

An earlier discarded report-writing shell incident has no retained command
transcript, event capture, stderr, or exit artifact. It is outside this
report's affirmative conclusion, and this record draws no inference about its
commands or effects.

## Observed plan and confirmation boundary

The command printed the complete procedure and marks live stages as
`[mutating]`. It describes read-only fixture and sellable-baseline reads, plus
planned order placement, replay/conflict probes, cancellation, amendment,
conditional-order registration/modification/cancellation, and optional
holdings-dependent probes. It also identifies the deferred
`conditional-trigger` step as an opt-in action that can sell one share at
market.

Before any live request, a non-list run requires one typed, expiring
confirmation string covering the displayed plan; `--confirm-each` instead
requires a separate confirmation before each mutation. The printed procedure
states that no flag can answer either prompt and that unlisted work stops for a
new approval. This confirms the human-confirmation boundary relevant to the
engine interlock diagnosis.

## Safety result and next dependency

The reviewed `--list` invocation only rendered the plan. Under the verified
list-branch contract, it did not send a broker request or reach record or
mutation handling. It did not itself create, cancel, amend, or modify an
order, start or restart the engine, change a trading toggle, alter a
config/profile, change systemd, or start/reset a survey. No task checkbox was
changed by this scoped command.

The prior engine launch remains blocked on incomplete capability-attestation
coverage. This planning check shows what a later `verify run` would do, but
cannot create that coverage. A live `verify run` remains a separate
human-only action requiring explicit approval for the displayed actual
order/conditional-order mutations. After successful live verification, the
engine's same-profile proof must still be established before the 4.2 unit
activation evidence can be reconsidered.
