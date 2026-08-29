# Function Logic Map: `(*official.Client).VerifyAuthoritativeAccountIdentity`

- Source: `internal/official/account_identity_authority.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input | Invariant | Authority | Failure |
| --- | --- | --- | --- |
| client | sealed constructor-owned production endpoint/transport | `Client.AuthorityOrigin` | `ErrAuthorityOrigin`, zero HTTP |
| account ID | canonical positive decimal sequence | a072 configured account identity | typed invalid identity |
| official account list | exactly one matching record and exact selected sequence | official `/api/v1/accounts` | mismatch/refusal |

## Branches and early returns

| Branch | Mutation / I/O | Result | Test |
| --- | --- | --- | --- |
| nil/context/invalid ID | none | refusal | invalid input table |
| configured/non-production origin | none | `ErrAuthorityOrigin` | pre-HTTP origin test |
| account read fails | one read attempt | wrapped read error | fake/TLS origin test |
| duplicate/missing/mismatched selection | read only | scope refusal | account matrix |
| exact unique selected account | read only; opaque local seal | opaque valid identity | success test |

No broker, journal, activation, toggle or write call exists. Configuration is sealed before the
method begins, so the origin precheck and account read remain one immutable client boundary.

## Calls and live bindings

| Callee | Purpose | Error/retry contract |
| --- | --- | --- |
| `canonicalAccountSequence` | reject noncanonical account scope before I/O | pure, no fallback |
| `Client.AuthorityOrigin` | prove immutable official endpoint and transport before/after read | exact refusal, no retry |
| `Client.Accounts` | production account-list re-observation | context-bound official read |
| `Client.SelectedAccountSeq` | prove account-scoped header selection | pure current value |

## State mutations and fallbacks

- The existing official account discovery may prime only the already-defined unresolved account
  sequence; this method creates no new mutable authority.
- No WTS/configured-origin fallback exists.

## Safety conclusion

- `Client.VerifyAuthoritativeAccountIdentity` in `internal/official/account_identity_authority.go`
  releases only success/error, not reusable raw identity material.
- The function cannot place, amend or cancel orders.
