# Function Logic Map: `Client.StrictMinuteCandles`

- Source: `internal/official/strict_minute_candles.go`
- Source SHA-256: `d32181a939f298db306f492b488468b5925ac0ba97dad3f82cb1cb3286254ced` (current worktree; `sha256sum` verified equal to `ast.json` `source_sha256`, 2026-08-18)
- Signature: `(c *Client) StrictMinuteCandles(ctx context.Context, market, symbol string, count int, before string) (StrictMinutePage, error)`
- Source range: `131:1`–`196:2` (ast.json `start`/`end`)
- AST evidence: `ast.json` regenerated 2026-08-18 against the decision-30 sources; branches 11, returns 9, calls 25, assignments 11, defers 0, go statements 0.
- Disposition: New function (lot L1b, not in the frozen base 016da624); AST regenerated 2026-08-18 against the decision-30 sources; branch enumeration is the evidence for the L1b acceptance record.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- One strict page read for both markets (decision 13). `RawMinuteCandles` is untouched, so a047's 5-minute identity and its US-refusal regression keep their goldens; the KR identity preserved here is the endpoint, the parameters and the raw DTO (`/api/v1/candles`, `interval=1m`, `adjusted=false`, raw decimal and timestamp strings).
- Arguments are validated **before** any request so a malformed call never spends the shared credential's quota: `market` must be exactly `KR` or `US` (D3 — the producer upper-cases), `symbol` must match `^[0-9]{6}$` (KR) or `^[A-Z][A-Z0-9.]{0,9}$` (US), `count` must be 1..200, and `before`, when given, must satisfy the timestamp grammar and name an instant that exists.
- **Decision 30 (2026-08-18), documentation-only for this function.** Every instant that crosses this boundary — each returned `RawMinuteCandle.Timestamp`, the `before` argument and the returned `NextBefore` — is the bar's **close** instant, not its open (human-run US probe, 2026-08-18 03:29 KST: the bar labelled `03:30:00` was still growing while the wall clock sat inside `[03:29, 03:30)`; review.md). This reader is a transport DTO and **converts nothing**: it hands the broker's string through verbatim, and turning a label into an open instant (`− 60 s`) is the producer's job (`internal/officialbars`, `internal-officialbars--adoptpage`). The convention is now documented at the field (`StrictMinutePage.Candles`, source 66–73) and at this function (source 126–130); no branch, argument or return value changed, and the AST counts are identical to the pre-correction bundle.
- Because `before` is an **inclusive** upper bound on that close instant, a request with `before = now` already excludes the still-forming bar: the forming bar's close instant has not happened yet. The producer keeps decision 6's "never admit `bars[0]`" rule anyway, since decision 6 wants an *observed* successor as proof.
- Query encoding is `url.Values.Encode()`: single percent-encoding, `+` as `%2B`, canonical order `adjusted, before, count, interval, symbol` — the measured request literal.
- Token, retry and transport stay on the ordinary production path (`c.get → send → doRequest`). This function adds no retry, never touches token/refresh/exchange, and passes `out == nil` so `unwrapAndDecode` returns without a standard decode; it reads the **bytes** through a **chained** `AttemptObserver` so any observer already on the context still fires (the M0 raw-bytes precedent must not be shadowed).
- The page is built from one attempt only: `ReadAt` is that attempt's `BodyReadComplete`, `StatusCode` its status, `BodyDigest` the `sha256:` of exactly those bytes. `Budget` is the advisory `RateBudget` for `/api/v1/candles` read after the call.
- `StrictMinutePage` is an exported plain struct so the producer's fake reader can build one; a static guard confines the composite literal to this file.

## Branches and early returns

Exact AST return nodes: `133`, `137`, `140`, `143`, `150`, `173`, `178`, `182`, `185`.

| Branch | AST kind | Source location | Meaning (one short clause) | Test disposition |
|---|---|---|---|---|
| B1 | if | 132:2 | nil receiver → `CLIENT_MISSING` | `TestStrictMinuteCandlesRefusesANilClient` |
| B2 | if | 136:2 | market is not exactly `KR`/`US` (D3) → `MARKET_INVALID`; also fixes the expected currency | `TestStrictMinuteCandlesRefusesBadArgumentsBeforeAnyRequest` (subtests `lower case market`, `unknown market`) |
| B3 | if | 139:2 | symbol fails the market's grammar → `SYMBOL_INVALID` | `TestStrictMinuteCandlesRefusesBadArgumentsBeforeAnyRequest` (subtests `kr symbol with letters`, `kr symbol too short`, `us symbol lower case`, `us symbol too long`, `us symbol leading digit`, `empty symbol`) |
| B4 | if | 142:2 | `count` outside 1..200 → `COUNT_INVALID` | `TestStrictMinuteCandlesRefusesBadArgumentsBeforeAnyRequest` (subtests `count zero`, `count above the page cap`, `count negative`) |
| B5 | if | 147:2 | a `before` was supplied → validate it before the request | taken: `TestStrictMinuteCandlesSendsTheCanonicalQueryAndReturnsThePage`; untaken: `TestStrictMinuteCandlesAcceptsTheKoreanMarket` (query without `before`) |
| B6 | if | 149:3 | `before` fails the grammar or names an instant that does not exist → `BEFORE_INVALID`, no request | `TestStrictMinuteCandlesRefusesBadArgumentsBeforeAnyRequest` (subtests `before with a zulu offset`, `before without an offset`, `before with four fractional digits`, `before with a named zone`, `before that does not exist`) |
| B7 | if | 159:2 | a `before` was supplied → add it to the query | `TestStrictMinuteCandlesSendsTheCanonicalQueryAndReturnsThePage` (asserts `before=2026-08-15T05%3A00%3A00.000%2B09%3A00`), `TestStrictMinuteCandlesAcceptsTheKoreanMarket` (absent) |
| B8 | if | 168:3 | an observer already on the context is still called (chained, not shadowed) | `TestStrictMinuteCandlesUsesTheLastSuccessfulAttemptAndChainsTheOuterObserver` (outer observer sees both the 401 and the 200) |
| B9 | if | 172:2 | `c.get` failed (transport error or `classifyStatus`) → propagate, no page | `TestStrictMinuteCandlesPropagatesTransportClassification` (subtests `rate limited`, `server error`; asserts the error is not a contract refusal and the page is empty) |
| B10 | if | 177:2 | no usable attempt (ruling 28: the last attempt must be 2xx) → `NO_SUCCESSFUL_ATTEMPT` | not-applicable: unreachable through today's `send` (D6, declared) — see `internal-official--strictminutefinalattempt`; the untaken side is pinned by `TestStrictMinuteCandlesUsesTheLastSuccessfulAttemptAndChainsTheOuterObserver` |
| B11 | if | 181:2 | the strict decode refused the body → propagate the typed refusal | `TestStrictMinuteCandlesRefusesMalformedBodies` (45 body subtests, each asserting the exact `StrictReason*`) |

## Calls and live bindings

| Callee expression | Source location | Evidence |
|---|---|---|
| `strictMinuteMarketCurrency(market)` | 135 | market → `KRW`/`USD`, the currency every candle must carry |
| `strictMinuteCheckSymbol(market, symbol)` | 139 | per-market symbol grammar |
| `strictMinuteInstant(before)` | 148 | the `before` grammar is applied before the request, not after |
| `query.Set` ×5 | 155–160 | canonical query; literal asserted by `TestStrictMinuteCandlesSendsTheCanonicalQueryAndReturnsThePage` |
| `ctx.Value(attemptObserverKey{})`, `WithAttemptObserver` | 164, 166 | **live binding** — chains onto any existing observer (M0 precedent `M0ReadSource.ConditionalOrderRaw/OrderRawByID`) |
| `c.get(traced, PathMinuteCandles, query, nil)` | 172 | **live binding** — the ordinary token path; `out == nil` means `unwrapAndDecode` returns without a standard decode (`internal-official--unwrapanddecode` B1), and `send`'s ≤2 refresh-on-401 loop stays production behaviour (`internal-official--client.send`) |
| `strictMinuteFinalAttempt(attempts)` | 176 | ruling 28 last-attempt-must-be-2xx selection |
| `strictMinuteDecode(used.Body, count, currency, beforeInstant)` | 180 | the whole body contract (decision 14) |
| `sha256.Sum256(used.Body)`, `hex.EncodeToString` | 184, 193 | `BodyDigest` over the exact bytes; asserted against the served body in `TestStrictMinuteCandlesSendsTheCanonicalQueryAndReturnsThePage` |
| `c.RateBudget(PathMinuteCandles)` | 194 | advisory budget of the same response; `TestStrictMinuteCandlesReportsTheRateBudgetOfTheSameResponse` |

## State mutations and fallbacks

- Locals only (11 AST assignments): `currency`, `beforeInstant`, `query`, `outer`, `attempts`, `traced`, `used`, `candles`/`terminal`/`cursor`, `digest`. The client is not mutated here; no goroutines, no defers, no clock read (`ReadAt` comes from the attempt trace, not from `time.Now`).
- The only side effect is the outbound GET on the ordinary path, and it happens only after every argument has passed. Refusals at B1–B6 send nothing (asserted by `harness.hitCount() == 0`).
- No fallback of any kind: a refused body is never "fixed up", never partially returned. Every failing path returns the zero `StrictMinutePage`.

## Safety conclusion

- High-risk adjacency: this is the L1b entry point onto the official client GET/token path, so a call can drive `send`'s ≤2 refresh-on-401 against a credential shared with neighbouring products (token-war memory). It fails closed in both directions — a malformed argument never reaches the broker, and a body that differs from the measured contract is refused rather than coerced.
- Read-only by construction: no order, cancel, stop-loss or toggle surface is touched, and `RawMinuteCandles` and the rest of the client are unmodified. Nothing in production calls this yet; the producer wires it only when L5 does, under human approval.
- Recorded residuals (review.md 2026-08-17): `doRequest` reads the body uncapped and the 2 MiB cap is applied post-read (transport-wide, accepted); an absent `nextBefore` is refused although the documented schema lists only `candles` as required (fail-closed until a terminal page is measured); `[]json.RawMessage` is allocated before the count bound, inside that 2 MiB cap; ruling 26 makes a single off-minute bar refuse the whole page; and no test drives `c.rates.record` through `doRequest`, so the shared-quota accounting is unit-tested in isolation only.
