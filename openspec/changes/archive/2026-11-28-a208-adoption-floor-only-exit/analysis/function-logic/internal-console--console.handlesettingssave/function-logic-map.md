# Function Logic Map: `Console.handleSettingsSave`

- Source: `internal/console/settings.go` (238-275)
- Qualified: `Console.handleSettingsSave`
- AST evidence: `ast.json` (source_sha256 `5ba0d5d2ded1…`) — AST branches 4 · returns 4
- **편집 뒤 갱신(a208 3.3, 2026-11-28)**: `ast.json` 을 편집 뒤 소스로 재추출(`revision: current`), 좌표는 옛→새 AST 노드 순서 대응으로 옮김(분기 · 반환 개수 무변, 새로 생긴 분기 없음). 아래 「Safe edit boundary」는 편집 전 계획이며 그대로 구현됨.
- Risk scan: `risk-pattern-report.md`(설정 규칙 일치 0)
- 작성 시점: 편집 **전**(a208 tasks 1.1, 2026-11-28, HEAD `810aec6f7` — base `baeddfe31` 와 본문 동일).

**역할.** 콘솔 `/settings/save` — 폼에서 편입 블록을 새로 만들어 저장 seam 에 넘김. route `ANY /settings/save`(console.go:796).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `r` 폼 | enabled · exclude/include · 손절 % | 세션 + CSRF 통과 요청 | — |
| `c.opts.Settings` | 저장 seam(`Load` · `Save`) | cmd/tossctl `consoleAdoptionSettings` | nil → 501 |
| 불변식 | 폼에 없는 값을 저장이 바꾸면 안 됨(a208: `floor_only_exit`) | spec 「콘솔 저장이 보존한다」 | — |

## Branches and early returns

| Branch | Condition | At | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `if c.opts.Settings == nil` | 239:2 | — | refuse 501 (242:3) | `TestAnUnwiredSettingsSeamIsExplained` |
| B2 | `if err != nil` — `Load` 실패 | 245:2 | — | redirect (247:3) | 미측정 |
| B3 | `if err != nil` — 손절 % 해석 실패 | 258:2 | — | redirect (261:3) | `TestInvalidStopPercentWritesNothing` |
| B4 | `if err := c.opts.Settings.Save(next); err != nil` | 270:2 | — | redirect (272:3) | 미측정 |

Exact AST return positions: 242:3, 247:3, 261:3, 272:3

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `c.opts.Settings.Load` | 파일 읽기 가능 확인(값은 지금 버림 — claims C13) | 오류 → B2 | AST |
| `r.PostFormValue` · `splitSymbols` · `parseStopPercent` | 폼 해석 | 오류 → B3 | AST |
| `c.opts.Settings.Save` | 저장(+ cmd 쪽 저장 시점 audit) | 오류 → B4 | AST |
| `c.redirectSettings` · `c.refuse` · `effectNotice` | 응답 | — | AST |

Live config binding: 요청마다 seam 호출.

## State mutations and fallbacks

- seam 을 통한 config.json 쓰기 1 회(성공 시).

## Safety conclusion

- Safe edit boundary(a208): 버리던 `Load` 값을 `current` 로 받고 `next` 리터럴에 `FloorOnlyExit: current.FloorOnlyExit` 한 필드. 분기 · 반환 · 순서 무변. 화면 무변경.
- High-risk impact: no(설정 저장; 토글을 켜는 길을 만들지 않음 — 이월만).
