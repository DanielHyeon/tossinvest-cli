# Function Logic Map: `resolveAccountRef`

- Source: `internal/app/engine/interlock.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| official reader | non-nil concrete engine official client | `NewContext` | missing reader refuses startup |
| account list | ordered official records | `/api/v1/accounts` | request failure refuses startup |
| implicit default | first record only | official account-scope spec | blank number or non-positive sequence refuses |
| account identity | number and sequence from the same record | first `domain.Account` | never skip to a later record |
| selected sequence | positive and equal to first record | official client cache/explicit option | mismatch refuses before journal open |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | official reader is nil | none | no-client error | existing startup refusal test |
| B2 | account-list request fails | official read/audit refusal | error | existing rate/error startup tests |
| B3 | list is empty | none | no-account error | startup empty-list coverage |
| B4 | first account number is blank | none | invalid-default-account error | strict first-record test |
| B5 | first sequence parse fails or is nonpositive | none | invalid-default-account error | strict first-record test |
| B6 | selected client sequence is absent/nonpositive or differs from record | none | account-scope mismatch error | explicit mismatch/negative engine test |
| success | first number and sequence valid and selected sequence matches | public `Accounts` already primed the same sequence | first account number | actual engine recovery test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `reads.Accounts` | resolve the engine's one official account | caller context; no retry | CodeGraph + AST |
| `reads.SelectedAccountSeq` | confirm the actual request header scope | atomic read; no lock, HTTP, or retry | post-edit AST |
| `strconv.Atoi` | validate domain ID is a positive official sequence | parse/non-positive refuses | post-edit AST |
| `strings.TrimSpace` | validate and normalize the same record's account number | blank refuses | post-edit AST |

## State mutations and fallbacks

- This function does not set the sequence itself; the same successful
  `Client.Accounts` call primes zero during serialized discovery. It atomically
  reads back the selected sequence and demands equality, covering positive
  explicit options and invalid negative options.
- The loop/fallback to later account records is removed. A configurable
  multi-account selector is not inferred.
- Startup refusal remains audited by `NewContext`.

## Safety conclusion

- Safe edit boundary: validate only record zero and compare the atomically read
  selected sequence before returning its account number.
- High-risk impact: yes — the result scopes journal, Guardian, Gateway, and
  recovery while the sequence scopes official requests. Same-record tests are
  mandatory before shipping.
