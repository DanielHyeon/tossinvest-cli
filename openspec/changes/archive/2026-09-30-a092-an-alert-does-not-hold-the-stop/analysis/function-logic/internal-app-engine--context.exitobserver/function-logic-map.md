# Function Logic Map: `Context.ExitObserver`

- Source: `internal/app/engine/exitwiring.go`
- AST evidence: `ast.json` — **편집 뒤**, :319–356, 분기 6 · 반환 4 · 호출 6, source_sha256 `2b0696f0b0cc…`, 추출 커밋 `e55102f0`. 편집 전 번들은 `analysis/pre-edit/unit5/`(없으면 단위 ④ 번들이 편집 전).
- Risk scan: `risk-pattern-report.md`
- 편집(착지 단위 ⑤ `e55102f0`): `RecordOnly` 에 `Relay: c.NormalAlertRelay()`(C8) · `Alerts` · `Announcer` 를 호출자 값과 무관하게 덮음(B#4 — 두 조건 분기가 `c.Notifier != nil` 하나로).
- 재번호: difflib 정렬(편집 전 `b01e0cd0` 소스 대비): 교체 B5, B6 → B5; B7~B7 → B6~B6(같은 분기, 줄 이동).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 | 편집 전 계약 + 위 편집 | 호출자 | 아래 분기 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if c == nil` (:320) | — | — | (미실행) |
| B2 | `if !c.Automation.Verified` (:323) | — | — | `TestTheExitObserverIsUnavailableWithoutAVerifiedGate` |
| B3 | `if !ok` (:327) | — | — | (미실행) |
| B4 | `if opts.Names == nil` (:343) | — | — | `TestA092ExitObserverGetsRecordOnlyAlertPaths`, `TestA092TheExitObserverOverridesACallersSyncAlertPath` |
| B5 | `if c.Notifier != nil` (:348) | — | — | `TestA092ExitObserverGetsRecordOnlyAlertPaths`, `TestA092TheExitObserverOverridesACallersSyncAlertPath` |
| B6 | `if opts.Floor == nil` (:352) | — | — | `TestA092ExitObserverGetsRecordOnlyAlertPaths`, `TestA092TheExitObserverOverridesACallersSyncAlertPath` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `NewExitObserver`, `c.NormalAlertRelay`, `exitSideFloor`, `exitSideRetrier`, `fmt.Errorf` | 편집 뒤 호출 | 위 편집 참조 | AST |

## State mutations and fallbacks

- 위 편집 요약이 전부 — 나머지는 편집 전과 같음(편집 전 번들 대조).

## Safety conclusion

- Safe edit boundary: 위 편집만.
- High-risk impact: High-risk(exit 알림 입구).
