# Function Logic Map: `Adapter.Fetch`

- Source: `internal/strategyevidence/source.go`
- AST evidence: `ast.json` (extracted at the re-baselined comparison base)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `a.policy` | a sealed, minted official policy | `MintSourcePolicy` + `validateOfficial` | the policy's own typed refusal, zero transport calls |
| `a.transport` | non-nil | `NewOfficialAdapter` | `ErrSourceDisabled` |
| `request` | operation id set; every bound at or under the policy's | the caller | `ErrSourceBoundExceeded` before any request |
| credential | present when the policy requires one | `CredentialProvider` | `ErrSourceCredential`; the secret never enters `RequestMetadata` |
| transport error text | never carried outward | `redactTransportError` | only the sentinel class survives |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | the sealed official policy does not validate | none | the policy's typed error | `TestEveryPolicyFieldHasAZeroCallRefusal` |
| B2 | no transport is bound | none | `ErrSourceDisabled` | `TestEveryPolicyFieldHasAZeroCallRefusal` |
| B3 | `PageLimit` was left zero | fills it from the policy | falls through | `TestAdapterRejectsRuntimeBoundExpansionBeforeTransport` |
| B4 | `ResponseByteLimit` was left zero | fills it from the policy | falls through | `TestAdapterRejectsRuntimeBoundExpansionBeforeTransport` |
| B5 | `Concurrency` was left zero | fills it with 1 | falls through | `TestAdapterRejectsConcurrentCallBeforeSecondTransportRequest` |
| B6 | `RequestDeadline` was left zero | fills it from the policy | falls through | `TestAdapterRejectsRuntimeBoundExpansionBeforeTransport` |
| B7 | `OperationDeadline` was left zero | fills it from the policy | falls through | `TestAdapterRejectsRuntimeBoundExpansionBeforeTransport` |
| B8 | `Page` was left zero | fills it with 1 | falls through | `TestSECOfficialAdapterCollectsFrozenPaginatedFixtureWithDeclaredIdentity` |
| B9 | `PageSize` was left zero | fills it from the policy | falls through | `TestSECOfficialAdapterCollectsFrozenPaginatedFixtureWithDeclaredIdentity` |
| B10 | the operation id is blank after trimming | none | `ErrSourceBoundExceeded` | `TestAdapterRejectsRuntimeBoundExpansionBeforeTransport` |
| B11 | any request bound exceeds the policy's | none | `ErrSourceBoundExceeded` | `TestAdapterRejectsRuntimeBoundExpansionBeforeTransport` |
| B12 | the policy requires a credential | resolves one through the provider | falls through | `TestAdapterKeepsCredentialOutOfRequestMetadata` |
| B13 | the provider is absent | none | `ErrSourceCredential` | `TestOfficialAdaptersDoNotCommitIncompleteOrUntrustedBatches` |
| B14 | the provider fails or returns a blank secret | none | `ErrSourceCredential` | `TestOfficialAdaptersDoNotCommitIncompleteOrUntrustedBatches` |
| B15 | the concurrency slot cannot be acquired | none | `ErrSourceRateLimited` | `TestAdapterRejectsConcurrentCallBeforeSecondTransportRequest` |
| B16 | attempt loop, bounded by `MaxRetries` | consumes one call from the rate window per attempt | falls through to `ErrSourceIncomplete` | `TestAdapterEnforcesResponseBytesWindowAndRetryAfter` |
| B17 | the call window is exhausted | none | `ErrSourceRateLimited` | `TestAdapterEnforcesResponseBytesWindowAndRetryAfter` |
| B18 | the transport returned an error | none | see B19/B20 | `TestTransportErrorsCarryNoCredentialOrURL` |
| B19 | that error is a cancellation or a deadline | none | the sentinel, with the transport's own message discarded | `TestTransportErrorsCarryNoCredentialOrURL` |
| B20 | retries are exhausted on a transport error | none | `ErrSourceRetriesExhausted` wrapping the redacted class | `TestTransportErrorsCarryNoCredentialOrURL` |
| B21 | the body is larger than the request byte limit | none | `ErrSourceBoundExceeded` | `TestAdapterEnforcesResponseBytesWindowAndRetryAfter` |
| B22 | the status is 2xx | copies the body out | the fetch result | `TestSECOfficialAdapterCollectsFrozenPaginatedFixtureWithDeclaredIdentity` |
| B23 | the status is 401 or 403 | none | `ErrSourceCredential` | `TestOfficialAdaptersDoNotCommitIncompleteOrUntrustedBatches` |
| B24 | the status is not retryable | none | `ErrSourceIncomplete` | `TestOfficialAdaptersDoNotCommitIncompleteOrUntrustedBatches` |
| B25 | the status is retryable and attempts remain | none | falls through to the wait | `TestAdapterEnforcesResponseBytesWindowAndRetryAfter` |
| B26 | the status is retryable and this was the last attempt | none | see B27 | `TestAdapterEnforcesResponseBytesWindowAndRetryAfter` |
| B27 | that last retryable status was 429 | none | `ErrSourceRateLimited` | `TestAdapterEnforcesResponseBytesWindowAndRetryAfter` |
| B28 | Retry-After is negative or beyond the operation deadline | none | `ErrSourceBoundExceeded` | `TestAdapterEnforcesResponseBytesWindowAndRetryAfter` |
| B29 | the bounded wait is cancelled | none | the context error | `TestAdapterEnforcesResponseBytesWindowAndRetryAfter` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `a.policy.validateOfficial` | refuse an unsealed or out-of-contract policy before anything else | pure; zero transport calls on refusal | AST + `TestEveryPolicyFieldHasAZeroCallRefusal` |
| `a.credentials.Credential` | resolve the secret at the boundary, never into metadata | any error is `ErrSourceCredential` | AST + `TestAdapterKeepsCredentialOutOfRequestMetadata` |
| `a.acquire` / `a.consumeCall` / `a.release` | hold the concurrency slot and the absolute call window | `ErrSourceRateLimited`; released by `defer` | AST + `TestAdapterRejectsConcurrentCallBeforeSecondTransportRequest` |
| `a.transport.Do` | the one outbound request, under a per-request deadline | errors are redacted before they leave this function | AST + `TestTransportErrorsCarryNoCredentialOrURL` |
| `redactTransportError` | keep the credential-bearing URL out of every returned error | pure; preserves only context sentinels | AST + `TestTransportErrorsCarryNoCredentialOrURL` |
| `a.waiter.Wait` | honour a bounded Retry-After | context cancellation is returned as-is | AST + `TestAdapterEnforcesResponseBytesWindowAndRetryAfter` |

## State mutations and fallbacks

- Local only: the request struct is defaulted in place and the adapter's call/concurrency counters move.
- No journal, broker, Guardian or toggle call occurs; the import closure of this package cannot reach one.
- The completion pass removed one fallback: the transport error's own message. `%v`-flattening it into `ErrSourceRetriesExhausted` both leaked the OpenDART `crtfc_key` query parameter and made the cause unrecoverable downstream; it is now `%w` over a redacted class.

## Safety conclusion

- High-risk: this is the only function in the change that makes an outbound request, and the only one that has ever held a credential.
- The leak was latent — the module contains no `Transport` implementation — but `net/http` bakes the full URL into `*url.Error`, and OpenDART authenticates by query parameter, so the first real transport would have logged the key (issues.md I15).
- B19/B20 are now measured: `TestTransportErrorsCarryNoCredentialOrURL` asserts the returned error contains neither the credential nor the host, and still classifies as its sentinel.
