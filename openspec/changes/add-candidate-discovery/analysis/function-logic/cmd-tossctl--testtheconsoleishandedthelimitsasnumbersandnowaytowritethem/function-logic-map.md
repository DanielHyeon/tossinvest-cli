# Function Logic Map: `TestTheConsoleIsHandedTheLimitsAsNumbersAndNoWayToWriteThem`

- Source: `cmd/tossctl/console_test.go`
- AST evidence: `ast.json` (revision=current, L471–490, 분기 4개)
- Risk scan: `risk-pattern-report.md`
- 이 change의 base: `583772c4` — base에는 이 함수가 없었다. 현재 본문을 diff hunk가 덮으므로 evidence가 요구된다 (revision=current)

Guardian 한도 seam의 형상을 고정한다: 메서드 하나(`GateLimits`), 그리고 `console.Options`가 그 필드를 받는다.

두 번째 메서드는 콘솔이 브라우저에서 위험 한도를 옮길 수 있게 되는 것이고, 콘솔 spec은 그 능력이 없다고 문서로 적고 있다. `*config.Service`(Init과 외과적 writer 보유)도, 게이트 타입 자체도 넘어가지 않는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `consoleGateLimitsSeam(&rootOptions{configDir: t.TempDir()})` | non-nil | console.go | nil이면 `t.Fatal` — 해석 가능한 디렉터리에서 seam이 없다는 뜻 |
| `consoleOptionFields(t)["GateLimits"]` | true | console.go 소스 | false면 개요가 영원히 미배선 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | seam이 nil | — | `t.Fatal` | 이 테스트 |
| B2 | 메서드가 `[GateLimits]` 하나가 아님 | — | `t.Fatalf` | 동일 |
| B3 | 메서드 이름 수집 루프 | — | — | 동일 |
| B4 | `GateLimits` 필드 부재 | — | `t.Error` | 동일 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `reflect.TypeOf(seam)` | 메서드 집합을 런타임에서 센다 | 쓰기 메서드 추가를 즉시 잡는다 | L469 |
| `consoleOptionFields` | 배선 여부를 소스에서 읽는다 | 런타임이 실행하지 않을 실패를 잡는다 | L478 |

## State mutations and fallbacks

- 테스트 — `t.TempDir()`만 사용, 실계좌·실 config 접촉 없음.

## Safety conclusion

- Safe edit boundary: 메서드 수 기대값과 필드 존재 검사.
- High-risk impact: yes (Guardian 경로의 집행자) — 콘솔이 한도를 쓸 수 없다는 주장의 자동 검사다. 실계좌 부작용은 없다.
