# Function Logic Map: `ReconcileDriver.adoptOne`

- Source: `internal/app/engine/adoption.go` (373-432)
- Qualified: `ReconcileDriver.adoptOne`
- AST evidence: `ast.json` (source_sha256 `7257fb1da221…`) — AST branches 3 · returns 3
- **편집 뒤 갱신(a208 3.3, 2026-11-28)**: `ast.json` 을 편집 뒤 소스로 재추출(`revision: current`), 좌표는 옛→새 AST 노드 순서 대응으로 옮김(분기 · 반환 개수 무변, 새로 생긴 분기 없음). 아래 「Safe edit boundary」는 편집 전 계획이며 그대로 구현됨.
- **F3 재편집(2026-11-28)**: 알림 본문 호출 인자 한 줄 `adoptionAlertBody(d.opts.Adoption.FloorOnlyExit)` → `adoptionAlertBody(d.opts)`(분기 · 반환 · 좌표 무변, `ast.json` 재추출). 사본 적용 여부를 토글 값이 아니라 「토글 ∧ 공통 정책에 사본 있음」으로 판정하기 위함(새 파일 함수 `adoptionFloorOnlyPolicy`).
- Risk scan: `risk-pattern-report.md`(설정 규칙 일치 0)
- 작성 시점: 편집 **전**(a208 tasks 1.1, 2026-11-28, HEAD `810aec6f7` — base `baeddfe31` 와 본문 동일).

**역할.** 편입 한 건을 영속하고 exit state 를 연다. 편입 기록의 정책 ID 를 정하는 **유일한** 자리(claims C1). 호출자 `ReconcileDriver.adopt`(adoption.go:215) 하나(CodeGraph 1.6.0 callers).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `c candidate` | 후보 판정 통과(신선 · 대사 · include/exclude) | `adopt` | — |
| `observed` | 편입 직전 시세(15초 이내) | `adopt` 의 가격 읽기 | — |
| `d.opts.Adoption` | 병합된 `config.Adoption`(거부 블록은 전면 0) | `Context.ReconcileDriver` reconcileloop.go:383 | — |
| `d.opts.CommonPolicy` | 빈 값 또는 등록된 공통 정책 ID | `NewReconcileDriver` 검증 reconcileloop.go:337-341 | 미등록 ID 는 드라이버 생성 거절 |
| 불변식 | 편입 기록 영속이 exit state 열기보다 먼저(크래시 계약) | 이 함수 주석 | — |

## Branches and early returns

| Branch | Condition | At | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `if err != nil` — `SyntheticStop` 유도 실패 | 375:2 | `logDeferred` | return false (377:3) | 기존 편입 시험군(이름 미측정) |
| B2 | `if err != nil` — `AdoptPosition` 거절 | 394:2 | `logDeferred` | return false (396:3) | 기존 편입 시험군(이름 미측정) |
| B3 | `if _, err := OpenAdoptedExitState(...); err != nil` | 399:2 | `logDeferred`, 계속 진행(알림은 보냄) | — | 미측정(크래시 계약 자체는 `TestCrashBetweenAdoptionAndExitStateIsCompleted` 가 원장 수준에서 잼 — 이 분기를 지나지 않음) |

Exact AST return positions: 377:3, 396:3, 431:2

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `exitpolicy.SyntheticStop` | 합성 손절 = 관측가 × (1 − pct) | 오류 → B1 | AST |
| `Journal.AdoptPosition` | 편입 기록 영속(정책 ID 포함) | 오류 → B2, 정책 ID 는 등록 공통 정책인지만 검사(journal/adoption.go:379-382) | AST · HEAD |
| `journal.RFC3339` · `d.clk.Now` | 관측 시각 | — | AST |
| `Journal.OpenAdoptedExitState` | exit state 열기 — 정책 ID 로 RATCHET/LADDER 결정(C2), 정체성 고정(C3) | 오류 → B3(기록은 남고 다음 주기가 완결) | AST · HEAD |
| `d.alert` · `d.label` | 편입 알림(본문은 리터럴) | 알림 실패는 판정에 영향 없음 | AST |
| `delete(d.unmanaged, …)` | 무관리 래치 해제 | — | AST |

Live config binding: `d.opts.CommonPolicy`(기동 때 로드된 `engine.exit_policy.common_policy`), `d.opts.Adoption.DefaultStopPct`. a208 이후 `d.opts.Adoption.FloorOnlyExit` 도 같은 기동 스냅숏.

## State mutations and fallbacks

- 원장: 편입 기록 1 행 + exit state 1 행(B3 실패 시 0). 메모리: `d.unmanaged` 래치 삭제.
- fallback: B3 실패는 다음 주기 · 관측기 working set 이 완결(크래시 계약).

## Safety conclusion

- Safe edit boundary(a208): `ExitPolicyID:` 값 한 자리와 알림 `Body:` 값 한 자리를 **새 파일의 함수 호출**로 바꾼다(새 분기 0 — 분기 재번호 없음). 손절 유도(B1) · 순서 · 반환은 무편집.
- 토글 거짓이면 두 값이 a208 이전과 같다(T3 바이트 단언).
- High-risk impact: yes — 편입 포지션의 익절 · 보호선 정책을 정한다. 손절 기준(합성 손절) 무변.
