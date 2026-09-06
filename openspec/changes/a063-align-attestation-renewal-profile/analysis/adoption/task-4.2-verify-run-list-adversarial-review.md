# a063 task 4.2 `verify run --list` adversarial review

**Verdict: CLEAR — provenance correction closes the prior blocking gap for this limited planning evidence.**

Reviewed read-only. No `tossctl` command, service action, engine action, survey, configuration action, or order command was invoked during this review.

## Current evidence

- Corrected evidence report SHA-256: `dea5177e3b062bf0ad47c125cf2800de5fc772125019d30a5fd31e67171cccf6`.
- Retained transient output remains 15,658 bytes with SHA-256 `67aaae6b5c9ae3785cb706a7a138bf9415f950a3e40c73b4bdca4bcc61fa9a2e`.
- Reviewed candidate remains a regular non-symlink file with recorded SHA-256 `d9ef7f5ecac62a9e93159ebfc5cf5cf50d1ac1ee93b68c7ff9d48704777f6a6b`.
- Current `cmd/tossctl/verify.go` takes the `opts.list` branch before credentials, records, network, or mutation handling: it writes steps and returns. Its request-observer test asserts that `verify run --list` sends no request.

## Re-review of the blocking condition

The corrected report now expressly confines its affirmative result to the scoped candidate list command, retained output digest, and current source/test contract. It identifies the earlier discarded unquoted-heredoc report-writing incident as lacking a transcript, event capture, stderr, and exit artifact, places it outside the affirmative conclusion, and draws no inference about its commands or effects.

That removes the previous unsupported global assertion that the incident did not invoke `tossctl` or cause a mutation. The remaining safety claims are correctly scoped to the evidenced `--list` branch and static command contract.

The document still states that task 4.2 remains open and does not conflate planning output with real verification, attestation coverage, engine proof, unit activation, or human approval for live orders. It retains the human-only boundary for a non-list `verify run`.

This CLEAR concerns the evidence wording only. It does not establish any claim about the discarded shell incident and does not authorize a live verification, engine retry, systemd action, survey, gate, or archive.
