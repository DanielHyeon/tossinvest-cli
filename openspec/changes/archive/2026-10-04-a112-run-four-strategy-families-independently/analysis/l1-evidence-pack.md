# L1 evidence pack — official source / evidence authority (read-only, pre-edit)

Scope: a112 lot **L1** (`tasks.md:10`). This file is the Manager's pre-implementation
evidence pack for the Terra L1 implementer. It records CodeGraph hard evidence and
current-worktree line citations for every symbol L1 touches or must not touch. It is
not permission to edit anything; ownership is `tasks.md:5-16`.

- Worktree: `feat/a112-four-family-runtime` @ `016da624` with untracked L2/M-B0/M-B1
  files present (`git status --short`: `internal/breakoutlane/`, `internal/official/a112_mbus_*`,
  `tools/a112-mb-us-source/`, `openspec/changes/a112-…/`).
- CodeGraph: `codegraph_status` reports 1930 files / 33 853 nodes / 110 383 edges, WAL,
  Go 1618 files; the index already contains the untracked files (e.g. `NewClosedBar`
  resolves to `internal/breakoutlane/types.go:110`; `A112MBUSCandle` callers include
  `tools/a112-mb-us-source/official_reader.go:20`). Every CodeGraph row below was
  re-confirmed by `Read`/`grep` on the current worktree; rows that CodeGraph could not
  answer are marked **grep-derived (advisory)**.
- Market-string caveat that recurs everywhere below: `marketclock.Market` is lowercase
  `"kr"/"us"` (`internal/clock/market.go:24-26`); `breakoutlane.Market` and the L0 golden
  setup-ID vector use uppercase `"KR"/"US"` (`internal/breakoutlane/types.go:10-11`,
  `analysis/goldens/breakout-evidence-and-sizing-v1.json` `known_vector.market`).

---

## 1. L1 target symbols — CodeGraph hard evidence

### 1.1 `internal/strategyevidence`

Package imports only `internal/clock` (`go list`), so it has no dependency on `official`,
`scheduler`, `strategycandle` or `breakoutlane` today. Production importers of the package
are `internal/continuationlane`, `internal/reversallane`, `internal/weeklyvaluelane`,
`internal/strategyproposal` (`go list -f '{{.Imports}}' ./...`). No `cmd/` or
`internal/app` file references `strategyevidence` (grep: 0 hits outside tests).

| Symbol | Definition | Signature / shape | Callers | Callees | Impact radius (what breaks if changed) |
|---|---|---|---|---|---|
| `EvidenceKind` + consts | `internal/strategyevidence/model.go:18-26` | `type EvidenceKind string`; `KindTradability="tradability"`, `KindDisclosureRisk="disclosure_risk"`, `KindDisclosure="disclosure"`, `KindKRNetFlow="kr_net_flow"`, `KindUSParticipation="us_participation"` | consumed by `kindSupportsMarket` (model.go:217), `Requirement.Kind` (projection.go:20), `Refusal.Kind` (projection.go:15) | — | Adding a const is additive; the DB `evidence_kind` column is free `TEXT NOT NULL` (store.go:129) with no CHECK, so no migration is needed. |
| `SourceAuthority` + consts | `model.go:28-35` | `AuthorityOpenDART="opendart"`, `AuthorityKRX="krx"`, `AuthoritySEC="sec_edgar"`, `AuthorityTossOpenAPI="toss_open_api"` | `authorityValid` (model.go:195), `authoritySupportsMarket` (model.go:204), `MintSourcePolicy` (source.go:76-117) | — | `AuthorityTossOpenAPI` already exists and serves KR **and** US (model.go:210-211); L1 bars/quotes should use it, no new authority const required. |
| `authoritySupportsMarket` | `model.go:204-215` | `(authority SourceAuthority, market marketclock.Market) bool` | `normalizeHeader` (model.go:166); test `TestOfficialAuthorityMarketMatrix` (model_test.go:110) | — | Toss OpenAPI branch (`:210-211`) returns true for KR and US. **No edit needed for L1** unless a new authority is minted. |
| `kindSupportsMarket` | `model.go:217-228` | `(kind EvidenceKind, market marketclock.Market) bool`; exact switch, `default: false` (`:225-226`) | `normalizeHeader` (model.go:168), `choose` (projection.go:115); CodeGraph impact (depth 2): 7 symbols — `normalizeHeader`, `NewEnvelope`, `choose`, `Project` | — | **Must be edited** for any new kind: an unknown kind is refused at `NewEnvelope` → `normalizeHeader:168-169` with `EVIDENCE_UNSUPPORTED`. FLM exists and is current (see §3). |
| `Header` | `model.go:83-104` | fields: `EvidenceID, Market marketclock.Market, Symbol, IssuerIdentity, IssuerMappingVersion, Kind, SchemaVersion, Authority, SourceRecordID, RevisionIdentity, SupersedesRevisionIdentity, MarketEffectiveDate, SourceEventAt, SourceAvailableAt, ObservedAt, IngestedAt, Currency, Unit, Availability, Confidence` | every producer/replayer | — | Struct equality is used by `sameImmutableProvenance` (store.go:265-269); adding a field changes idempotency semantics — **do not add fields**. |
| `NewEnvelope` | `model.go:118-132` | `(header Header, payload []byte) (Envelope, error)`; order: `normalizeHeader` → `canonicalJSON` → `validateTypedPayload(h.Kind, canonical)` → sha256 digest | CodeGraph: `Append` (store.go:204), `Snapshot.Valid` (store.go:302), `scanEnvelope` (store.go:484) + 11 test fixtures across continuationlane/reversallane/weeklyvaluelane/strategyproposal | `normalizeHeader`, `canonicalJSON`, `validateTypedPayload` | Any strictness added inside `validateTypedPayload` is re-applied on Append, on replay (`scanEnvelope`) and on `Snapshot.Valid` — this is the only hook that makes the strict decoder authoritative on the read path too. |
| `normalizeHeader` | `model.go:142-193` | `(in Header) (Header, error)`; trims/upper-cases Symbol+Currency (`:145,153`); requires KR/US (`:156`), identity (`:158`), source identity (`:160`), no self-supersede (`:162`), authority valid/market (`:164-167`), **kind supports market (`:168`)**, `SourceAvailableAt` non-zero (`:170`), all four clocks non-zero (`:172`), ordering `SourceEventAt <= SourceAvailableAt <= ObservedAt <= IngestedAt` (`:174-175`), currency/unit/availability/confidence (`:176-183`), `MarketEffectiveDate` = `YYYY-MM-DD` (`:185`), then UTC-normalizes the four times (`:188-191`) | `NewEnvelope` | `authorityValid`, `authoritySupportsMarket`, `kindSupportsMarket` | Not to be edited by L1; its `:174` ordering is the constraint L1's producer must satisfy for every bar/quote header. |
| `canonicalJSON` | `model.go:230-248` | `([]byte) ([]byte, error)`; `UseNumber`, top-level object required (`:237-239`), single value (`:241-246`), re-marshals map (sorted keys) | `NewEnvelope` (model.go:123), `decodeNoDuplicateJSON` (official_source.go:357) — CodeGraph aggregated with unrelated same-name funcs in httpapi/journal | `decodeCanonicalValue` | Numbers become `canonicalNumber` text; JSON *strings* pass through untouched → decimal-string prices survive canonicalization byte-for-byte. |
| `decodeCanonicalValue` | `model.go:259-312` | recursive; **duplicate keys rejected** (`:278-280`); arrays allowed (`:291-303`); numbers → `normalizeJSONNumber` (`:307-308`) | `canonicalJSON` | `normalizeJSONNumber` | Provides duplicate-key refusal for free to any payload. |
| `normalizeJSONNumber` | `model.go:314-396` | `(string) (canonicalNumber, error)`; bounds 1024 bytes / exponent 1e6; strips leading/trailing zeros; emits `digits[e±exp]` | `decodeCanonicalValue` | — | JSON numbers are canonicalized (e.g. `1.50`→`15e-1`) — this is why the L1 strict decoder must require **integer JSON numbers** (no fraction/exponent) for minor/PPM fields and reject decimal *numbers*; decimal *strings* from the broker must be converted to integer minor units before payload assembly. |
| `validateTypedPayload` | `model.go:398-428` | `(kind EvidenceKind, canonical []byte) error`; decodes to `map[string]any` (`:401-404`), `rejectSecretFields` (`:405`), then a shallow optional-field type map `{blocked:bool, code:string, score_ppm:number, flow:number, value:number, market:string}` (`:408`); unknown fields are **not** rejected | `NewEnvelope` (model.go:127); CodeGraph impact (depth 3): 53 symbols incl. `Append`, `Snapshot.Valid`, `scanEnvelope`, `SealSnapshot`, `Replay` and every lane fixture test | `rejectSecretFields` | **Must be edited** to dispatch by kind to a strict breakout decoder; today an unknown/float/unbounded payload for a new kind would pass. FLM exists and is current (see §3). |
| `rejectSecretFields` | `model.go:430-450` | recursive over map/array; refuses keys containing `credential|secret|password|token|api_key`, `crtfc_key`, `authorization` after `-`/space→`_` and lower-casing (`:434-436`) | `validateTypedPayload` only (CodeGraph) | — | Reusable as-is by the strict decoder; note `token` substring — L1 field names must avoid e.g. `token_size`. |
| `Store.migrate` | `store.go:97-120` | `(ctx) error`; `PRAGMA user_version` gate: `> SchemaVersion(1)` → `ErrSchemaTooNew` (`:102-104`), `== 1` → no-op (`:105-107`), else applies `schemaV1` in one tx (`:113-119`) | `Open` (store.go:74) | — | Schema (`store.go:122-189`): `evidence_records` (`:123-147`, `evidence_kind TEXT NOT NULL` `:129`, `UNIQUE(authority, source_record_id, revision_identity)` `:146`, STRICT), `evidence_asof_idx(market,symbol,issuer_identity,issuer_mapping_version,source_available_at,ingested_at,market_effective_date)` (`:148`), `evidence_supersedes_idx` (`:149`), `revision_conflicts` (`:150-162`), `snapshots(evaluation_at, ingestion_cutoff, snapshot_digest UNIQUE)` (`:163-172`), `snapshot_items` (`:173-180`), 8 append-only/immutable triggers (`:181-188`). **Kind column is a free string** — additive kind needs no schema change. Bumping `SchemaVersion` would break `OpenReadOnly` (readonly.go:44 requires `version == SchemaVersion`) for existing DBs → avoid. |
| `Store.Append` | `store.go:194-263` | `(ctx, evidence Envelope) (AppendResult, error)`; overrides `IngestedAt` with the store's trusted clock (`:198-203`, refuses if before `ObservedAt`); re-runs `NewEnvelope` (`:204`); same identity+digest+provenance → idempotent (`:220-221`); same identity, different digest/provenance → quarantines into `revision_conflicts` and returns `ErrRevisionConflict` (`:223-235`); `SupersedesRevisionIdentity` must exist and keep market/symbol/issuer/mapping/kind (`:237-245`); INSERT (`:246-255`) | tests only today (CodeGraph: `store_test.go`, `review_hardening_test.go`, lane fixtures) — **no production writer exists** | `NewEnvelope`, `readByIdentity`, `sameImmutableProvenance`, `stamp` | Correction semantics for L1: a corrected bar MUST come as a **new `RevisionIdentity`** with `SupersedesRevisionIdentity` = previous; a re-append of the same revision with different bytes is a conflict, not a correction (spec `breakout-retest-strategy-lane/spec.md:69` "Correction은 기존 row를 덮지 않고 additive revision"). |
| `sameImmutableProvenance` | `store.go:265-269` | `(left, right Header) bool`; blanks `EvidenceID` and `IngestedAt`, then `left == right` | `Append` (`:220`) | — | Struct equality → see `Header` row. |
| `SnapshotQuery` | `store.go:271-278` | `{Market, Symbol, IssuerIdentity, IssuerMappingVersion, EvaluationAt, IngestionCutoff}` — **no Kind filter** | `SealSnapshot`, `snapshotDigest`, `Snapshot.Valid`, `Replay` | — | A breakout snapshot for `(us, AAPL, issuer, mapping)` will include every kind stored under that scope (tradability, disclosure, participation **and** every bar/quote revision ≤ cutoffs). See §5 risk. |
| `Snapshot` / `Snapshot.Valid` | `store.go:280-310` | `{ID, Digest, Market, Symbol, IssuerIdentity, IssuerMappingVersion, EvaluationAt, IngestionCutoff, Items []Envelope}`; `Valid()` rebuilds each item via `NewEnvelope` and recomputes `snapshotDigest` (`:294-310`) | `Replay` consumers | `NewEnvelope`, `snapshotItemMatchesQuery`, `snapshotDigest` | Any strict decoder in `validateTypedPayload` is re-checked here. |
| `Store.SealSnapshot` | `store.go:312-379` | `(ctx, query SnapshotQuery) (Snapshot, error)`; requires both cutoffs (`:316-318`); `localDate = Market.TradingDay(EvaluationAt)` (`:321`); SQL (`:330-343`) selects rows with `market_effective_date <= localDate AND source_event_at <= EvaluationAt AND source_available_at <= EvaluationAt AND ingested_at <= IngestionCutoff` and **no superseding row visible under the same four predicates** (`:334-340`); ordered by `authority, source_record_id, revision_identity, evidence_id` (`:341`); `INSERT OR IGNORE` into `snapshots`/`snapshot_items` (`:364-374`) | tests only today (13 CodeGraph callers, all `_test.go`) | `envelopeColumns`, `scanEnvelope`, `snapshotDigest` | **Dual cutoff mapping to design (`design.md:142,168`)**: design `source_observed_at` ↔ header `SourceEventAt`/`SourceAvailableAt` gated by `EvaluationAt`; design `recorded_at` ↔ header `IngestedAt` gated by `IngestionCutoff` (store-trusted clock, `store.go:198`). `ObservedAt` (client wall clock) is **not** part of the SQL cutoff; it is only used by `Append`'s backdating check (`:200`) and by projection freshness (`projection.go:179`). |
| `snapshotDigest` (strategyevidence) | `store.go:381-402` | `(query SnapshotQuery, items []Envelope) string`; length-prefixed fields (`writeSnapshotDigestField` `:404-409`) over query + every header field + payload digest, items sorted by `EvidenceID` (`:387`) | `SealSnapshot`, `Snapshot.Valid`, `Replay` (consumer.go:68) | — | Distinct from `breakoutlane.snapshotDigest` (arithmetic.go:78). L1's evidence digest lives here; L2's evidence-input digest is computed later by L3 from the same bytes. |
| `OpenReadOnly` | `readonly.go:17-56` | `(ctx, options Options) (*Store, error)`; requires abs path, owner-only 0600 regular non-symlink (`readonly_owner_unix.go:9-19`), `mode=ro`, `query_only(1)`, `user_version == SchemaVersion` (`readonly.go:44`) | production: `LoadProductionAuthorityBatch` (`internal/strategyproposal/production.go:212`); tests | — | Only production reader of evidence.db today. |
| `NewDormantSnapshotReadPort` / `Replay` | `consumer.go:25-72` | `Replay(ctx, market, SnapshotReference) (Snapshot, error)`; SELECT-only (guarded by `TestDormantSnapshotReadPortIsStructurallySelectOnly` `consumer_static_test.go:12-55`, which also forbids `net/http`, `/broker`, `/dispatch`, `/execgw`, `/guardian`, `/operating`, `/runtime`, `/toggle` imports **in `consumer.go`**) | `strategyproposal/production.go:217,238` | `scanEnvelope`, `snapshotItemMatchesQuery`, `snapshotDigest` | L1 must not add imports to `consumer.go`. |
| `snapshotItemMatchesQuery` / `validSnapshotReference` | `consumer.go:74-92` | item scope + `SourceEventAt/SourceAvailableAt <= EvaluationAt`, `IngestedAt <= IngestionCutoff`, `MarketEffectiveDate <= TradingDay(EvaluationAt)`; reference must be `snapshot-<64 lowercase hex>` | `Replay`, `Snapshot.Valid` | — | — |
| `projection.go` API | `Refusal` (`:13`), `Requirement{Kind, Required, MaxAge, Authorities, Currency, Unit}` (`:19-26`), `ProjectionPolicy` (`:28-34`), `Project(snapshot, policy)` (`:55-112`), `choose` (`:114-176`), `evidenceStale` (`:178-195`) | `Project` picks one envelope per required kind by authority priority; conflicts among same-authority digests → `EVIDENCE_CONFLICT` (`:146-148`) | lanes (continuation/reversal/weeklyvalue production_proposal.go) | `kindSupportsMarket` (`:115`) | `choose` selects **exactly one** envelope per kind — it cannot express "an ordered series of 15–512 bars". A breakout kind that stores one row per bar cannot be consumed through `Project`/`choose`; L1's snapshot-assembly must provide its own bounded ordered accessor (additive), or store the assembled series as one envelope. See §5. |
| `OfficialAdapter` / `NewOfficialAdapter` / `Collect` / `fetchPage` | `official_source.go:54-136` | `NewOfficialAdapter(policy SourcePolicy, transport Transport, credentials CredentialProvider, budget *SharedRateBudget) (*OfficialAdapter, error)`; `Collect` switches **only** `AuthoritySEC → collectSEC` and `AuthorityOpenDART → collectDART`, `default → ErrSourceUnavailable` (`:106-113`) | CodeGraph: 3 callers, all `official_source_test.go` — **no production caller** | `Adapter.Fetch` (source.go:328) | Confirmed SEC/DART only. Not reusable for Toss OpenAPI candles without a new authority branch and a `Transport` implementation; `MintSourcePolicy` also refuses anything but SEC/DART (`source.go:95-117`) and refuses KRX outright (`:77`). |
| `decodeNoDuplicateJSON` | `official_source.go:356-368` | `(body []byte, target any) error`; runs `canonicalJSON` for duplicate-key/shape check, then a second `json.Decoder` into `target`, refuses trailing values | `collectSEC` (`:174,199`), `collectDART` (`:274`) | `canonicalJSON` | Reusable pattern for strict envelope decoding; note it does **not** reject unknown fields (`DisallowUnknownFields` not set). |
| `SourcePolicy` / `SourcePolicyConfig` / `MintSourcePolicy` / `validateOfficial` | `source.go:28-185` | contract-sealed policy (`contractSeal` `:51`, `officialSeal` `:132-141`); `Method="GET"`, `AccessContract="official"` (`:87`) | `NewOfficialAdapter`, `Adapter.Fetch` | — | Sealed to SEC/DART endpoint identities (`:98,107,152,156`); a Toss OpenAPI policy would need a new branch — outside L1's minimal path. |
| `Transport` / `TransportRequest` / `TransportResponse` | `source.go:247-266` | `Do(ctx, TransportRequest, Credential) (TransportResponse{Status, Body, RetryAfter, ObservedAt}, error)` | `Adapter.Fetch` (`:392`) | — | Interface is HTTP-agnostic; an official-client-backed `Transport` would be the "adapter" shape if L1 chose to reuse `Adapter.Fetch`'s bounded retry/budget (`:387-426`). |
| `SharedRateBudget` / `NewSharedRateBudget` | `source.go:289-302, 471-513` | per-key window `{from, calls, active}`; `acquire/release/consume` under one mutex; key = `authority/contractID` (official_source.go:71) | `NewOfficialAdapter`, `Adapter.acquire/consumeCall` | — | Process-local only; it does not read broker `X-RateLimit-*` headers. |
| `Adapter.Fetch` | `source.go:328-428` | bounded retry loop `attempt <= MaxRetries` (`:387`), 401/403 → `ErrSourceCredential` (`:409`), retryable set + `RetryAfter` wait (`:412-425`), byte cap (`:403`) | `OfficialAdapter.fetchPage` | `Transport.Do`, `Waiter.Wait` | Retries are policy-driven; L1 must decide whether bar collection is allowed to retry at all (M-B contract was no-retry). |

### 1.2 `internal/official`

Package imports `internal/domain`, `internal/orderintent`, `internal/trading` (`go list`), and
transitively `internal/config`, `internal/exitpolicy`, `internal/orderlineage`,
`internal/settingmeta`, `internal/version`. It contains order-write surfaces
(`orders_write.go`, `conditional_writes.go`). Importing `official` from `strategyevidence`
would create no import cycle (verified: `go list -deps ./internal/official` does not include
`strategyevidence`), but would put the order-capable package inside the evidence package's
dependency closure (task 8.2 dependency guard, `tasks.md:121`). See §3/§5.

| Symbol | Definition | Facts | Callers | Impact for L1 |
|---|---|---|---|---|
| `RawMinuteCandle` | `internal/official/candle_raw.go:13-15` | `struct{ Timestamp, Open, High, Low, Close, Volume, Currency string }` — decimal/timestamp bytes kept as Go strings | `RawMinutePage.Candles()`, strategycandle | Reusable DTO shape. |
| `RawMinutePage` | `candle_raw.go:24-43` | private fields `market, symbol, interval, source string; adjusted bool; candles []RawMinuteCandle; nextBefore string; valid bool`; getters `Valid/Market/Symbol/Interval/Adjusted/Source/NextBefore/Candles` | `strategycandle.AdaptOfficialMinutePage` | Constructible only inside `official`; `nextBefore string` collapses JSON `null`/absent/`""` (candle_raw.go:71 ← `apiCandlePage.NextBefore string` candle_reads.go:37) — pre-edit-targets row "M-B0 KR raw authority boundary" (`analysis/pre-edit-targets.md:10`). |
| `Client.RawMinuteCandles` | `candle_raw.go:47-78` | `(ctx, market, symbol string, count int, before string, adjusted bool) (RawMinutePage, error)`; **US rejection**: `if market != "KR" \|\| !krxRawCandleSymbol.MatchString(symbol) { return …, fmt.Errorf("official raw candles: canonical KRX market/symbol is required") }` (`:51-53`, regexp `^[0-9]{6}$` `:22`); query symbol/interval=1m/count/before/adjusted (`:54-63`); `c.get(ctx, "/api/v1/candles", q, &raw)` into `apiCandlePage` (`:64-66`) | CodeGraph: 5 callers, all tests (`candle_raw_test.go:11,47`, `strategycandle/adapter_test.go:15`, `strategymarket/bars_test.go:24`, `a112_mbus_read_test.go:305`) — **zero production callers** | US refusal is regression-guarded twice (`candle_raw_test.go:49-50`, `a112_mbus_read_test.go:305-309`). L1 must not widen it (`tasks.md:25`, `design.md:146`). |
| Raw decoding path used by `RawMinuteCandles` | `client.go:234-236` (`get`) → `getWithHeaders:392-411` → `send:320-366` → `doRequest:191-207` → `unwrapAndDecode:213-228` | `doRequest`: `io.ReadAll` with **no byte cap** (`:200`), records rate headers via `c.rates.record(readRateBudget(...))` (`:199`), emits `AttemptTrace` with raw body (`:195,202,205`); `send`: token via `c.tm.token(ctx)` (`:321`, may exchange/write cache — `token.go:60-80`), **retries up to twice on 401 with `refresh`** (`:344-361`); `unwrapAndDecode`: `json.Unmarshal(body,&apiEnvelope)` then `json.Unmarshal(env.Result, out)` (`:218,224`) — standard decoder: last-duplicate-key wins, unknown fields ignored, JSON *number* into a `string` field is a type error (fail-closed on bare numbers), `null` cursor → `""` | every official reader (`c.get(` appears in 18 non-test files) | Decimal strings **are** preserved (struct fields are `string`, `candle_reads.go:23-31`) but envelope strictness, cursor tri-state, body cap, no-retry and same-request rate-header binding are **not** provided by this path. |
| `Client.Candles` / `adaptCandles` | `candle_reads.go:48-65` / `:86-111` | `adaptCandles` → `domain.Candle{Open: parseDecimal(...), …}` (`:97-101`), `time.Parse(time.RFC3339)` zero-on-failure (`:90-94`), `FetchedAt: time.Now()` (`:108`) | console/cmd | **FORBIDDEN as L1 authority** (float, lossy). |
| `parseDecimal` | `market_reads.go:15-24` | `strconv.ParseFloat(s, 64)`, returns 0 on error | all official adapters | Proof of float conversion. |
| `Client.Prices` / `adaptPrices` | `market_reads.go:142-150` / `:168-181` | `Last: parseDecimal(p.LastPrice)` (`:173`), `FetchedAt: time.Now()` | discovery/console | **FORBIDDEN as L1 quote authority**. |
| `Client.Orderbook` / `adaptOrderbook` | `market_reads.go:209-217` / `:229-260` | `Price: parseDecimal(a.Price)` (`:236`), bids `:247`; DTOs `apiOrderbookEntry{Price, Volume string}` (`:188-191`), `apiOrderbook{Asks, Bids, Currency, Timestamp}` (`:201-206`) | console | **FORBIDDEN as L1 quote authority**; the *DTO* names match the M-B observed shape `{timestamp,currency,asks[],bids[]}` but level element schema is UNOBSERVED (receipt run4 `observed_broker_behaviour.orderbook`). |
| `Client.AuthorityOrigin` / `authorityOriginLocked` | `client.go:145-155` / `:157-161` | true only when `configurationSealed && authorityOrigin && base == defaultBaseURL && transport == authorityTransport` | M-B0 seam (`a112_mbus_read.go:118`), account/identity authorities | Reusable **opaque proof** for "production origin"; `AuthorityOrigin{production bool}` has private fields (`:137-141`). |
| `Client.get` / `send` / `doRequest` | `client.go:234, 320, 191` | see decoding row; `send` retries/refreshes | ~18 read files, `getAcct/postAcct/deleteAcct` | Read-only in M-B0 (`tasks.md:25`); L1's ownership row does not list `client.go` → treat as read-only in L1 too. |
| Token usage | `token.go:60-80` (`token`), `:83-85` (`isStillValid`, 60 s skew), `:108` (`refresh`), `:130` (`exchange`), `:178` (`loadCache`), `:204` (`saveCache`) | ordinary `get` may exchange and write the cache | `send` | Memory note `tossos-shared-openapi-credential-token-war`: do not touch `token.go`; a production L1 reader going through `c.get` inherits exchange/refresh behaviour and can participate in the token war. |
| `AttemptTrace` / `WithAttemptObserver` / `observeAttempt` | `trace.go:13-19` / `:29-34` / `:36-40` | ctx-carried observer receives `{RequestStart, BodyReadComplete, StatusCode, Body, Err}` **per attempt**; body is the exact bytes before decode | `doRequest` | Available hook to capture raw bytes/status from the ordinary path without editing it; note it does **not** carry response headers. |
| `RateBudget` / `Client.RateBudget` / `readRateBudget` / `headerInt` | `ratebudget.go:148-175` / `:217` / `:227-247` / `:289-299` | last-seen budget per `budgetKey(path)` (`:103-114`, `/api/v1/candles` has no id segment); `headerInt` uses `header.Get` (first value only) — **no cardinality check**; `Reported` false ⇒ absent, never zero (`:170-174`) | `doRequest:199` | Available as an advisory budget observer; not equivalent to M-B0's slice-cardinality validation (`a112_mbus_read.go:348-368`). |
| `ParseRateBudgetReset` | `ratebudget.go:250-273` | epoch vs delta heuristic, `ResetUnparsed` outside plausibility | scheduler/budget | Reset unit is opaque per receipt; L1 must treat Reset as opaque regardless. |
| `Client.MarketCalendar` | `calendar_reads.go:23-40` | `map[string]any` untyped | MCP/console | Not typed. |
| `Client.TypedMarketCalendar` + types | `typed_calendar_reads.go:11-54` | `MarketCalendarResponse{PreviousBusinessDay, Today, NextBusinessDay MarketCalendarDay}`; `MarketCalendarDay{Date, Integrated *…, PreMarket, RegularMarket, AfterMarket *MarketCalendarSession{StartTime, EndTime time.Time}}`; US uses `day.RegularMarket`, KR uses `day.Integrated.RegularMarket` (`internal/scheduler/calendar.go:112-117`) | `scheduler.AdaptOfficialCalendar` | Matches receipt calendar structure `result.today.{dayMarket,preMarket,regularMarket,afterMarket}` for US (dayMarket is not modelled in the typed struct — unknown field, ignored by std decoder). |
| `validateMarketCalendarDate` | `calendar_reads.go:42-54` | exact `YYYY-MM-DD` | `MarketCalendar`, `TypedMarketCalendar`, `A112MBUSCalendar` | reusable. |
| M-B0 seam exported symbols | `a112_mbus_read.go:19-56, 72-94` | `A112MBUSResult` (private fields; getters `Method/Path/CanonicalQuery/StatusCode/Body/CursorJSON/CursorValue/RateHeader/RateHeaders`), `A112MBUSHoldError`, `A112MBUSCandle(ctx, *Client, before []byte)`, `A112MBUSOrderbook(ctx, *Client)`, `A112MBUSCalendar(ctx, *Client, date string)`; private `a112MBUSRead/ReadAt/Cursor/StrictJSONObject/ValidateSuccessRateHeaders/Validate429/CachedToken` | CodeGraph `A112MBUSCandle`: 12 tests + `tools/a112-mb-us-source/official_reader.go:20` | **Not callable from L1** (see §2). |
| Closest reusable production pattern for L1's own raw reader | `RawMinuteCandles` KR path (`candle_raw.go:47-78`) for query construction/`RawMinutePage` shape; `a112MBUSStrictJSONObject` (`a112_mbus_read.go:301-346`) and `a112MBUSCursor` (`:199-231`) as the *reference algorithm* for strict envelope/cursor decode (must be re-implemented, not referenced, per §2); `WithAttemptObserver` for raw body capture; `Client.RateBudget` for advisory budget | — | — |

### 1.3 `internal/strategycandle` (`adapter.go`)

| Symbol | Definition | Facts | Callers |
|---|---|---|---|
| `AdaptOfficialMinutePage` | `internal/strategycandle/adapter.go:13-30` | rejects unless `page.Valid() && Market()=="KR" && Interval()=="1m" && Source()=="official-open-api" && !Adjusted()` (`:14-17`); copies candles into `strategymarket.RawMinuteCandle` and calls `strategymarket.SealAdaptedOfficialMinutePage` (`:26-29`) | `AdaptAndSealClosedKRXFiveMinute` only |
| `AdaptAndSealClosedKRXFiveMinute` | `adapter.go:32-38` | `(market, symbol string, page official.RawMinutePage, now time.Time) (strategymarket.VerifiedBar, error)` → `strategymarket.SealOfficialClosedKRXFiveMinuteFor` | CodeGraph: 1 caller, `adapter_test.go:42` — **zero production callers** |
| Why it rejects US | KR-only guard at `adapter.go:14`; downstream `strategymarket.SealOfficialClosedKRXFiveMinuteFor` requires `market == "KR"` (`internal/strategymarket/bars.go:119`) and `aggregateClosedKRXFiveMinute` hard-codes `Asia/Seoul` (`bars.go:138`), KRX 09:00–15:30 minute-of-day window (`bars.go:158-159`) and `KRW` (`bars.go:165`) | — |
| Static guard | `strategymarket.SealAdaptedOfficialMinutePage` may be referenced in production only from `internal/strategycandle/adapter.go` (`internal/strategymarket/adapter_guard_test.go:55-75`) | L1 must **not** call `SealAdaptedOfficialMinutePage` from any new file. |
| KR bar-label precedent | KR path treats the candle timestamp as **bar open** (bar-start): `minuteOfDay >= 15*60+30` is rejected (`bars.go:158-159`) and `closedAt := minutes[0].local.Add(5*time.Minute)` (`bars.go:189`); the schema comment also says "timestamp … bar open time" (`candle_reads.go:16`) | relevant to §4 bar-label decision |

What strategycandle "seals" for KR today: a 5-minute aggregate `VerifiedBar` (`bars.go:80-93`) — **not** a 1-minute closed bar and **not** an evidence envelope. It is a different product from L1's per-minute append-only evidence; reuse is limited to the DTO copy pattern.

### 1.4 `internal/breakoutlane` (untracked L2, accepted `review.md:76-100`)

Import allowlist is stdlib-only (`internal/breakoutlane/guard_test.go:90-103`: `crypto/sha256, encoding/hex, errors, math/bits, strconv, strings, unicode/utf8`), so `breakoutlane` cannot import `strategyevidence`; conversion evidence→`EvidenceInput` must live in a consumer package (L3 `strategyproposal.buildLaneInput`, `analysis/pre-edit-targets.md:18`).

Ownership table statements: L2 row — "L1 contract may be represented by an inert test-only port, never production authority" (`tasks.md:11`); L1 row lists only `internal/strategyevidence/**` and "additive read-only official KR/US raw bar/quote adapter files and their tests" (`tasks.md:10`); §9 module table puts "closed-bar snapshot tests" under `strategyevidence` and "KR/US adapters, fixtures" under `breakoutlane` (`design.md:218-219`). **Therefore L1 production code must not import `breakoutlane`**; L1 tests may reference the L2 field contract by value.

Exact field contract L1's evidence must be able to produce:

| L2 symbol | Definition | Contract |
|---|---|---|
| `ClosedBarInput` | `types.go:103-107` | `Sequence, Revision, IntervalMS, HighMinor, LowMinor, CloseMinor, RVOLPPM, UpperWickRangePPM uint64; ID, SessionID string; RegularSession, Closed, VolumeExpanded bool` |
| `NewClosedBar` | `types.go:110-115` | rejects `Sequence==0 \|\| Revision==0 \|\| IntervalMS != 60_000 \|\| !canonical(ID) \|\| !canonical(SessionID) \|\| !RegularSession \|\| !Closed \|\| High==0 \|\| Low==0 \|\| Close==0 \|\| Low>High \|\| Close<Low \|\| Close>High` |
| `canonical()` | `arithmetic.go:14-16` | `s != "" && utf8.ValidString(s) && strings.TrimSpace(s)==s && !ContainsRune('\x00')` |
| `validStructuralEvidenceInput` | `machine.go:140-158` | Market ∈ {KR,US}; canonical Symbol/SessionID/CalendarVersion/LaneID; `LaneVersion=="v1"`; `15 <= len(Bars) <= 512`; `ATRMinor>0`; lane ID by market; every bar `NewClosedBar` OK, `bar.SessionID == input.SessionID`, unique `ID`, `Sequence` strictly `prev+1` |
| `QuoteSealInput` / `QuoteSealDigest` / `NewQuoteSeal` | `types.go:117-131` | `BidMinor, AskMinor, LastMinor, SourceObservedAtMS, ReceivedAtMS uint64; Currency, Digest string`; digest = `hashFields("quote.v1", bid, ask, last, currency, srcObs, recv)`; requires all three prices > 0, `Bid <= Ask`, canonical currency, digest match |
| `EvidenceInput` / `NewEvidenceSnapshot` | `types.go:167-190` | `{Market, Symbol, SessionID, CalendarVersion, LaneID, LaneVersion, Config V1Config, Bars []ClosedBar, ATRMinor, EvaluatedAtMS uint64, Quote QuoteSeal, FX FXSeal, Sizing SizingInput}` |
| `snapshotDigest` (L2) | `arithmetic.go:78-86` | `"sha256:" + hashFields("snapshot.v1", …all fields…, per-bar 13 fields)` |
| `setupID` | `machine.go:200-202` | `"sha256:" + hashNUL("tossos.breakout.setup.v1", Market, Symbol, SessionID, CalendarVersion, Bars[0].ID, Bars[14].ID, LaneID, LaneVersion, Config.Digest())` — `opening_range_first_bar_id` **is `Bars[0].ID`**, i.e. the first bar L1 supplies |
| `closedBarContentDigest` / `validCorrectionLineage` | `machine.go:196-198` / `:169-186` | correction is accepted only if same `(sequence,id,sessionID)`, revision non-decreasing, and content changed ⇒ revision strictly increased |
| `Descriptors` | `types.go:50-55` | `KRLaneID="kr_short_breakout_retest_v1"`, `USLaneID="us_short_breakout_retest_v1"`, `LaneVersionV1="v1"` |

Units L1 must therefore emit: prices in **integer minor units** (uint64), ratios in **integer PPM**, `IntervalMS=60000`, timestamps as **uint64 ms**, session/bar identifiers as canonical strings, `Sequence` contiguous within a session, `Revision >= 1`.

### 1.5 `internal/clock` (marketclock)

| Symbol | Definition | Facts |
|---|---|---|
| `Market`, `MarketKR="kr"`, `MarketUS="us"` | `internal/clock/market.go:20-27` | lowercase; `ParseMarket` (`:31-38`) accepts any case |
| `tzName` | `:44-47` | KR `Asia/Seoul`, US `America/New_York` |
| `regularHours` | `:55-58` | KR 09:00–15:30, US 09:30–16:00 local, half-open `[open, close)`; holidays explicitly out of scope ("holidays come from the broker's calendar endpoint" `:49-54`) |
| `Market.Location()` | `:79-92` | cached `*time.Location` |
| `Session{Market, Day, Open, Close, Weekend}` / `Contains` | `:99-118` | half-open, weekend ⇒ false |
| `Market.RegularSession(t)` / `InRegularSession(t)` | `:122-143` / `:146-152` | weekday-only static hours; **not** holiday/early-close aware |
| `Market.TradingDay(t)` | `:158-164` | market-local `YYYY-MM-DD` — used by `SealSnapshot:321` and `snapshotItemMatchesQuery:82` for `market_effective_date` |
| `Market.UTCOffset(t)` | `:183-190` | DST-aware |
| `Clock` interface / `System()` / `NewFake` | `clock.go:35, 73` ; `fake.go` | store trusted clock (`store.go:33,59-62`) |

There is **no** `internal/marketcalendar` package. The official-calendar-derived, versioned
session authority is `internal/scheduler.CalendarSnapshot{Market, Version "sha256:…", Source
"official-openapi", FetchedAt, PreviousBusinessDay/Today/NextBusinessDay CalendarDay{Date,
Regular *SessionWindow{Open, Close}, EarlyClose}}` built by
`scheduler.AdaptOfficialCalendar(market, official.MarketCalendarResponse, fetchedAt)`
(`internal/scheduler/calendar.go:44-104`); its `Version` is fetch-time-independent
(`:85-98`) and is already the engine's `CalendarDigest`
(`internal/app/engine/strategy_proposal_authority.go:194`, `strategy_route_authority.go:183`).
`scheduler` imports `official` and `clock` only (`go list`), so it is import-safe from a new
adapter package. This is the natural source of `calendar_version`; `session_id` has no
existing producer (grep for `SessionID`/`session_id` outside `breakoutlane` finds only
journal weekly-reservation `SessionDate` strings) — L1 must define it (§5).

### 1.6 Engine wiring points (documentation only — L1 must not wire)

| Fact | Evidence |
|---|---|
| The only production reader of evidence.db is `strategyproposal.LoadProductionAuthorityBatch` → `strategyevidence.OpenReadOnly` + `NewDormantSnapshotReadPort` + `Replay` | `internal/strategyproposal/production.go:212-238` |
| Engine passes `evidencePath = <journal dir>/evidence.db` into the proposal loader | `internal/app/engine/strategy_entry_supervisor.go:278-283` |
| There is **no production writer** of evidence.db (no caller of `strategyevidence.Open`, `Store.Append`, `SealSnapshot`, `NewOfficialAdapter` outside tests) | CodeGraph callers of `SealSnapshot` (13, all tests), `NewOfficialAdapter` (3, all tests), `Append` (tests); grep `strategyevidence` in `cmd/`, `internal/app/` = 0 hits |
| Where a later lot would wire a producer: the L5 family-worker/coordinator runtime files and `Context.NewPairedStrategyEntryProductionAssembly` / `refreshPairedStrategyEntryProductionAssembly` (remote I/O must stay outside the assembly mutex) | `tasks.md:14` (L5 ownership), `analysis/pre-edit-targets.md:21-25`, `design.md:174,241`; task 3.6 (`tasks.md:75`) says "outside strategy refresh critical sections… scheduler capabilities" |
| The proposal loader converts snapshots to lane inputs in `buildLaneInput` (L3 target) | `internal/strategyproposal/production.go:269-382` per `analysis/pre-edit-targets.md:18` |

---

## 2. M-B0 static-guard constraints on L1

Source: `internal/official/a112_mbus_static_test.go`.

- Repository-wide walk (`TestA112MBUSStaticNoProductCallerOrForbiddenPath`, `:170-191`) parses every
  `.go` file under the repo root except `.git`/`vendor` (`:394-421`) and fails on any violation
  from `a112MBUSFindForbiddenRefs` (`:423-487`).
- **Allowed files** (`a112MBUSReferenceAllowed`, `:489-498`), verbatim logic:
  - any path with prefix `tools/a112-mb-us-source/`;
  - any path with prefix `internal/official/a112_mbus_` **and** suffix `_test.go`;
  - exactly `internal/official/a112_mbus_read.go`, `internal/official/a112_mbus_read_unix.go`,
    `internal/official/a112_mbus_read_unsupported.go`.
- **What counts as a violation** in every other file:
  - a selector `<alias>.<Name>` where `<alias>` imports
    `github.com/JungHoonGhae/tossinvest-cli/internal/official` and `<Name>` ∈
    `{A112MBUSResult, A112MBUSHoldError, A112MBUSCandle, A112MBUSOrderbook, A112MBUSCalendar}`
    (`:387-390`, `:462-471`);
  - a dot-import of `official` plus a bare reference to those names (`:480-482`);
  - **inside `internal/official/*.go` with `package official`**: any bare identifier equal to
    an exported surface name **or with prefix `a112MBUS`** (`:450`, `:476-478`) — so a new
    `internal/official/*.go` L1 file cannot call the private helpers either.
- `a112_mbus_read.go`/`_unix.go` are additionally string-scanned to contain none of
  `.token(`, `.refresh(`, `.exchange(`, `.saveCache(`, `AuthHeaders(`, `.get(`, `.send(`
  (`:180-190`) — this constrains the seam files, not L1 files.
- Ordinary `import "…/internal/official"` is **not** banned (`:451-453` returns nil when no
  alias/dot-import/same-package and design `design.md:148`).
- Consequence: `strategyevidence`, `strategycandle`, `breakoutlane`, engine, cmd and any new
  L1 adapter file **must not reference** the M-B0 exported symbols or `a112MBUS*` helpers.
  L1 must add its **own** production raw US reader (and calendar/quote readers) — the M-B0
  code is a reference algorithm, not a dependency (`tasks.md:28`, `design.md:148`,
  spec `breakout-retest-strategy-lane/spec.md:88,90`).

---

## 3. Existing functions L1 would edit vs. purely additive files

### 3.1 Existing bodies that MUST change

| Function | Why | FLM present? | `ast.json` freshness |
|---|---|---|---|
| `kindSupportsMarket` (`model.go:217-228`) | new kind(s) otherwise refused at `normalizeHeader:168` | yes — `analysis/function-logic/internal-strategyevidence--kindsupportsmarket/{ast.json,function-logic-map.md,branch-test-map.md,risk-pattern-report.md}` | `source_sha256 = c49652afe47c01d16963b359d59d9ed4745656509a9336a2157ebeaa4e7bb368` == `sha256sum internal/strategyevidence/model.go` (current) → **CURRENT** |
| `validateTypedPayload` (`model.go:398-428`) | only hook that runs on Append **and** replay/Valid; must dispatch new kind(s) to a strict decoder (unknown field, float/decimal number, secret, bounded arrays, enum) | yes — `analysis/function-logic/internal-strategyevidence--validatetypedpayload/…` | same SHA → **CURRENT** |

Recommended minimal shape (post-edit FLM/BTM refresh required per `tasks.md:123`): add the new
kind case(s) to the switch (`kindSupportsMarket`), and at the top of `validateTypedPayload`
after `rejectSecretFields`, `switch kind { case Kind…: return validate…(object) }` before the
generic shallow map so legacy kinds are byte-for-byte unaffected.

### 3.2 Existing bodies that need NOT change (cite, do not edit)

- `authoritySupportsMarket` — Toss OpenAPI already KR+US (`model.go:210-211`).
- `Store.migrate` / `schemaV1` — kind column is free text; `SchemaVersion` bump would break
  `OpenReadOnly` on existing DBs (`readonly.go:44`). Only if an additive index is proven
  necessary would `migrate` need a v2 step; today's `evidence_asof_idx` (`store.go:148`)
  covers `SealSnapshot`'s predicates.
- `normalizeHeader`, `NewEnvelope`, `Append`, `SealSnapshot`, `snapshotDigest`, `Replay`,
  `Project/choose` — reused as-is. If L1 needs a kind-scoped bounded ordered series read
  (see §5), add a **new** method rather than editing `SealSnapshot`/`SnapshotQuery` (struct
  field addition would change `snapshotDigest` inputs at `store.go:383`).
- `RawMinuteCandles`, `Client.get/send/doRequest`, `token.go`, `trace.go`, `ratebudget.go` —
  read-only (M-B0 row `tasks.md:8`, L1 row `tasks.md:10` does not list them).

### 3.3 Additive file plan (keeps L1 additive; names indicative)

| File | Content | Notes |
|---|---|---|
| `internal/strategyevidence/breakout_bar.go` | `Kind…` const(s) for closed 1m bar (and quote if separate), payload struct(s) with only integer/bool/string fields, strict decoder `validate…Payload(map[string]any) error` (unknown-field refusal, integer-only `json.Number` check with no `.`/`e`, bounded arrays, enum whitelist, digest presence) | new function bodies → new-function AST/FLM after GREEN |
| `internal/strategyevidence/breakout_snapshot.go` | closed-bar identity helpers (`(market, symbol, session_id, interval, open_at)` → `SourceRecordID`), revision/supersedes helpers, dual-cutoff assembly (query builder that pins `EvaluationAt` = session-close/`source_observed_at` and `IngestionCutoff` = `recorded_at`), bounded ordered read of bar rows for one session (additive method on `*Store` or a new read type) | must not touch `consumer.go` (static import guard `consumer_static_test.go:22`) |
| `internal/strategyevidence/breakout_bar_test.go`, `breakout_snapshot_test.go` | task 2.5 REDs (`tasks.md:62`): unknown field/enum, float/minor mismatch, secret-like, unbounded/duplicate/future/unfinished bars, append-only correction revision, dual-cutoff replay | fixtures hand-written, shape-only (see §4) |
| official raw KR/US bar + quote + calendar producer adapter files (location TBD — see §5 Q1): e.g. `internal/official/us_bars_raw.go` (+`kr_bars_raw.go`, `quote_raw.go`) **or** a new package such as `internal/officialbars/` | strict envelope/cursor decode (own implementation of the M-B0 reference algorithm), body cap, no-retry policy decision, same-response rate-header capture, `AuthorityOrigin` check, KST-offset timestamp → market-local via `marketclock.Market.Location()`, calendar join via `scheduler.AdaptOfficialCalendar` regular window (end-exclusive), decimal-string → integer minor conversion with explicit scale table, USD/KRW check | if placed in `internal/official`: cannot use `a112MBUS*` helpers (§2); if a new package: keeps `strategyevidence` free of the `official` closure |
| their `_test.go` with `httptest` | REDs mirroring receipt facts (§4) | no external network |

---

## 4. Contract facts L1 must encode from the M-B PASS receipt

Sources: `analysis/measurements/m-b-us-source/receipt-2026-08-16-run4.json` (PASS, sealed run B),
`receipt-2026-08-16-run3.json` (ACCEPT-HOLD, same body bytes), `receipt-schema-candle-crawl.md`,
`rate-budget.json`, `raw-us-bar-quote-payload.json`. Raw bodies live only in secure `/tmp`
(`design.md:154`, `tasks.md:30`); nothing below is a raw mirror.

| Aspect | Observed fact (receipt) | L1 encoding requirement |
|---|---|---|
| Request | `GET /api/v1/candles` with canonical query `adjusted=false&before=<cursor percent-encoded once>&count=200&interval=1m&symbol=AAPL` (run4 `observed_broker_behaviour.requests[0]`; run3 `request`); page 1 `before` = explicit `--before` literal (`2026-08-15T05:00:00.000+09:00`) | fixed `interval=1m`, `adjusted=false`, `count=200`; `before` optional; percent-encode decoded cursor bytes exactly once (`tasks.md:29`) |
| Page size | 200 bars/page × 4 pages = 800 unique strictly-descending timestamps (run4 `candles`) | decoder accepts ≤200 per page; order descending; enforce uniqueness |
| Cursor semantics | `nextBefore` raw JSON string; cursor == last bar − 1 min; page N+1 first bar == page N cursor (**inclusive**); 4 cursors + `--before` all distinct (run4 `cursor_continuity`); terminal `null` **not observed** — `cap_exhausted` recorded (run4 `terminal_null`, `receipt-schema-candle-crawl.md`) | strict cursor typing: JSON string or `null` only; absent/empty/number/object/array → fail closed (`design.md:150`); treat cursor as inclusive lower bound of the next page (dedupe the overlapping bar); loop detection |
| Bar field names/types | keys `timestamp/openPrice/highPrice/lowPrice/closePrice/volume/currency`; 3 200 price fields are JSON **strings** with 1–4 dp (20 integer-form), 800 volumes integer strings, 0 bare numbers; `currency` `USD` 800/800 (run4 `candles`; run3 `types`) | decode into string fields (as `apiCandle` `candle_reads.go:23-31` does), then exact decimal → integer minor with a **declared scale per currency** (USD 2 dp ⇒ minor = cents; KRW 0 dp); reject if the string has more fractional digits than the scale allows or contains sign/exponent; reject bare JSON numbers; reject currency ≠ expected |
| Timezone/timestamp | `YYYY-MM-DDTHH:MM:SS.000+09:00` (KST offset) for US bars (run3 `types`, run4 `page_ranges_kst`) | parse RFC3339 with mandatory offset (KR precedent: `offsetTimestamp` regexp `strategymarket/bars.go:103`, enforced at `:150`), convert with `marketclock.MarketUS.Location()`; never trust the offset as the market zone |
| Calendar structure | `result.today.date=2026-08-14`; `today/previousBusinessDay/nextBusinessDay` each with `dayMarket 09:00-17:00`, `preMarket 17:00-22:30`, `regularMarket 22:30-05:00(+1d)`, `afterMarket 05:00-08:50(+1d)` KST; regularMarket **end-exclusive** (run4 `regular_session_join`, `calendar_body`) | join via `official.TypedMarketCalendar` → `scheduler.AdaptOfficialCalendar` regular window `[Open, Close)`; `dayMarket` is not in the typed struct (ignored) |
| Regular-session coverage | 390/390 bars in `[22:30, 05:00)`; convention-agnostic because 391 consecutive bars `22:30..05:00` inclusive were all present; ≥1 pre-session bar; extended-hours minutes without trades are omitted, not zero-filled (run3 `extended_hours`) | L1 gap policy: inside regular session a missing minute is a defect (fail closed for that session's snapshot); outside session gaps are expected |
| Rate headers | cardinality one; keys `X-RateLimit-Limit/Remaining/Reset` (canonical `X-Ratelimit-*` on the wire); Retry-After absent; candles Limit 20 / Remaining 19,19,18,19 / Reset 1; orderbook 15/14/1; calendar 3/2/1; quota **shared with live TossOS containers**; Reset unit unstated (run4 `rate_headers`; run3 `rate_headers`) | capture headers of the same response; treat `Reset` as opaque; budget conservatively; do not derive timing from `Reset` |
| Orderbook | HTTP 200 `{result:{timestamp, currency USD, asks [], bids []}}` — empty book, market closed; **level element schema and decimal encoding UNOBSERVED** (run4 `orderbook`, `residuals_carried_to_l1_and_ledger[0]`) | quote adapter must **fail closed until level schema is observed** under a human-approved probe; do not infer from `apiOrderbookEntry{price,volume}` (`market_reads.go:189-192`) which is float-adapted and unproven for US |
| Envelope | body `{"result": …}`; strict object; duplicate keys refused by the seam (`a112_mbus_read.go:298-346`) | replicate strict-object decode |
| Rate/quota static contract | `rate-budget.json`: `MARKET_DATA_CHART` documented 5 req/s; client semantics per `internal/official/ratebudget.go` (`source_sha256 5c8466cb…`) | advisory only; the M-B observed Limit=20 for candles differs from the documented 5/s — treat both as opaque |
| Design residuals L1 must resolve by explicit decision | (a) bar-label convention → `opening_range_first_bar_id`; (b) variable decimal scale; (c) Reset unit opaque; (d) quote schema unobserved (run4 `residuals_carried_to_l1_and_ledger`; `tasks.md:38`) | (a) decide **bar-start (open_at)** unless a documented reason overrides: matches design bar identity `(market, symbol, session_id, interval, open_at)` (`design.md:142`), the KR precedent (`bars.go:158-159,189`) and the schema comment (`candle_reads.go:16`); with bar-start, the first regular bar of `2026-08-14 US` is `22:30 KST` and the 05:00 KST bar is after-hours; record the decision in the evidence payload (e.g. `bar_label:"open_at"`) so replay cannot silently flip; (b) explicit per-currency scale + refusal of over-precise strings; (c) opaque; (d) fail closed |
| Finality | `CandlePageResponse` carries no `closed/finalized` flag or server-observed timestamp (`design.md:168`); M-B PASS does not prove finality (`tasks.md:38`, run4 `decision`) | `Closed=true` may be asserted only when `open_at + 60s <= source_observed_at` **and** the bar is inside the calendar regular window; `source_observed_at` must come from an official-response-bound instant, not the local clock as authority (`design.md:166`) — L1 must define which header field carries it (see §5 Q4) |

---

## 5. Risks / open questions for the Manager

1. **Ownership of new files under `internal/official`.** L1 row verbatim: "`internal/strategyevidence/**`, additive read-only official KR/US raw bar/quote adapter files and their tests" (`tasks.md:10`). It does not name a directory. If the adapters go into `internal/official`, they are same-package with the M-B0 seam and the guard forbids any `a112MBUS*` identifier use (§2), and they sit beside order-write code; if they go into `internal/strategyevidence`, that package gains the `official` → `trading/config/orderintent` closure (task 8.2 dependency guard, `tasks.md:121`). A third option is a new package (e.g. `internal/officialbars`) importing `official` + `scheduler` + `strategyevidence` — please state which the table permits.
2. **`SnapshotQuery` has no kind filter and `Project/choose` selects one envelope per kind** (`store.go:271-278`, `projection.go:114-176`). One-row-per-bar storage would put hundreds of rows into every symbol snapshot and be unreadable via `choose`. Decision needed: (a) one envelope per bar + a new bounded ordered read method, or (b) one envelope per *assembled session series* (bounded ≤512 bars, revisioned as a whole). Option (b) fits `choose` and `snapshotDigest` unchanged but makes a single-bar correction a full-series revision. Both remain additive; both need the `validateTypedPayload` dispatch.
3. **`IssuerIdentity`/`IssuerMappingVersion` are mandatory header scope** (`model.go:158-159`, `SealSnapshot:316`). Existing US evidence uses SEC CIK, KR uses DART corp code (`official_source.go:227,311`). L1 must state what a bar's issuer identity is (symbol-only evidence has no natural issuer) and how the L3 loader will know it.
4. **Which header field is `source_observed_at`?** Candidates: `SourceEventAt` = bar `open_at`, `SourceAvailableAt` = bar close (`open_at+60s`) or response-observed instant, `ObservedAt` = client receipt time. `SealSnapshot` cuts on `SourceEventAt`/`SourceAvailableAt` (`store.go:333`), not `ObservedAt`. `design.md:166` forbids the process clock as finality authority, but the header ordering at `model.go:174` requires `ObservedAt >= SourceAvailableAt` and `Append` overrides `IngestedAt` with the store clock (`store.go:198-203`). Decision needed before the strict decoder freezes field names.
5. **`session_id` and `calendar_version` sources.** No production producer of `session_id` exists; `calendar_version` should be `scheduler.CalendarSnapshot.Version` (`calendar.go:46,98`) which the engine already treats as `CalendarDigest`. Golden vector uses `session_id "KRX:2026-08-18"`, `calendar_version "krx-calendar-v1"` (`goldens/breakout-evidence-and-sizing-v1.json known_vector`) — examples, not decisions. L1 must freeze the format (recommend `<MARKET-UPPER>:<exchange-local YYYY-MM-DD>` and the scheduler digest) and reconcile lowercase `marketclock.Market` vs uppercase golden `market`.
6. **KR path reuse vs duplication.** `RawMinuteCandles` (KR) has zero production callers and goes through the retrying `send` (`client.go:344-361`) with unbounded body read and non-strict cursor. Reusing it for KR bars gives KR a weaker contract than US. Recommend one strict raw reader for both markets (own implementation), keeping `RawMinuteCandles` untouched for regression.
7. **Retry/token policy of the production reader.** M-B was no-retry/cache-only; production `send` refreshes/exchanges (`token.go:60-80,108`) and the credential is shared (memory: token war). L1 must state whether the bar producer uses the ordinary token path (accepting refresh side effects) or a cache-only read; the latter cannot reuse `a112MBUSCachedToken` (§2) and would need its own FD-safe reader — a large surface for L1.
8. **Schema-migration risk.** No migration is required for a new kind (`store.go:129`), and bumping `SchemaVersion` would make `OpenReadOnly` refuse existing DBs (`readonly.go:44`). If a per-kind index becomes necessary for 390×N rows/day, it must be a v2 additive migration with a `user_version==1` upgrade path and a `readonly` acceptance of both versions — out of scope unless proven.
9. **Test data.** Repository fixtures must be hand-written, sanitized, **shape-only** (field names/types/offset format/empty orderbook/rate header names) — never copies of the `/tmp` receipt bodies (`design.md:154`, `tasks.md:30`; `raw-us-bar-quote-payload.json` classification "not a test fixture"). Existing `internal/strategyevidence/testdata/*.json` are SEC/DART shapes and are unrelated. Golden `analysis/goldens/breakout-evidence-and-sizing-v1.json` must be consumed hash-identical, not rewritten (`tasks.md:46`).
10. **US bar-label decision is L1's, and it fixes the setup ID forever.** `setupID` uses `Bars[0].ID` (`machine.go:201`); if L1 picks bar-start and a later measurement proves bar-end, every KR/US setup identity changes. Recommend recording the convention inside the evidence payload and the config digest so it is versioned, and requesting a small human-approved probe (one GET during a live US session comparing the latest bar's timestamp against wall time) before L1 acceptance rather than after.

---

## Verification commands (read-only, all run from `/mnt/D/project/axipient/TossOS`)

```
git status --short ; git log --oneline -3 ; git branch --show-current
codegraph_status ; codegraph_callers RawMinuteCandles|SealSnapshot|NewOfficialAdapter|NewEnvelope|canonicalJSON|rejectSecretFields|authoritySupportsMarket|OpenReadOnly|A112MBUSCandle|AdaptAndSealClosedKRXFiveMinute|get
codegraph_impact kindSupportsMarket (depth 2) ; codegraph_impact validateTypedPayload (depth 3) ; codegraph_node NewClosedBar
sed -n / cat -n on: internal/strategyevidence/{model,store,official_source,source,consumer,projection,readonly,readonly_owner_*}.go, internal/strategyevidence/consumer_static_test.go
cat -n internal/official/{candle_raw,candle_reads,client,a112_mbus_read,a112_mbus_read_unix,a112_mbus_read_unsupported,a112_mbus_static_test,trace,calendar_reads,typed_calendar_reads}.go ; sed -n on internal/official/{market_reads,ratebudget,token}.go
grep -n "func (c \*Client)" internal/official/market_reads.go ; grep -c "c\.get(" internal/official/*.go
cat -n internal/strategycandle/adapter.go ; sed -n 1,200p internal/strategymarket/bars.go ; internal/strategymarket/adapter_guard_test.go
cat -n internal/breakoutlane/{types,rules,arithmetic}.go ; sed -n 129,205p internal/breakoutlane/machine.go ; sed -n 90,110p internal/breakoutlane/guard_test.go
cat -n internal/clock/market.go ; grep -n "^func|^type" internal/clock/clock.go ; sed -n 1,175p internal/scheduler/calendar.go
grep -rn "RawMinuteCandles|AdaptOfficialMinutePage|SealAdaptedOfficialMinutePage" --include=*.go internal cmd | grep -v _test.go
grep -rn "strategyevidence\." --include=*.go internal cmd | grep -v "^internal/strategyevidence/" | grep -v _test.go
grep -rn "evidence.db|EvidencePath" --include=*.go cmd internal/app internal/strategyproposal | grep -v _test.go
go list -f '{{.ImportPath}}: {{join .Imports " "}}' ./internal/strategyevidence ./internal/official ./internal/scheduler ./internal/strategycandle ./internal/breakoutlane ./internal/strategymarket
go list -deps ./internal/official ./internal/scheduler | grep -c strategyevidence   # 0
go list -f '{{.ImportPath}} {{join .Imports " "}}' ./... | grep internal/strategyevidence
sha256sum internal/strategyevidence/model.go ; python3 (read source_sha256 from analysis/function-logic/internal-strategyevidence--{kindsupportsmarket,validatetypedpayload}/ast.json)
Read: openspec/changes/a112-…/{tasks.md, design.md 120-262, specs/breakout-retest-strategy-lane/spec.md 60-110, analysis/current-main-evidence.md 57-61, analysis/pre-edit-targets.md 1-40, analysis/measurements/m-b-us-source/{receipt-2026-08-16-run4.json, receipt-2026-08-16-run3.json, receipt-schema-candle-crawl.md, rate-budget.json, raw-us-bar-quote-payload.json}, analysis/goldens/breakout-evidence-and-sizing-v1.json, review.md 76-100 + grep "L1"}
```

No file other than this evidence pack was written; nothing staged, committed, executed from
`/tmp`, or fetched from the network; `~/.config/tossctl/*` and `.env` files were not read.
