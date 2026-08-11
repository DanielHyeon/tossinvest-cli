# Function Logic Map: `Gateway.checkProtection`

- Source: `internal/execgw/protection.go` (L89-110)
- AST evidence: `ast.json` — 분기 5, return 6, 외부 호출 8
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `plan.raisesExposure` | bool | 호출부가 만든 `mutationPlan` | false면 **무조건 통과** — 노출을 줄이는 주문은 보호 여부를 묻지 않는다 |
| `g.protectionCheckForTest` | nil(프로덕션) 또는 test seam | `export_test.go`만 대입. 프로덕션 setter 없음 | non-nil이면 이후 모든 검사를 **우회**한다 |
| `g.protectionReadiness` | non-nil 필요 | Gateway 생성자 | nil이면 fail-closed 거부 |
| `plan.quantity` | canonical positive integral | `canonicalProtectionQuantity` | 아니면 fail-closed 거부 |
| `previous.adapter` | 직전 checkpoint | 호출부가 보관 | `Check`에 그대로 전달 |

**불변식**: 이 함수는 노출을 **늘리는** mutation만 막는다. 줄이는 mutation은 B1에서 즉시 통과한다.
이 비대칭이 안전 불변식 §0-3(청산 즉시성 불약화)의 구현이다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error |
|---|---|---|---|
| B1 (L90) | `!plan.raisesExposure` | 없음 | `(zero, nil)` — 통과 |
| B2 (L93) | `g.protectionCheckForTest != nil` | 없음 | seam 위임. **프로덕션에서 nil** |
| B3 (L96) | `g.protectionReadiness == nil` | 없음 | `protectionNotWired(plan, "market readiness provider is missing")` |
| B4 (L100) | `!ok` (canonical quantity 실패) | 없음 | `protectionNotWired(plan, "…canonical positive integral quantity")` |
| B5 (L106) | `refusal != nil` | 없음 | `protectionNotWired(plan, refusal.Error())` |
| fall-through (L109) | 위 전부 미해당 | 없음 | `(protectionCheckpoint{adapter: checkpoint}, nil)` |

**side effect 없음** — 이 함수는 순수 판정이다. AST의 `assignments`가 quantity/checkpoint 지역변수뿐이다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry 계약 | Evidence |
|---|---|---|---|
| `g.protectionCheckForTest` | test seam | 프로덕션 nil (`protection.go` 주석: "no production setter") | AST L94 |
| `protectionNotWired` | 거부 생성 | 순수 | AST L97/101/107 |
| `canonicalProtectionQuantity` | float64 → uint64 정규화 | 순수, `(uint64, bool)` | AST L99 |
| `g.protectionReadiness.Check` | **실제 판정 주체** | `*RejectedError` 반환. retry 없음 | AST L103 |
| `g.clk.Now` | 판정 시각 | 주입 clock | AST L105 |

## State mutations and fallbacks

- 상태 변경 없음. fallback 없음. 모든 실패는 fail-closed 거부다.
- **a100이 바꿀 지점**: `g.protectionReadiness.Check`가 성공을 돌려줄 수 있는 실제 provider가
  현재 프로덕션에 없다. 이 함수 자체는 이미 올바르며, **채워야 할 것은 provider다.**

## Safety conclusion

- Safe edit boundary: **이 함수는 편집하지 않는 것이 목표다.** a100은 `protectionReadiness` 뒤의
  provider를 배선한다. 함수 본문 편집이 필요해지면 그 자체가 설계 재검토 신호다.
- High-risk impact: **yes** — 진입 허용의 최종 관문이다.
