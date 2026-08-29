# Function Logic Map: `(*officialfx.ProductionAuthorityService).Collect`

- Source: `internal/officialfx/production.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input | Invariant | Authority | Failure |
| --- | --- | --- | --- |
| service/context | non-nil, configured account/base/clock | package-owned constructor | paired typed unavailable |
| frozen time | one canonical UTC value shared by KR and US | injected production clock | affected market refusal |
| KR source | production-origin exact official account identity | `official.Client.AuthoritativeAccountIdentity` | KR-only refusal |
| US source | current signed pinned policy plus `ReadOfficial` | local manifest/state + official client | US-only refusal |

## Branches and early returns

| Branch | Mutation / I/O | Result | Test |
| --- | --- | --- | --- |
| invalid service/clock/context | no evidence minted | paired refusal | invalid input table |
| KR succeeds, US succeeds | account read + FX read + US state commit | two opaque evidences | paired success |
| KR fails, US succeeds | both attempted; US state commit | isolated KR refusal | paired isolation |
| KR succeeds, US fails | account attempted; US may stop pre-FX | isolated US refusal | paired manifest table |
| both fail | both attempted where globally possible | paired refusals | paired failure test |

The two paths share only the frozen clock and immutable account scope. A market-local failure is not
used as a cancellation signal for its peer. The method has no LIVE/broker/journal capability.

## Calls and live bindings

| Callee | Purpose | Error/retry contract |
| --- | --- | --- |
| configured `Now` | freeze one common evaluation time | one call, zero/invalid refuses |
| `collectKR` | official account recheck + bounded identity mint | market-local result |
| `collectUS` | signed policy/state + existing `ReadOfficial` | market-local result |

## State mutations and fallbacks

- KR performs read-only official identity verification.
- US alone may advance the local trusted-time/generation state after valid official evidence.
- No peer cancellation and no quote-currency/default-haircut fallback exists.

## Safety conclusion

- `ProductionAuthorityService.Collect` in `internal/officialfx/production.go` always joins both
  outcomes before returning and exposes only opaque evidence by value.
- No broker, journal, activation or operating-toggle dependency is present.
