# a091 freeze 재리뷰 4라운드 — Claude 보이스 A (적대 Eng) 원문 요지

- 2026-10-01, 같은 에이전트 재개, 스냅숏 `/tmp/claude-1000/a091-r4-tree`(= `8870c6c9`) read-only.

VERDICT: APPROVE — P0 · P1 없음. P2 둘은 구현 착수 전 수정 또는 명명.

(1) 3라운드 미결: ③ 동치 CLOSED(D3:93-104 · 3.3a 칸, 머리 「⟺」 잔존 nit) · #6 루프 몫 PARTIAL(P2 — 5.3 의 Acknowledge 칸에 밀린 행 수 N 미고정; 출처 과장 — 아래 3) ·
#8 계좌 CLOSED(`o.alert` 우회 · `escalate` `:433` · `:440` 필드 제거 · 번들 · 호출자 `:238` · `:261` · `record_only.go:157` · 카나리; 배달 실행자 자기 `escalate`
`alertdelivery.go:466-495` 는 이미 계좌 없음) · P3 셋 CLOSED.
(2) 취소 규칙 건전: `Runtime.Run` 이 `loopCtx` 취소(`runtime.go:345`) 뒤 `wg.Wait()`(`:347`), exit 는 감독 루프(`cmd/tossctl/engine.go:720-726`), 원장은 `Run` 뒤
`ectx.Close()`(`:224` → `engine.go:618`) — `WithoutCancel` 기록은 원장이 닫히기 전에 끝난다. 억제 출처: HTTP 중 취소는 `*url.Error` 가 `context.Canceled` 를 감쌈,
`localOpenSells` 는 `%w` 감쌈, HTTP 시한은 `DeadlineExceeded` → ①. P3: 재시도 대기 중 취소는 직전 일시 오류(`retry.go:383-385`) → ① 행(재시작 뒤 배달) — 명명,
`WithoutCancel` 기록은 기한 없음(`n.mu` · 풀) — 「로컬 트랜잭션만큼」에 「기한 없음」 추가.
(3) 750ms 대입: 수치 충실(a092 `design.md:1094-1098`), 출처 과장 P2 — (a) `alertLoopShare` 는 a092 21판에서 철회 · 미착지, (b) 356.1ms 탐침은 쓰기 ·
`EnqueueAlert` 경합(`design.md:1685-1692`)이고 `Acknowledge` 의 `n.mu` 보유 백로그는 재지 않았다 — 5.3 멈춤 규칙이 실제 방벽.
(4) P3: D5 「`o.alert` → RecordOnly 를 탄다」 와 D8 「`o.alert` 우회」 모순 — 진입점 명시(`o.opts.Alerts.Notify(context.WithoutCancel(ctx), e)`; `RecordCritical(…,0)` 로
구현하면 1h 재무장 상실) · `runtime.go:344` → `:345` · D1 「뿐」 에 `escalate` 필드 제거의 꺼짐 엔진 로그 변화 누락 · 운영 로그 수치 미검증(스냅숏 밖).
