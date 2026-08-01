# Function Logic Map: `EvaluateRatchet`

- Source: `internal/exitpolicy/ratchet.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Decimal price/state inputs | positive prices; entry > initial stop; ratios in [0,1] | journal exit state + observed quote + validated config | `ErrRefused`, no state/order side effect |
| Pending action | empty or existing proposal action | journal exit state | breach may displace partial; duplicate breach suppressed |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | custom/default config invalid | none | refusal | existing invalid-config tests |
| B2 | price/risk/state decimal invalid | none | refusal | `TestUnusableInputsAreRefused` |
| B3 | observation exceeds high-water | local watermark advances | continue | monotone property test |
| B4 | trigger creates stronger candidate | local baseline/level advance | continue | trigger table tests |
| B5 | observed below newly composed baseline | full-exit proposal, or duplicate suppression | early decision | breach/pending tests |
| B6 | partial trigger; already taken/pending/free | suppression or 40% proposal | decision | once/pending tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `riskOf`, decimal parsers | validate canonical decimal inputs | pure, no retry | CodeGraph + AST |
| `ratchetCandidate` | select level/candidate/partial from the frozen table | pure/refusal | CodeGraph + AST |
| `ComputeProtectedStop` | monotone max composition | fail closed | CodeGraph + AST |

## State mutations and fallbacks

- The function is pure. It mutates only a local decision and never journal/broker state.
- Breach is evaluated after protection promotion and outranks partial profit-taking.

## Safety conclusion

- Safe edit boundary: preserve trigger values, monotone composition, breach-first precedence, and pending suppression; wrap its result in one immutable snapshot.
- High-risk impact: yes — stop/take-profit judgement.
