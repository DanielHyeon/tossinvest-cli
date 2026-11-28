# Function Logic Map: `Service.SaveEngineAdoption`

- Source: `internal/config/adoption_io.go` (80-126)
- Qualified: `Service.SaveEngineAdoption`
- AST evidence: `ast.json` (source_sha256 `6f1df446b991…`) — AST branches 9 · returns 9
- Risk scan: `risk-pattern-report.md`(설정 규칙 일치 0)
- 작성 시점: 편집 **전**(a208 tasks 1.1, 2026-11-28, HEAD `810aec6f7` — base `baeddfe31` 와 본문 동일).

**역할.** 콘솔 저장 — 블록을 검증한 뒤 `engine.adoption` 값 구간만 바이트로 교체. 호출자 `consoleAdoptionSettings.Save`(cmd/tossctl/adoptionsettings.go:63).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `a Adoption` | 콘솔이 만든 블록 | `handleSettingsSave` · `handleSettingsInclude` | — |
| 불변식 | 엔진이 0 으로 만들 블록은 쓰지 않음 · 블록 밖 바이트 보존 | 파일 머리 주석 · `TestSaveReplacesOnlyTheAdoptionBlock` | — |

## Branches and early returns

| Branch | Condition | At | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `if why := next.validate(); why != ""` | 85:2 | — | return 거절 (86:3) | `TestSaveRefusesAnInvalidBlock` |
| B2 | `if err != nil` — 잠금 실패 | 90:2 | — | return err (91:3) | 미측정 |
| B3 | `switch` ReadFile 결과 | 96:2 | — | — | 아래 case |
| B4 | `case errors.Is(err, os.ErrNotExist)` | 97:2 | 뼈대 파일 생성 | return installBytes (106:3) | `TestSaveCreatesASkeletonOnlyWhenTheFileIsAbsent` |
| B5 | `if err != nil` — 뼈대 직렬화 실패 | 103:3 | — | return err (104:4) | 미측정 |
| B6 | `case err != nil` | 107:2 | — | return err (108:3) | 미측정 |
| B7 | `if !json.Valid(data)` | 111:2 | — | return 거절 (112:3) | `TestSaveRefusesACorruptFile` |
| B8 | `if err != nil` — 블록 직렬화 실패 | 117:2 | — | return err (118:3) | 미측정 |
| B9 | `if err != nil` — splice 실패 | 122:2 | — | return err (123:3) | 미측정 |

Exact AST return positions: 86:3, 91:3, 104:4, 106:3, 108:3, 112:3, 118:3, 123:3, 125:2

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `normaliseSymbols` · `next.validate` | 정규화 · 검증 | 거절 → B1 | AST |
| `acquireConfigLock` · `lock.release` | flock 직렬화 | 오류 → B2 | AST |
| `os.ReadFile` · `json.Valid` · `json.Marshal(Indent)` | 읽기 · 검사 · 직렬화 | 오류 → B5~B8 | AST |
| `spliceAdoption` · `s.installBytes` | 값 구간 교체 · 원자 쓰기 | 오류 → B9 / return | AST |

Live config binding: 호출 때마다 파일을 읽고 씀(콘솔 요청 시).

## State mutations and fallbacks

- config.json 원자 교체(engine.adoption 값 구간만). 실패면 파일 무변.

## Safety conclusion

- Safe edit boundary(a208): `next` 리터럴에 `FloorOnlyExit: a.FloorOnlyExit` 한 필드. `omitempty` 라 거짓이면 직렬화 바이트가 a208 이전과 같다. 분기 0 추가.
- High-risk impact: no.
