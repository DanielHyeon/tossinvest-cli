# Function Logic Map: `strategyProposalAuthorityLoader.collectMarket`

- Source: `internal/app/engine/strategy_proposal_authority.go` (165-245)
- Function: `strategyProposalAuthorityLoader.collectMarket` in package `engine`
- Signature: `strategyProposalAuthorityLoader.collectMarket(params=5, results=1)`
- File SHA-256: `61724700e3ecd848b656a37d5f8632a17aed7e02639179d1a5f62a1f508bd27e`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 13.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

한 시장의 경로 항목들을 봉인된 제안 항목으로 바꾼다. 태스크 5.4 이후, 한 종목이 여러 가족을 제안하면
`batch.LanesFor(symbol)` 로 그 종목의 모든 가족 제안을 받아 `arbitrateProposalScope` 에 넘기고 보정 중재가 하나를 고른다.
중재가 닫히면 그 종목만 빼지 않고 **시장 전체를 닫는다** — 하나를 빼면 목록이 둘에서 하나로 줄어
아래 세 소비자가 공유하는 `len(entries) != 1` 관문이 오히려 만족되고, 막으려던 것과 상관없는 다른 종목이 대신 풀린다(리뷰 지적 C2).
제안을 하나도 내지 않은 종목은 중재 대상이 아니라 거절 수에만 든다 — 이는 5.4 이전 `batch.For` 가 `!ok` 를 세던 것과 같은 동작이다.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode.
- engine tagged suite: `go test -count=1 -tags tossos_testseams -covermode=count -coverpkg=./internal/app/engine,./internal/strategyarbiter,./internal/strategyproposal,./internal/strategyrouter ./internal/app/engine/`
- engine untagged suite: `go test -count=1 -covermode=count -coverpkg=./internal/app/engine,./internal/strategyarbiter,./internal/strategyproposal,./internal/strategyrouter ./internal/app/engine/`
- proposal tagged suite: `go test -count=1 -tags tossos_testseams -covermode=count -coverpkg=./internal/app/engine,./internal/strategyarbiter,./internal/strategyproposal,./internal/strategyrouter ./internal/strategyproposal/`
- Per-test attribution set: the seven `Test*` functions that can reach `strategyProposalAuthorityLoader.collectMarket` — the six in `a112_arbitration_test.go` and `strategy_proposal_authority_test.go` plus none elsewhere, because no other engine test constructs a proposal loader or a production assembly. This is the complete reaching set, not a sample.
- Measured entry: executed 6x in the engine tagged suite; never in the engine untagged suite or in the proposal package suite, because the fixtures that build a proposal loader live behind the seam tag.

Exact AST return positions: 168:3, 171:3, 174:3, 177:3, 182:3, 194:4, 206:3, 228:4, 232:44, 237:3, 243:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 170:2 | arm entered 1x (engine tagged suite); arm not entered (engine untagged suite); no cover block (proposal tagged suite); entered by `TestAProposalMeasuredAgainstAnotherSymbolsRouteAuthorityIsRefused` |
| B2 | if | 173:2 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no cover block (proposal tagged suite); no per-test profile in the attribution set entered it |
| B3 | if | 176:2 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no cover block (proposal tagged suite); no per-test profile in the attribution set entered it |
| B4 | if | 181:2 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no cover block (proposal tagged suite); no per-test profile in the attribution set entered it |
| B5 | if | 185:2 | arm entered 6x (engine tagged suite); arm not entered (engine untagged suite); no cover block (proposal tagged suite); entered by `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket`, `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestAnUncalibratedMarketRefusesEvenASingleProposal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently`, `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestSymbolsWithNoProposalAtAllAreCountedRefusedRatherThanArbitrated` |
| B6 | range | 191:2 | arm entered 17x (engine tagged suite); arm not entered (engine untagged suite); no cover block (proposal tagged suite); entered by `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket`, `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestAnUncalibratedMarketRefusesEvenASingleProposal`, `TestAProposalMeasuredAgainstAnotherSymbolsRouteAuthorityIsRefused`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently`, `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestSymbolsWithNoProposalAtAllAreCountedRefusedRatherThanArbitrated` |
| B7 | if | 193:3 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no cover block (proposal tagged suite); no per-test profile in the attribution set entered it |
| B8 | if | 205:2 | arm entered 1x (engine tagged suite); arm not entered (engine untagged suite); no cover block (proposal tagged suite); entered by `TestStrategyProposalAuthorityKeepsMarketFailureLocal` |
| B9 | range | 215:2 | arm entered 14x (engine tagged suite); arm not entered (engine untagged suite); no cover block (proposal tagged suite); entered by `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket`, `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestAnUncalibratedMarketRefusesEvenASingleProposal`, `TestAProposalMeasuredAgainstAnotherSymbolsRouteAuthorityIsRefused`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently`, `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestSymbolsWithNoProposalAtAllAreCountedRefusedRatherThanArbitrated` |
| B10 | if | 217:3 | arm entered 3x (engine tagged suite); arm not entered (engine untagged suite); no cover block (proposal tagged suite); entered by `TestSymbolsWithNoProposalAtAllAreCountedRefusedRatherThanArbitrated` |
| B11 | if | 222:3 | arm entered 3x (engine tagged suite); arm not entered (engine untagged suite); no cover block (proposal tagged suite); entered by `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestAnUncalibratedMarketRefusesEvenASingleProposal`, `TestAProposalMeasuredAgainstAnotherSymbolsRouteAuthorityIsRefused` |
| B12 | if | 233:2 | arm entered 2x (engine tagged suite); arm not entered (engine untagged suite); no cover block (proposal tagged suite); entered by `TestSymbolsWithNoProposalAtAllAreCountedRefusedRatherThanArbitrated` |
| B13 | range | 240:2 | arm entered 8x (engine tagged suite); arm not entered (engine untagged suite); no cover block (proposal tagged suite); entered by `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket`, `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestAnUncalibratedMarketRefusesEvenASingleProposal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently`, `TestStrategyProposalAuthorityKeepsMarketFailureLocal` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `len` | 168:144 |
| `len` | 170:31 |
| `fail` | 171:10 |
| `fail` | 174:10 |
| `fail` | 177:10 |
| `strings.TrimSpace` | 179:13 |
| `loader.getenv` | 179:31 |
| `DecodeString` | 180:14 |
| `base64.StdEncoding.Strict` | 180:14 |
| `base64.StdEncoding.EncodeToString` | 181:19 |
| `len` | 181:72 |
| `fail` | 182:10 |
| `strings.TrimSpace` | 188:12 |
| `loader.getenv` | 188:30 |
| `make` | 189:13 |
| `len` | 189:58 |
| `make` | 190:14 |
| `len` | 190:59 |
| `entry.approved.Symbol` | 192:13 |
| `bySymbol.approved.Valid` | 193:22 |
| `fail` | 194:11 |
| `append` | 197:13 |
| `entry.route.Request` | 197:97 |
| `loader.load` | 199:16 |
| `strategyrouter.Market` | 200:42 |
| `strings.TrimSpace` | 200:111 |
| `loader.getenv` | 200:129 |
| `ed25519.PublicKey` | 201:15 |
| `strings.TrimSpace` | 204:23 |
| `loader.getenv` | 204:41 |
| `batch.ManifestDigest` | 205:19 |
| `fail` | 206:10 |
| `make` | 213:13 |
| `batch.Len` | 213:55 |
| `batch.LanesFor` | 216:12 |
| `route.approved.Symbol` | 216:27 |
| `len` | 217:6 |
| `arbitrateProposalScope` | 221:14 |
| `fail` | 223:14 |
| `string` | 225:41 |
| `append` | 230:13 |
| `sort.Slice` | 232:2 |
| `entries.route.approved.Symbol` | 232:51 |
| `entries.route.approved.Symbol` | 232:88 |
| `len` | 233:5 |
| `fail` | 234:13 |
| `sha256.New` | 239:7 |
| `h.Write` | 241:10 |
| `(unnamed)` | 241:18 |
| `entry.route.approved.Symbol` | 241:25 |
| `entry.authority.Proposal` | 241:66 |
| `len` | 244:16 |
| `len` | 244:52 |
| `hex.EncodeToString` | 244:144 |
| `h.Sum` | 244:163 |

## State mutations and fallbacks

- AST assignments: 29. Defers: 0. Goroutine statements: 0.

## Safety conclusion

중재는 이전보다 *덜* 거절하지 않는다. 이전에는 모호한 종목이 있으면 무조건 시장을 닫았고, 이제는 승인된
보정 점수로 유일한 최고점을 고를 수 있을 때만 통과시킨다. 고르지 못하면 여전히 시장을 닫는다.
새로 통과할 수 있게 된 유일한 경우는 '봉인된 매니페스트가 승인한 점수로 유일 승자가 정해지는' 경우다.
`ResultAuthority`, `strategyAccountAuthorityLoader.collectMarket`, `strategyProjectionFromAssembly` 세 소비자가
이 함수가 만든 목록을 읽으므로, 정정이 있어야 할 자리는 여기 하나다.
