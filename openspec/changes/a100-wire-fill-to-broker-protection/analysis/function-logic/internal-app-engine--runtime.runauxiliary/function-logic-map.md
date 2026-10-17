# Function Logic Map: `Runtime.runAuxiliary`

> **a100 R0 재동결(2026-10-04) — 채택 번들.** a100 편집 전 번들(base 882a0b49 기준)은 분기 구조가 바뀌어 낡았다. 이 메서드를 마지막으로 바꾼 a092(아카이브 2026-09-30)의 번들을 그대로 가져온다 — 현재 소스와 SHA-256 일치(7d941fa5…, 90-121). 분기 · 시험 인용은 a092 편집 뒤 기준이다. a100 의 T 로트가 이 메서드를 편집할 때 이 번들 위에 Branch Test Map 을 덧붙인다.

- Source: `internal/app/engine/auxiliary.go`
- AST evidence: `ast.json` — **편집 뒤**, :90–121, 분기 3 · 반환 2 · 호출 5, source_sha256 `7d941fa53886…`, 추출 커밋 `e55102f0`. 편집 전 번들은 `analysis/pre-edit/unit5/`(없으면 단위 ④ 번들이 편집 전).
- Risk scan: `risk-pattern-report.md`
- 편집(착지 단위 ⑤ `e55102f0`): 정지 로그 이벤트를 `aux.StopEvent`(빈 값이면 `EventAlertUndelivered`)로 — 분기 하나 추가(K13).
- 재번호: difflib 정렬(편집 전 `b01e0cd0` 소스 대비): B1~B1 → B1~B1(같은 분기, 줄 이동); 새 분기 B2(:110); B2~B2 → B3~B3(같은 분기, 줄 이동).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 | 편집 전 계약 + 위 편집 | 호출자 | 아래 분기 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if r.gracefulStop(ctx, err)` (:92) | — | — | `TestEveryAuxiliaryExecutorIsStarted`, `TestTheEngineStopsPromptlyWithTheDelivererAttached` |
| B2 | `if event == ""` (:110) | — | — | `TestA092TheDelivererStopKeepsItsEvent`, `TestADeadAuxiliaryExecutorIsNotRestarted` |
| B3 | `if aux.OnStop == nil` (:117) | — | — | `TestA092TheDelivererStopKeepsItsEvent`, `TestA092TheRelayStopIsNotAnUndeliveredCriticalAlert` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `aux.OnStop`, `describe`, `r.gracefulStop`, `r.log`, `runAuxiliaryBody` | 편집 뒤 호출 | 위 편집 참조 | AST |

## State mutations and fallbacks

- 위 편집 요약이 전부 — 나머지는 편집 전과 같음(편집 전 번들 대조).

## Safety conclusion

- Safe edit boundary: 위 편집만.
- High-risk impact: 중간 — 보조 실행자 정지 기록.
