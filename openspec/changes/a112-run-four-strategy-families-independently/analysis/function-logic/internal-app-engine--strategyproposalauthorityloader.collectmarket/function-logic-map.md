# Function Logic Map: `strategyProposalAuthorityLoader.collectMarket`

- Source: `internal/app/engine/strategy_proposal_authority.go` (209-320)
- Function: `strategyProposalAuthorityLoader.collectMarket` in package `engine`
- Signature: `strategyProposalAuthorityLoader.collectMarket(params=5, results=1)`
- File SHA-256: `b5e266ae728fcfc400e041bcdc82d2af4b68ec73542c2d6a525fa342523bf06f`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 15.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

한 시장의 경로 항목들을 봉인된 제안 항목으로 바꾼다. 태스크 5.4.2 이후 한 종목의 여러 가족 제안을
그 자리에서 훑지 않는다 — `coordinateMarketProposals` 가 시장 조정자에 넣고, 조정자가 소유자 범위마다 하나를 고른다.
목록의 순서도 조정자가 정한다(소유자 범위 사전순이라 종목 오름차순). 그 순서가 `ProposalSetDigest` 에 그대로 들어간다.

조정자가 한 범위를 닫으면 그 종목만 빼지 않고 **시장 전체를 닫는다** — 하나를 빼면 목록이 둘에서 하나로 줄어
아래 세 소비자가 공유하는 `len(entries) != 1` 관문이 오히려 만족되고, 막으려던 것과 상관없는 다른 종목이 대신 풀린다.

**이 불변식은 5.4.2 까지 중재 결함에만 성립했다. 태스크 5.4.3 이 나머지 절반을 채웠다.**
`batch.LanesFor` 가 빈 목록을 돌려주는 일은 "이 종목은 원래 제안이 없다"와 "이 종목의 제안이
고장으로 사라졌다"를 구별하지 못했고, 그래서 고장 하나가 목록을 줄이고 줄어든 목록이
`len(entries) != 1` 관문을 오히려 만족시켜 상관없는 다른 종목을 풀어 줬다(5.4.2 리뷰 P1-1,
시연으로 확인). 지금은 `strategyproposal` 이 그 둘을 갈라 형(型)으로 들고 나오고, 이 함수가
배치를 받자마자 `batch.Fault()` 를 먼저 본다 — 고장이면 조정에 넣지 않고 시장을 닫는다.

고장의 기준은 "제안이 안 나왔다"가 아니라 **"약속받은 봉인된 입력을 얻지 못했다"** 다.
시장 상태 때문에 평가가 거절한 것(스프레드·호가·사이징·보호적이지 않은 목표가)은 고장이
아니라 매일 일어나는 정상 거절이므로 예전처럼 거절로 세고 시장은 열어 둔다. 그 경계는
처음에 잘못 그었고 `TestAnEvaluationRefusalIsAbsenceNotFault` 가 잡았다.

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

Exact AST return positions: 212:3, 215:3, 218:3, 221:3, 226:3, 238:4, 250:3, 259:3, 277:3, 286:3, 295:3, 304:3, 311:3, 317:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 214:2 | arm entered 3x (engine tagged suite); arm not entered (engine untagged suite); `TestALostProposalClosesTheMarketInsteadOfReleasingTheOtherSymbol`, `TestAMarketWithMoreScopesThanTheQueueHoldsClosesInsteadOfDroppingOne`, `TestAProposalMeasuredAgainstAnotherSymbolsRouteAuthorityIsRefused` |
| B2 | if | 217:2 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no per-test profile in the attribution set entered it |
| B3 | if | 220:2 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no per-test profile in the attribution set entered it |
| B4 | if | 225:2 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no per-test profile in the attribution set entered it |
| B5 | if | 229:2 | arm entered 7x (engine tagged suite); arm not entered (engine untagged suite); `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestAnUncalibratedMarketRefusesEvenASingleProposal`, `TestEntriesComeBackInOwnerScopeOrderNotRouteOrder`, `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently`, `TestSymbolsWithNoProposalAtAllAreCountedRefusedRatherThanArbitrated`, `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket` |
| B6 | range | 235:2 | arm entered 10023x (engine tagged suite); arm not entered (engine untagged suite); `TestALostProposalClosesTheMarketInsteadOfReleasingTheOtherSymbol`, `TestAMarketWithMoreScopesThanTheQueueHoldsClosesInsteadOfDroppingOne`, `TestAProposalMeasuredAgainstAnotherSymbolsRouteAuthorityIsRefused`, `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestAnUncalibratedMarketRefusesEvenASingleProposal`, `TestEntriesComeBackInOwnerScopeOrderNotRouteOrder`, `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently`, `TestSymbolsWithNoProposalAtAllAreCountedRefusedRatherThanArbitrated`, `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket` |
| B7 | if | 237:3 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no per-test profile in the attribution set entered it |
| B8 | if | 249:2 | arm entered 1x (engine tagged suite); arm not entered (engine untagged suite); `TestStrategyProposalAuthorityKeepsMarketFailureLocal` |
| B9 | if | 254:2 | arm entered 1x (engine tagged suite); arm not entered (engine untagged suite); `TestALostProposalClosesTheMarketInsteadOfReleasingTheOtherSymbol` |
| B10 | if | 271:2 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no per-test profile in the attribution set entered it |
| B11 | if | 280:2 | arm entered 1x (engine tagged suite); arm not entered (engine untagged suite); `TestAMarketWithMoreScopesThanTheQueueHoldsClosesInsteadOfDroppingOne` |
| B12 | if | 288:2 | arm entered 3x (engine tagged suite); arm not entered (engine untagged suite); `TestAProposalMeasuredAgainstAnotherSymbolsRouteAuthorityIsRefused`, `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestAnUncalibratedMarketRefusesEvenASingleProposal` |
| B13 | if | 298:2 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no per-test profile in the attribution set entered it |
| B14 | if | 306:2 | arm entered 2x (engine tagged suite); arm not entered (engine untagged suite); `TestSymbolsWithNoProposalAtAllAreCountedRefusedRatherThanArbitrated` |
| B15 | range | 314:2 | arm entered 11x (engine tagged suite); arm not entered (engine untagged suite); `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestAnUncalibratedMarketRefusesEvenASingleProposal`, `TestEntriesComeBackInOwnerScopeOrderNotRouteOrder`, `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently`, `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `len` | 212:144 |
| `len` | 214:31 |
| `fail` | 215:10 |
| `fail` | 218:10 |
| `fail` | 221:10 |
| `strings.TrimSpace` | 223:13 |
| `loader.getenv` | 223:31 |
| `DecodeString` | 224:14 |
| `base64.StdEncoding.Strict` | 224:14 |
| `base64.StdEncoding.EncodeToString` | 225:19 |
| `len` | 225:72 |
| `fail` | 226:10 |
| `strings.TrimSpace` | 232:12 |
| `loader.getenv` | 232:30 |
| `make` | 233:13 |
| `len` | 233:58 |
| `make` | 234:14 |
| `len` | 234:59 |
| `entry.approved.Symbol` | 236:13 |
| `bySymbol.approved.Valid` | 237:22 |
| `fail` | 238:11 |
| `append` | 241:13 |
| `entry.route.Request` | 241:97 |
| `loader.load` | 243:16 |
| `strategyrouter.Market` | 244:42 |
| `strings.TrimSpace` | 244:111 |
| `loader.getenv` | 244:129 |
| `ed25519.PublicKey` | 245:15 |
| `strings.TrimSpace` | 248:23 |
| `loader.getenv` | 248:41 |
| `batch.ManifestDigest` | 249:19 |
| `fail` | 250:10 |
| `batch.Fault` | 254:22 |
| `fail` | 255:13 |
| `absence.String` | 257:37 |
| `coordinateMarketProposals` | 270:26 |
| `fail` | 272:13 |
| `fail` | 281:13 |
| `fail` | 289:13 |
| `string` | 291:40 |
| `arbitration.entries` | 297:23 |
| `fail` | 299:13 |
| `len` | 306:5 |
| `fail` | 307:13 |
| `sha256.New` | 313:7 |
| `h.Write` | 315:10 |
| `(unnamed)` | 315:18 |
| `entry.route.approved.Symbol` | 315:25 |
| `entry.authority.Proposal` | 315:66 |
| `len` | 318:16 |
| `len` | 318:52 |
| `hex.EncodeToString` | 319:34 |
| `h.Sum` | 319:53 |

## State mutations and fallbacks

- AST assignments: 48. Defers: 0. Goroutine statements: 0.

## Safety conclusion

이 함수는 거절을 느슨하게 한 적이 없다. 5.4.2 가 닫히는 경우를 셋 늘렸고 — 계보 신원
충돌(B10), 큐 넘침(B11), 되돌릴 자리 없음(B13) — 5.4.3 이 넷째를 더했다: 배치가 받아들인
스코프의 제안을 잃음(B9). 넷 다 이전에는 존재하지 않던 닫힘이며 전부 보수 방향이다.
B10 은 아직 진입 0 회다. 이것은 통과가 아니라 커버리지 구멍이며 그대로 남겨 보고한다.

B9 를 **조정 앞**에 둔 이유: 조정까지 가면 잃은 종목은 그냥 없는 종목처럼 보이고,
짧아진 목록이 아래 관문을 오히려 만족시킨다. 닫으려면 줄어들기 전에 봐야 한다.

`ProductionFault` 는 어느 종목의 어느 레인이 무엇 때문에 사라졌는지를 한 줄로 싣는다.
`QueueDropCount` 와 마찬가지로 **아직 운영자에게 닿지 않는다** — 투영은 태스크 7.3 이다.

닫힘 가지 넷은 이제 `RefusedCount` 에 `RoutedCount` 를 그대로 싣는다. 시장이 닫히면 경로에 오른
종목 전부가 못 나간 것이기 때문이다. 앞선 판본의 `refused + 1` 은 10,001 종목이 하나도 못 나간
주기를 "1건 거절"로 보고했다.

`QueueDropCount` 는 지금 배선에서 **언제나 0** 이다. 계수기가 오르는 두 곳(접힘·넘침)이 둘 다
구조적으로 도달 불가이기 때문이다 — 종목 중복이 위(B7)에서 먼저 거절되고 매니페스트가
(종목, 레인)마다 하나만 실어 같은 칸이 두 번 오지 않으며, 큐 상한이 매니페스트 상한과 같아
정상 매니페스트는 넘칠 수 없다. 그러므로 이 수를 "조용한 유실이 없음의 증거"로 인용하지 않는다.
