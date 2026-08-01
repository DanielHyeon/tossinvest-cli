# Branch Test Map: `ParkerConservativeLane.Evaluate`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | approved snapshot invalid | zero-approval row | indirect only | pass |
| B2 | valid approved scope is not exact KR/symbol | `TestParkerRejectsApprovedUnsupportedMarketBeforeSource` | mapped incorrectly | pass |
| B3 | frozen source proof invalid | zero/forged source rows | yes | yes |
| B4 | market input bundle invalid or mismatched | zero market row | yes | yes |
| B5 | candidate not active or outside approved life | current/future/exact-expiry table | yes | yes |
| B6 | non-trading day | translated StockOS session table | missing source reason | pass |
| B7 | open-auction half-open window | translated session + exact boundary tables | collapsed into after-hours | pass |
| B8 | close-auction half-open window, including early close from authoritative close | translated session + exact boundary tables | collapsed into cutoff | pass |
| B9 | fixed 15:40 KST after-hours start | translated session + exact boundary tables | broad outside-session check | pass |
| B10 | before open+10m opening skip | opening exact boundary and translated session rows | yes | yes |
| B11 | after close-minus-45m entry cutoff | regular/StockOS early-close exact/+1ns and post-close gap rows | wrong 15:20 fixture | pass |
| B12 | official unadjusted closed-bar proof invalid | zero-bar row | yes | yes |
| B13 | bar closes after evaluation/session close | direct future-close row | indirect only | pass |
| B14 | official fresh normal-state proof invalid | zero-state row | yes | yes |
| B15 | official no-position proof invalid | zero-position row | yes | yes |
| B16 | OHLCV invalid after opaque bar proof | upstream `TestAggregateClosedKRXFiveMinuteFailsClosed`; impossible to forge cross-package proof | structural invariant | pass |
| B17 | aggregate volume is exactly zero | direct zero-volume row | missing | pass |
| B18 | required indicator decimal malformed despite same-package opaque-state forgery | forged indicator table | missing | pass |
| B19 | close is not above VWAP | frozen boundary table | yes | yes |
| B20 | VWAP slope below 0.08 | frozen boundary table | yes | yes |
| B21 | close/low fails EMA9 pullback; touch ceiling exact/+epsilon | EMA gate plus 0.25% exact boundary tests | partial | pass |
| B22 | bar is not bullish after earlier gates | direct fake-breakout row | missing | pass |
| B23 | LVN space below 1.2 | frozen boundary table | yes | yes |
| B24 | tangled score below 0.35 | frozen boundary table | yes | yes |
| B25 | optional band expansion present/absent | present baseline plus optional-absent row | indirect only | pass |
| B26 | present expansion malformed despite same-package opaque-state forgery | forged indicator table | missing | pass |
| B27 | expansion exceeds 1.8 | frozen boundary table | yes | yes |
| B28 | RR `<1.5` after LVN gate | structural minimum `12/7` test | falsely marked refusal-covered | pass as structural invariant |
| B29 | optional HVN present/absent | present baseline plus optional-absent row | indirect only | pass |
| B30 | present HVN malformed despite same-package opaque-state forgery | forged indicator table | missing | pass |
| B31 | HVN distance below LVN forward space | frozen boundary table | yes | yes |
| B32 | signal age outside `[0,15s]` | nanosecond age table | yes | yes |
| B33 | optional live price present/absent | present baseline plus missing-price fallback row | yes | yes |
| B34 | present live price malformed/nonpositive | forged malformed plus nonpositive rows | partial | pass |
| B35 | negative live-entry delta | negative exact/+1 boundary rows | missing | pass |
| B36 | derived absolute drift exceeds 0.20% | positive/negative +1 boundary rows | yes | yes |
| B37 | fixed-schema JSON identity marshal error | `DecisionRecord` contains only JSON-safe scalars; error branch is structurally unreachable | overclaimed direct coverage | pass as structural invariant |
| B38 | final mint refuses internally inconsistent record | direct exhaustive mint boundary table; synthetic/translated valid mints | indirect only | pass |
| Order | source refusal precedence through EMA → fake breakout → LVN and all later gates | simultaneous-failure tests | partial | pass |
| Success | synthetic derivation plus translated StockOS final-bar/indicator arithmetic | synthetic and translated parity tests | synthetic-only overclaim | pass |
