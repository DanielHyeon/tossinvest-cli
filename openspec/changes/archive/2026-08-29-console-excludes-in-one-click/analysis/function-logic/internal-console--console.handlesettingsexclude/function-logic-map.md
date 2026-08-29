# Function Logic Map: `Console.handleSettingsExclude`

- Source: `internal/console/settings.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Narrative context: `../../function-logic-map.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `symbol` (POST) | 비어 있지 않은 문자열 | 폼(행이 이미 아는 심볼) | 빈 값은 400 |
| `c.opts.Settings` | nil 가능 | 주입 seam | nil은 501 |
| `current` | 파일 원문 블록 | `LoadRawEngineAdoption` | 판독 실패는 저장하지 않는다 |
| `next.DefaultStopPct` | 읽은 값 그대로 | — | **이 경로는 손절폭을 쓰지 않는다**(`validate()`가 exclude 목록에는 손절폭을 요구하지 않는다) |
| `remove` (POST) | `"1"` 또는 부재 | 폼 | 부재는 추가 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `Settings == nil` | 없음 | 501 refuse | `TestExcludeRefusesWhatItCannotDo` |
| B2 | `symbol == ""` | 없음 | 400 refuse | `TestExcludeRefusesWhatItCannotDo` |
| B3 | `Load()` 오류 | 없음 | redirect + 사유 | 기존 seam 오류 경로와 동일 |
| B4 | `remove == "1"` | exclude에서 그 심볼만 제거 후 Save | redirect "제외 해제됨" | `TestReleasingAnExclusion` |
| B5 | 해제 경로의 Save 오류 | 없음 | redirect "제외 해제 안 됨" | 저장 거부 경로 |
| B6 | `!next.Excludes(symbol)` | exclude에 추가 + 정렬 | 계속 | `TestExcludingASymbolFromThePositionsScreen` |
| B7 | `next.Included(symbol)` | **include에서 함께 제거** | 계속 + 공지에 명시 | `TestExcludingADesignatedSymbolDropsTheDesignation` |
| B8 | Save 오류 | 없음 | redirect "제외 안 됨" | 저장 거부 경로 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `AdoptionSettings.Load` | 서버가 스스로 읽는다 — 폼이 되돌려 보내지 않은 값은 폼이 잃을 수 없다 | 오류는 저장 포기 | CodeGraph + AST |
| `AdoptionSettings.Save` | 외과적 기록 + `console.adoption.exclude_symbols` 감사 | 엔진이 zeroing할 블록은 거부 | CodeGraph + AST |
| `withoutSymbol` | B4·B7의 제거 | 순수 | CodeGraph + AST |
| `config.Adoption.Excludes` / `.Included` | 멱등·상호배제 판정 | 순수 | CodeGraph + AST |
| `Console.engineRunning` | 반영 시점 안내 | advisory 마커, 실패는 false | CodeGraph + AST |

## State mutations and fallbacks

- 쓰는 것은 `ExcludeSymbols`, 그리고 B7에서만 `IncludeSymbols`다. `Enabled`·`DefaultStopPct`는 **읽은 값 그대로 되돌려 쓴다**.
- `next.Rejected = ""` — 거부 사유를 다시 기록하지 않는다. 블록이 실제로 무효면 `Save`가 거부하므로 엔진이 zeroing할 파일은 쓰이지 않는다.
- 공지는 현재형 보장을 하지 않는다: 가동 중 엔진은 기동 스냅샷으로 돌므로 다음 대사 주기에 그 심볼을 편입할 수 있다(적대적 리뷰 A1). 효력 시점은 `effectNotice`만 말한다.
- 알려진 한계: `Load`가 `config.Service`의 flock 밖이라 두 탭의 동시 저장은 lost update가 된다 — seam(Load/Save)의 선행 성질이며 `/settings/save`·`/settings/include`도 같다.

## Safety conclusion

- Safe edit boundary: 신규 함수. 계좌·journal·브로커 무접촉이고 `engine.adoption` 블록만 쓴다.
- High-risk impact: yes — 제외 여부가 미편입 보유에 합성 손절이 생기는지를 정한다. 다만 `judgeHoldings`가 `ExitEligible()`에서 먼저 반환하므로 **이미 관리 중인 포지션의 손절에는 도달하지 못한다**(§0.4).
