# Function Logic Map: `Client.ensureAccountSeq`

- Source: `internal/official/client.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| context | live or cancelled | account-scoped caller | request error propagates |
| `accountSeq` | atomic zero means unresolved; positive means selected; negative is invalid | `Client` constructor or first account discovery | positive returns without mutex; negative refuses without HTTP |
| account list | zero or more adapted official accounts | shared account-list helper | empty/non-positive first ID fails closed |
| mutex | exactly one first discovery at a time | `Client.mu` | no duplicate concurrent resolution |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | first atomic load is nonzero | none | enters cached/invalid fast path | cached/explicit tests |
| B2 | first atomic load is positive rather than negative | none | cached sequence or invalid-sequence error | cached contention and negative explicit tests |
| B3 | post-lock atomic recheck is nonzero | none | enters concurrent-completion path | concurrent first-use tests |
| B4 | post-lock atomic recheck is positive rather than negative | none | selected sequence or invalid-sequence error | concurrent first-use tests |
| B5 | account-list request fails | HTTP/rate budget only | wrapped lazy-resolution error | failure/cancel/malformed tests |
| B6 | successful list is empty | none | no-accounts error | empty discovery test |
| B7 | first account ID is not numeric | none | parsing error | structural integer-adapter guard |
| B8 | parsed first sequence is not positive | none | invalid-sequence error | zero/negative discovery test |
| B9 | helper-selected atomic value differs from the parsed first record | none | invariant error | structural invariant |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| shared locked account-list helper | avoid recursive lock while preserving one discovery | caller context; no retry | CodeGraph + planned AST |
| `strconv.Atoi` | recover official integer key from domain ID | parse failure refuses | current AST |

## State mutations and fallbacks

- The atomic positive/negative fast path avoids the discovery mutex entirely.
- Only unresolved zero enters the mutex; it rechecks after locking before I/O.
- It will call the non-locking helper because public `Accounts` also acquires the
  same mutex; calling the public method while locked would deadlock.
- No sequence is guessed or selected from a later account. A discovered
  zero/negative value is refused; a configured negative value is rejected
  immediately and is never emitted as a header.

## Safety conclusion

- Safe edit boundary: return selected atomic state before locking, call the
  lock-assuming helper only for unresolved zero, reject invalid values, and
  verify the helper selected the same positive first-record sequence.
- High-risk impact: yes — a wrong sequence could scope reads or orders to the
  wrong account. Exact first-account and explicit override tests are mandatory.
