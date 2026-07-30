# Function Logic Map: `newConsoleCmd`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json` (revision=current, L95–160, 분기 0개)
- Risk scan: `risk-pattern-report.md`
- 이 change의 base: `137cc8d0` — 본문 변경: Long 도움말에 `signals` 세 줄 추가 (revision=current)

`tossctl console` 명령의 정의. `mutating: true`이고 그 이유가 이 change 묶음에서 바뀌지 않았다는 것이 요점이다 — 콘솔은 verify 러너를 통해 **실제 주문을 낼 수 있다**(타이핑 승인 뒤).

**이제 배선된 것**: 도움말이 화면 목록을 다시 적는다 — `/dashboard`(개요), `/`(콘솔), `/signals`(발굴). `/signals`가 "저장소를 읽고 원천을 부르지 않는다"고 명시된다.
**여전히 배선되지 않은 것**: 세션 토큰·승인·바인딩 인터페이스를 바꾸는 플래그는 하나도 없다. `--port` 하나뿐이고 `--confirm-each`도 제공하지 않는다. `verify run`의 confirmer는 여전히 터미널 전용이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `root *rootOptions` | non-nil (root가 주입) | `newRootCmd` | — |
| `opts.port` | 0 또는 포트 번호 | `--port` | 0이면 OS가 고른다 |
| `Annotations` | `source=official`, `mutating=true` | 이 파일 | `mutating`을 내리면 도구 계층의 자동 실행 금지가 풀린다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| (무분기) | — | cobra 명령 구성 + `--port` 등록 | `*cobra.Command` | `TestConsoleIsRegisteredAndAnnotated`, `TestConsoleOffersOnlyThePortFlag` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.TrimSpace` | Long 도움말 정리 | — | ast.json calls |
| `cmd.Flags().IntVar` | `--port` 단 하나 | 소스 기준 플래그 수 1개를 테스트가 고정 | `TestConsoleOffersOnlyThePortFlag`의 `declaredFlags` |
| `runConsole` (RunE closure) | 실행 본체 | 에러는 cobra로 | L153 |

## State mutations and fallbacks

- 명령 객체만 만든다. 부작용 없음.
- 도움말에 `127.0.0.1`, `session token`, `CSRF`, `read-only`가 반드시 등장한다 — 운영자가 무엇을 여는지 알아야 한다.

## Safety conclusion

- Safe edit boundary: Long 도움말 텍스트와 플래그 등록. 플래그를 하나라도 늘리면 승인 표면의 손잡이가 된다.
- High-risk impact: yes (주문 경로) — `mutating: true`인 명령의 정의다. 승인·바인딩·토큰을 여는 플래그가 하나라도 생기면 사람 승인 없는 실주문 경로가 열린다. 현재 그런 플래그는 없다.
