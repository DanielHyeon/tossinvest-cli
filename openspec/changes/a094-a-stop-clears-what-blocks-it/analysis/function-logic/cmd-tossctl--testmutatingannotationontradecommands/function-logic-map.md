# Function Logic Map: `TestMutatingAnnotationOnTradeCommands`

- Source: `cmd/tossctl/help_convention_test.go` (`96`–`157`)
- Qualified: `TestMutatingAnnotationOnTradeCommands`
- AST evidence: `ast.json` (`source_sha256` b22792a94e54b07b…)
- Risk scan: `risk-pattern-report.md`

**역할.** 기존 시험 — mutating 표지 목록에 `tossctl engine attempt-resolve` 한 줄(완화 명령 가족 계약).

## Inputs and invariants

시험 픽스처(임시 원장). 생산 코드 아님 — 비례 원칙상 이 번들은 게이트의 요구 집합(수정된 기존 함수)을 채우는 기록이다.

## Branches and early returns

| Branch | 종류 | 조건 (원문) | 진입 실측 |
|---|---|---|---|
| B1 | range | `:147` `for _, c := range leafCommands(newRootCmd()) {` | — |
| B2 | if | `:150` `if wantMutating[path] && !isMut {` | — |
| B3 | if | `:153` `if !wantMutating[path] && isMut {` | — |

## Calls and live bindings

시험 대상 API(원장 · 관측 루프)만 부른다. 브로커 호출 0.

## State mutations and fallbacks

임시 원장만.

## Safety conclusion

- High-risk impact: no(시험). 단언은 약화되지 않았다 — 바뀐 것은 새 호출 형태(기대 intent) 또는 이름 붙은 대가의 +1 관측뿐.
