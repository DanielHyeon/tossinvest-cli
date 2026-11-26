# Function Logic Map: `dispatch`

- Source: `tools/a140-breakout-signal-shadow/main.go` (55-67, base `8877d7e01`)
- AST evidence: `ast.json`(편집 전 추출, source_sha256 `56fab140558f…` — 편집 뒤 재추출로 갱신함)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: Low — 오프라인 측정 도구의 하위 명령 분기. 생산 경로 · 주문 · 설정 쓰기 없음.
- 호출자: `main`(main.go:47-53) 하나 — `os.Args[1:]` · `os.Stdout` 을 넘김. a166 census `shapes.json` 가 이 함수 본문을 고정함(design §2 · M2).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `args` | 첫 원소 = 하위 명령 이름, 나머지는 그 명령 플래그 | `os.Args[1:]` | 비면 B1 오류, 모르는 이름이면 B5 오류(GET 0) |
| `ctx` | 시그널 취소 ctx | `main` | 하위 명령에 그대로 넘김 |
| `stdout` | 보고 출력 | `main` | `report` 에만 넘김 |

## Branches and early returns — 편집 전(base `8877d7e01`)

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if 56 `len(args) == 0` | 없음 | `a subcommand is required: "run" or "report"` | 기존 시험 없음(편집 뒤 시험 추가) |
| B2 | switch 59 `args[0]` | 없음 | — | — |
| B3 | case 60 `"run"` | `runCommand` 호출 | 그 결과 | a166_budget_test.go(runCommand 직접) |
| B4 | case 62 `"report"` | `reportCommand` 호출 | 그 결과 | 기존 report 시험 |
| B5 | default 64 | 없음 | `unknown subcommand %q; use "run" or "report"` | 편집 뒤 시험 추가 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `runCommand` | 장중 그림자 측정 | 오류를 그대로 돌려줌 | AST calls |
| `reportCommand` | 원장 보고 | 오류를 그대로 돌려줌 | AST calls |
| `errors.New` · `fmt.Errorf` | B1 · B5 오류 | — | AST calls |

## State mutations and fallbacks

- 없음. 폴백 없음 — 모르는 이름은 오류.

## a203 편집 계획(design §2 · C10)

- B4 와 default 사이에 `case "replay": return replayCommand(ctx, args[1:], stdout)` 1개, B1 · B5 오류 문구에 `"replay"` 를 더함. 다른 갈래 불변.

## Safety conclusion

- Safe edit boundary: case 1개 + 오류 문구. `run` · `report` 갈래와 그 인자 불변.
- High-risk impact: no — 도구 하위 명령 분기.
