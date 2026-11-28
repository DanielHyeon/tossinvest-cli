# a200 1.4 · 1.5 재현 — 색인

편집 전 코드(base `884be2a74` = origin/main `9738a0c32` + 등록 문서)의 **현재 동작**을 기록하는 시험. 이름의 `CurrentBehaviour` 는
구현 단계(2.2)가 W4 · W5 를 RED 로 뒤집는다는 표지. 전부 `t.TempDir()` 임시 원장, 합성 값, 운영 경로 접근 0.

| 원문 | 명령 | 결과 |
|---|---|---|
| `w4.txt` | `go test ./internal/journal -run 'TestA200' -count=1 -v` | 8 PASS(범위 개정 뒤 「조정을 사이에 둔 부분 체결」 추가 — design §2 의 대가, 구현 뒤에도 AF 로 남아야 함) |
| `cycles.txt` | `go test ./internal/reconcile -run 'TestA200' -count=1 -v` | 6 PASS |

## W4 원장 관측(본 사고 모양 — `w4.txt` 첫 시험)

엔진 진입 · 관리 포지션 31, 손절 발의 무장 + 매도 주문 `AttachExitIntent`, 수렴과 같은 모양의 조정(UNKNOWN 31→0) 커밋, 같은 초에 그 매도 31 종결 체결.

| 표 | 관측 |
|---|---|
| `reconcile_states` | `ATTRIBUTION_FAILED` 활성 1 — evidence `position projection refused transition X59 (SELL_ON_CLOSED) … [order o-stop, delta 31, cumulative 31]` |
| `positions` | CLOSED / 0, `closed_at` = 조정 시각 |
| `exit_events` | `OPENED · BASELINE_BREACH_EXIT · ADJUSTMENT_CLOSED · PROPOSAL_FILLED`, `COMPLETED` 0 |
| `exit_states` | completed=1, `pending_intent_id` 조정 직후 `i-stop` → 종결 체결 뒤 빈 값, taken 1 |
| `trade_outcomes` | 0 |
| `position_adjustments` | `UNKNOWN 31→0` 1 |
| 시각 | 조정 `created_at` = 체결 `committed_at` 같은 초 |
| 성과 탐침(엔진 진입) | 동결 함수가 이 원장에서 성공 — 값이 W0(조정 없이 체결 먼저) 동결 값과 같음(같은 초) |
| 성과 탐침(편입) | `post_adoption_disposal` 거절 |

## 실제 Tracker 창(`cycles.txt`)

| 창 | 남는 것(일치 비교 2~5 주기 뒤) |
|---|---|
| W4(조정 뒤 · Refresh 전 체결) | `ATTRIBUTION_FAILED` 활성, operator_only, 진입 막힘 |
| W5(Refresh 뒤 · persist 전 체결) | 그 주기 Observe 오류 `durable scope is already owned by cause ATTRIBUTION_FAILED`, 이후 AF 고착 |
| W6(QM persist 뒤 체결) | AF 행 없음, QM 이 첫 일치에서 `ADJUSTMENT_APPLIED` 로 풀림, 차단 0 |
| W2 · W3 · 부분 조정 | `out-of-scope.md` |

사전 증거 서술과 **다른 점**: W2 · W3 고착이 추론이 아니라 재현됨. 그리고 재매수 뒤 옛 체결이 새 인스턴스에 닿음(전제 (f) — `claim-check.md`).
