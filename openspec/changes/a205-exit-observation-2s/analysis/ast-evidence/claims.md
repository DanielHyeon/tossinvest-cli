# a205 분기 주장의 AST 근거

design 이 근거로 쓰는 **기존 함수 내부 분기 · 호출**은 이 디렉터리의 `tools/logic-map` 산출물로 열거함(.claude/CLAUDE.md 필수 진입 5).
여기 함수들은 이 change 가 **편집하지 않는다** — FLM 번들이 아니라 주장의 근거라서 `analysis/function-logic/` 이 아니라 이 자리에 둠
(선례 a140 · a203 `analysis/ast-evidence/`).

- 기준: base `e2921fc00`(origin/main), go1.26, 2026-11-27. 편집 전 소스.
- 생성: `GOFLAGS=-trimpath go run ./tools/logic-map --file <파일> --func <함수> > engine--<함수>.ast.json`
- 각 파일의 `source_sha256` 이 그 시점 소스 지문임. 이 change 는 `exitloop.go` 의 상수 · 주석만 고치므로 GREEN 뒤 `exitloop.go` 지문은 달라지고
  `exit_lower_limit.go` · `tracer.go` 지문은 그대로임.

| 산출물 | 파일:줄(base) | source_sha256(앞 12) | 분기 | return |
|---|---|---|---|---|
| `engine--ExitObserver.Run` | internal/app/engine/exitloop.go:428-439 | 2ed5bf8fceff | 3 (B1 for · B2 if · B3 if) | 2 (:431 · :436) |
| `engine--ExitObserver.Interval` | exitloop.go:401-406 | 2ed5bf8fceff | 1 | 2 |
| `engine--exitLowerLimitBoard.cycleComplete` | internal/app/engine/exit_lower_limit.go:142-150 | 9f676675fcda | 2 (B1 if · B2 select) | 1 |
| `engine--exitLowerLimitBoard.Run` | exit_lower_limit.go:166-183 | 9f676675fcda | 4 | 2 |
| `engine--exitLowerLimitBoard.refresh` | exit_lower_limit.go:185-242 | 9f676675fcda | 13 | 7 |
| `engine--lowerLimitReadTimeout` | exit_lower_limit.go:244-253 | 9f676675fcda | 2 | 2 |
| `engine--Tracer.Run` | internal/app/engine/tracer.go:276-325 | 181d2038475c | 12 | 7 |

## 주장

- **C1 — 관측 주기는 `Run` 이 `Interval()` 로 두 번 읽는 값 하나다.** `ExitObserver.Run` 호출 열거: `o.reportCycle(o.ObserveOnce)` :433 →
  `o.opts.lowerLimitBoard.cycleComplete(o.Interval())` :434 → `o.clk.Sleep(ctx, o.Interval())` :435(B3 — 오류면 :436 반환). `go_statements` 0 ·
  `defers` 0. `Interval` 은 B1(`opts.Interval > 0`) 아니면 `DefaultExitObservationInterval`(:405). 생산 배선(`cmd/tossctl/engine.go:675`)은
  `Interval` 을 넘기지 않음(`rg -n "Interval:" cmd/tossctl/engine.go` 의 ExitObserverOptions 블록에 0) → 생산 주기 = 상수.
- **C2 — 하한가 갱신은 손절 판정과 같은 goroutine 에 있지 않다.** `Run` 에서 하한가 쪽 호출은 `cycleComplete` 하나이고, `cycleComplete` 의
  B2 는 `select`(기본 갈래가 있는 보내기 — 채널 용량 1, `newExitLowerLimitBoard` `tokens: make(chan time.Duration, 1)`)라 막히지 않음.
  GET 은 `exitLowerLimitBoard.Run`(보조 실행자 `ExitLowerLimitRefresher`, `cmd/tossctl/engine.go` `Auxiliary`)의 B3 `select` 가 토큰을 받아
  `b.refresh` :180 에서 함. 그러므로 주기를 바꿔도 갱신기의 시간이 손절 판정을 늦추지 않음(안전 불변식 4).
- **C3 — 갱신 한 번의 시간 상한은 주기에서 유도된다.** `refresh` :186 `lowerLimitReadTimeout(interval)` → :187 `prefetchLimit := interval - readLimit`,
  B1 :188(≤0 이면 readLimit) → :191 사전조회 context. 브로커 읽기 :221 `context.WithTimeout(ctx, readLimit)` → :222 `ProtectivePriceLimits`.
  `lowerLimitReadTimeout` B1 :246(`interval > 0 && interval/2 < 2s` → interval/2), B2 :249(≤0 → 1ms). 갱신 한 번은 GET 을 최대 1건 하고
  반환(:240 publish 뒤 return, 실패 갈래 :215 · :225 · :229). 주기 2s 면 읽기 1s · 사전조회 1s, 합 2s = 주기. 5s 면 2s · 3s, 합 5s = 주기.
- **C4 — tracer 는 같은 `Interval()` 을 잔다.** `Tracer.Run` :315 `t.opts.Clock.Sleep(ctx, t.opts.Observer.Interval())`. tracer 는 생산 호출자가
  없음(`rg -n "NewTracer|Tracer\{|&Tracer" --glob '!*_test.go' cmd internal` → `tracer.go` 의 정의 2줄뿐 — 시험 전용). `MaxCycles` · `MaxDuration` 둘 다 호출자 값이라 주기가 짧아지면
  같은 `MaxCycles` 의 벽시계 시간이 줄 뿐 판정은 같음.

## 확인 안 한 것

- `refresh` 의 사전조회 context 가 반복문 안 `HasProtectiveExitAttempt` 에도 쓰인다는 것(:213 `prefetchCtx`)은 호출 열거로 확인, 실제 지연 분포는 미측정.
- `/price-limits` 단건 GET 지연 분포는 운영 미측정(design D2 남은 위험).
