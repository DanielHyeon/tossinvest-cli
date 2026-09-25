# Function Logic Map: `alertDeliverer.cycle`

- Source: `internal/app/engine/alertdelivery.go`
- AST evidence: `ast.json` — **편집 전**(base 논리), :145–164, 분기 3 · 반환 3 · 호출 7.
  source_sha256 `89a491c9c259…`, 추출 HEAD `463cc895` (2026-09-25).
- Risk scan: `risk-pattern-report.md`

**이 번들은 proposal 이 이 함수의 분기를 근거로 쓰기 때문에 문서보다 먼저 만들었다**
(FLM-before-claiming). a124 는 이 함수를 편집한다 — 편집 뒤 `revision: current` 로 재추출한다(tasks 1.3).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx` | 취소 가능 | `Run` :122 이 넘긴다 | B3 |
| `d.Journal` | non-nil | `NewRuntime` 배선 | nil 이면 패닉 — 배선이 보장, 이 함수는 검사하지 않음 |
| `d.batch()` | `Batch > 0` 이면 그 값, 아니면 `alertDeliveryBatch` = 10 (:74) | :150 | — |
| `d.heldReported` | 현재 경합 중인 행만 | `forgetLapsedHeld` :149 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `PendingAlerts` 오류 (:151) | 없음 | `fmt.Errorf("listing the critical alert backlog: %w")` (:152) — `Run` :124-128 이 로그만 남기고 계속 돈다 | `a098_the_outbox_gets_emptied_test.go` (파일 단위) |
| B2 | `range pending` (:154) | 행마다 `deliverOne` (:161) | — | `a098_one_cycle_takes_a_batch_test.go` |
| B3 | 행 사이 `ctx.Err() != nil` (:158) | 남은 행 미처리 | `nil` (:159) | a098 취소 전파 시험 |
| 종단 | 배치 소진 | — | `nil` (:163) | — |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `d.forgetLapsedHeld(d.Clock.Now())` :149 | 만료된 경합 기록 제거 | 없음 | AST |
| `d.Journal.PendingAlerts(ctx, d.batch())` :150 | 미전달 행 나열 — `WHERE state = ? ORDER BY id` + `LIMIT ?` (`internal/journal/outbox.go:518-521`) | 오류 → B1 | AST + 소스 원문 |
| `d.deliverOne(ctx, alert)` :161 | 행 하나의 전달 | **반환값 없음** — 결과가 이 함수에 돌아오지 않는다 | AST |

## State mutations and fallbacks

- 이 함수는 `deliverOne` 의 결과를 받지 않는다. 배치 전부가 실패한 사이클과 전부 성공한 사이클이
  같은 `nil` (:163) 로 끝난다 — **지속 실패를 셀 자리가 이 함수에 없다** (a124 R1 의 근거).
- 행 선택은 `id` 오름차순 `LIMIT batch` 뿐이다. 시도를 다 쓴 오래된 미전달 행이 batch 개 이상 쌓이면
  새 행은 이 함수에 **도달하지 않는다** (a124 R2 의 근거, a092 20라운드 B-2).
- `Run` :122 는 사이클 오류를 로그로만 다루고 루프를 계속 돈다 — 실행자 정지 판정은 `Run` 의 반환값이고
  이 함수의 `nil` 은 그것을 절대 만들지 않는다.
