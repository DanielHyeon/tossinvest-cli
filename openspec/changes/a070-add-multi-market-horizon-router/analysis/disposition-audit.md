# a070 처분 감사 (2026-10-01, 분석 전용 — 코드 변경 0)

측정 시점: `disposition-audit/measured-at.txt`(HEAD sha · codegraph 1.6.0). 근거는 CodeGraph callers(`disposition-audit/callers-*.txt`),
Go AST(`tools/logic-map` — `route.ast.json` · `routeset.ast.json`), 비시험 파일의 `strategyrouter.` 선택자 전수(`external-production-symbols.txt`),
그리고 grep 으로 보강한 호출 자리(CodeGraph 가 engine 의 `NewOwnerKey` 호출 넷을 못 잡아서 — 아래 표에 명시).

## 1. 결론 한 줄

a070 의 **타입 · owner 봉인 · market record 봉인 · 요청/결정 타입은 a072 · a112 생산 경로의 기반으로 살아 있고**, **`Route()` 는 a112 가
대체 · 차단했으며**, **스케줄러 durable record(CAS · rollback · store · legacy migration)와 `QuotaAuthority` 는 생산 호출자 0 인 섬**이다.
남은 태스크 3.1 · 3.4 · 4.2 · 4.3 중 3.1 은 a072 3.23 + a112 RouteSet/조정자가 이미 했고, 3.4 의 대상(공유 quota 소진 · migration 거절)은
배선된 적이 없다. → **권고 ②(부분 대체) — `--skip-specs` 아카이브로 닫기**.

## 2. 조각별 대조 (a070 코어 커밋 `7b57fd24` 의 파일)

| a070 조각 | 생산 호출자 (영수증) | a112 · a072 쪽 대응물 | 판정 |
|---|---|---|---|
| `types.go` — `Market` · `Horizon` · `DesiredState` · `RuntimeState` · `RefusalCode` | 비시험 외부 선택자 `strategyrouter.Market` 37 · `MarketKR/US` 42 · `Horizon*` 35 · `DesiredState` 2 · `RuntimeState` 3 · `RefusalNone` 4 (`external-production-symbols.txt`) | 그대로 씀 | **살아 있음** |
| `router.go` `OwnerKey` · `NewOwnerKey` · `validOwnerKey` | `strategyrouter.OwnerKey` 27 · `NewOwnerKey` 5 — `strategyflow/flow.go:166`(CodeGraph) · `app/engine/strategy_first_leg_owner_scope.go:30,38` · `strategy_owner_scope_authority.go:34`(grep) · `OwnerKey{}` 리터럴 `strategy_lane_runtime.go:236` · `strategy_market_coordinator.go:70` · `strategyarbiter/arbiter.go:118` | a112 owner-scope 축(조정자 intake 키 · first-leg owner scope)이 **이 타입을 그대로** 씀. `riskbucket.OwnerKey{AccountID,Market,Symbol,ProspectiveGeneration string}` 는 a066 의 별개 원장 키(generation 이 문자열 · 의미가 prospective) — 둘은 병존하고 다리는 engine 의 owner-scope 권한 파일 | **살아 있음** |
| `router.go` `Owner` · `OwnerSnapshot` · `newOwnerSnapshot` · `ownerSnapshotSeal` | `newOwnerSnapshot` ← `LoadProductionRouteAuthorityBatch`(`production.go:318`, a072 3.23 — `risk_bucket_owners` 에서 재구성, `PRAGMA user_version` 고정) · `ownerSnapshotSeal` ← `RouteSet`(`routeset.go:54`) · `OwnerSnapshot` ← `strategyarbiter/arbiter.go` | a072 생산 router 권한의 owner 재구성 · a112 RouteSet 전단 검증 | **살아 있음** |
| `scheduler.go` `MarketRecord` · `newMarketRecord` · `EvaluateMarketLifecycle` | `newMarketRecord` · `EvaluateMarketLifecycle` ← `LoadProductionRouteAuthorityBatch`(`production.go:359,364` — **서명 매니페스트에서 매 로드 새로 만듦**) · `EvaluateMarketLifecycle` ← `RouteSet`(`routeset.go:100`) | a072 서명 route 매니페스트(3.23) | **살아 있음(값 타입으로만)** |
| `router.go` `Candidate` · `RouteRequest` · `RouteDecision` | `RouteRequest` 4 · `RouteDecision` 6 — `strategyflow` · `strategyproposal/production.go` · `strategyarbiter` · `production.go` | a112 RouteSet 입력/출력 | **살아 있음** |
| `registry.go` `Descriptor` · `Descriptors` · `ValidateDescriptors` | a072 가 `production.go` 와 함께 확장(`8022f578`) | a112 4.1 이 6→8 descriptor | **살아 있음(a072/a112 가 소유)** |
| `router.go` **`Route()`** | **생산 호출자 0**(CodeGraph `callers-Route.txt` 12 건 전부 `_test.go`) | a112 4.3.1 이 `RouteSet` 으로 대체, **4.3.2 의 AST/import 가드가 생산 폐포에서 `Route` 선택자 호출을 금지**(결정 47: `router.go` 는 한 줄도 안 고침). AST 대조(`route.ast.json` 25 분기 ↔ `routeset.ast.json` 23 분기): 앞 20 분기(키 · 봉인 · 범위 · owner revision · 신선도 · active owner 수 · market record · revision · lifecycle · 기존 owner 보존)는 **원문 동일**, 갈리는 것은 끝의 원시 `Score` 최고점 선택 · 동점 거절 5 분기 ↔ 전부 보존 · 레인 중복 거절 3 분기 | **대체됨** |
| `scheduler.go` `SchedulerState` · `NewSchedulerState` · `CASMarketRecord` · `RollbackMarketRecord` · `MarketRecordStore` · `DefaultMarketRecord` · `ValidSchedulerState` | **생산 폐포 0**: `CASMarketRecord` 의 비시험 호출자는 `RollbackMarketRecord`(scheduler.go:166) · `MarketRecordStore.CAS`(:204) 뿐이고 둘 다 생산 호출자 0, `NewSchedulerState` 는 `MigrateLegacy` · `NewMarketRecordStore` 뿐이고 둘 다 0 | 시장 desired 상태 = `internal/scheduler.RestorePairedProduction`(`strategy_schedule_authority.go:135`, a072 3.17), 시장 활성화 = 서명 매니페스트 → `newMarketRecord`(a072 3.23), 가족 활성화 = `strategyrouter.LoadProductionFamilyActivation`(`strategy_family_activation.go:183`, a112 8.7.1), 레인 cadence · deadline · backoff · latch = `strategyworker.Lane`(a112 5.3.1/5.3.2) + 원장 latch(5.3.3, schema v32) | **섬(생산 0)** |
| `migration.go` `MigrateLegacy` · `LegacyState` | **생산 호출자 0**(`callers-MigrateLegacy.txt` 2 건 다 시험) | 이관할 legacy durable record 자체가 생산에 없음(위 행) | **섬(생산 0)** |
| `quota.go` `QuotaAuthority` · `AcquireRequest` · `Capability` | **생산 호출자 0**(`callers-NewQuotaAuthority.txt` · 비시험 `strategyrouter.Quota*`/`Capability` 선택자 0). 스냅숏 생성자 `newQuotaSnapshot` 은 패키지 비공개라 생산에서 `Install` 할 값도 못 만듦 | 정본 market-aware-scheduler 「polling은 API 예산을 침범하지 않는다」의 구현은 **`internal/scheduler.BudgetCoordinator`**(`d4c73400`, 2026-08-01 — a070 이전) — 다만 이것도 진입 경로 생산 호출자 0(`callers-Decide.txt` · `callers-NewBudgetCoordinator.txt`). a112 7.1(미착수)이 family subscope 를 얹을 자리 | **섬(생산 0) + 정본과 어긋남**(아래 §4) |

## 3. 남은 태스크의 실제 상태

| a070 태스크 | 실측 |
|---|---|
| 3.1 a067–a069 레인 · a066 owner snapshot 통합 | **이미 됨 — 다른 change 가**: a072 3.23(`LoadProductionRouteAuthorityBatch` 가 `risk_bucket_owners` 에서 owner snapshot 을 재구성해 `RouteRequest` 를 봉인) + a072 3.24(레인 proposal) + a112 RouteSet · 조정자(owner 우선 보존 · 동점/다중 owner 거절 — a112 four-family-strategy-runtime 델타). a067 · a068 · a069 · a066 · a072 전부 아카이브 |
| 3.4 공유 quota 소진 · migration 거절 · 한 시장 장애에서 안전 루프 지속 | 앞 둘은 **대상 코드가 배선된 적 없음**(§2 섬). 한 시장 장애 · OFF 의 안전 루프 지속은 a072 4.4 · strategy-runtime 정본 「entry OFF는 safety lifecycle을 중단하지 않는다」가 소유. 남은 실재 공백(예산 소진 시 안전 루프)은 a112 2.8(미착수)이 소유 |
| 4.2 넓은 회귀 · 4.3 sdd/gate | 대상이 위 둘이라 독립 의미 없음 |

## 4. 델타 spec 7 요구의 행방 (정본 `openspec/specs/` 에 a070 요구 0 — 제목 전수 grep)

| a070 요구 | 덮는 것 | 적용하면 |
|---|---|---|
| router: owning lane 하나 결정(동점 → ambiguity) | **충돌** — a112 strategy-engine 델타 「production route authority는 평가 전 cross-family winner를 고르지 않는다」(raw `Candidate.Score` 사전 선택 MUST NOT). owner 키 · 보존 · 다중 owner 거절 부분은 a112 four-family-strategy-runtime 「market coordinator는 owner와 calibrated score로 최대 한 proposal만」이 승계 | 정본이 a112 와 모순 |
| router: KR/US routing lifecycle 독립 | 정본 market-aware-scheduler 「KR과 US scheduler binding은 동시에 독립 진행한다」 · strategy-runtime 「KR과 US evaluation은 독립 scope」 | 중복 |
| router: 순수 · OFF 무 mutation | 정본 strategy-engine 「production strategy dispatch는 router와 campaign lineage를 보존한다」(router mutation MUST NOT) · a112 「lane worker는 mutation 권위 없는 sealed proposal만」 | 중복 |
| scheduler: KR/US record 독립 revision · activation(durable CAS) | 생산은 서명 매니페스트 + `scheduler.Restore` — durable CAS record 미구현 | **구현되지 않은 SHALL 이 정본에** |
| scheduler: legacy state 이관 | 미구현(§2) | 동상 |
| scheduler: KR/US session 독립 | 정본 「KR과 US scheduler binding은 동시에 독립 진행한다」 | 중복 |
| scheduler: market · horizon capability = 공유 quota subscope | a112 market-aware-scheduler 델타 「entry polling capability는 market horizon family subscope에 결합된다」(family 를 더한 상위집합) — 미구현(a112 7.1) | a112 와 이중 정본 |

→ 어느 요구도 지금 정본에 들어갈 자격이 없다(중복 · 충돌 · 미구현). a123 선례(「구현되지 않은 SHALL 이 정본에 들어가면 안 된다」 → `--skip-specs`)와 같은 모양.

## 5. 선택지와 영수증

**① 잔여 통합을 a112-뒤 세계로 재작성 후 계속 — 비권고.** 3.1 은 a072 3.23 · a112 RouteSet/조정자가 이미 했다(§3). 계속하려면 대상이
배선된 적 없는 스케줄러 섬 · `QuotaAuthority` 를 생산에 붙이는 일이 되는데, 시장/가족 활성화는 서명 매니페스트(a072 3.17 · 3.23, a112 8.7.1)와
`strategyworker.Lane`(a112 5.3.x)이 이미 소유한다 — 두 번째 권한 표면을 만드는 것이고 a112 L4 · L5 와 소유권이 겹친다.

**② 부분 대체 — 권고.** 살릴 조각(§2 「살아 있음」)은 **이미 생산에 있어** 따로 옮길 것이 없다. 처분은 문서뿐:
1. status.md 를 아래 §6 으로 갱신(현 status 의 Scheduler · Quota 줄은 생산에 없는 것을 「GREEN」처럼 적고 있음).
2. `openspec archive a070-add-multi-market-horizon-router --skip-specs` — 미완료 3.1 · 3.4 · 4.2 · 4.3 은 「대체/대상 없음」 처분으로 닫고
   review 에 이 감사를 인용(a123 선례: 미완료 11 · `--skip-specs`).
3. a112 쪽 정리(a112 소유자 판단): dependency-matrix 의 「a070 router/scheduler INCOMPLETE → L6 · L7 차단」 행을 「대체됨 — router 계약은
   a112 RouteSet + 조정자, 스케줄러 계약은 a072 서명 매니페스트 · `scheduler.Restore` · a112 FamilyActivation」로. 이 행이 지금 L6 의 형식상
   선행 조건이다(`analysis/goldens/dependency-matrix.json`, a112 proposal 「a070 router/scheduler contract」).
4. **섬 코드의 처분은 별건**: `Route()`(a112 시험 다수가 참조 대조로 부름) · 스케줄러 섬 · `QuotaAuthority`. 지금 지우면 a112 시험을
   건드린다. a112 7.1 이 quota 권한을 하나로 정한 뒤 정리 change 하나로 묶는 것을 권한다.

**③ 불구현 아카이브(a112 가 capability 를 대체) — 부정확.** a070 코어의 타입 · owner/market 봉인 · 요청/결정 타입은 a112 가 **대체한 것이 아니라
그 위에 선** 것이다(§2: 패키지 밖 비시험 선택자 250 건, `RouteSet` 의 앞 20 분기가 `Route` 와 원문 동일). 대체된 것은 `Route()` 의 선택 꼬리와
spec 의 단일 승자 요구뿐이다. ③ 의 문언으로 닫으면 기록이 사실과 다르다 — 행동(`--skip-specs` 아카이브)은 ② 와 같고 사유가 다르다.

## 6. status.md 갱신 제안

```markdown
# Status — a070-add-multi-market-horizon-router

- 처분(2026-10-01 감사 `analysis/disposition-audit.md`): 부분 대체 — 아카이브 대상(`--skip-specs`)
- 살아 있음(a072 · a112 생산 경로의 기반): `types.go`, `OwnerKey`/`NewOwnerKey`, `Owner`/`OwnerSnapshot` 봉인,
  `MarketRecord`/`EvaluateMarketLifecycle`(서명 매니페스트에서 매 로드 생성), `RouteRequest`/`Candidate`/`RouteDecision`, descriptor registry
- 대체됨: `Route()` 의 원시 점수 단일 승자 선택 → a112 `RouteSet` + 보정 점수 조정자(4.3.1, 생산 폐포 금지 가드 4.3.2)
- 생산 호출자 0(섬): durable `SchedulerState`/CAS/rollback/`MarketRecordStore`, `MigrateLegacy`, `QuotaAuthority`
  — 시장 · 가족 활성화는 서명 매니페스트(a072 3.17 · 3.23, a112 8.7.1), 레인 cadence/latch 는 `strategyworker.Lane`(a112 5.3.x)
- 남은 태스크: 3.1 은 a072 3.23 · a112 가 수행, 3.4 의 quota · migration 대상은 배선된 적 없음, 4.2 · 4.3 은 독립 의미 없음
- spec delta: 적용하지 않음 — 중복(정본 market-aware-scheduler · strategy-runtime · strategy-engine), 충돌(a112 「평가 전 cross-family winner 금지」),
  미구현(durable CAS · legacy migration · quota subscope → a112 7.1)

No market was selected, no lane was enabled, and no live order/toggle/activation was created.
```

## 7. 사람 판단이 필요한 것

- **quota 권한의 단일화(a112 7.1 의 입력).** 정본 「polling은 API 예산을 침범하지 않는다」는 암호학적 난수 capability · 발급 기억 ·
  observation cycle · reset generation 증명을 요구하고, 그 구현은 `internal/scheduler.BudgetCoordinator` 다. a070 `QuotaAuthority` 는
  토큰이 요청 필드 + 스냅숏 digest 의 sha256(`quota.go` `capabilityToken`)이라 결정적이고, reset generation 진행 · observation cycle ·
  `SafetyReserve` 규칙(50% 올림 · 최소 5)이 없으며 reserve 는 호출자가 넣는다. a112 L4 원장은 `internal/strategyrouter/{router,quota,scheduler}.go`
  를 7.1 의 편집 대상으로 적었다 — 7.1 이 `QuotaAuthority` 에 family 를 얹으면 정본보다 약한 두 번째 예산 권한이 생산에 선다.
  권고: 7.1 은 `scheduler.BudgetCoordinator` 에 `(market, horizon, family)` subscope 를 얹는다.
- 섬 코드 삭제 시점(§5 ②-4).
</content>
</invoke>
