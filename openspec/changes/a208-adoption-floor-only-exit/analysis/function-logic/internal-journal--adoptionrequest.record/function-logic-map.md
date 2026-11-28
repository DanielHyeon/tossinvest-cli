# Function Logic Map: `AdoptionRequest.record`

- Source: `internal/journal/adoption.go` (357-426)
- Qualified: `AdoptionRequest.record`
- AST evidence: `ast.json` (source_sha256 `6adc78fcb71d…`) — AST branches 13 · returns 11
- Risk scan: `risk-pattern-report.md`(일치 0)
- 작성 시점: 편집 **전**(a208 F3 재작업, 2026-11-28, HEAD `468ec87ce` — base `baeddfe31` 와 이 함수 본문 동일)

**역할.** 편입 요청을 검증해 저장할 행을 만든다. 호출자 `Journal.AdoptPosition` 하나(CodeGraph 1.6.0 — record ← 1 caller).
정책 ID 검증(B5 · B6)이 편입 기록에 적을 수 있는 정책 ID 의 유일한 원장 쪽 문이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `r.PositionID` · Symbol · Market · ObservedAt | 비어 있지 않음 | 호출자(adoptOne) | ErrInvalidRequest(B2~B4) |
| `r.ExitPolicyID` | 빈 값(RATCHET) 또는 등록 공통 정책 ID | `exitpolicy.CommonPolicyByID` | 미등록 → ErrInvalidRequest(B6) |
| Quantity · ObservedPrice · SyntheticStop | 양수, 손절 < 관측가 | canonical 10 진 | ErrInvalidRequest(B7~B12) |
| 불변식 | 저장 전 t0 쌍이 쓸 수 있는 값 | 함수 주석 | — |

## Branches and early returns

| Branch | Condition | At | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `switch` 필수 필드 | 367:2 | — | — | — |
| B2 | PositionID 없음 | 368:2 | — | return 오류 (369:3) | 미측정 |
| B3 | Symbol/Market 없음 | 371:2 | — | return 오류 (372:3) | 미측정 |
| B4 | ObservedAt 없음 | 374:2 | — | return 오류 (375:3) | 미측정 |
| B5 | `if a.ExitPolicyID != ""` | 379:2 | — | — | a208 T1 · T3(정책 ID 있는 편입) |
| B6 | `if _, ok := CommonPolicyByID(...); !ok` | 380:3 | — | return 미등록 정책 (381:4) | 미측정(이름 특정 안 함) |
| B7 | 수량 정규화 실패 | 387:2 | — | return (388:3) | 미측정 |
| B8 | 수량 0 | 390:2 | — | return (391:3) | 미측정 |
| B9 | 관측가 실패 | 394:2 | — | return (395:3) | 미측정 |
| B10 | 합성 손절 실패 | 397:2 | — | return (398:3) | 미측정 |
| B11 | 비교 실패 | 405:2 | — | return (406:3) | 미측정 |
| B12 | 손절 ≥ 관측가 | 409:2 | — | return (410:3) | 미측정 |
| B13 | `if a.CostBasis != ""` | 416:2 | CostBasisSource | — | 미측정 |

Exact AST return positions: 369:3, 372:3, 375:3, 381:4, 388:3, 391:3, 395:3, 398:3, 406:3, 410:3, 425:2

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `exitpolicy.CommonPolicyByID` | 정책 ID 등록 여부 | ok 불리언 | AST |
| `canonicalQuantity` · `positivePrice` · `riskcalc.CompareDecimal` | 수 검증 | ErrInvalidRequest | AST |
| `adoptionDigest` | 프리이미지 digest(정책 ID 포함) | — | AST |

Live config binding: 없음.

## State mutations and fallbacks

- 반환 값만 만듦(DB 쓰기는 호출자). fallback 없음.

## Safety conclusion

- Safe edit boundary(a208 F3): B6 의 조건에 `&& !exitpolicy.IsAdoptedFloorOnlyPolicyID(a.ExitPolicyID)` 를 더함 — 편입 보호선 사본 ID 하나만 더 받음.
  분기 수 무변. 다른 미등록 ID 는 지금처럼 거부. 사본 ID 가 편입이 아닌 exit state 로 열리는 길은 seedPolicyIdentity(→ LegacyLadderPolicyIdentity 의 adopted 검사)가 막음.
- High-risk impact: yes(편입 기록의 정책 ID = 익절 · 보호선 정책).
