# a201 1.4 — 실제 `RunOnce` 재현 (편집 전 코드의 현재 동작)

- HEAD `fffb03d67`(base `f9b818a0e` + base 기록), go1.26.5. production Go 편집 0.
- 시험: `internal/app/engine/a201_stale_quantity_comparison_repro_test.go`(새 파일, 6 PASS). 원문 `current-behaviour.txt`.
- 명령: `go test ./internal/app/engine -run 'TestA201CurrentBehaviour' -count=1 -timeout 30m -v`.

## 끼워 넣은 방법 — production 편집 없음

`ReconcileDriver.RunOnce` 를 그대로 돌림(수집 2회 · 안정화 · `LocalStateFromJournal` · `Compare` · Ingest · 수렴 · Refresh ·
Observe · 편입 판정). 체결 커밋 지점은 **이미 주입 가능한 인터페이스 필드 둘**을 감싸서 만듦.

| 지점 | 감싼 필드 | 끼우는 순간 |
|---|---|---|
| W2 | `Converger.Journal`(`reconcile.ConvergeStore`, converge.go:118) | `Positions` 위임 **직전** — 로컬 읽기(reconcileloop.go:434) 뒤 · 수렴의 `Positions`(converge.go:193) 전 |
| W3-a | 같은 필드 | `Positions` 반환 **직후** — `FillWatermark`(converge.go:205) 전 |
| tx 직전 | 같은 필드 | `FillWatermark` 반환 **직후** — 조정 tx(converge.go:209) 전 |
| 수렴 뒤 · Observe 전 | `Tracker.Journal`(`reconcile.ReconcileStore`, mismatch.go:297) | Refresh 의 `ActiveReconcileStates`(mismatch.go:1027) |

조립은 운영 리터럴과 같은 필드(apply 훅 `Project`·`Campaign`·`Exit`·`Costs` — gateway.go:201-214, Tracker `Terminal` 포함 —
gateway.go:271-278, Converger `Credit` — gateway.go:369-371). 다른 점: 진입 게이트 신선도 요구를 비움(대사 차단만 보이게),
Converger `Alert` 없음, 운영 `Period` 대신 주기마다 가짜 시계 60초 전진.

## 관측

| 시험 | 그 주기 | 이후 깨끗한 주기 5번 | 비고 |
|---|---|---|---|
| W2 (`Positions` 전 체결) | 오류 없음, converged 0, QM 행 1 활성, 진입 거절 | 미해제(awaiting 1), 진입 거절 | 재매수(계좌 5 · 로컬 0) 주기: folded 1 · adopted 0 — **편입 거절**, QM 유지 |
| W3-a (`Positions` 뒤 체결) | `cycle.Err` = 수량 재검증 stale(「expected quantity 31 but found 0」), QM 1 | 미해제 | |
| tx 직전 체결(`FillWatermark` 뒤) | **여전히 수량 재검증 (a) stale** — 워터마크 (b) 아님 | 미해제 | 수량을 움직이는 체결은 어느 지점이든 (a)가 먼저 잡음(재검증 순서 B15@253 → B17@267) |
| W3-b (워터마크만 stale) | tx 직전에 **다른 계좌의 같은 종목 체결**(수량 불변 fill_events 행)로 (b) stale(「expected fill watermark 1 but found 2」), 엔진 체결은 Refresh 때 커밋 → QM 1 | 미해제 | (b) 단독 stale 은 수량 불변 행이 있어야만 생김 — 계좌 조건 없는 종목 워터마크 |
| 앞 주기 크레딧 위의 W2 | 외부 매도 주기: 31→20 조정 · 크레딧 · QM. 다음 W2 주기: 낡은 비교가 그 종목을 분쟁으로 봐 크레딧 반박 | 미해제 | 이미 선 QM 도 같은 경로로 영구화 |
| 판 중단 뒤 종목(설계 범위 밖) | 종목 A W3-a stale → 판 중단(converge.go:228-229) → B 미수렴, A·B QM 2 | B 체결을 **그 주기 뒤**에 커밋 → A·B 모두 미해제 | B 의 체결은 재판정보다 뒤 — a201 의 재판정이 못 잡는 모양 |

## 사전 증거와 다른 점

1. 사전 증거(a200 손밟기)는 W3 을 (a)로만 재현했고 (b) 단독은 재현 못 했음 — 여기서 (b) 단독 stale 을 만드는 입력을 찾음:
   **조정 tx 직전의 수량 불변 fill_events 행**(다른 계좌의 같은 종목 체결 · 투영이 거절한 체결 류). 엔진 매도 체결 자신은
   어느 지점에 와도 (a)로 걸림. 그래서 (b) 단독 stale 뒤의 고착은 「그 뒤 다른 시점에 커밋된 엔진 체결」과의 합성임.
2. 판 중단(G1) 뒤 종목의 체결이 Observe **뒤**에 오면 같은 고착 — 재판정 시점보다 늦은 체결이라 design D1 의 자리로는 못 잡음.
   design §4 「잔여는 재판정~Observe 의 메모리 구간」보다 넓음(claim-check C-R1).
3. 앞 주기에 이미 선 QM 도 W2 로 크레딧을 잃고 영구화함 — a201 이 그 항목을 빼면 그 주기 비교가 깨끗해져 **앞 크레딧으로 해제**가
   일어날 수 있음(1.5 (e) 탐침).
4. 시작 주기의 adopted=1 은 픽스처의 엔진 진입에 exit state 가 없어 편입 판정이 편입한 것 — 재현 결론과 무관.
