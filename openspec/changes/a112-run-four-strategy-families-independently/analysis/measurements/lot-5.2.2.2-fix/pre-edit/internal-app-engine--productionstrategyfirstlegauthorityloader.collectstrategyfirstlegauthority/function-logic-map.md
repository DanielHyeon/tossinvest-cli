# Function Logic Map: `collectStrategyFirstLegAuthority`

- Source: `internal/app/engine/strategy_account_first_leg_authority.go`
- Source SHA-256: `d0d6281292dafcc979edce741a3a2bf98ed348f023267d8198d2436c71ec7291`
- Signature: `productionStrategyFirstLegAuthorityLoader.collectStrategyFirstLegAuthority(params=2, results=2)`
- Source range: `255:1`–`348:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 5.2.2.2).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 범위 거절 타입(`strategyScopeRefusal`)을 만드는 자리는 B5 · B6 둘뿐이다 — 언급 census `a112_scope_refusal_census_test.go` `TestTheScopeRefusalTypeIsMadeOnlyWhereTheCensusSaysItIs`.
- 시장 단위 개수 관문이 지키던 것(관문 전수표 (a)~(e))은 review 「5.2.2.2」 절의 대체 수단 · 시험으로 옮겼다.

## Branches and early returns

- Exact AST return nodes: `257:3, 264:3, 270:3, 274:3, 282:3, 286:3, 291:3, 296:3, 301:3, 318:3, 327:4, 331:4, 333:3, 337:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 256:2 | loader · ctx · 시계 · 원장 · Guardian 부재 → 발급 불가 |
| B2 | if | 262:2 | 시장 권한(위험 · 환율 · 계좌 · 일정) 준비 미완 → `paired production authority is incomplete for market`(개수 조건은 6.2 에서 이미 뗌) |
| B3 | if | 269:2 | 소유자 범위 선택 실패(0 또는 복수) → `…owner scope is not uniquely authorized by the assembly` — **범위 거절 타입 아님**(위조 의심 → 주기 멈춤) |
| B4 | if | 273:2 | 봉인된 identity 대조(편집 전 B5) — 불일치는 **타입 없는 오류**(범위 거절 아님) |
| B5 | if | 281:2 | **(새)** 그 범위의 준비된 위험 권한 없음 → `*strategyScopeRefusal`(그 범위만 거절 · 봉투 폴백 없음, J3) |
| B6 | if | 285:2 | **(새)** 그 범위의 준비된 계좌 권한 없음 → `*strategyScopeRefusal` |
| B7 | if | 290:2 | **(새)** 계보 시장 통화 미지(KR/US 밖) → 발급 불가 — 통화는 봉투가 아니라 `Lineage.Market` 에서 유도(A#3) |
| B8 | if | 294:2 | 위험 권한 범위 불일치(편집 전 B6) — 이제 **그 범위의** 번들로 대조 |
| B9 | if | 299:2 | 포지션 캠페인 CAS 변경(편집 전 B7) |
| B10 | range | 306:2 | 위험 버킷 항목 순회(편집 전 B8) — 범위 번들 |
| B11 | if | 317:2 | 가격 단위 무효(편집 전 B9) |
| B12 | if | 326:3 | 노출 스냅숏 만료(collect 클로저, 편집 전 B10) — 범위 계좌 권한의 FreshUntil |
| B13 | if | 330:3 | 예약 버전 읽기 실패(collect 클로저, 편집 전 B11) |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `errors.New` | 257:51 |
| `StrategyMarket` | 259:12 |
| `loader.proposals.forMarket` | 260:52 |
| `loader.risk.forMarket` | 260:88 |
| `loader.fx.forMarket` | 261:3 |
| `loader.accounts.forMarket` | 261:32 |
| `loader.schedule.forMarket` | 261:67 |
| `errors.New` | 264:51 |
| `proposal.authorityForOwnerScope` | 268:31 |
| `errors.New` | 270:51 |
| `proposalAuthority.Proposal` | 272:12 |
| `result.ExecutionTerms.Identity` | 273:68 |
| `accepted.result.ExecutionTerms.Identity` | 273:104 |
| `errors.New` | 274:51 |
| `strategyOwnerKeyOf` | 279:12 |
| `riskAuthority.forScope` | 280:28 |
| `account.forScope` | 284:37 |
| `errors.New` | 291:51 |
| `riskBundle.Scope` | 293:11 |
| `riskBundle.Validate` | 294:12 |
| `string` | 295:3 |
| `string` | 295:27 |
| `scope.AsOf.Equal` | 295:102 |
| `errors.New` | 296:51 |
| `loader.journal.CurrentPositionCampaignCAS` | 298:14 |
| `string` | 298:88 |
| `uint64` | 300:40 |
| `errors.New` | 301:51 |
| `riskBundle.Entries` | 303:13 |
| `make` | 304:13 |
| `len` | 304:50 |
| `make` | 305:16 |
| `len` | 305:63 |
| `append` | 307:13 |
| `append` | 309:16 |
| `MajorDecimal` | 314:25 |
| `result.ExecutionTerms.Entry` | 314:25 |
| `MajorDecimal` | 315:23 |
| `result.ExecutionTerms.EffectiveStop` | 315:23 |
| `MajorDecimal` | 316:27 |
| `result.ExecutionTerms.Target` | 316:27 |
| `errors.New` | 318:51 |
| `strategyFirstLegBindingDigest` | 320:19 |
| `riskBundle.Digest` | 320:75 |
| `strings.TrimPrefix` | 321:38 |
| `strings.TrimPrefix` | 322:37 |
| `UTC` | 325:10 |
| `loader.clk.Now` | 325:10 |
| `readCtx.Err` | 326:24 |
| `now.IsZero` | 326:48 |
| `now.After` | 326:64 |
| `accountAuthority.FreshUntil` | 326:74 |
| `errors.New` | 327:38 |
| `loader.journal.ReservationVersion` | 329:26 |
| `accountAuthority.ObservedAt` | 333:40 |
| `accountAuthority.OpenExposure` | 333:103 |
| `string` | 337:89 |
| `accountAuthority.AccountState` | 339:12 |
| `riskBundle.Policy` | 340:80 |
| `loader.guardian.PolicyVersion` | 342:26 |
| `loader.guardian.LimitsDigest` | 342:81 |
| `strings.TrimPrefix` | 345:81 |
| `strings.TrimPrefix` | 346:42 |
| `strings.TrimPrefix` | 347:39 |
| `proposalAuthority.WeeklyBinding` | 347:99 |

## State mutations and fallbacks

- 이 함수 자신은 원장을 쓰지 않는다. 봉투 · 시장 값으로의 폴백 없음(범위 권한이 없으면 거절 또는 준비 안 됨).

## Safety conclusion

- High-risk(1차 레그 발급). 편집은 개수 관문을 걷어 **범위 둘인 활성화 시장의 발급을 연다** — 각 범위는 자기 위험 · 계좌 권한으로만 발급되고, 권한이 없는 범위는 그 범위만 타입 거절된다. 활성화 없는 시장은 결과 · 계좌 권한이 여전히 항목 하나를 요구하므로 새 통과 입력 0.
- 위조 다섯 축은 범위 하나 · 둘(두 순서) 쌍에서 모두 거절(행동 시험), 오분류 · 폴백 변이 X05~X11 · X21 · X22 CAUGHT.
