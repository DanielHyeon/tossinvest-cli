# Function Logic Map: `writeCandidateTable`

- Source: `cmd/tossctl/candidate.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| cycle result and candidate report | read-only scan projection | candidate cycle/report builder | all degradation, veto, shadow, retention and space states remain visible |
| veto code order | copied fixed D3 array | `candidate.OrderedVetoCodes` | callers cannot mutate the package order |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B13 | halted/source/read/recorded variations | writes human-readable lines | writer error returned at end | candidate CLI table tests |
| B14 | iterate D3 veto codes | writes raised/unmeasured counts | continue | candidate CLI veto output tests |
| B15-B28 | reasons, alarms, sightings, shadow distributions | writes human-readable lines | continue | candidate CLI table tests |
| B29-B34 | retention and free-space state switches | writes explicit state line | nil or writer error | retention/space CLI tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `fmt` writers and sorting/wrapping helpers | deterministic read-only table | no retry; writer error returned | CodeGraph + AST |
| `candidate.OrderedVetoCodes` | fixed D3 code block | pure array copy | AST B14 |

## State mutations and fallbacks

- Writes only to the supplied writer. No config, candidate, threshold, order, or RiskIntent mutation.
- The only follow-up change is B14's ordering source; all other branches are byte-for-byte behavior-preserving context.

## Safety conclusion

- Safe edit boundary: replace access to exported mutable ordering with copy accessor.
- High-risk impact: no; read-only CLI rendering and existing branch coverage remain unchanged.
