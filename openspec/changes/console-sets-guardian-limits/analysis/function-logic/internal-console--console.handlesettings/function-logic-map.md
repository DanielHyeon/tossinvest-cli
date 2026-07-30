# Function Logic Map: `Console.handleSettings`

- Source: `internal/console/settings.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Narrative context: `../../function-logic-map.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `c.opts.Settings` | nil 가능 | `Options` | nil이면 `Wired=false`, 편입 폼 미렌더 |
| `c.opts.Limits` | nil 가능 | `Options` | nil이면 `LimitsWired=false`, 한도 폼 미렌더 (이 change가 추가) |
| seam의 Load 오류 | 임의 오류 | 파일시스템 | 각자의 `*LoadErr`에 담기고 화면은 계속 렌더 |
| `r.URL.Query()["notice"]` | 임의 문자열 | 리다이렉트 | 템플릿이 이스케이프 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `c.opts.Settings != nil` | `page.Wired=true`, 편입 블록 적재 | 없음 | `TestWithoutASeamNeitherControlRenders` |
| B2 | 편입 Load 오류 | `page.LoadErr` | 없음 (렌더 계속) | `TestTheSettingsScreenShowsTheRawBlockAndTheVerdict` |
| B3 | `c.opts.Limits != nil` (신규) | `page.LimitsWired=true`, 게이트 블록 적재 | 없음 | `TestWithoutASeamTheLimitEditorRefusesRatherThanPretends` |
| B4 | 한도 Load 오류 (신규) | `page.LimitsLoadErr` | 없음 (렌더 계속) | `TestAnUnreadableConfigDoesNotHideTheRestOfTheScreen` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `AdoptionSettings.Load` | 편입 블록 원문 + 검증 판정 | 오류는 화면 문장 | CodeGraph + AST |
| `LimitSettings.Load` | 게이트 블록 원문 (신규) | 오류는 화면 문장 | CodeGraph + AST |
| `Console.engineRunning` | 반영 시점 안내 | advisory 마커, 실패는 false | CodeGraph + AST |
| `Console.render` | 템플릿 | 없음 | CodeGraph + AST |

## State mutations and fallbacks

- 로컬 `page` 구조체만 채운다. 이 함수는 쓰기를 하지 않는다.
- 두 seam의 실패는 **독립적**이다. B4가 B1·B2의 결과를 지우지 않으므로, 한도를 못 읽어도 편입 설정 화면은 그대로 뜬다 — 이것이 B3·B4를 B1 안에 중첩하지 않은 이유다.
- 한도 블록은 `page.Gate`에 통째로(=`enabled` 포함) 실린다. 표시 전용이며, 이 페이지의 어떤 폼도 그 값을 되돌려보내지 않는다.

## Safety conclusion

- Safe edit boundary: 두 개의 독립적인 seam 적재 블록. 어느 하나의 nil·오류도 다른 하나를 가리지 않아야 한다.
- High-risk impact: no — 읽기 전용 렌더. 다만 이 함수가 화면에 싣는 `Gate`가 §0.5 경로의 설정이므로, 파일에 없는 값을 여기서 채워 넣으면(암묵 기본값) 화면과 엔진이 갈라진다. 채우지 않는다(design D1).
