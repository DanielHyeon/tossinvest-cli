# Function Logic Map: `Console.accountPanelFrom`

- Source: `internal/console/overview.go`
- Post-edit AST evidence: `ast.json` (17 branches; revision `current`; source SHA-256 recorded by extractor)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `now` | dashboard cache lookup/policy-cache time | overview request | never serves as the final marker/freshness authority inside the shared decorator |
| broker snapshot | cold, failed, stale, or known cache value | `holdings.peek` | dashboard performs no broker refresh and never reports unreadable as empty |
| journal view/live exits | readable or explicitly failed | journal read | unreadable/absent state stays unmeasured or explicit |
| rows | broker/journal join | `joinPositions` then `decoratePositionRows` | all actionable exit values use the decorator's later post-marker response time |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | range over decorated rows at `internal/console/overview.go:661` | inspects request-local rows | no external side effect | `TestA111HoldingsRoutesRecheckFreshnessAfterPolicyCacheMiss` |
| B2 | readable broker row absent from journal at `internal/console/overview.go:662` | sets `AnyJournalAbsent` and breaks | absence is explicit, not unmanaged | `TestAHoldingIsNotCalledUnmanagedWhenTheJournalCouldNotBeRead` |
| B3 | range over configured overview markets at `internal/console/overview.go:669` | accumulates market panels | markets remain currency-separated | `TestTheOverviewNeverAddsAcrossMarkets` |
| B4 | broker snapshot unusable at `internal/console/overview.go:671` | appends blocked/unreadable market row | continue without fabricated totals | `TestAFailedBrokerReadIsNotTheSameAsAnEmptyCache` |
| B5 | range over joined rows at `internal/console/overview.go:682` | accumulates per-market totals/counts | request-local only | `TestTheOverviewNeverAddsAcrossMarkets` |
| B6 | row not broker-backed or belongs to another market at `internal/console/overview.go:683` | skips row | prevents cross-market addition | `TestTheOverviewNeverAddsAcrossMarkets` |
| B7 | management classification switch at `internal/console/overview.go:689` | chooses one count bucket | unknown is never coerced to managed/unmanaged | `TestAHoldingIsNotCalledUnmanagedWhenTheJournalCouldNotBeRead` |
| B8 | unknown row case at `internal/console/overview.go:690` | increments unknown | forces counts unmeasured below | `TestAHoldingIsNotCalledUnmanagedWhenTheJournalCouldNotBeRead` |
| B9 | managed row case at `internal/console/overview.go:692` | increments managed | no change to exit evidence | `TestAnAdoptedHoldingRendersAsManagedWithItsBasis` |
| B10 | remaining row default at `internal/console/overview.go:694` | increments unmanaged/other | no implicit adoption | `TestAnUnmanagedHoldingIsLabelledExactlyOnce` |
| B11 | market has no symbols at `internal/console/overview.go:699` | appends explicit empty row | does not report zero-valued invented holdings | `TestTheOverviewNeverAddsAcrossMarkets` |
| B12 | one or more classifications unknown at `internal/console/overview.go:708` | marks managed/unmanaged counts unmeasured | fail-closed count truth | `TestAHoldingIsNotCalledUnmanagedWhenTheJournalCouldNotBeRead` |
| B13 | all classifications known at `internal/console/overview.go:711` | emits measured counts | measured only from known rows | `TestAnAdoptedHoldingRendersAsManagedWithItsBasis` |
| B14 | broker snapshot usable at `internal/console/overview.go:720` | scans for out-of-market holdings | unusable snapshot skips derivative claims | `TestAHoldingInNeitherMarketIsNamedRatherThanDropped` |
| B15 | range over rows for other-market detection at `internal/console/overview.go:722` | collects market:symbol labels | request-local only | `TestAHoldingInNeitherMarketIsNamedRatherThanDropped` |
| B16 | row not broker-backed or already KR/US at `internal/console/overview.go:723` | skips pair | avoids duplicate/false other-market rows | `TestTwoMarketsAloneProduceNoOtherRow` |
| B17 | at least one other-market pair at `internal/console/overview.go:728` | appends blocked explanatory row | unknown currency is never summed | `TestAHoldingInNeitherMarketIsNamedRatherThanDropped` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `holdings.peek` | consume cache without broker I/O | cold/failed/stale truth preserved; never refreshes | `TestTheOverviewReadsTheBrokerCacheWithPeekAndNothingElse` |
| `joinPositions` | combine broker and journal identities | disagreement remains explicit | existing overview/portfolio tests |
| `decoratePositionRows` | share policy, marker and exit projection with `/positions` | blocking policy work is followed by pre/read/post marker authority | both named A111 holdings-route REDs |
| market aggregation helpers | build read-only totals and explanations | no cross-currency total or mutation | overview aggregation tests |

## State mutations and fallbacks

- Mutates only an `accountPanel` and request-local joined rows.
- Dashboard uses the same decorator as `/positions`; the caller's `now` remains cache time, while exit lines use the decorator's post-marker `responseAt`.
- Broker/journal/management unknown states remain blocked or unmeasured, never empty or unmanaged by inference.

## Safety conclusion

- Safe edit boundary: dashboard read projection and aggregation only.
- High-risk impact: yes for displayed exit lines; shared-route REDs prove post-policy freshness and stopped-marker non-resurrection.
