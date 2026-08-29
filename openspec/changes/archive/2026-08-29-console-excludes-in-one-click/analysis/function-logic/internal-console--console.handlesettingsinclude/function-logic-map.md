# Function Logic Map: `Console.handleSettingsInclude`

- Source: `internal/console/settings.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Narrative context: `../../function-logic-map.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `symbol` (POST) | 비어 있지 않은 문자열 | 폼 | 빈 값은 400 |
| `c.opts.Settings` | nil 가능 | 주입 seam | nil은 501 |
| `current` | 파일 원문 블록 | `LoadRawEngineAdoption` | 판독 실패는 저장하지 않고 사유를 표시 |
| `next.DefaultStopPct` | `[0.02, 1)` | config 검증 | 범위 밖이면 콘솔 기본값 5%를 **명시적으로** 기록 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `Settings == nil` | 없음 | 501 refuse | `TestAnUnwiredSettingsSeamIsExplained` |
| B2 | `symbol == ""` | 없음 | 400 refuse | 기존 refuse 테스트 |
| B3 | `Load()` 오류 | 없음 | redirect + 사유 | 기존 seam 오류 테스트 |
| B4 | `remove == "1"` | include에서 제거 후 Save | redirect "지정 해제됨" | `TestRemovingADesignationOnlyAffectsTheFuture` |
| B5 | 해제 경로의 Save 오류 | 없음(저장 실패) | redirect "해제 안 됨" | 기존 거부 테스트 |
| B6 | `!next.Included(symbol)` | include에 추가 + 정렬 | 계속 | `TestDesignatingASymbolFromThePositionsScreen` |
| B7 | `next.Excludes(symbol)` — **이 change가 추가** | 없음 — 공지 문구만 바꾼다 | 계속 | `TestDesignatingAnExcludedSymbolSaysTheExclusionWins` |
| B8 | `DefaultStopPct` 범위 밖 | 5% 기록 | 계속 | `TestDesignationAppliesTheDefaultStopFraction` |
| B9 | Save 오류 | 없음 | redirect "지정 안 됨" | `TestAnInvalidSaveWritesNothing` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `AdoptionSettings.Load` | 파일 원문 블록 | 오류는 저장 포기 | CodeGraph + AST |
| `AdoptionSettings.Save` | 외과적 기록 + 감사 | 엔진이 zeroing할 블록은 거부 | CodeGraph + AST |
| `withoutSymbol` | 해제 경로 | 순수 | CodeGraph + AST |
| `config.Adoption.Excludes` | B7 판정 | 순수 | CodeGraph + AST |
| `Console.engineRunning` | 반영 시점 안내 | advisory 마커, 실패는 false | CodeGraph + AST |

## State mutations and fallbacks

- 이 change가 바꾼 것은 **B7 한 갈래와 그것이 고르는 공지 문자열**뿐이다. 쓰기 경로(B4·B6·B8·B9)는 무변경이다.
- B4의 제거 루프를 `withoutSymbol` 호출로 바꿨다 — 같은 필터의 무동작 리팩터이고 `TestRemovingADesignationOnlyAffectsTheFuture`가 그대로 덮는다.
- fallback: 제외되지 않은 심볼(`B7` 거짓)의 공지는 변경 전과 **바이트 단위로 같다**.

## Safety conclusion

- Safe edit boundary: 공지 문자열 선택. 저장되는 블록의 내용은 이 change로 달라지지 않는다.
- High-risk impact: yes — 이 블록이 합성 손절의 생성 여부를 정한다. 그래서 바꾼 것을 공지로 한정했다.
