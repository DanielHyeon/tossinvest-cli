# Function Logic Map: `TestTheLimitsReadIsBounded`

- Source: `cmd/tossctl/console_test.go`
- AST evidence: `ast.json` (revision=current, L500–516, 분기 4개)
- Risk scan: `risk-pattern-report.md`
- 이 change의 base: `47672c6f` — base에는 이 함수가 없었다. 현재 본문을 diff hunk가 덮으므로 evidence가 요구된다 (revision=current)

seam이 context를 받지 않으므로 데드라인이 **함수 안에** 있어야 한다는 것을 소스에서 고정한다.

다른 쪽에 있는 것은 30초마다 자기를 다시 로드하고 하루 종일 열려 있는 화면이다. 멈춘 파일시스템에서 `context.Background()`로 config를 읽으면 HTTP 핸들러가 무한히 잡히고, 돌아오지 않는 렌더는 개요가 유일하게 문장으로 표현할 수 없는 실패다 — 나머지 실패는 전부 화면 위의 문장이 된다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `os.ReadFile("console.go")` | 패키지 디렉터리 | 디스크 | 읽기 실패는 `t.Fatalf` |
| `consoleGateLimitsTimeout` | > 0 | console.go 상수 | 0 이하는 즉시 만료 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 소스 읽기 실패 | — | `t.Fatalf` | 이 테스트 |
| B2 | `s.svc.Load(context.Background())` 발견 | — | `t.Error` — 무한 대기 | 동일 |
| B3 | `context.WithTimeout(context.Background(), consoleGateLimitsTimeout)` 부재 | — | `t.Error` | 동일 |
| B4 | `consoleGateLimitsTimeout <= 0` | — | `t.Errorf` | 동일 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.Contains` | 두 문자열의 존재/부재를 검사 | 정확 문자열 — 리팩터가 우회하면 실패로 드러난다 | L498, L501 |

## State mutations and fallbacks

- 테스트 — 파일 읽기만.

## Safety conclusion

- Safe edit boundary: 검사하는 두 소스 문자열과 상수 부호.
- High-risk impact: yes (Guardian 경로) — 한도 읽기가 무한히 잡히면 개요가 응답하지 않고, 운영자는 문제를 파악하려는 바로 그 순간 화면을 잃는다. 실계좌 부작용은 없다.
