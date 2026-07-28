# Function Logic Map: `TestOpeningEveryConsoleReadScreenResolvesTheAccountOnce`

- Source: `cmd/tossctl/console_test.go`
- AST evidence: `ast.json` (revision=current, L836–1030, 분기 37개)
- Risk scan: `risk-pattern-report.md`
- 이 change의 base: `137cc8d0` — base에는 이 함수가 없었다. seam별 테스트가 통과하는 동안에도 참이 아니었던 **콘솔 단위**의 주장을 고정하려고 추가되었다 (revision=current)

주장의 단위가 seam이 아니라 **콘솔 세션**이다. `TestTheOrdersSeamResolvesTheAccountOnceAndBuildsNoSecondClient`는 seam 하나 안에서 구축 횟수를 세므로, 포지션 화면과 /orders가 각자 해석하던 동안에도 통과했다. 이 테스트는 그 상태에서 실패한다.

두 절반으로 되어 있다.

1. **런타임**: `runConsole`과 같은 방식으로 배선한 seam들(공유 resolver 1개, 기동 시 1회 생성)을 열어 본다. 화면마다 2회, 그다음 새 resolver로 두 화면을 **동시에**. 기대값은 각각 계좌 해석 1회다. 동시 절반은 `-race`에서 락의 직렬화까지 확인한다.
2. **정적**: `console.go`를 파싱해 ① `verifyBrokerFactory`를 부르는 함수 집합이 `consoleBrokerBuildSites`(이유가 적힌 map)와 정확히 일치하는지, ② `runConsole`이 공유 resolver를 **정확히 하나** 만들고 모든 읽기 seam에 그 식별자를 넘기는지 검사한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `verifyBrokerFactory` (패키지 변수) | 테스트가 교체 — 호출 1회 = 계좌 해석 1회 | verify.go `buildVerifyBroker` | 복원은 `t.Cleanup` |
| `built` 카운터 | `mu`로 보호 | 이 테스트 | 동시 절반이 있으므로 뮤텍스 없이는 `-race`에서 경합 |
| `consoleBrokerBuildSites` | 이유 문자열이 비지 않은 map | 이 파일 | 빈 이유 = 논증 없는 예외 — FAIL |
| `readSource(t, "console.go")` | 패키지 디렉터리의 소스 | 디스크 | 파싱 실패는 `t.Fatalf` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1/B2 | 라운드 2회 × 화면 2종 열기 루프 | `built` 증가는 factory 안에서 | — | 이 테스트 |
| B3 | 화면 열기가 에러 | — | `t.Fatalf` | 동일 |
| B4 | 순차 개시 후 해석 횟수 != 1 | — | `t.Errorf` — 화면당 해석은 세션당 `/api/v1/accounts` 2회 | 동일 (**이 테스트의 RED가 걸린 곳**) |
| B5–B8 | 동시 개시 goroutine 기동과 화면별 호출 분기(`switch`/두 `case`) | goroutine 2개 | — | 동일 (`-race`) |
| B9/B10 | 동시 개시 에러 수집과 보고 | — | `t.Fatalf` | 동일 |
| B11 | 동시 개시가 해석을 1회 넘게 함 | — | `t.Errorf` — 구축이 직렬화되지 않았다 | 동일 (`-race`) |
| B12 | `console.go` 파싱 실패 | — | `t.Fatalf` — 가드가 아무것도 읽지 못함 | 동일 |
| B13–B17 | 선언 순회, 함수만 선별, 수신자 이름 조립, 호출 노드 선별, `verifyBrokerFactory` 호출 집계 | `sites` map | — | 동일 |
| B18/B19 | 논증되지 않은 구축 지점 | — | `t.Errorf` | 동일 |
| B20/B21 | `consoleBrokerBuildSites`에 있으나 더 이상 구축하지 않는 항목 | — | `t.Errorf` — 쓰이지 않는 예외는 다음 사람이 재활용한다 | 동일 |
| B22/B23 | 선언 순회에서 `runConsole`만 선별 | — | — | 동일 |
| B24–B29 | `runConsole` 본문의 대입문에서 `newConsoleBroker` 호출과 그 좌변 식별자 수집 | `builds`, `holder` | — | 동일 |
| B30–B33 | 인자 1개짜리 호출 중 읽기 seam 생성자의 인자 표현식 수집 | `seamArgs` | — | 동일 |
| B34 | `builds != 1` | — | `t.Fatalf` — 두 번째 resolver는 두 번째 해석 | 동일 |
| B35–B37 | seam 배선 누락 / 인자가 공유 식별자가 아님 | — | `t.Errorf` | 동일 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newConsoleBroker` / `newConsoleHoldings` / `consoleOrdersSeam` | `runConsole`과 같은 배선을 만든다 | 정적 절반이 이 동일성을 검사한다 | console.go L216, L232, L235 |
| `newVerifyServer(t)` | 세 읽기에 답하는 httptest 서버 | 실계좌 접촉 없음 | verify_test.go |
| `parser.ParseFile` + `ast.Inspect` | 런타임이 절대 실행하지 않을 배선 실패를 잡는다 | 소스가 곧 증거 | 이 테스트의 정적 절반 |

## State mutations and fallbacks

- 패키지 변수 `verifyBrokerFactory`를 교체하고 `t.Cleanup`으로 복원한다. 그 외의 상태 변이는 없다.
- 실계좌·원장·설정 어디에도 쓰지 않는다.

## Safety conclusion

- Safe edit boundary: 기대값 1회와 `consoleBrokerBuildSites`. map에 이유 없이 항목을 더하는 것이 이 가드가 조용히 짧아지는 방식이다.
- High-risk impact: yes (주문 경로 — rate 예산) — 계좌 해석이 늘어나면 라이브 검증이 429로 실주문 스텝을 잃는다(M4). 이 테스트는 그 증가를 콘솔 단위에서 잡는 유일한 자동 검사다. 테스트 자체는 실계좌 부작용이 없다.
