# Function Logic Map: `strategyProposalAuthorityLoader.collectMarket`

- Source: `internal/app/engine/strategy_proposal_authority.go` (195-297)
- Function: `strategyProposalAuthorityLoader.collectMarket` in package `engine`
- Signature: `strategyProposalAuthorityLoader.collectMarket(params=5, results=1)`
- File SHA-256: `f4f6627f945825c7a199bd38f9152122decd51da410acd8660ea8463093413e0`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 14.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

한 시장의 경로 항목들을 봉인된 제안 항목으로 바꾼다. 태스크 5.4.2 이후 한 종목의 여러 가족 제안을
그 자리에서 훑지 않는다 — `coordinateMarketProposals` 가 시장 조정자에 넣고, 조정자가 소유자 범위마다 하나를 고른다.
목록의 순서도 조정자가 정한다(소유자 범위 사전순이라 종목 오름차순). 그 순서가 `ProposalSetDigest` 에 그대로 들어간다.

조정자가 한 범위를 닫으면 그 종목만 빼지 않고 **시장 전체를 닫는다** — 하나를 빼면 목록이 둘에서 하나로 줄어
아래 세 소비자가 공유하는 `len(entries) != 1` 관문이 오히려 만족되고, 막으려던 것과 상관없는 다른 종목이 대신 풀린다.

**이 불변식은 중재 결함에만 성립한다. 제안이 없는 종목에는 성립하지 않는다.** B6 안의 B7 위쪽
(`batch.LanesFor` 가 빈 목록) 은 그 종목을 `refused++` 로 세고 건너뛴다. 그런데 그 빈 목록은
"이 종목은 원래 제안이 없다"와 "이 종목의 제안이 고장으로 사라졌다"를 구별하지 못한다 —
`LoadProductionAuthorityBatch` 안에 조용히 빈 목록을 만드는 `continue` 가 일곱 개 있다.
그래서 고장 하나가 목록을 줄이고, 줄어든 목록이 `len(entries) != 1` 관문을 오히려 만족시켜
상관없는 다른 종목을 풀어 준다. 5.4.2 적대적 리뷰가 시연으로 확인했고(P1-1), 이 `continue` 는
5.4.2 이전부터 같은 자리에 있었다. 고치는 일은 태스크 5.4.3 이다. 여기서는 **선언과 달성이
다르다는 사실을 적어 둔다.**

큐가 넘치면 중재 거절 코드를 빌려 쓰지 않고 `PROPOSAL_QUEUE_OVERFLOW` 라는 엔진 자신의 이름으로 보고한다.
동결 골든 `refusal_enums` 에 큐 코드가 없으므로, 빌려 쓰면 큐가 넘친 일이 봉인이 깨진 일로 보고된다.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode.
- 모든 실행은 `systemd-run --user --scope -p MemoryMax=… -p MemorySwapMax=0` cgroup 안에서 돌렸다.
  묶지 않고 돌린 이 패키지가 커널 OOM 으로 데스크톱을 세 번 죽였기 때문이다(`engine.test`, anon-rss 약 36GB).
  원인은 이 lot 이 고친 조정자 용량이며, 측정 방법이 아니라 측정 대상이 문제였다.
- engine tagged suite: `go test -c -tags tossos_testseams -covermode=count -coverpkg=./internal/app/engine ./internal/app/engine/`
  뒤에 그 바이너리를 `-test.coverprofile` 로 실행. 스위트 전체 PASS, 73.8% of statements.
- engine untagged suite: 같은 명령에서 `-tags tossos_testseams` 만 뺀 것. 스위트 전체 PASS, 63.6% of statements.
- Per-test attribution set: 같은 태그 바이너리를 `-test.run '^<Test>$'` 로 하나씩 돌린 열 개의 프로파일.
- **귀속 완전성은 주장이 아니라 측정이다.** 아래 모든 분기에서 테스트별 진입 수의 합이 스위트 전체 진입 수와
  정확히 같다. 이 집합 밖의 테스트가 어느 arm 이든 들어갔다면 그 등식이 깨진다. 깨진 행은
  `ATTRIBUTION MISMATCH` 로 표시되며 아래에는 하나도 없다.

Exact AST return positions: 198:3, 201:3, 204:3, 207:3, 212:3, 224:4, 236:3, 254:3, 263:3, 272:3, 281:3, 288:3, 294:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 200:2 | arm entered 2x (engine tagged suite); arm not entered (engine untagged suite); `TestAMarketWithMoreScopesThanTheQueueHoldsClosesInsteadOfDroppingOne`, `TestAProposalMeasuredAgainstAnotherSymbolsRouteAuthorityIsRefused` |
| B2 | if | 203:2 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no per-test profile in the attribution set entered it |
| B3 | if | 206:2 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no per-test profile in the attribution set entered it |
| B4 | if | 211:2 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no per-test profile in the attribution set entered it |
| B5 | if | 215:2 | arm entered 7x (engine tagged suite); arm not entered (engine untagged suite); `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestAnUncalibratedMarketRefusesEvenASingleProposal`, `TestEntriesComeBackInOwnerScopeOrderNotRouteOrder`, `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently`, `TestSymbolsWithNoProposalAtAllAreCountedRefusedRatherThanArbitrated`, `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket` |
| B6 | range | 221:2 | arm entered 10021x (engine tagged suite); arm not entered (engine untagged suite); `TestAMarketWithMoreScopesThanTheQueueHoldsClosesInsteadOfDroppingOne`, `TestAProposalMeasuredAgainstAnotherSymbolsRouteAuthorityIsRefused`, `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestAnUncalibratedMarketRefusesEvenASingleProposal`, `TestEntriesComeBackInOwnerScopeOrderNotRouteOrder`, `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently`, `TestSymbolsWithNoProposalAtAllAreCountedRefusedRatherThanArbitrated`, `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket` |
| B7 | if | 223:3 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no per-test profile in the attribution set entered it |
| B8 | if | 235:2 | arm entered 1x (engine tagged suite); arm not entered (engine untagged suite); `TestStrategyProposalAuthorityKeepsMarketFailureLocal` |
| B9 | if | 248:2 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no per-test profile in the attribution set entered it |
| B10 | if | 257:2 | arm entered 1x (engine tagged suite); arm not entered (engine untagged suite); `TestAMarketWithMoreScopesThanTheQueueHoldsClosesInsteadOfDroppingOne` |
| B11 | if | 265:2 | arm entered 3x (engine tagged suite); arm not entered (engine untagged suite); `TestAProposalMeasuredAgainstAnotherSymbolsRouteAuthorityIsRefused`, `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestAnUncalibratedMarketRefusesEvenASingleProposal` |
| B12 | if | 275:2 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no per-test profile in the attribution set entered it |
| B13 | if | 283:2 | arm entered 2x (engine tagged suite); arm not entered (engine untagged suite); `TestSymbolsWithNoProposalAtAllAreCountedRefusedRatherThanArbitrated` |
| B14 | range | 291:2 | arm entered 11x (engine tagged suite); arm not entered (engine untagged suite); `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestAnUncalibratedMarketRefusesEvenASingleProposal`, `TestEntriesComeBackInOwnerScopeOrderNotRouteOrder`, `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently`, `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `len` | 198:144 |
| `len` | 200:31 |
| `fail` | 201:10 |
| `fail` | 204:10 |
| `fail` | 207:10 |
| `strings.TrimSpace` | 209:13 |
| `loader.getenv` | 209:31 |
| `DecodeString` | 210:14 |
| `base64.StdEncoding.Strict` | 210:14 |
| `base64.StdEncoding.EncodeToString` | 211:19 |
| `len` | 211:72 |
| `fail` | 212:10 |
| `strings.TrimSpace` | 218:12 |
| `loader.getenv` | 218:30 |
| `make` | 219:13 |
| `len` | 219:58 |
| `make` | 220:14 |
| `len` | 220:59 |
| `entry.approved.Symbol` | 222:13 |
| `bySymbol.approved.Valid` | 223:22 |
| `fail` | 224:11 |
| `append` | 227:13 |
| `entry.route.Request` | 227:97 |
| `loader.load` | 229:16 |
| `strategyrouter.Market` | 230:42 |
| `strings.TrimSpace` | 230:111 |
| `loader.getenv` | 230:129 |
| `ed25519.PublicKey` | 231:15 |
| `strings.TrimSpace` | 234:23 |
| `loader.getenv` | 234:41 |
| `batch.ManifestDigest` | 235:19 |
| `fail` | 236:10 |
| `coordinateMarketProposals` | 247:26 |
| `fail` | 249:13 |
| `fail` | 258:13 |
| `fail` | 266:13 |
| `string` | 268:40 |
| `arbitration.entries` | 274:23 |
| `fail` | 276:13 |
| `len` | 283:5 |
| `fail` | 284:13 |
| `sha256.New` | 290:7 |
| `h.Write` | 292:10 |
| `(unnamed)` | 292:18 |
| `entry.route.approved.Symbol` | 292:25 |
| `entry.authority.Proposal` | 292:66 |
| `len` | 295:16 |
| `len` | 295:52 |
| `hex.EncodeToString` | 296:34 |
| `h.Sum` | 296:53 |

## State mutations and fallbacks

- AST assignments: 43. Defers: 0. Goroutine statements: 0.

## Safety conclusion

이 lot 은 거절을 느슨하게 하지 않았다. 닫히는 경우가 셋 늘었다 — 계보 신원 충돌(B9),
큐 넘침(B10), 되돌릴 자리 없음(B12). 셋 다 이전에는 존재하지 않던 닫힘이며 전부 보수 방향이다.
B9 는 아직 진입 0 회다. 이것은 통과가 아니라 커버리지 구멍이며 그대로 남겨 보고한다.

닫힘 가지 넷은 이제 `RefusedCount` 에 `RoutedCount` 를 그대로 싣는다. 시장이 닫히면 경로에 오른
종목 전부가 못 나간 것이기 때문이다. 앞선 판본의 `refused + 1` 은 10,001 종목이 하나도 못 나간
주기를 "1건 거절"로 보고했다.

`QueueDropCount` 는 지금 배선에서 **언제나 0** 이다. 계수기가 오르는 두 곳(접힘·넘침)이 둘 다
구조적으로 도달 불가이기 때문이다 — 종목 중복이 위(B7)에서 먼저 거절되고 매니페스트가
(종목, 레인)마다 하나만 실어 같은 칸이 두 번 오지 않으며, 큐 상한이 매니페스트 상한과 같아
정상 매니페스트는 넘칠 수 없다. 그러므로 이 수를 "조용한 유실이 없음의 증거"로 인용하지 않는다.
