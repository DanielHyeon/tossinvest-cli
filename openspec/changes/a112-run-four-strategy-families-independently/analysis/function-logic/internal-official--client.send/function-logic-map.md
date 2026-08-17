# Function Logic Map: `Client.send`

- Source: `internal/official/client.go`
- Source SHA-256: `6d145916a6797f3a32842561346c8767585a951d8883ebd81bf8af4632106639` (current worktree; verified with `sha256sum` 2026-08-17, equal to `source_sha256` in `ast.json`)
- Signature: `(c *Client) send(ctx context.Context, makeReq func(tok string) (*http.Request, error), out any) error` (`ast.json`: `Client.send(params=3, results=1)`)
- Source range: `320:1`–`366:2`
- AST counts: branches 9, returns 8, calls 8, defers 0, go statements 0 (`ast.json` generated 2026-08-17 by `go run ./tools/logic-map`).
- Risk scan: `risk-pattern-report.md`.
- Citation-only bundle: this function is NOT edited by a112; its branch enumeration is evidence for the L1b brief (official raw reader + bar producer). Any later body edit requires a fresh RED/BTM.

## Inputs and invariants

- One authenticated-request policy for every verb. `makeReq` is the only per-verb part (method, body, query, per-request headers) and must be callable twice, because the retry rebuilds the request so the new token is applied and a consumed body reader is not reused (`getWithHeaders` 392–411, `postWithHeaders` 427–449, `deleteAcct` 369–385).
- The token comes from `c.tm.token(ctx)` at 321 — shared token-manager state, not a parameter. The 401 answer is bounded by `for attempt := 0; attempt < 2 && code == http.StatusUnauthorized; attempt++` (344): at most two refresh/re-issue passes, and the second cannot adopt, because `refresh` never returns the token it was refused on.
- `adopted` distinguishes a token taken from the shared cache file (liveness only inferred from its expiry) from one exchanged with the broker. `!adopted` breaks after a single retry (358–360); an adopted token that is also refused is allowed one more pass so the request still ends on a minted token.
- Terminal classification is fixed: non-2xx goes to `classifyStatus(code, body)` (362–363), 2xx goes to `unwrapAndDecode(body, out)` (365) with `out` allowed to be nil.
- Relevant to the L1b brief: a112's own reader is structurally forbidden from calling `.send(` (`a112_mbus_static_test.go`), so this retry/auth policy is cited as the contrast case — the a112 raw reader makes exactly one attempt and never refreshes.

## Branches and early returns

Exact AST return nodes: `323, 327, 331, 348, 352, 356, 363, 365`.

| Branch | AST kind | Source location | Meaning (one short clause) | Test disposition |
|---|---|---|---|---|
| B1 | if | 322:2 | `c.tm.token(ctx)` failed → return the token error unwrapped | not-applicable: no test injects a token-acquisition failure inside `send`; the failure modes themselves are pinned on the manager by `TestTokenExchangeError` and `TestTokenEmptyAccessTokenErrors` |
| B2 | if | 326:2 | first `makeReq(tok)` failed → return | not-applicable: defensive — `makeReq` only fails when `http.NewRequestWithContext` rejects the URL or method (client.go:376–379, 399–402, 436–439) |
| B3 | if | 330:2 | first `c.doRequest(req)` failed (transport) → return | not-applicable: transport-failure path, identical to `doRequest` B1; no test in this package injects one |
| B4 | for | 344:2 | bounded 401 loop: at most two passes, and only while `code == 401` | taken: `TestGetUnwrapsEnvelopeAndRetriesOn401`, `TestDeleteAcctRetriesOn401`, `TestAttemptObserverTraces401RefreshThenSuccessfulRetry`, `TestARefusedProcessAdoptsTheTokenAnotherProcessAlreadyGot`, `TestARotationThatLandsMidRequestCostsNoToken`; untaken (2xx first try): `TestPostSendsJSONAndUnwrapsEnvelope` |
| B5 | if | 347:3 | `c.tm.refresh(ctx, tok)` failed → return | not-applicable: no test makes the token endpoint fail during a `send` retry |
| B6 | if | 351:3 | retry `makeReq(tok)` failed → return | not-applicable: same defensive path as B2 |
| B7 | if | 355:3 | retry `c.doRequest(req)` failed (transport) → return | not-applicable: same transport path as B3 |
| B8 | if | 358:3 | `!adopted` → break, spending exactly one retry on a freshly exchanged token | taken: `TestGetUnwrapsEnvelopeAndRetriesOn401` and `TestDeleteAcctRetriesOn401` both assert `calls == 2`; `TestARefusedProcessWithNothingToAdoptStillExchanges` exchanges because there is nothing to adopt. Untaken (adopted, so the loop runs a second pass): `TestAnAdoptedTokenThatIsAlsoRefusedStillEndsOnAMintedOne` |
| B9 | if | 362:2 | `code < 200 || code >= 300` → `classifyStatus(code, body)` | `TestGetNon2xxReturnsClassifiedError` (500), `TestRawReadsClassifyErrorsLikeEveryOtherRead` (429/401/500 through the same `send`) |

## Calls and live bindings

| Callee expression | Source location | Evidence |
|---|---|---|
| `c.tm.token(ctx)` | 321 | shared token manager; cold load, cache hit and exchange are covered by `TestTokenExchangeAndCache`, `TestTokenColdLoadFromDiskCache` |
| `makeReq(tok)` | 325, 350 | per-verb request builder, called once per attempt so the retry carries the new token |
| `c.doRequest(req)` | 329, 354 | the only HTTP hop; see the `internal-official--client.dorequest` bundle (rate headers, attempt trace, unbounded body read) |
| `c.tm.refresh(ctx, tok)` | 346 | returns `(tok, adopted, err)`; adoption semantics pinned by `TestARefusedProcessAdoptsTheTokenAnotherProcessAlreadyGot` and `TestASiblingGoroutineThatAlreadyReplacedTheTokenIsNotOutbought` |
| `classifyStatus(code, body)` | 363 | maps status onto sentinels without leaking the body (`TestAnAuthRefusalDoesNotCarryTheResponseBody`, `TestClassifyStatus`) |
| `unwrapAndDecode(body, out)` | 365 | 2xx envelope unwrap; see the `internal-official--unwrapanddecode` bundle |

## State mutations and fallbacks

- Locals only inside the frame: `tok`, `req`, `code`, `body`, `err`, `attempt`, `adopted` (8 AST assignments). Shared state is mutated indirectly: `c.tm.refresh` rewrites the process-wide token cache and the on-disk cache file shared with sibling processes, and `c.doRequest` writes the per-`Client` rate-budget map.
- No fallback inside the function: every error path returns, and the loop cannot run more than twice regardless of how many 401s arrive. There is no sleep, no backoff and no jitter — a 429 is not retried here at all, it goes straight to `classifyStatus`.
- The retry rebuilds the request rather than reusing it, so a POST body is re-read from the encoded byte slice captured in `postWithHeaders` (430) rather than from a drained reader.

## Safety conclusion

- Safe edit boundary: none of the loop's four exit conditions may be widened without a fresh RED. Raising the bound above two passes, or removing the `!adopted` break, converts a shared-cache rotation into an exchange storm — the exact production defect a082 measured (a 24-hour token re-exchanged seven times a minute).
- High-risk impact: yes. This is the authentication path for every order, cancel, and protection write (`orders_write.go`, `conditional_writes.go` all reach it), and `ErrAuth` raised here latches the execution entry gate. Weakening the classification at B9, or letting a 401 fall through as a success, would be a High-risk change.
- Untested branches are six: B1, B2, B3, B5, B6, B7 — one token-acquisition failure, two request-builder defensive paths, and three transport-failure paths. None of them changes a decision; each returns the underlying error unchanged. The decision branches (B4, B8, B9) are covered. Package suite green (`go test ./internal/official -count=1`, 2026-08-17).
