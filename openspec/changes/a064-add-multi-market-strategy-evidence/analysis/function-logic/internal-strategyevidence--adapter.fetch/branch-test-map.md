# Branch Test Map: `Adapter.Fetch`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | each policy field, mutated and re-sealed, refuses with zero transport calls | `TestEveryPolicyFieldHasAZeroCallRefusal` | nine subtests failed before the contract caps existed | PASS |
| B2 | an adapter with no transport refuses | `TestEveryPolicyFieldHasAZeroCallRefusal` | existing coverage | PASS |
| B3 | an unset page limit takes the policy's | `TestAdapterRejectsRuntimeBoundExpansionBeforeTransport` | existing coverage | PASS |
| B4 | an unset byte limit takes the policy's | `TestAdapterRejectsRuntimeBoundExpansionBeforeTransport` | existing coverage | PASS |
| B5 | an unset concurrency takes 1 | `TestAdapterRejectsConcurrentCallBeforeSecondTransportRequest` | existing coverage | PASS |
| B6 | an unset request deadline takes the policy's | `TestAdapterRejectsRuntimeBoundExpansionBeforeTransport` | existing coverage | PASS |
| B7 | an unset operation deadline takes the policy's | `TestAdapterRejectsRuntimeBoundExpansionBeforeTransport` | existing coverage | PASS |
| B8 | an unset page takes 1 | `TestSECOfficialAdapterCollectsFrozenPaginatedFixtureWithDeclaredIdentity` | existing coverage | PASS |
| B9 | an unset page size takes the policy's | `TestSECOfficialAdapterCollectsFrozenPaginatedFixtureWithDeclaredIdentity` | existing coverage | PASS |
| B10 | a blank operation id refuses before the transport | `TestAdapterRejectsRuntimeBoundExpansionBeforeTransport` | existing coverage | PASS |
| B11 | each runtime bound above the policy refuses before the transport | `TestAdapterRejectsRuntimeBoundExpansionBeforeTransport` | existing coverage | PASS |
| B12 | OpenDART resolves its credential at the boundary | `TestAdapterKeepsCredentialOutOfRequestMetadata` | existing coverage | PASS |
| B13 | a credential-requiring policy with no provider refuses | `TestOfficialAdaptersDoNotCommitIncompleteOrUntrustedBatches` | existing coverage | PASS |
| B14 | a blank secret refuses | `TestOfficialAdaptersDoNotCommitIncompleteOrUntrustedBatches` | existing coverage | PASS |
| B15 | a second concurrent fetch is refused before its transport call | `TestAdapterRejectsConcurrentCallBeforeSecondTransportRequest` | existing coverage | PASS |
| B16 | a retryable status is retried within `MaxRetries` | `TestAdapterEnforcesResponseBytesWindowAndRetryAfter` | existing coverage | PASS |
| B17 | the absolute call window refuses the call after the ceiling | `TestAdapterEnforcesResponseBytesWindowAndRetryAfter` | existing coverage | PASS |
| B18 | a transport error is classified rather than returned verbatim | `TestTransportErrorsCarryNoCredentialOrURL` | the raw `*url.Error` came back and its message carried `crtfc_key=<secret>` | PASS |
| B19 | a deadline/cancellation keeps its sentinel and loses the URL | `TestTransportErrorsCarryNoCredentialOrURL` | `errors.Is` held but the message contained the credential | PASS |
| B20 | retry exhaustion wraps a redacted cause and stays `errors.Is`-recoverable | `TestTransportErrorsCarryNoCredentialOrURL` | `%v` both leaked the URL and made the cause unrecoverable | PASS |
| B21 | an oversized body refuses | `TestAdapterEnforcesResponseBytesWindowAndRetryAfter` | existing coverage | PASS |
| B22 | a 2xx body is copied out with its observed time | `TestSECOfficialAdapterCollectsFrozenPaginatedFixtureWithDeclaredIdentity` | existing coverage | PASS |
| B23 | 401/403 becomes a credential refusal | `TestOfficialAdaptersDoNotCommitIncompleteOrUntrustedBatches` | existing coverage | PASS |
| B24 | a non-retryable status becomes incomplete | `TestOfficialAdaptersDoNotCommitIncompleteOrUntrustedBatches` | existing coverage | PASS |
| B25 | a retryable status with attempts left waits and retries | `TestAdapterEnforcesResponseBytesWindowAndRetryAfter` | existing coverage | PASS |
| B26 | a retryable status on the last attempt does not wait | `TestAdapterEnforcesResponseBytesWindowAndRetryAfter` | existing coverage | PASS |
| B27 | a final 429 is reported as rate limiting, not as exhaustion | `TestAdapterEnforcesResponseBytesWindowAndRetryAfter` | existing coverage | PASS |
| B28 | an out-of-range Retry-After refuses instead of sleeping | `TestAdapterEnforcesResponseBytesWindowAndRetryAfter` | existing coverage | PASS |
| B29 | a cancelled bounded wait returns the context error | `TestAdapterEnforcesResponseBytesWindowAndRetryAfter` | existing coverage | PASS |
