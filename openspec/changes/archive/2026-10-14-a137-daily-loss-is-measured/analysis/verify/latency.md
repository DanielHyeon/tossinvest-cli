# a137 task 2.4 — 일일 손실 원장 읽기 지연

- 사본: 연결 워크트리 `bd723b7f`(scratchpad `wt-latency`), 전용 `GOCACHE`, `-tags tossos_testseams`. 측정 시험은 사본에만 둔 `internal/journal/a137_latency_measure_test.go`(이 디렉터리에 사본 보관) — 커밋하지 않음.
- 측정 순간: 2026-10-15 00:38 KST, 20 코어, load average 5.0~6.5(병행 세션 부하 — a136 게이트 시작 직전; **상한 쪽으로 치우칠 수 있는 값**).
- 방법: 한 계좌에 CLOSED 관리 포지션 n 개(진입 결정 · 성과 행 포함, 과거로 90분 간격, 마지막 20 개만 오늘 일 키 창 안)를 한 트랜잭션으로 심고,
  각 읽기를 30 회 재서 p50 · p95 · max. 실 원장은 읽지 않음.

## 픽스처 크기를 고른 근거(사람이 확인할 수 있는 수에서 유도)

- 하루 진입 상한: 하루 매수 총액 기본 10,000,000 KRW(a136) ÷ 주문당 notional 상한 1,000,000 KRW = KR 10 건 + US(USD 축 1,500/500) ≈ 3 건 → 하루 청산 ≲ 20 건(부분 청산 · 재진입 여유 포함).
- 1년 ≈ 250 거래일 → ≲ 5,000 CLOSED 포지션/년. `positions` 는 보존 정리가 없다(성과 표만 180 일) → 4 년 ≈ 20,000, 스트레스 50,000.
- 사람이 실제 값을 확인하는 질의(원장 **사본**에 `sqlite3 -readonly`): `SELECT count(*) FROM positions WHERE state = 'CLOSED';`

## 결과

| CLOSED 관리 포지션 n | countMalformedClosedAt(전수 스캔) p50 / p95 | dailyLossCandidates p50 / p95 | DailyLossRows 전체 p50 / p95 / max |
|---|---|---|---|
| 1,000 | 0.66 ms / 0.69 ms | 0.71 ms / 0.79 ms | 1.35 ms / 1.37 ms / 1.39 ms |
| 5,000 | 3.4 ms / 3.5 ms | 2.3 ms / 2.3 ms | 5.6 ms / 6.1 ms / 7.7 ms |
| 20,000 | 15.4 ms / 15.9 ms | 9.8 ms / 13.4 ms | 24.6 ms / 25.9 ms / 26.4 ms |
| 50,000 | 41.7 ms / 57.0 ms | 27.5 ms / 28.6 ms | 69.0 ms / 70.3 ms / 70.9 ms |

- 둘 다 계좌의 CLOSED 행 수에 선형(행당 ≈ 0.7~1.4 µs). `closed_at` 색인이 없어 `dailyLossCandidates` 도 계좌 범위를 훑음(account_ref 선두 색인 + state 거름).
- 단일 연결 점유: 측정 한 번이 원장 연결을 위 시간만큼 잡음(그동안 체결 반영 쓰기가 기다림). 측정은 진입 후보가 첫 다리 승인에 닿을 때만 일어남(주기마다가 아님).
- 판단: 1년 규모(≈ 5,000)에서 6 ms, 4년(≈ 20,000)에서 26 ms — 손절 · 청산 지연 기준(초 단위)에 비해 작음. 50,000 을 넘기 전에 `positions(account_ref, state, closed_at)` 색인을 후속으로 검토(스키마 변경 = 별도 change).

## N+1 경로(재리뷰 P3-2) — 오늘 후보 수를 바꿔 재측정(같은 사본, 2026-10-15 01:01 KST, load 5.1~6.2)

`DailyLossRows` 는 계좌의 CLOSED 행을 positionSelect 로 한 번 읽고 Position.ExitEligible 로 거른 뒤, **오늘 창 안 후보마다** 성과 행 · 마지막 조정을 점 질의
하나로 읽는다(열 이름 단일 철자 가드 때문에 조인 대신 이 모양 — 4fc5d4a4). 위 표는 이미 이 경로다. 후보 수만 늘려 그 몫을 분리:

| n | 오늘 후보 | dailyLossCandidates p50 / p95 | DailyLossRows 전체 p50 / p95 / max |
|---|---|---|---|
| 5,000 | 20 | 2.3 ms / 4.2 ms | 5.4 ms / 5.5 ms / 5.6 ms |
| 5,000 | 200 | 4.8 ms / 5.6 ms | 8.1 ms / 8.8 ms / 9.0 ms |
| 20,000 | 20 | 9.5 ms / 13.8 ms | 23.1 ms / 24.7 ms / 30.1 ms |
| 20,000 | 200 | 13.4 ms / 20.1 ms | 31.2 ms / 42.5 ms / 44.7 ms |

- 후보 하나당 점 질의 몫 ≈ 14 µs((4.8 − 2.3) ms ÷ 180). 하루 후보 상한(≈ 20)에서는 N+1 몫이 ≈ 0.3 ms 로 전수 스캔보다 작다.
