# Function Logic Map: `Client.accountsLocked`

- Source: `internal/official/reads.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| caller mutex ownership | `Client.mu` already held | `Accounts` or `ensureAccountSeq` | calling without ownership would race |
| official response | fully decoded account-list envelope | `/api/v1/accounts` | classified/cancel error; no cache update |
| current sequence | atomic zero unresolved, positive selected, negative invalid | `Client.accountSeq` | only zero may be primed |
| selection provenance | explicit positive or implicit discovery | immutable `accountSeqExplicit` | implicit drift refuses |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | account-list request fails | HTTP/rate budget only | exact error; cache unchanged | failure/cancel tests |
| B2 | cache is zero with a nonempty positive first record | atomic store of first sequence | continues to adapted list | priming/concurrency tests |
| B3 | priming condition is false | none | enters preservation/drift path | explicit/invalid/drift tests |
| B4 | selected sequence is implicit positive | none | validate latest first record | drift tests |
| B5 | latest first record is empty, invalid, or different | none | identity-drift error | drift table |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `c.get` | fetch and fully decode the official list | caller context; auth retry only; classified HTTP errors | CodeGraph + AST |
| `adaptAccounts` | preserve public domain response | pure conversion | existing adapter tests |

## State mutations and fallbacks

- The only durable mutation is zero → positive first-record sequence.
- Positive explicit and negative invalid values are never overwritten.
- A later public response cannot silently move or contradict an implicit cache.
- Empty/zero/negative/error responses do not prime; there is no later-record
  fallback and no account-list cache.

## Safety conclusion

- Safe edit boundary: run only while the discovery mutex is held, prime only a
  fully decoded positive first record, and reject later implicit-selection drift.
- High-risk impact: yes — this state scopes every official read and mutation.
