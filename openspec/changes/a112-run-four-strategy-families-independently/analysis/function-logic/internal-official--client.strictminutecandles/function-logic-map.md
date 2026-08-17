# Function Logic Map: `Client.StrictMinuteCandles`

- Source: `internal/official/strict_minute_candles.go`
- Source SHA-256: `441bed46f81bc928cab03d512b3ff1305c0c663cb1b58027986e2e91b739977d` (current worktree; `sha256sum` verified equal to `ast.json` `source_sha256`, 2026-08-17)
- Signature: `(c *Client) StrictMinuteCandles(ctx context.Context, market, symbol string, count int, before string) (StrictMinutePage, error)`
- Source range: `120:1`–`185:2` (ast.json `start`/`end`)
- AST evidence: `ast.json` generated 2026-08-17; branches 11, returns 9, calls 25, assignments 11, defers 0, go statements 0.
- Disposition: New function (lot L1b, not in the frozen base 016da624); AST generated 2026-08-17 after GREEN; branch enumeration is the evidence for the L1b acceptance record.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- One strict page read for both markets (decision 13). `RawMinuteCandles` is untouched, so a047's 5-minute identity and its US-refusal regression keep their goldens; the KR identity preserved here is the endpoint, the parameters and the raw DTO (`/api/v1/candles`, `interval=1m`, `adjusted=false`, raw decimal and timestamp strings).
- Arguments are validated **before** any request so a malformed call never spends the shared credential's quota: `market` must be exactly `KR` or `US` (D3 — the producer upper-cases), `symbol` must match `^[0-9]{6}$` (KR) or `^[A-Z][A-Z0-9.]{0,9}$` (US), `count` must be 1..200, and `before`, when given, must satisfy the timestamp grammar and name an instant that exists.
- Query encoding is `url.Values.Encode()`: single percent-encoding, `+` as `%2B`, canonical order `adjusted, before, count, interval, symbol` — the measured request literal.
- Token, retry and transport stay on the ordinary production path (`c.get → send → doRequest`). This function adds no retry, never touches token/refresh/exchange, and passes `out == nil` so `unwrapAndDecode` returns without a standard decode; it reads the **bytes** through a **chained** `AttemptObserver` so any observer already on the context still fires (the M0 raw-bytes precedent must not be shadowed).
- The page is built from one attempt only: `ReadAt` is that attempt's `BodyReadComplete`, `StatusCode` its status, `BodyDigest` the `sha256:` of exactly those bytes. `Budget` is the advisory `RateBudget` for `/api/v1/candles` read after the call.
- `StrictMinutePage` is an exported plain struct so the producer's fake reader can build one; a static guard confines the composite literal to this file.

## Branches and early returns

Exact AST return nodes: `122, 126, 129, 132, 139, 162, 167, 171, 174`.

| Branch | AST kind | Source location | Meaning (one short clause) | Test disposition |
|---|---|---|---|---|
| B1 | if | 121:2 | nil receiver → `CLIENT_MISSING` | `TestStrictMinuteCandlesRefusesANilClient` |
| B2 | if | 125:2 | market is not exactly `KR`/`US` (D3) → `MARKET_INVALID`; also fixes the expected currency | `TestStrictMinuteCandlesRefusesBadArgumentsBeforeAnyRequest` (subtests `lower case market`, `unknown market`) |
| B3 | if | 128:2 | symbol fails the market's grammar → `SYMBOL_INVALID` | `TestStrictMinuteCandlesRefusesBadArgumentsBeforeAnyRequest` (subtests `kr symbol with letters`, `kr symbol too short`, `us symbol lower case`, `us symbol too long`, `us symbol leading digit`, `empty symbol`) |
| B4 | if | 131:2 | `count` outside 1..200 → `COUNT_INVALID` | `TestStrictMinuteCandlesRefusesBadArgumentsBeforeAnyRequest` (subtests `count zero`, `count above the page cap`, `count negative`) |
| B5 | if | 136:2 | a `before` was supplied → validate it before the request | taken: `TestStrictMinuteCandlesSendsTheCanonicalQueryAndReturnsThePage`; untaken: `TestStrictMinuteCandlesAcceptsTheKoreanMarket` (query without `before`) |
| B6 | if | 138:3 | `before` fails the grammar or names an instant that does not exist → `BEFORE_INVALID`, no request | `TestStrictMinuteCandlesRefusesBadArgumentsBeforeAnyRequest` (subtests `before with a zulu offset`, `before without an offset`, `before with four fractional digits`, `before with a named zone`, `before that does not exist`) |
| B7 | if | 148:2 | a `before` was supplied → add it to the query | `TestStrictMinuteCandlesSendsTheCanonicalQueryAndReturnsThePage` (asserts `before=2026-08-15T05%3A00%3A00.000%2B09%3A00`), `TestStrictMinuteCandlesAcceptsTheKoreanMarket` (absent) |
| B8 | if | 157:3 | an observer already on the context is still called (chained, not shadowed) | `TestStrictMinuteCandlesUsesTheLastSuccessfulAttemptAndChainsTheOuterObserver` (outer observer sees both the 401 and the 200) |
| B9 | if | 161:2 | `c.get` failed (transport error or `classifyStatus`) → propagate, no page | `TestStrictMinuteCandlesPropagatesTransportClassification` (subtests `rate limited`, `server error`; asserts the error is not a contract refusal and the page is empty) |
| B10 | if | 166:2 | no usable attempt (ruling 28: the last attempt must be 2xx) → `NO_SUCCESSFUL_ATTEMPT` | not-applicable: unreachable through today's `send` (D6, declared) — see `internal-official--strictminutefinalattempt`; the untaken side is pinned by `TestStrictMinuteCandlesUsesTheLastSuccessfulAttemptAndChainsTheOuterObserver` |
| B11 | if | 170:2 | the strict decode refused the body → propagate the typed refusal | `TestStrictMinuteCandlesRefusesMalformedBodies` (45 body subtests, each asserting the exact `StrictReason*`) |

## Calls and live bindings

| Callee expression | Source location | Evidence |
|---|---|---|
| `strictMinuteMarketCurrency(market)` | 124 | market → `KRW`/`USD`, the currency every candle must carry |
| `strictMinuteCheckSymbol(market, symbol)` | 128 | per-market symbol grammar |
| `strictMinuteInstant(before)` | 137 | the `before` grammar is applied before the request, not after |
| `query.Set` ×5 | 144–149 | canonical query; literal asserted by `TestStrictMinuteCandlesSendsTheCanonicalQueryAndReturnsThePage` |
| `ctx.Value(attemptObserverKey{})`, `WithAttemptObserver` | 153, 155 | **live binding** — chains onto any existing observer (M0 precedent `M0ReadSource.ConditionalOrderRaw/OrderRawByID`) |
| `c.get(traced, PathMinuteCandles, query, nil)` | 161 | **live binding** — the ordinary token path; `out == nil` means `unwrapAndDecode` returns without a standard decode (`internal-official--unwrapanddecode` B1), and `send`'s ≤2 refresh-on-401 loop stays production behaviour (`internal-official--client.send`) |
| `strictMinuteFinalAttempt(attempts)` | 165 | ruling 28 last-attempt-must-be-2xx selection |
| `strictMinuteDecode(used.Body, count, currency, beforeInstant)` | 169 | the whole body contract (decision 14) |
| `sha256.Sum256(used.Body)`, `hex.EncodeToString` | 173, 182 | `BodyDigest` over the exact bytes; asserted against the served body in `TestStrictMinuteCandlesSendsTheCanonicalQueryAndReturnsThePage` |
| `c.RateBudget(PathMinuteCandles)` | 183 | advisory budget of the same response; `TestStrictMinuteCandlesReportsTheRateBudgetOfTheSameResponse` |

## State mutations and fallbacks

- Locals only (11 AST assignments): `currency`, `beforeInstant`, `query`, `outer`, `attempts`, `traced`, `used`, `candles`/`terminal`/`cursor`, `digest`. The client is not mutated here; no goroutines, no defers, no clock read (`ReadAt` comes from the attempt trace, not from `time.Now`).
- The only side effect is the outbound GET on the ordinary path, and it happens only after every argument has passed. Refusals at B1–B6 send nothing (asserted by `harness.hitCount() == 0`).
- No fallback of any kind: a refused body is never "fixed up", never partially returned. Every failing path returns the zero `StrictMinutePage`.

## Safety conclusion

- High-risk adjacency: this is the L1b entry point onto the official client GET/token path, so a call can drive `send`'s ≤2 refresh-on-401 against a credential shared with neighbouring products (token-war memory). It fails closed in both directions — a malformed argument never reaches the broker, and a body that differs from the measured contract is refused rather than coerced.
- Read-only by construction: no order, cancel, stop-loss or toggle surface is touched, and `RawMinuteCandles` and the rest of the client are unmodified. Nothing in production calls this yet; the producer wires it only when L5 does, under human approval.
- Recorded residuals (review.md 2026-08-17): `doRequest` reads the body uncapped and the 2 MiB cap is applied post-read (transport-wide, accepted); an absent `nextBefore` is refused although the documented schema lists only `candles` as required (fail-closed until a terminal page is measured); `[]json.RawMessage` is allocated before the count bound, inside that 2 MiB cap; ruling 26 makes a single off-minute bar refuse the whole page; and no test drives `c.rates.record` through `doRequest`, so the shared-quota accounting is unit-tested in isolation only.
