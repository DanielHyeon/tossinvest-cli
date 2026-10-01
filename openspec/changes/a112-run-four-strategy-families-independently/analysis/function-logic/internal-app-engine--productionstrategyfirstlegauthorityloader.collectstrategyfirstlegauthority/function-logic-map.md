# Function Logic Map: `collectStrategyFirstLegAuthority`

- Source: `internal/app/engine/strategy_account_first_leg_authority.go`
- Source SHA-256: `03beca9c77991951b18a6a97237327425a48188aa56bcb3c72bfb5c2c299cac5`
- Signature: `productionStrategyFirstLegAuthorityLoader.collectStrategyFirstLegAuthority(params=2, results=2)`
- Source range: `288:1`–`397:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 5.2.2.2 리뷰 수리).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 범위 거절 타입을 만드는 자리는 B8 의 범위 국소 갈래 하나뿐 — census `TestTheScopeRefusalTypeIsMadeOnlyWhereTheCensusSaysItIs` · `TestNoEngineErrorTypeImplementsAs`.
- 활성화 밖에서 마지막 경계의 수용 집합은 80ae96a5 이전(7ab8cd12)과 같다(B5).

## Branches and early returns

- Exact AST return nodes: `290:3, 297:3, 303:3, 307:3, 312:3, 319:3, 327:4, 329:3, 334:3, 340:3, 345:3, 350:3, 367:3, 376:4, 380:4, 382:3, 386:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 289:2 | loader · ctx · 시계 · 원장 · Guardian 부재 |
| B2 | if | 295:2 | 시장 권한 준비 미완 |
| B3 | if | 302:2 | 소유자 범위 선택 실패 — 위조 의심(타입 없음) |
| B4 | if | 306:2 | 봉인된 identity 대조 — 불일치는 타입 없는 오류 |
| B5 | if | 311:2 | **(새)** 서명 활성화 **없는** 시장에서 항목이 정확히 하나가 아님 → `paired production authority is incomplete for market`(편집 전 문구) |
| B6 | if | 318:2 | **(새)** 범위 키 정규화 실패 → 결함(타입 없음) |
| B7 | if | 324:2 | 그 범위의 준비된 위험 권한 없음 |
| B8 | if | 326:3 | **(새)** 원인이 범위 국소가 아님(원장 결함 · 무결성 · 항목 부재) → 타입 없는 결함 `production risk authority fault …: %w`(주기 멈춤 · 원인 보존); 범위 국소(`riskbucket.ErrProductionRiskScopeRefused`)면 범위 거절 타입(원인 Unwrap) |
| B9 | if | 332:2 | 그 범위의 준비된 계좌 권한 없음 → **언제나 결함**(타입 없는 오류 `production account authority fault …: %w` — 주기 멈춤 · 원인 보존). 계좌 매니페스트는 시장 단위 파일이라 범위 국소 원인이 없다(codex 재확인 P1 → Manager 판정 (A) — 앞 판의 「ctx 만 결함」 대체) |
| B10 | if | 339:2 | 계보 시장 통화 미지 |
| B11 | if | 343:2 | 위험 권한 범위 불일치(범위 번들) |
| B12 | if | 348:2 | 포지션 캠페인 CAS 변경 |
| B13 | range | 355:2 | 위험 버킷 항목 순회 |
| B14 | if | 366:2 | 가격 단위 무효 |
| B15 | if | 375:3 | 노출 스냅숏 만료(collect 클로저) |
| B16 | if | 379:3 | 예약 버전 읽기 실패(collect 클로저) |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `errors.New` | 290:51 |
| `StrategyMarket` | 292:12 |
| `loader.proposals.forMarket` | 293:52 |
| `loader.risk.forMarket` | 293:88 |
| `loader.fx.forMarket` | 294:3 |
| `loader.accounts.forMarket` | 294:32 |
| `loader.schedule.forMarket` | 294:67 |
| `errors.New` | 297:51 |
| `proposal.authorityForOwnerScope` | 301:31 |
| `errors.New` | 303:51 |
| `proposalAuthority.Proposal` | 305:12 |
| `result.ExecutionTerms.Identity` | 306:68 |
| `accepted.result.ExecutionTerms.Identity` | 306:104 |
| `errors.New` | 307:51 |
| `Verified` | 311:6 |
| `proposal.familyActivation` | 311:6 |
| `len` | 311:48 |
| `errors.New` | 312:51 |
| `strategyOwnerKeyOf` | 317:16 |
| `errors.New` | 319:51 |
| `riskAuthority.forScope` | 323:28 |
| `riskAuthority.riskScopeCause` | 325:24 |
| `fmt.Errorf` | 327:52 |
| `account.forScope` | 331:37 |
| `fmt.Errorf` | 334:51 |
| `account.accountScopeCause` | 335:4 |
| `errors.New` | 340:51 |
| `riskBundle.Scope` | 342:11 |
| `riskBundle.Validate` | 343:12 |
| `string` | 344:3 |
| `string` | 344:27 |
| `scope.AsOf.Equal` | 344:102 |
| `errors.New` | 345:51 |
| `loader.journal.CurrentPositionCampaignCAS` | 347:14 |
| `string` | 347:88 |
| `uint64` | 349:40 |
| `errors.New` | 350:51 |
| `riskBundle.Entries` | 352:13 |
| `make` | 353:13 |
| `len` | 353:50 |
| `make` | 354:16 |
| `len` | 354:63 |
| `append` | 356:13 |
| `append` | 358:16 |
| `MajorDecimal` | 363:25 |
| `result.ExecutionTerms.Entry` | 363:25 |
| `MajorDecimal` | 364:23 |
| `result.ExecutionTerms.EffectiveStop` | 364:23 |
| `MajorDecimal` | 365:27 |
| `result.ExecutionTerms.Target` | 365:27 |
| `errors.New` | 367:51 |
| `strategyFirstLegBindingDigest` | 369:19 |
| `riskBundle.Digest` | 369:75 |
| `strings.TrimPrefix` | 370:38 |
| `strings.TrimPrefix` | 371:37 |
| `UTC` | 374:10 |
| `loader.clk.Now` | 374:10 |
| `readCtx.Err` | 375:24 |
| `now.IsZero` | 375:48 |
| `now.After` | 375:64 |
| `accountAuthority.FreshUntil` | 375:74 |
| `errors.New` | 376:38 |
| `loader.journal.ReservationVersion` | 378:26 |
| `accountAuthority.ObservedAt` | 382:40 |
| `accountAuthority.OpenExposure` | 382:103 |
| `string` | 386:89 |
| `accountAuthority.AccountState` | 388:12 |
| `riskBundle.Policy` | 389:80 |
| `loader.guardian.PolicyVersion` | 391:26 |
| `loader.guardian.LimitsDigest` | 391:81 |
| `strings.TrimPrefix` | 394:81 |
| `strings.TrimPrefix` | 395:42 |
| `strings.TrimPrefix` | 396:39 |
| `proposalAuthority.WeeklyBinding` | 396:99 |

## State mutations and fallbacks

- 이 함수 자신은 원장을 쓰지 않는다. 봉투 · 시장 값으로의 폴백 없음(범위 권한이 없으면 거절 또는 준비 안 됨).

## Safety conclusion

- High-risk(1차 레그). 편집은 거절을 **늘리기만** 한다(활성화 없는 다항목 쌍 · 범위 국소가 아닌 결함은 이제 주기를 멈춤). 새로 통과하는 입력 0.
