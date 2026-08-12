# Function Logic Map: `NewRuntime`

- Source: `internal/app/engine/runtime.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

> **왜 이 산출물이 a098에 있는가.** 사용자 결정 9-2가 배달 실행자를 **감독 집합 밖의
> 보조 실행자**로 만들라고 했다. 그러려면 `RuntimeOptions`에 새 필드가 생기고,
> **그 필드를 조립 시점에 검증하는 자리가 이 함수**다(design D8.3).
> `NewRuntime`은 기존 High-risk 함수이므로 면제할 수 없다.
>
> a092·a098의 `Runtime.Run` 번들과 **같은 파일·같은 HEAD**다(`source_sha256` 동일).
> 다른 것은 함수와 이 문서다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `opts.Loops` | **비어 있으면 안 된다** | `:184` | `ErrRuntimeUnavailable` — *"a runtime with nothing to supervise is a process that would report success while trading nothing"* |
| `loop.Name` | 공백 아님 · **집합 안에서 유일** | B4·B5 | 각각 즉시 반환 |
| `loop.Run` | non-nil | B6 | 즉시 반환 |
| `loop.Health` | nil 허용 | B7 | nil이면 **아래 둘을 검사하지 않는다** — `continue` |
| `loop.Trigger` | `Health`가 있으면 **journal의 닫힌 열거 안** | B8 · `journal.AutomaticTrigger` | 즉시 반환. doc `:180-182`: *"a trigger the journal's closed enumeration does not know"* |
| `opts.AccountRef` | `Health`가 있으면 공백 아님 | B9 | 즉시 반환 |
| `opts.Clock`·`Threshold`·`HealthInterval` | 0/nil 허용 | B10·B11·B12 | **기본값으로 채운다** — 거부가 아니다 |

> **이 함수의 검증 정책은 두 종류다.** B1·B4~B9는 **거부**(반환)이고
> B10~B12는 **보정**(기본값)이다. a098이 더하는 검증은 **거부 쪽**이다 —
> 이름 없는 보조 실행자는 죽었을 때 로그가 누가 죽었는지 못 적는다(B4와 같은 이유).

## Branches and early returns

`ast.json`의 열거를 그대로 쓴다 — AST 분기 12 · 이탈 7 · 호출 13.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 `:184` | `len(opts.Loops) == 0` | 없음 | `:185` 반환 | `TestTheRuntimeRefusesWiringItCannotSupervise` — `"no loops"` |
| B2 `:189` | `range opts.Loops` | `seen[name]` 채움 | — | 같은 테스트 전체 |
| B3 `:191` | `switch` (조건 없음) | 없음 | — | — |
| B4 `:192` | `name == ""` | 없음 | `:193` 반환 | 같은 테스트 — `"a loop with no name"` |
| B5 `:195` | `seen[name]` | 없음 | `:196` 반환 | 같은 테스트 — `"two loops with one name"` |
| B6 `:197` | `loop.Run == nil` | 없음 | `:198` 반환 | 같은 테스트 — `"a loop with no Run"` |
| B7 `:201` | `loop.Health == nil` | 없음 | **`continue`** — 아래 둘을 건너뛴다 | 기존 조립 경로 |
| B8 `:204` | `!journal.AutomaticTrigger(loop.Trigger)` | 없음 | `:205` 반환 | 같은 테스트 — `"a trigger the journal does not enumerate"` |
| B9 `:209` | `AccountRef`가 공백 | 없음 | `:210` 반환 | 같은 테스트 — `"an escalating loop with no account"` |
| B10 `:222` | `r.clk == nil` | `clock.System()` 대입 | — | 기존 |
| B11 `:225` | `r.threshold <= 0` | `DefaultDegradationThreshold` | — | 기존 |
| B12 `:228` | `r.interval <= 0` | `DefaultHealthInterval` | — | 기존 |
| (분기 아님) `:231` | 위 거부 다섯이 다 거짓 | `&Runtime{…}` 구성 | `:231` 반환 | 전부 |

**a098이 더하는 갈래는 B2의 루프가 끝난 뒤와 B10 사이다.** 보조 실행자 슬라이스를
같은 모양으로 훑어 **이름 공백·중복·`Run` nil**을 거부한다. 중복 판정의 집합은
**`seen`을 이어 쓴다** — 보조 실행자와 감독 루프가 같은 이름을 쓰면 죽음 로그의
`"loop"` 필드가 어느 쪽인지 못 가른다.

> **B1은 안 바꾼다.** 보조 실행자만 있고 `Loops`가 비면 **여전히 거부**여야 한다.
> 배달 실행자만 도는 프로세스는 *"trading nothing"*의 정확한 정의다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.TrimSpace` `:190`·`:209` | 이름·계좌의 공백 정규화 | 순수 | `ast.json` calls |
| `journal.AutomaticTrigger` `:204` | 닫힌 열거 조회 | 순수 · bool | `ast.json` calls |
| `journal.AutomaticTriggers` `:207` | 거부 메시지에 열거를 싣는다 | 순수 | `ast.json` calls |
| `fmt.Errorf` ×5 | 거부 사유 | — | `ast.json` calls |
| `clock.System()` `:223` | 기본 시계 | — | `ast.json` calls |

**이 함수는 I/O를 하지 않는다.** 네트워크·원장·파일 호출이 0이다 — 조립 시점 검증만 한다.

## State mutations and fallbacks

- 지역 `seen` 맵만 쓴다. **패키지 전역을 안 건드린다.**
- `r.clk`·`r.threshold`·`r.interval`은 **0값을 기본값으로 덮는 폴백**이다(B10~B12).
- `escalated` 맵을 빈 채로 만든다(`:220`) — `Runtime.Run`이 아니라 여기서 만든다.
- **`Recover`·`Alerts`·`Escalate`·`Announcer`·`Log`는 검증하지 않는다.** 전부 optional이고
  nil이면 그 결과가 없을 뿐이다. **a098의 보조 실행자 필드도 같은 성질이면 안 된다** —
  이름 없는 실행자는 조용히 도는 것이 아니라 **죽었을 때 못 알아보는 것**이므로 거부다.

## Safety conclusion

- **Safe edit boundary**: a098은 **B2 루프 뒤에 같은 모양의 검증 루프 하나**를 더한다.
  B1과 B4~B12의 조건·반환·기본값은 **한 자도 안 바꾼다.**
  편집 후 AST의 분기가 12보다 크고 **B1·B4~B12의 줄 의미가 그대로면** 의도한 편집이다.
- **High-risk impact**: **yes** — 이 함수가 거부하지 못한 배선은 **필요한 순간에 발견된다.**
- **a098이 여기서 지는 제약**:

  | 제약 | 왜 | 관측 |
  |---|---|---|
  | 보조 실행자의 이름은 **감독 루프 이름과도** 충돌하면 안 된다 | 죽음 로그의 `"loop"` 필드가 한 이름 공간이다 | tasks §3의 R |
  | `Loops`가 비면 **여전히 거부** | 배달 실행자만 도는 엔진은 매매를 안 한다 | 기존 `"no loops"` 케이스가 **그대로 통과해야 한다** |
  | 보조 실행자가 **없어도** 조립이 성공한다 | 오늘의 모든 호출자가 그 필드를 안 넘긴다 | 기존 열한 테스트 전부 |
