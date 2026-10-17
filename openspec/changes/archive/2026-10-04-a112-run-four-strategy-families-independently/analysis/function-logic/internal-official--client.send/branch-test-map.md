# Branch Test Map: `Client.send`

- Source: `internal/official/client.go`, SHA-256 `6d145916a6797f3a32842561346c8767585a951d8883ebd81bf8af4632106639`; branch IDs follow `ast.json` (generated 2026-08-17).
- AST counts: branches 9, returns 8, calls 8, defers 0, go statements 0. Source range `320:1`–`366:2`. Signature `(c *Client) send(ctx context.Context, makeReq func(tok string) (*http.Request, error), out any) error`.
- Citation-only bundle: this function is NOT edited by a112; its branch enumeration is evidence for the L1b brief (official raw reader + bar producer). Any later body edit requires a fresh RED/BTM.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 322:2 — the token manager cannot supply a token | not-applicable: no test injects that failure inside `send`; the manager's own failures are covered by `TestTokenExchangeError` and `TestTokenEmptyAccessTokenErrors` | n/a (not edited) | not-applicable |
| B2 | if at 326:2 — the first `makeReq` cannot build a request | not-applicable: defensive, reachable only through a malformed URL or method | n/a (not edited) | not-applicable |
| B3 | if at 330:2 — the first attempt fails in transport | not-applicable: no transport-failure injection in this package | n/a (not edited) | not-applicable |
| B4 | for at 344:2 — first attempt answers 401, the loop refreshes and re-issues; a 2xx first attempt must not enter it | `TestGetUnwrapsEnvelopeAndRetriesOn401`, `TestDeleteAcctRetriesOn401`, `TestAttemptObserverTraces401RefreshThenSuccessfulRetry`, `TestARefusedProcessAdoptsTheTokenAnotherProcessAlreadyGot`, `TestARotationThatLandsMidRequestCostsNoToken`, `TestPostSendsJSONAndUnwrapsEnvelope` (untaken side) | n/a (not edited) | existing suite green |
| B5 | if at 347:3 — `refresh` fails inside the loop | not-applicable: no test fails the token endpoint mid-retry | n/a (not edited) | not-applicable |
| B6 | if at 351:3 — the retry `makeReq` cannot build a request | not-applicable: defensive, same as B2 | n/a (not edited) | not-applicable |
| B7 | if at 355:3 — the retry fails in transport | not-applicable: same as B3 | n/a (not edited) | not-applicable |
| B8 | if at 358:3 — an exchanged (not adopted) token breaks after one retry; an adopted-but-dead token is allowed a second pass that ends on a minted one | `TestGetUnwrapsEnvelopeAndRetriesOn401` and `TestDeleteAcctRetriesOn401` (both assert exactly two data calls), `TestARefusedProcessWithNothingToAdoptStillExchanges`, `TestAnAdoptedTokenThatIsAlsoRefusedStillEndsOnAMintedOne` (untaken side) | n/a (not edited) | existing suite green |
| B9 | if at 362:2 — non-2xx is classified rather than decoded (500, 429, and a 401 that survives the loop) | `TestGetNon2xxReturnsClassifiedError`, `TestRawReadsClassifyErrorsLikeEveryOtherRead` | n/a (not edited) | existing suite green |

Cross-process properties the loop exists for (calls, not branches): `TestTwoProcessesSharingOneCacheFileStopBuyingTokensFromEachOther` bounds twelve rounds across two clients to three exchanges, and `TestASiblingGoroutineThatAlreadyReplacedTheTokenIsNotOutbought` covers two goroutines refused on the same stale token.

Verification: `go test ./internal/official -count=1` green on 2026-08-17 (exit 0). No RED round applies — a112 does not edit this function.
