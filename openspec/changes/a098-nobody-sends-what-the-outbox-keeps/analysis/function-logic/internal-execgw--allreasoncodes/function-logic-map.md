# Function Logic Map: `AllReasonCodes`

- Source: `internal/execgw/failclosed.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

> **왜 이 산출물이 a098에 있는가.** 사용자 결정 8-1이 **여섯 번째 진입 차단 트리거**를
> 승인했고, 그 트리거는 **자기 ReasonCode를 가져야 한다**(design D8.2). 새 코드는
> 이 함수의 리터럴에 한 줄로 들어간다 — 즉 **기존 함수 본문 편집**이다.
> `.claude/CLAUDE.md`의 규칙상 면제할 수 없다. 분기가 없는 함수여도
> **「분기가 없다」를 AST로 확인한 것**과 **안 본 것**은 다르다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 입력 | **없다** — `params=0` | `ast.json` signature | — |
| 반환 | `[]ReasonCode`, **정렬됨** | `:299`의 `sort.Slice` | 정렬이 빠지면 golden이 깨진다 |
| 리터럴의 원소 | `reason.go`가 선언한 상수만 | `internal/execgw/reason.go` | 미선언 식별자는 **컴파일 실패** |
| **열거의 완결성** | 이 빌드가 낼 수 있는 코드 **전부** | 함수 doc `:249-253` | 빠뜨려도 **컴파일은 통과한다** — 아래 ⚠ |

> **⛔⛔ 이 함수의 doc은 지금 **거짓**이다 — 5라운드 A-T2. 실측했다.**
>
> | | 값 | 근거 |
> |---|---|---|
> | `reason.go`가 선언한 `ReasonCode` 상수 | **45** | `rg -c 'ReasonCode = ' internal/execgw/reason.go` |
> | `AllReasonCodes()`가 등록한 것 | **29** | `wc -l internal/execgw/testdata/reason_codes.golden` |
> | **차이** | **16** | Guardian 변조·replay 계열·protection/reservation/strategy 계열 |
>
> doc(`:249-253`)은 *"every reason code this build can emit"*라고 적지만 **아니다.**
> 상수를 선언만 하고 여기 안 넣어도 빌드는 초록이고, `:283-288`의 주석이 그 전례를
> 스스로 적는다 — *"They were declared with the rest of the enum but never registered here."*
> **그 문제는 그때 고쳐진 것이 아니라 다섯 개만 고쳐진 것이었다.**
>
> **a098에 대한 뜻**: 새 코드를 `reason.go`에만 선언하고 여기 안 넣으면
> golden이 안 깨지고 아무도 모른다 — 그래서 §5.2b가 필요하다.
> **동시에, §5.2b의 「길이가 하나 늘었다」는 이 열거의 완결성을 증명하지 않는다.**
> 그것이 증명하는 것은 **새 코드가 등록됐다**뿐이다. 남은 16개는 a098의 범위가 아니고,
> **미배정 후속으로 이름을 붙여 둔다** — 침묵한 생략이 아니다.

## Branches and early returns

`ast.json`의 열거를 그대로 쓴다 — AST 분기 0 · 이탈 2 · 호출 1 · 대입 1.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | **없다.** 조건 분기가 0개다 | 슬라이스 리터럴 하나 구성(`:255`) | `:300` — 정렬된 슬라이스 | `TestReasonCodeEnumIsStable` |

- 이탈 둘 중 `:299`는 **`sort.Slice`에 넘긴 비교 클로저의 반환**이고(`codes[i] < codes[j]`),
  함수 자체의 이탈은 `:300` **하나**다. AST가 둘로 세는 것은 클로저를 함께 훑기 때문이다.
- **조건 분기가 없으므로 「어떤 입력에서 다르게 도는가」가 없다.** 이 함수의 행동은
  **리터럴의 내용 그 자체**이고, 그래서 검증이 golden 대조로 되어 있다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `sort.Slice` `:299` | 반환 순서를 결정론으로 만든다 | 오류 없음. 순수 함수 | `ast.json` calls |

**호출자 실측** (`rg`로 전수):

| 호출자 | 무엇을 하는가 | a098에 대한 뜻 |
|---|---|---|
| `internal/execgw/failclosed_test.go:40` | golden 대조 + 중복 검사 (`TestReasonCodeEnumIsStable`) | **새 코드를 넣는 순간 RED가 된다** |
| `internal/execgw/gen_golden_test.go:20` | `TOSSOS_UPDATE_GOLDEN=1`일 때만 fixture 재생성 (`TestWriteReasonCodeGolden`) | GREEN으로 가는 **유일한 정규 경로** |

`internal/risk/reason.go:184`에도 같은 이름의 함수가 있지만 **다른 패키지의 다른 열거**다.
a098은 그것을 안 건드린다.

## State mutations and fallbacks

- 전역 상태를 **안 바꾼다.** 매 호출마다 새 슬라이스를 만든다(`assignments=1`).
- **폴백 없음.**
- **이 함수 밖의 짝 자료구조 둘**이 같은 코드 목록을 따로 들고 있다 — 실측:

  | 자리 | 무엇 | 새 코드를 안 넣으면 |
  |---|---|---|
  | `retry.go:584` `latchOrder` | `checkAccountEntryLocked`(`:540`)의 **결정론적 우선순위** | **차단은 여전히 걸린다.** `:545-547`이 목록에 없는 래치도 반환한다 — 다만 **순서가 비결정적**이 된다 |
  | `testdata/reason_codes.golden` | 문자열 고정 | `TestReasonCodeEnumIsStable`이 **깨진 채로 남는다** |

  `latchOrder`는 **함수가 아니라 패키지 변수**이므로 Function Logic Map 대상이 아니다
  (`not-applicable` — 리터럴 append이고 분기가 없다). 그 append가 바꾸는 **함수**는
  `checkAccountEntryLocked`이고, a098은 **그 본문을 안 건드린다.**

## Safety conclusion

- **Safe edit boundary**: a098은 리터럴에 **원소 하나**를 더한다. 분기 0은 편집 후에도 0이고,
  이탈 2도 그대로다. **편집 후 AST에 분기가 생기면 그것은 a098이 의도한 편집이 아니다.**
- **High-risk impact**: **yes** — 이 문자열은 원장과 운영자 알림에 그대로 들어간다
  (doc `:250-253`). 이름을 잘못 고르면 되돌리기가 **데이터 마이그레이션**이다.
- **a098이 여기서 지는 제약**:

  | 제약 | 관측 |
  |---|---|
  | 새 코드를 `reason.go`와 **이 함수 둘 다**에 넣는다 | `TestReasonCodeEnumIsStable` |
  | golden은 `TOSSOS_UPDATE_GOLDEN=1`로만 갱신한다 | `TestWriteReasonCodeGolden` |
  | 기존 스물여덟 문자열은 **한 자도 안 바꾼다** | 같은 테스트 |
