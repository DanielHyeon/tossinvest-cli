# Function Logic Map: `Console.handlePositions`

- Source: `internal/console/portfolio_pages.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Narrative context: `../../function-logic-map.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `c.opts.Settings` | nil 가능 | 주입 seam | nil이면 `CanDesignate=false`, 어떤 컨트롤도 렌더하지 않는다 |
| `Settings.Load()` | `(config.Adoption, verdict, error)` | config 파일 원문 | 오류면 스탬프하지 않는다 — 읽지 못한 목록으로 행을 칠하지 않는다 |
| `page.Snap.Rows` | 브로커+원장 조인 결과 | `Console.positions` | 비어 있어도 루프가 무동작 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `c.opts.Settings != nil` | 없음 | 거짓이면 스탬프 없이 렌더 | `TestWithoutASeamNeitherControlRenders` |
| B2 | `Load()` 오류 없음 | `CanDesignate = true` | 오류면 `CanDesignate`가 false로 남는다 | 기존 seam 오류 테스트 |
| B3 | `range page.Snap.Rows` | 행마다 `Designated`·`Excluded` 스탬프 | 없음 | `TestAnExcludedRowIsLabelledAndOffersRelease` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Console.positions` | 브로커·원장 조인 스냅샷 | 캐시 TTL·검증 중 브로커 호출 유예 | CodeGraph + AST |
| `AdoptionSettings.Load` | 두 목록 — **호출 1회** | 오류는 스탬프 포기 | CodeGraph + AST |
| `config.Adoption.Included` / `.Excludes` | 행별 등재 판정 | 순수 | CodeGraph + AST |
| `Console.render` | 화면 | 없음 | CodeGraph + AST |

## State mutations and fallbacks

- 이 change의 변경은 같은 루프 안에서 `Excluded` 스탬프 **1줄 추가**다.
- `Load()`는 여전히 **1회**다. 두 번 읽으면 두 목록이 서로 다른 스냅샷에서 와서, 한 행이 두 목록 어디에도 없거나 양쪽 모두에 있는 것처럼 그려질 수 있다.
- fallback: seam이 없거나 읽히지 않으면 두 스탬프 모두 zero value로 남고 화면은 이 change 이전과 같다.

## Safety conclusion

- Safe edit boundary: 스탬프 1줄. 읽기 핸들러이며 아무것도 쓰지 않는다.
- High-risk impact: no — GET 읽기 경로다. 계좌·원장·config에 쓰지 않는다.
