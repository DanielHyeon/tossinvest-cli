# Function Logic Map: `configServiceFor`

- Source: `cmd/tossctl/adoptionsettings.go`
- AST evidence: `ast.json` (revision=current, L41–54, 분기 4개)
- Risk scan: `risk-pattern-report.md`
- 이 change의 base: `47672c6f` — base에는 이 함수가 없었다. 현재 본문을 diff hunk가 덮으므로 evidence가 요구된다 (revision=current)

프로필별 config 파일 경로를 해석해 `*config.Service`를 만든다. **한도를 보여주는 화면과 편입 블록을 편집하는 화면이 서로 다른 파일을 읽는 일이 없도록** 두 seam이 이 함수 하나를 공유한다.

이 함수 자체는 아무것도 콘솔에 넘기지 않는다 — `*config.Service`는 `consoleGateLimitsSeam`과 `newAdoptionSettingsSeam` 안에 갇히고, 콘솔에는 각각 숫자 5개와 `Load`/`Save` 두 메서드만 나간다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `root.configDir` | 빈 문자열 또는 디렉터리 | `--config-dir` | 공백만 있으면 기본 경로로 폴백 |
| `config.DefaultPaths()` | 성공/실패 | OS 경로 규칙 | 실패하면 `nil` — 두 seam 모두 미배선 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `switch {}` 진입 | — | — | `TestTheSeamSavesAuditsAndPreservesTheFile` |
| B2 | `root != nil && TrimSpace(root.configDir) != ""` | `path = <dir>/config.json` | 계속 | `TestTheConsoleIsHandedTheLimitsAsNumbersAndNoWayToWriteThem` (`configDir: t.TempDir()`) |
| B3 | default — 기본 경로 | `paths.ConfigFile` | 계속 | `TestTheSeamSavesAuditsAndPreservesTheFile` |
| B4 | `config.DefaultPaths()` 실패 | 없음 | `nil` — seam 미배선 | `TestATypedNilSeamNeverReachesTheInterface` (nil 전파 형태) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.TrimSpace` | 공백뿐인 `--config-dir`를 지정으로 오인하지 않는다 | — | L44 |
| `filepath.Join` | 프로필 경로 조립 | — | L45 |
| `config.DefaultPaths` | 기본 경로 | 에러는 nil 반환으로 흡수 | L47 |
| `config.NewService` | 서비스 생성 | 파일 없음은 여기서 실패가 아니다 — Load 시점의 문제 | L53 |

## State mutations and fallbacks

- 파일을 만들지도 읽지도 않는다. 경로 결정만 한다.
- nil 반환이 "미배선"의 유일한 신호이고, 두 seam이 그것을 구체 타입에서 확인한다.

## Safety conclusion

- Safe edit boundary: 경로 결정 규칙. 두 seam이 이 함수를 공유한다는 사실이 깨지면 한 화면이 다른 파일을 읽는다.
- High-risk impact: yes (Guardian·손절 파라미터 경로) — 이 함수가 고르는 파일이 자동매매 게이트 한도와 편입 블록의 `default_stop_pct`를 담는다. 잘못된 파일을 고르면 개요는 존재하지 않는 한도를 보여주고 편입 저장은 엔진이 읽지 않는 파일에 쓴다. 주문 능력은 넘기지 않는다.
