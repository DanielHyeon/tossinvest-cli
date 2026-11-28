# Function Logic Map: `dispatch`

- Source: `tools/a140-breakout-signal-shadow/main.go` (55-67, base `8877d7e01`)
- AST evidence: `ast.json`(**편집 뒤** 재추출, revision current, 56-70, source_sha256 `e2aacba1bc7a…` — 편집 전 갈래는 아래 「편집 전」 표와 커밋 `f71ad60ba` 의 같은 번들)
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
| `replayCommand` | a203 과거 재생(편집 뒤) | 오류를 그대로 돌려줌, M9 는 그 안에서 종료 코드 2 | AST calls(편집 뒤) |
| `errors.New` · `fmt.Errorf` | B1 · B5 오류 | — | AST calls |

## State mutations and fallbacks

- 없음. 폴백 없음 — 모르는 이름은 오류.

## a203 편집 — 적용됨(design §2 · C10)

- B4 와 default 사이에 `case "replay": return replayCommand(ctx, args[1:], stdout)` 1개, B1 · default 오류 문구에 `"replay"` 를 더함. 다른 갈래 불변.

## Branches and early returns — 편집 뒤(현재, 56-70)

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if 57 `len(args) == 0` | 없음 | `a subcommand is required: "run", "report" or "replay"`(58) | `TestA203DispatchRoutesReplay` |
| B2 | switch 60 `args[0]` | 없음 | — | 아래 B3~B6 |
| B3 | case 61 `"run"` | `runCommand` | 그 결과(62) | `TestA166Breakout*`(runCommand 직접) |
| B4 | case 63 `"report"` | `reportCommand` | 그 결과(64) | `TestA203DispatchRoutesReplay`(빈 원장 report 가 오류 없이 돎) |
| B5 | case 65 `"replay"` — **a203 새 갈래** | `replayCommand`(a203_replay.go) | 그 결과(66) | `TestA203DispatchRoutesReplay`(「replay needs a step」 오류가 replay 쪽에서 나옴) |
| B6 | default 67 | 없음 | `unknown subcommand %q; use "run", "report" or "replay"`(68) | `TestA203DispatchRoutesReplay` |

반환 좌표(AST, 편집 뒤): 58 · 62 · 64 · 66 · 68. 편집 뒤 실측: 분기 6, 반환 5(편집 전 5 · 4 에 case 1개).

## Safety conclusion

- Safe edit boundary: case 1개 + 오류 문구. `run` · `report` 갈래와 그 인자 불변. a166 census `shapes.json` 의 이 함수 고정 본문을 같은 커밋에서 갱신(M2).
- High-risk impact: no — 도구 하위 명령 분기.
