# Function Logic Map: `collectStrategyFirstLegAuthority`

- Source: `internal/app/engine/strategy_account_first_leg_authority.go`
- Source SHA-256: `e6c12de7902b15de91de03a8da004f1ad167ce37d37580cf52766f4707ed0a4d`
- Signature: `productionStrategyFirstLegAuthorityLoader.collectStrategyFirstLegAuthority(params=2, results=2)`
- Source range: `213:1`–`293:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 6.2 봉인 로트).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 입력: loader 자기 권한 쌍(조립이 새로 고침 때 중재한 제안 · 위험 · 환율 · 계좌 · 일정)과 건너온 `accepted`(봉투가 나른 결과).
- **봉인 불변식(6.2):** 발급하는 1차 레그의 결과는 건너온 값이 아니라 **조립의 권한 쌍에서 소유자 범위로 다시 꺼낸 항목**이고, 그 항목의 봉인된 identity 가 accepted 와 같아야 한다. 범위는 accepted 계보에서 읽되 선택 기준일 뿐 대조 대상이 아니다(자기 참조 함정 회피 — `authorityForOwnerScope` 머리말).
- 시장 단위 개수 관문(B4)은 봉인이 아니라 상한이고 5.2.2.2 가 걷어 낸다. 이 편집은 판정을 더 엄격하게만 한다(새로 통과하는 입력 0).

## Branches and early returns

- Exact AST return nodes: `215:3, 222:3, 228:3, 232:3, 236:3, 241:3, 246:3, 263:3, 272:4, 276:4, 278:3, 282:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 214:2 | loader · ctx · 시계 · 원장 · Guardian 부재 → 발급 불가 |
| B2 | if | 220:2 | 위험 · 환율 · 계좌 · 일정 준비 · 활성화 부재 → `paired production authority is incomplete for market`. **편집: 개수 조건을 떼어 냄**(아래 B4) |
| B3 | if | 227:2 | **(새) 소유자 범위 선택 실패** — 조립의 권한 쌍에 accepted 범위가 정확히 하나가 아님(0: 미선택 범위 · 다른 시장, 2+: 범위당 하나 붕괴) → `production proposal identity changed: owner scope is not uniquely authorized by the assembly` |
| B4 | if | 231:2 | **(새, 편집 전 B2 에서 분리) 시장 단위 개수 관문** `len(proposal.entries) != 1` — 봉인이 아니라 시장당 하나 상한, 걷어 내는 일은 5.2.2.2 |
| B5 | if | 235:2 | 봉인된 identity 대조(편집 전 B3, 조건 불변) — 선택된 항목의 `Lineage.Identity` · `ExecutionTerms.Identity()` 와 accepted 비교 |
| B6 | if | 239:2 | 위험 권한 범위 불일치(편집 전 B4) |
| B7 | if | 244:2 | 포지션 캠페인 CAS 변경(편집 전 B5) |
| B8 | range | 251:2 | 위험 버킷 항목 순회(편집 전 B6) |
| B9 | if | 262:2 | 가격 단위 무효(편집 전 B7) |
| B10 | if | 271:3 | 노출 스냅숏 만료(collect 클로저, 편집 전 B8) |
| B11 | if | 275:3 | 예약 버전 읽기 실패(collect 클로저, 편집 전 B9) |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `errors.New` | 215:51 |
| `StrategyMarket` | 217:12 |
| `loader.proposals.forMarket` | 218:52 |
| `loader.risk.forMarket` | 218:88 |
| `loader.fx.forMarket` | 219:3 |
| `loader.accounts.forMarket` | 219:32 |
| `loader.schedule.forMarket` | 219:67 |
| `errors.New` | 222:51 |
| `proposal.authorityForOwnerScope` | 226:31 |
| `errors.New` | 228:51 |
| `len` | 231:5 |
| `errors.New` | 232:51 |
| `proposalAuthority.Proposal` | 234:12 |
| `result.ExecutionTerms.Identity` | 235:68 |
| `accepted.result.ExecutionTerms.Identity` | 235:104 |
| `errors.New` | 236:51 |
| `riskAuthority.bundle.Scope` | 238:11 |
| `riskAuthority.bundle.Validate` | 239:12 |
| `string` | 240:3 |
| `string` | 240:27 |
| `scope.AsOf.Equal` | 240:102 |
| `errors.New` | 241:51 |
| `loader.journal.CurrentPositionCampaignCAS` | 243:14 |
| `string` | 243:88 |
| `uint64` | 245:40 |
| `errors.New` | 246:51 |
| `riskAuthority.bundle.Entries` | 248:13 |
| `make` | 249:13 |
| `len` | 249:50 |
| `make` | 250:16 |
| `len` | 250:63 |
| `append` | 252:13 |
| `append` | 254:16 |
| `MajorDecimal` | 259:25 |
| `result.ExecutionTerms.Entry` | 259:25 |
| `MajorDecimal` | 260:23 |
| `result.ExecutionTerms.EffectiveStop` | 260:23 |
| `MajorDecimal` | 261:27 |
| `result.ExecutionTerms.Target` | 261:27 |
| `errors.New` | 263:51 |
| `strategyFirstLegBindingDigest` | 265:19 |
| `riskAuthority.bundle.Digest` | 265:76 |
| `strings.TrimPrefix` | 266:38 |
| `strings.TrimPrefix` | 267:37 |
| `UTC` | 270:10 |
| `loader.clk.Now` | 270:10 |
| `readCtx.Err` | 271:24 |
| `now.IsZero` | 271:48 |
| `now.After` | 271:64 |
| `account.authority.FreshUntil` | 271:74 |
| `errors.New` | 272:38 |
| `loader.journal.ReservationVersion` | 274:26 |
| `account.authority.ObservedAt` | 278:40 |
| `account.authority.OpenExposure` | 278:104 |
| `string` | 282:89 |
| `account.authority.AccountState` | 284:12 |
| `riskAuthority.bundle.Policy` | 285:80 |
| `loader.guardian.PolicyVersion` | 287:26 |
| `loader.guardian.LimitsDigest` | 287:81 |
| `strings.TrimPrefix` | 290:81 |
| `strings.TrimPrefix` | 291:42 |
| `strings.TrimPrefix` | 292:39 |
| `proposalAuthority.WeeklyBinding` | 292:99 |

## State mutations and fallbacks

- 이 함수 자신은 원장을 쓰지 않는다(읽기: 캠페인 CAS · 예약 버전). 발급 값은 호출자가 q_final 입장에서 쓴다.

## Safety conclusion

- High-risk impact: yes(1차 레그 발급). 편집은 거절 갈래를 하나 더하고(B3) 선택 기준을 위치에서 소유자 범위로 바꿨다 — 편집 전 통과하던 입력 중 새로 통과하는 것은 없다(단일 항목 쌍에서 범위가 같으면 편집 전과 같은 항목을 고르고, 다르면 편집 전에도 identity 대조에서 거절됐다).
- 위조 축 다섯 · 기제 둘은 행동 시험, 선택 기제 · 대조 약화 · 재유도 제거는 변이 S01~S07 이 CAUGHT.
