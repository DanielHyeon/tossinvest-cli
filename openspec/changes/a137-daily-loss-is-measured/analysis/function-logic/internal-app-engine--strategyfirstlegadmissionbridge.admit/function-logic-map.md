# Function Logic Map: `strategyFirstLegAdmissionBridge.admit`

- Source: `internal/app/engine/strategy_first_leg_admission.go` (74-110, **편집 전** — HEAD `f030e27e`, `source_sha256` 254b4a6a…)
- AST evidence: `ast.json` — branches 8 · returns 7. 편집 뒤 재추출 예정(tasks 1.3)
- Risk scan: `risk-pattern-report.md`
- 생산 호출자: `strategyDispatchCycle.dispatch`(strategy_dispatch_cycle.go:151, `cycle.firstLeg.admit`) 하나. 시험 호출: strategy_first_leg_admission_test.go 7 · a112_family_gate_test.go 2

**역할.** 전략 첫 다리 결과 하나를 Guardian 의 precheck → 원자 발급으로 승인하고, 실패를 코드(`StrategyFirstLegAdmissionCode`)와 원인(`cause`)으로 분류해
돌려줌. 이 파일은 원장 · Gateway 능력을 갖지 않는다(`TestStrategyFirstLegAdmissionSourceHasOnlyGuardianCapability` 가 `*journal.Journal` 등 금지 문자열을 고정).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `result` | `strategyflow.Result` — 완결 계보 · 유효 집행 조건 · 시장 KR/US | 전략 흐름 | B1 → `validateStrategyFirstLegResult` 의 거절 그대로 |
| `b`, `b.loader`, `b.guardian` | 셋 다 non-nil | 엔진 조립 | B2 → `AUTHORITY_UNAVAILABLE` |
| `ctx` | 취소 가능 | 감독자 주기 | loader · 발급이 존중 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (76) | `refusal.Code != ""` | 없음 | 결과 검증 거절 | `TestStrategyFirstLegAdmissionRejectsInvalidBeforeAuthority` |
| B2 (79) | `b == nil \|\| b.loader == nil \|\| b.guardian == nil` | 없음 | `AUTHORITY_UNAVAILABLE` | `TestStrategyFirstLegAdmissionDormantPairedKRUS` |
| B3 (83) | 권한 수집 오류 | 없음 | `AUTHORITY_COLLECTION_FAILED`, **cause = err** | `TestStrategyFirstLegAdmissionClassifiesPairedFailures/*-collection` |
| B4 (88) | 권한 검증 실패 | 없음 | `AUTHORITY_MISMATCH`, cause 없음 | `TestStrategyFirstLegAdmissionRejectsAuthorityTamperPairedKRUS` |
| B5 (92) | precheck 오류 | 없음(precheck 는 mutation-free) | `AUTHORITY_MISMATCH` — **원인 버림**(B6·B7 외) | `…ClassifiesPairedFailures/*-precheck` |
| B6 (98) | B5 안: `errors.As(err, *QFinalRefusal)` 이고 `BUCKET_CAP_EXHAUSTED` | 없음 | B7 로 | a112 6.2 시험 |
| B7 (99) | B6 안: 소유자 범위 키를 유도함 | 없음 | cause = `*strategyScopeRefusal`(범위 거절 → 다음 범위 계속) | a112 6.2 시험 |
| B8 (106) | 원자 발급 오류 | 발급 트랜잭션 롤백(호출 대상 쪽) | `ATOMIC_ADMISSION_FAILED`, cause 없음 | `…ClassifiesPairedFailures/*-atomic` |
| (끝) | 발급 성공 | 결정 · 예약 · 캠페인 커밋(호출 대상 쪽) | `ADMITTED` + receipt | `TestStrategyFirstLegAdmissionUsesGuardianOnlyAllEightLanes` |

Exact AST return positions: 77 · 80 · 86 · 89 · 103 · 107 · 109

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `validateStrategyFirstLegResult` | 결과 형식 검증 | 거절 결과(오류 아님) | AST calls |
| `b.loader.collectStrategyFirstLegAuthority` | 계좌 · 위험 · FX · 스케줄 권한 조립 | 오류는 B3 로 원인째 | strategy_account_first_leg_authority.go |
| `validateStrategyFirstLegAuthority` | 조립 권한과 결과의 결속 | B4 | AST calls |
| `b.guardian.PrecheckQFinalCampaignFirstLeg` | mutation-free precheck(체인 포함) | 오류 = 체인 거절 · q_final 거절 · 계보 거절 | riskguardian_first_leg.go:49 → `PrecheckQFinalEntry` |
| `b.guardian.IssuePrecheckedQFinalCampaignFirstLeg` | 원자 발급 | B8 | riskguardian_first_leg.go:117 |

## State mutations and fallbacks

- 이 함수 자체는 상태를 바꾸지 않음. 발급 성공 시 원장 변경은 Guardian 쪽 트랜잭션.
- fallback 없음. B5 의 `DAILY_LOSS_LIMIT_REACHED` 체인 거절은 오늘 원인 없는 `AUTHORITY_MISMATCH` 로 접힌다(design F7 · F8).

## a137 편집 계획 (Q6)

- B5 안에 새 갈래: `errors.As(err, *execgw.IssueRefusal)` 이고 `Reason == DAILY_LOSS_LIMIT_REACHED`(타입 · 코드로, 문구 아님) →
  Guardian 의 새 강화 입구(ctx, ENTRY_BLOCKED 영속)를 부르고 새 코드 `DAILY_LOSS_LIMIT` + cause = 그 거절(강화 오류는 `errors.Join`). 범위 거절이 아님.
- B5 의 나머지 거절도 cause 를 싣는다(F8 — 하류 `errors.As` 가능). 문구(Detail)는 그대로.
- 인터페이스 `strategyFirstLegGuardian` 에 강화 메서드 하나 추가 — 원장 능력을 이 파일에 들이지 않는다(금지 문자열 시험 유지).

## Safety conclusion

- Safe edit boundary: B5 의 분류와 원인 운반만. B1~B4 · B6~B8 무변경.
- High-risk impact: yes(진입 승인) — 진입을 더 막는 방향만(강화 추가). 매도 · 손절 경로는 이 함수를 지나지 않음.
