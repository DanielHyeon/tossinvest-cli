# a163 증거 조정 (SDD 5단계)

| 질문 | CodeGraph 1.6.0 | CGC | AST / HEAD 원문 | 결론 |
| --- | --- | --- | --- | --- |
| `adaptPrices` 의 호출자 | `Client.Prices` 하나 | 다른 사본만 | `official--client.prices.ast.json` `:149` | 일치 — 어댑터 편집의 생산 도달은 `Prices` 경유뿐 |
| `adaptPrices` 가 브로커 시각을 읽나 | callees `parseDecimal` 만 | — | AST 호출 텍스트에 `Timestamp` 없음, `:175 time.Now` | 일치 — 읽지 않는다 |
| `quoteUsable` 호출자 | 6 | 줄 번호 다른 사본 | 6 자리 모두 AST 분기 좌표로 확인(`quote-time-evidence.md` C3) | 일치 |
| `observe` 호출자·피호출자 | 진짜 1·5 + 거짓 엣지 3·4 | — | AST `calls` 에 거짓 엣지 대상 없음 | 거짓 엣지 제외 후 일치 |
| `stableObservationID` 가 `FetchedAt` 을 쓰나 | 호출자 `snapshotContext` | — | AST `:1208` `FetchedAt.UTC` · B2 `:1209` | 일치 — D2 의 근거 |

불일치 0(거짓 엣지는 도구 한계로 분류, 편집 차단 사유 아님). CGC 는 이 체크아웃을 보지 못했다(advisory 미반영).
편집 착수 시(tasks 1.x) `make sdd-sync` 뒤 같은 질의를 다시 돌려 이 표를 갱신한다.

## tasks 0.7 재질의 — 2026-10-22, 격리 worktree `/tmp/tossos-a163-impl-20261022`(HEAD `1eab9b931`)

- `make sdd-sync` rc=0(`[sdd-sync] all indexes current`, 5m00s). 이 worktree 에는 `.codegraph/` 가 없어 `codegraph init .`(1.6.0)로 새로 색인.
  GBrain 은 advisory 단계에서 `Sync BLOCKED at 1eab9b93: 12512 file(s) failed to parse`(88 pages 만 import) — advisory 라 편집 차단 사유 아님,
  CGC 는 `codegraphcontext index .` 실행. 원문 질의 결과: `codegraph-callers-0.7-20261022.txt`.
- 재질의 결과(비시험 호출자만, 이름 충돌 거짓 엣지 제외):

| 심볼 | 비시험 callers | HEAD 원문 대조 | 결론 |
| --- | --- | --- | --- |
| `official.adaptPrices` | `Client.Prices` (`market_reads.go:142`) | 일치 | 0.6 과 같음 |
| `ExitObserver.observe` | `ObserveOnce` (`exitloop.go:457`) + 거짓 엣지(`filldetect.pollLocked` 등) | AST `calls` 에 없음 | 0.6 과 같음 |
| `ExitObserver.ObserveOnce` | 시험만 표시(상한 20) — 생산 `Run` `:400` · `Tracer.Run` 은 HEAD grep 으로 확인 | 일치 | 단일 goroutine 전제 유지(design D4) |
| `quoteUsable` | `ObserveOnce` · `judge` · `judgeRatchet` · `judgeLadder` · `refreshObservation` · `record` | 일치 | 렌더 해시 대상 |
| `stableObservationID` | `snapshotContext` `:1190` | **자유 함수**(수신자 없음) | design D7 의 「수신자 ExitObserver」 표기와 다름 — issues.md I1, 시험은 수신자 없음으로 고름 |
| `config.mergeEngine` (D8) | `service.load` `service.go:240` | 일치 | 생산 호출자 1 |
| `engine.recordGateSettings` (D8) | `NewContext`(`engine.go:472`) + 시험 4(`a074_notification_audit_test.go`) | 일치 | design D8 과 같음 — **단 `NewContext` 와 시험 함수 넷이 편집 대상이 되어 FLM 번들 5 를 더함**(tasks 1.1 목록 밖 — 아래) |
| `engine.Context.ExitObserver` (D8) | `cmd/tossctl.engineRuntime` `engine.go:655` | 일치 | — |

- `codegraph affected <file> --filter '*_test.go'` 건수: `exitloop.go` 1198 · `market_reads.go` 1192 · `config/engine.go` 1198 · `interlock.go` 1173 ·
  `exitwiring.go` 1137 · `cmd/tossctl/engine.go` 1143 · `domain/models.go` 1197. 전이 범위가 넓어 시험 선택 근거로 쓰지 않음(tasks 4.2 와 같음).
- 불일치: 둘(issues.md I1 · I2) — 둘 다 시험 범위 문제이고 편집 대상 · 판정 경로 사실과 충돌하지 않아 편집 차단 사유 아님(Manager 확인 요청).
