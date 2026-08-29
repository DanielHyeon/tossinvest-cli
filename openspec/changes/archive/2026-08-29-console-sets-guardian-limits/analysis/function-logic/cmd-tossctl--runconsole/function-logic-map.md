# Function Logic Map: `runConsole`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Narrative context: `../../function-logic-map.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `root` | nil 가능 | 호출자 | `configServiceFor`가 nil seam을 돌려주고 화면이 미배선을 말한다 |
| 경로 해석 5종 | 실패 가능 | 사용자 데이터 디렉터리 | verify/soak/attestation은 오류 반환, journal·engine 마커는 경고 후 빈 문자열 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `ctx == nil` | `ctx = context.Background()` | 없음 | 기존 |
| B2 | `resolveVerifyRecord` 오류 | 없음 | 오류 반환 | 기존 |
| B3 | `resolveVerifyRecordFor` 오류 | 없음 | 오류 반환 | 기존 |
| B4 | `resolveSoakRecord` 오류 | 없음 | 오류 반환 | 기존 |
| B5 | `resolveSoakAttestationPath` 오류 | 없음 | 오류 반환 | 기존 |
| B6 | `journal.DefaultPath` 오류 | 경고 출력, `journalPath=""` | 계속 | 기존 |
| B7 | `engineJournalDir` 성공 | 마커 경로 설정 | 계속 | 기존 |
| B8 | `engineJournalDir` 실패 | 경고 출력, 마커 빈 문자열 | 계속 | 기존 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `consoleSettingsSeam` | 편입 설정 seam | nil이면 화면이 미배선 | CodeGraph + AST |
| `consoleGateLimitsSeam` | 개요의 한도 **읽기** seam | nil이면 개요가 미배선 | CodeGraph + AST |
| `consoleLimitSettingsSeam` | 한도 **편집** seam (이 change가 추가) | nil이면 한도 폼 미렌더 | CodeGraph + AST |
| `console.Run` | 서버 기동 | 오류 반환 | CodeGraph + AST |

## State mutations and fallbacks

- 이 change의 변경은 `Options`에 `Limits:` 한 줄을 더한 것뿐이다. 기존 8개 분기와 그 순서·처리는 무변경이다.
- 읽기 seam(`GateLimits`)과 편집 seam(`Limits`)을 **둘 다** 넘긴다. 같은 `configServiceFor(root)`가 두 seam의 파일을 해석하므로, 개요가 보여주는 값과 설정 화면이 편집하는 값은 같은 파일이다.

## Safety conclusion

- Safe edit boundary: `Options` 리터럴 한 줄. 분기 로직 무변경.
- High-risk impact: no — 배선만 한다. 넘기는 편집 seam의 위험은 그 seam(`consoleLimitSettings`)과 config writer의 map이 다룬다.
