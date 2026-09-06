# Function Logic Map: `OfficialAdapter.collectSEC`

- Source: `internal/strategyevidence/official_source.go`
- AST evidence: `ast.json` (extracted at the re-baselined comparison base)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `request.EntityID` | exactly ten digits | the caller | `ErrSourceSchemaDrift`, zero requests |
| first page body | the frozen `sec-submissions-v1` shape whose `cik` equals the requested one | SEC | `ErrSourceSchemaDrift` |
| `filings.files[].name` | `CIK<the requested CIK>-submissions-<3 digits>.json` | the remote body — untrusted | `ErrSourceSchemaDrift` before the page is requested |
| byte and page budget | within the request's and the policy's | `SourcePolicy` | `ErrSourceBoundExceeded` / `ErrSourceIncomplete` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | the CIK is not ten digits | none | `ErrSourceSchemaDrift` | `TestOfficialAdaptersDoNotCommitIncompleteOrUntrustedBatches` |
| B2 | the first page fetch fails | none | the fetch error | `TestOfficialAdaptersDoNotCommitIncompleteOrUntrustedBatches` |
| B3 | the first page is not the frozen shape, or its CIK/name/tickers/exchanges do not hold | none | `ErrSourceSchemaDrift` | `TestOfficialAdaptersDoNotCommitIncompleteOrUntrustedBatches` |
| B4 | the recent-filings block does not parse | none | the parse error | `TestOfficialAdaptersDoNotCommitIncompleteOrUntrustedBatches` |
| B5 | the pages the body demands exceed the request or policy page limit | none | `ErrSourceIncomplete` | `TestOfficialAdaptersDoNotCommitIncompleteOrUntrustedBatches` |
| B6 | walk each historical page the first page declares | accumulates records and bytes | falls through | `TestSECOfficialAdapterCollectsFrozenPaginatedFixtureWithDeclaredIdentity` |
| B7 | a declared page name is not the frozen SEC resource shape for this CIK, or its filing count is negative | none — the page is never requested | `ErrSourceSchemaDrift` | `TestSECHistoricalPageResourceIsNotChosenByTheRemoteBody` |
| B8 | a historical page fetch fails | none | the fetch error | `TestOfficialAdaptersDoNotCommitIncompleteOrUntrustedBatches` |
| B9 | the accumulated bytes exceed the request limit | none | `ErrSourceBoundExceeded` | `TestOfficialAdaptersDoNotCommitIncompleteOrUntrustedBatches` |
| B10 | a historical page does not parse | none | `ErrSourceSchemaDrift` | `TestOfficialAdaptersDoNotCommitIncompleteOrUntrustedBatches` |
| B11 | a historical page's records do not parse, or its count contradicts the declared one | none | `ErrSourceSchemaDrift` | `TestOfficialAdaptersDoNotCommitIncompleteOrUntrustedBatches` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `a.fetchPage` | one bounded request per page, under the remaining byte budget | errors surface unwrapped and stop the collection | AST + `TestSECOfficialAdapterCollectsFrozenPaginatedFixtureWithDeclaredIdentity` |
| `validSECHistoricalPageName` | refuse a page path the remote body chose | pure | AST + `TestSECHistoricalPageResourceIsNotChosenByTheRemoteBody` |
| `decodeNoDuplicateJSON` | reject duplicate keys and trailing JSON | any error is schema drift | AST + `TestOfficialAdaptersDoNotCommitIncompleteOrUntrustedBatches` |
| `secRecords` | map one page into immutable official records | any error is schema drift | AST + `TestSECOfficialAdapterCollectsFrozenPaginatedFixtureWithDeclaredIdentity` |
| `completeOfficialBatch` | seal the batch and digest it | returns the marshal error | AST + `TestSECOfficialAdapterCollectsFrozenPaginatedFixtureWithDeclaredIdentity` |

## State mutations and fallbacks

- In-memory only: the record slice and the byte counter grow. Nothing is written to evidence.db here.
- There is no fallback endpoint: a page whose declared name is not the frozen shape is refused, not fetched from somewhere else.

## Safety conclusion

- High-risk in the sense the spec means: `filings.files[].name` is read out of a remote response and becomes the next request's `Resource`. The old guard only asked whether it was blank, so a spoofed upstream chose the path — the spec's 'do not use an unofficial endpoint as a fallback' prohibition (issues.md I16).
- Fail-closed scope, stated: B7 accepts exactly `CIK<the requested CIK>-submissions-<3 digits>.json`. The frozen fixture and real SEC submissions both use that form, so it refuses no normal input; `TestSECHistoricalPageResourceAcceptsTheFrozenFixtureShape` is the positive control. If SEC changes the form, collection stops with `ErrSourceSchemaDrift`, which is the correct outcome for an evidence source.
