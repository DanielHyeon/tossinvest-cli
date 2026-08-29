# Function Logic Map: `TestRawReadsClassifyErrorsLikeEveryOtherRead`

- Source: `internal/official/orders_raw_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

이 함수는 이 change가 **고치는 대상이 아니라 부딪힌 것**이다. raw read가 다른 read와
같은 sentinel을 돌려주는지 보는 표인데, 판정을 `err == ErrAuth`로 하고 있었다.
a082가 그 sentinel을 감싸면서 `==`가 깨졌다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 표의 세 행 | 429·401·500 | 같은 파일 | — |
| `tc.check` | 오류가 기대한 sentinel인가 | **계약은 `errors.Is`다** — `execgw/retry.go:60`, `classify.go:111`, `failclosed.go:210`, `cmd/tossctl/soak.go:613`, `openapi.go:56` 전부 그렇게 판정한다 | `==`는 계약보다 **엄격**해서, 감싼 오류를 production이 받아들이는데 이 테스트만 거부한다 |
| 함수 주석 | "the retry matrix classifies by those sentinels" | 같은 파일 186번째 줄 근처 | 그 matrix가 `errors.Is`를 쓴다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `range` — 세 오류 종류 | 없음 (httptest) | 없음 | 자기 자신 |
| B2 | `err == nil \|\| !tc.check(err)` | 없음 | `t.Errorf` | 자기 자신 |

**편집은 `tc.check` 세 개의 본문뿐이다.** `==`를 `errors.Is`로 바꾼다. 표의 행,
상태 코드, 기대 sentinel, 하네스, 단정 구조는 그대로다.

세 행을 **모두** 바꾼 이유: 인증 행만 바꾸면 같은 표 안에서 두 판정 방식이
섞이고, 읽는 사람이 그 차이를 의미 있는 것으로 읽는다. 감싸지 않은 sentinel에
대해 `errors.Is(x, x)`는 `x == x`와 같으므로 나머지 두 행의 판정력은 변하지 않는다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `rawTestClient` | 상태 코드를 고정한 httptest client | 네트워크 없음 | AST calls |
| `c.OrdersPageRaw` | 판정 대상 read | 오류를 그대로 올린다 | AST calls |
| `errors.Is` | **신규** — sentinel 판정 | 순수 | design D4 |

## State mutations and fallbacks

- 상태를 바꾸지 않는다. httptest 서버 하나를 세우고 내린다.
- fallback 없음.

## Safety conclusion

- Safe edit boundary: **`tc.check` 세 개의 본문뿐.** 표의 데이터와 단정 구조는
  글자 그대로 보존한다.
- 이 편집은 판정력을 **약화하지 않는다**. `errors.Is`는 production 전체가 쓰는
  기준이고, `==`는 그보다 좁아 감싼 오류를 잘못 거부했다. 약화 여부의 근거는
  `TestNothingDecidesAnAuthRefusalByReadingItsMessage`가 따로 고정한다 — 같은
  문구를 가진 무관한 오류는 `errors.Is`를 만족하지 않는다.
- High-risk impact: **no.** 테스트 전용이고 production 코드에 도달하지 않는다.
