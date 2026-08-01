# Candidate shadow observation report

- Change: `a046-approve-candidate-veto-thresholds`
- Report state: `incomplete / not approval-eligible`
- Numeric activation: absent
- Runtime outcome: `unapproved / passed=0 / verdict inactive`

No human-approved candidate-store snapshot or market/session evidence export was
provided to this change. The report therefore freezes the absence as
`not_measured`; it does not turn an empty population into a zero missing rate or
a zero markout.

| Market | Session | Sample count | Missing rate | 5m | 15m | 30m |
|---|---|---:|---|---|---|---|
| KR | regular | 0 | `not_measured` | `not_measured` | `not_measured` | `not_measured` |
| US | regular | 0 | `not_measured` | `not_measured` | `not_measured` | `not_measured` |

## Contract fixture (not evidence activation)

`internal/markout/testdata/a049_markout.golden.json` freezes the transport-neutral
5/15/30 minute selection contract against synthetic existing observations. Its
SHA-256 is:

`sha256:7d3a8dd3a0aa39b3e6669f388442ea449be93550c3d99d0708032416e80e8cec`

The fixture demonstrates first-observation-at-or-after-target selection, inclusive
`+60s` tolerance, and `not_measured` when no observation falls inside the window.
It is reusable by a049, but it is not market evidence and cannot approve a numeric
threshold set.

The activation fixtures in `internal/candidate/thresholdset_test.go` are likewise
synthetic protocol fixtures. They demonstrate digest and time binding only; their
decimal values and approver label are not human approval evidence and MUST NOT be
copied into a runtime registry.

## Missing activation inputs

- non-zero market/session sample count;
- measured missing rate;
- real 5/15/30 minute markout distribution;
- immutable evidence artifact and digest;
- separate human activation record selecting one version.

Until all inputs exist, the loader rejects a set and the UI remains read-only.
