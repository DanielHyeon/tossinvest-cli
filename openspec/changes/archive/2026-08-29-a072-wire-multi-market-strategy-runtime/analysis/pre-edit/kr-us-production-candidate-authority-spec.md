# KR/US production candidate authority — pre-edit specification

## Scope

This slice consumes the already-existing candidate discovery store and the a046 immutable
threshold/approval types. It does not scan a market, create or edit approval files, change a desired
state, mint an execution lease, call the Guardian, or reach a broker. KR and US are delivered as one
paired implementation wave; neither market is a prerequisite for the other.

## Production files and trust pins

For each market `M in {KR,US}`, the engine consumes exactly these regular files from the resolved
absolute config directory:

- `candidate-thresholds-M.json`
- `candidate-threshold-evidence-M.bin`
- `candidate-threshold-activation-M.json`

Every file must be owned by the current process owner, have exact mode `0400`, and be opened without
following a final-component symlink. Threshold and activation JSON are limited to 1 MiB; opaque
evidence is limited to 16 MiB. The engine accepts no caller-provided filename or relative config
directory.

The activation file is pinned by the market-local environment value
`TOSSOS_CANDIDATE_THRESHOLD_M_ACTIVATION_SHA256`. The pin is canonical
`sha256:<64 lowercase hex>`. The existing strict `candidate.LoadActivationRecord` binds the approved
version, market/session, canonical threshold-set digest, evidence digest, approval time and actor;
`candidate.LoadThresholdSet` then repeats those bindings against the exact set and evidence bytes at
the one frozen engine time. The loader exposes no approval constructor, writer, signer or toggle.

## Paired collection contract

1. Freeze the engine clock once.
2. If a market's scheduler authority is not ready, do not open its candidate authority files or read
   its discovery store. Classify only that market as `SCHEDULE_NOT_READY`.
3. For scheduler-ready markets, load and validate KR and US authority in independent goroutines.
   A failure or panic in one result must not cancel or overwrite its peer result.
4. Open the one discovery store read-only for the collection boundary and assess each ready market at
   the frozen instant using its own approved threshold set.
5. For every verdict, call the audited value-only `strategy` handoff. Only an exact measured-and-clear
   verdict becomes an immutable `strategy.ApprovedSnapshot`; raised, unmeasured, stale, mixed-life or
   cross-market candidates remain typed refusals and never reach a lane.
6. Return only market-keyed scalar readiness/count/digest observations across the command boundary.
   Opaque threshold sets, approval records, raw evidence, candidates and store handles stay private.
7. The production worker remains `Effective=false` until risk snapshot authority, lane input
   authorities, first-leg admission, protection and the official Gateway cycle are all assembled.

## Required paired RED scenarios

- Valid KR and US fixtures load concurrently and preserve distinct activation/set/evidence digests.
- Invalid KR owner/mode/symlink/pin/scope/evidence/approval time refuses KR while valid US remains
  ready; the mirrored US-invalid/KR-valid case also passes.
- Schedule OFF for either market performs zero file and store reads for that market.
- One market loader panic is recovered as a market-local refusal.
- Candidate assessment returns only immutable approved snapshots; unmeasured/raised candidates are
  counted as refusals and cannot make either worker effective.
- The public production assembly contains no threshold set, activation record, raw evidence, store,
  journal, Guardian, Gateway or broker capability.

## Completion boundary

This slice is complete only when the same test run proves both KR and US paths. It is not an
automatic-trading completion milestone and does not authorize LIVE orders or activation.
