# Function Logic Map: `TallyVetoes`

- Source: `internal/candidate/veto.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| candidate chase slice | zero or more three-state verdicts | caller-owned values | every candidate lands in vetoed, unmeasured, or passed exactly once |
| veto order | copied fixed array | `OrderedVetoCodes` | all per-code maps are initialized even at zero |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | initialize all D3 code keys | fresh maps | zero-valued columns | bucket/tally tests |
| B2 | iterate candidates | fresh tally only | continue | bucket/tally tests |
| B3 | iterate D3 codes per candidate | map counts | continue | per-code tests |
| B4 | classify state | map counts | one state branch | veto tests |
| B5 | dangerous | increment Raised | continue | raised tests |
| B6 | unmeasured | increment NotMeasured and Reasons | continue | missing tests |
| B7 | classify candidate | aggregate counts | exclusive branch | bucket test |
| B8 | vetoed candidate | increment Vetoed | continue | raised tests |
| B9 | otherwise unmeasured | increment Unmeasured | continue | absent threshold tests |
| B10 | fully clear | increment Passed | continue | clear test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `OrderedVetoCodes`, `State`, `Vetoed`, `HasUnmeasured` | fixed-order exclusive tally | pure; no I/O/retry | CodeGraph + AST |

## State mutations and fallbacks

- Mutates only newly allocated maps/counters. No threshold fallback or external state.

## Safety conclusion

- Safe edit boundary: order source changed from exported array to returned copy.
- High-risk impact: no; partition semantics and counts are unchanged.
