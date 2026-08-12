# Function Logic Map: `NewRuntime`

- Source: `internal/app/engine/runtime.go` (183-232)
- AST evidence: `ast.json` — branches 12, returns 7, calls 13, assignments 2,
  **defers 0, go_statements 0**
- Risk scan: `risk-pattern-report.md`

**17판이 통과해야 하는 검문소.** a092는 이 함수를 편집하지 않지만, 배달 루프를
`SupervisedLoop`으로 등록하려면 **여기의 거부 조건 전부를 만족해야 한다.**
그 조건이 무엇인지가 이 열거다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `opts.Loops` | 비어 있으면 안 됨 | 호출자 배선 | B1 `:184` → 오류 `:185` |
| `loop.Name` | 공백 아님, **중복 아님** | 같은 위 | B4 `:192`, B5 `:195` |
| `loop.Run` | nil이면 안 됨 | 같은 위 | B6 `:197` |
| `loop.Health` | **nil이면 나머지 검증을 건너뛴다** | 같은 위 | B7 `:201` `continue` |
| `loop.Trigger` | `journal.AutomaticTrigger`가 아는 값 | `journal` 열거 | B8 `:204` |
| `opts.AccountRef` | `Health`가 있으면 공백이면 안 됨 | 같은 위 | B9 `:209` |
| `opts.Clock`·`Threshold`·`HealthInterval` | 0이면 기본값 | B10·B11·B12 | 실패 없음 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return |
|---|---|---|---|
| B1 `:184` | `len(opts.Loops) == 0` | 없음 | 오류 `:185` |
| B2 `:189` | `range opts.Loops` | — | — |
| B3 `:191` | `switch` | — | — |
| B4 `:192` | 이름이 공백 | 없음 | 오류 `:193` |
| B5 `:195` | **이름이 중복** | 없음 | 오류 `:196` |
| B6 `:197` | `loop.Run == nil` | 없음 | 오류 `:198` |
| **B7 `:201`** | **`loop.Health == nil`** | `seen[name] = true`는 이미 됨 `:200` | **`continue` — B8·B9를 건너뛴다** |
| B8 `:204` | `Trigger`가 열거에 없다 | 없음 | 오류 `:205` |
| B9 `:209` | `AccountRef`가 공백 | 없음 | 오류 `:210` |
| B10 `:222` | `r.clk == nil` | `clock.System()` `:223` | — |
| B11 `:225` | `r.threshold <= 0` | 기본값 `:226` | — |
| B12 `:228` | `r.interval <= 0` | 기본값 `:229` | — |
| — `:231` | — | `*Runtime` | `r, nil` |

**B7이 17판의 선택지를 만든다.** `Health`를 nil로 두면 `Trigger`와 `AccountRef`
검증이 건너뛰어진다 — 즉 **배달 루프를 "감독은 받되 승격은 안 하는" 루프로 등록할
수 있다.** 반대로 `Health`를 주면 `Trigger`가 `journal.AutomaticTriggers()`의
닫힌 열거 안에 있어야 한다.

**그 열거에 쓸 값이 이미 있다**(측정): `operating_mode.go:514-523`이
`ModeTriggerCriticalAlertUndelivered`를 포함하고, `TargetModeForTrigger:537-545`가
그것을 `ModeEntryBlocked`로 보낸다. **자동으로 `HALT_ALL`에 가는 trigger는 없다**
(`:533-534`). 그러므로 17판은 **열거를 늘리지 않고** `Health`를 붙일 수 있고,
승격의 도달점은 `ENTRY_BLOCKED` — 진입 차단이지 정지가 아니다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `len` `:184` | 루프 수 | 순수 | AST calls |
| `strings.TrimSpace` `:190`·`:209` | 이름·계좌 정규화 | 순수 | AST calls |
| `journal.AutomaticTrigger` `:204` | trigger가 열거에 있나 | 순수 판정 | AST calls |
| `journal.AutomaticTriggers` `:207` | 오류 메시지용 열거 | 순수 | AST calls |
| `clock.System` `:223` | 기본 시계 | — | AST calls |
| `fmt.Errorf` ×5 | 거부 사유 | — | AST calls |

**I/O 없음. goroutine 없음.** 순수 검증 + 구조체 조립이다.

## State mutations and fallbacks

| 대상 | 자리 | 성격 |
|---|---|---|
| `seen` map | `:200` | 지역 — 이름 중복 검출용 |
| `*Runtime` | `:215-221` | 새 값 |
| `r.clk`·`r.threshold`·`r.interval` | B10·B11·B12 | **기본값 대입** — 유일한 fallback |

- **`ErrRuntimeUnavailable`로 전부 감싼다.** 배선 실수는 시작 시점에 죽는다.
  doc comment `:180-182`가 근거다 — 감독자가 필요한 순간에 발견하는 것보다
  낫다.

## Safety conclusion

- **Safe edit boundary**: a092는 이 함수를 편집하지 않는다.
  **편집하지 않고 통과할 수 있는지**가 17판의 조건이다.
- **High-risk impact**: yes — 엔진의 루프 집합이 여기서 확정된다.
  루프를 하나 더하면 **`Runtime.Run`의 "첫 정지가 전부를 내린다"**가
  그 루프에도 적용된다. 즉 **배달 루프가 죽으면 엔진 전체가 내려간다.**
  이것이 17판이 받아들이는 결합이고, `Runtime.Run` 산출물이 그 대가를 적는다.
- **B5(이름 중복)가 실질적 제약이다**: 배달 루프의 이름은 기존 루프
  (`LoopNames()`가 출력하는 집합)와 겹치면 안 된다.
- **열거는 확인했고 선택은 남았다.** `ModeTriggerCriticalAlertUndelivered`가
  이미 있으므로 B8을 통과하는 배선이 가능하다. `Health`를 붙일지(승격까지)
  nil로 둘지(감독만)는 **§7 설계 확정의 결정**이고, 이 산출물은 두 길이
  **둘 다 열려 있다**는 사실만 확정한다.
- `AccountRef`는 이미 배선돼 있다 — `Health`를 붙여도 B9가 새 요구를 만들지
  않는다(기존 루프들이 이미 그 경로를 지난다).
