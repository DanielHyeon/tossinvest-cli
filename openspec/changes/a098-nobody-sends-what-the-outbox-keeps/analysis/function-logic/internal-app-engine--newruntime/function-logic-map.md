# Function Logic Map: `NewRuntime`

- Source: `internal/app/engine/runtime.go` (**191-260**) — 편집 전에는 `183-232`
- AST evidence: `ast.json` — 편집 후 branches **17** / returns **10** / calls 17
  (편집 전 12 / 7 / 13)
- Risk scan: `risk-pattern-report.md`
- source SHA-256: 편집 전 `4bcd9ee7…` → **편집 후 `306276bd…`**

## ✅ 편집 후 재측정 — 방향은 맞고 **수와 id 예측이 둘 다 틀렸다** (2026-08-12, task 4.1)

| | 편집 전 | 예측 | **편집 후 실측** | |
|---|---:|---:|---:|---|
| 분기 | 12 | *"는다"* (내가 센 것은 **+4**) | **17 (+5)** | **틀림** |
| 이탈 | 7 | +3 | **10 (+3)** | 맞음 |
| 호출 | 13 | — | 17 | |

**+5인 이유**: 추출기는 `switch` **자체와 각 `case`를 따로 센다.** 새 블록은
`range` 1 + `switch` 1 + `case` 3 = **5**다. 내가 센 것은 `range` + `case` 3 = 4로,
**`switch` 노드를 빼먹었다.**

> **셀 수 있었다.** 바로 위의 감독 루프 블록이 **똑같은 모양**이고 그 열거가
> B2(range)·B3(switch)·B4~B6(case)로 이미 이 문서 표에 적혀 있다 — 즉 **같은 문서
> 안에 답이 있었다.** buildGateway 번들의 `fmt.Errorf` 누락과 같은 형태다:
> **못 셀 것이 아니라 옆을 안 본 것이다.**

> **⛔ *"B1과 B4~B12는 안 바뀐다"*는 조건에 대해서는 참이고 **id 에 대해서는 거짓**이다.**
> 새 블록을 **가운데**(감독 루프 검증 뒤 · 기본값 보정 앞) 넣었으므로 뒤의 id 가 밀린다:
>
> | 조건 | 편집 전 | **편집 후** |
> |---|---|---|
> | 루프 검증 (`range`·`switch`·`case`×3) | B2~B6 | **B2~B6 그대로** |
> | `Health`·`Trigger`·`AccountRef` | B7·B8·B9 | **B7·B8·B9 그대로** |
> | **보조 실행자 검증 (신설)** | — | **B10~B14** |
> | `clk == nil` | **B10** | **B15** |
> | `threshold <= 0` | **B11** | **B16** |
> | `interval <= 0` | **B12** | **B17** |
>
> **판정은 한 자도 안 바뀌었고 번호만 밀렸다.** 그런데 branch-test-map 은 번호로
> 칸을 가리키므로 그 표를 안 고치면 **「B10 = 시계 기본값」이 신설 검증을 가리킨다.**

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

`ast.json`의 열거를 그대로 쓴다 — **편집 후** AST 분기 17 · 이탈 10 · 호출 17.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 `:192` | `len(opts.Loops) == 0` | 없음 | `:193` 반환 | `TestTheRuntimeRefusesWiringItCannotSupervise` — `"no loops"` |
| B2 `:197` | `range opts.Loops` | `seen[name]` 채움 | — | 같은 테스트 전체 |
| B3 `:199` | `switch` (조건 없음) | 없음 | — | — |
| B4 `:200` | `name == ""` | 없음 | 반환 | 같은 테스트 — `"a loop with no name"` |
| B5 `:203` | `seen[name]` | 없음 | 반환 | 같은 테스트 — `"two loops with one name"` |
| B6 `:205` | `loop.Run == nil` | 없음 | 반환 | 같은 테스트 — `"a loop with no Run"` |
| B7 `:209` | `loop.Health == nil` | 없음 | **`continue`** — 아래 둘을 건너뛴다 | 기존 조립 경로 |
| B8 `:212` | `!journal.AutomaticTrigger(loop.Trigger)` | 없음 | 반환 | 같은 테스트 — `"a trigger the journal does not enumerate"` |
| B9 `:217` | `AccountRef`가 공백 | 없음 | 반환 | 같은 테스트 — `"an escalating loop with no account"` |
| **B10 `:227` (a098 신설)** | `range opts.Auxiliary` | **같은 `seen`을 이어 쓴다** | — | `TestTheRuntimeRefusesAMisWiredAuxiliaryExecutor` |
| **B11 `:229` (신설)** | `switch` (조건 없음) | 없음 | — | — |
| **B12 `:230` (신설)** | `name == ""` | 없음 | 반환 | 같은 테스트 — `"이름이 없다"` |
| **B13 `:233` (신설)** | `seen[name]` — **감독 루프 이름도 포함** | 없음 | 반환 | 같은 테스트 — `"보조끼리"` · `"감독 루프와"` 둘 다 |
| **B14 `:236` (신설)** | `aux.Run == nil` | 없음 | 반환 | 같은 테스트 — `"Run 이 없다"` |
| B15 `:249` | `r.clk == nil` | `clock.System()` 대입 | — | 기존 |
| B16 `:252` | `r.threshold <= 0` | `DefaultDegradationThreshold` | — | 기존 |
| B17 `:255` | `r.interval <= 0` | `DefaultHealthInterval` | — | 기존 |
| (분기 아님) `:260` | 위 거부 여덟이 다 거짓 | `&Runtime{…}` 구성 | `:260` 반환 | 전부 |

**a098이 더한 갈래는 B2의 루프가 끝난 뒤와 (옛) B10 사이다** — 실측상 B10~B14.
보조 실행자 슬라이스를 같은 모양으로 훑어 **이름 공백·중복·`Run` nil**을 거부한다.
중복 판정의 집합은 **`seen`을 이어 쓴다** — 보조 실행자와 감독 루프가 같은 이름을
쓰면 죽음 로그의 `"loop"` 필드가 어느 쪽인지 못 가른다.

> **`seen` 공유가 실제로 성질을 지는지 뮤테이션으로 확인했다 (E).**
> 보조 검증 앞에 `seen = map[string]bool{}` 한 줄을 넣으면 **네 하위 케이스 중
> `"감독 루프와 이름이 겹친다"` 하나만** FAIL 한다. 즉 그 케이스가 **공유를 단독으로
> 지고 있고**, 나머지 셋은 공유와 무관하다 — 넷을 한 덩어리로 세면 안 되는 이유다.

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
