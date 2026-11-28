# Function Logic Map: `TestMain`

- Source: `cmd/tossctl/help_convention_test.go`
- AST evidence: `ast.json`(편집 뒤 `revision: current` 로 다시 뽑음). 편집 전 AST 는 같은 도구로 먼저 뽑아 아래 열거의 근거로 씀
  — 분기 `null` · return `null` · 호출 3 개(`i18n.SetLang` :17, `os.Exit` :18, `m.Run` :18). 편집 뒤 `ast.json`(revision current)은 주석 한 줄이 끼어 `i18n.SetLang` :17, `os.Exit` :19, `a210RunBoxed` :19.
- Risk scan: `risk-pattern-report.md`(일치 0)

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `m *testing.M` | go test 가 넘기는 값 | testing 패키지 | 없음 |
| 프로세스 env(`TOSSOS_DATA_DIR` · `XDG_DATA_HOME` · `XDG_CONFIG_HOME` · `XDG_CACHE_HOME`) | 편집 전: 개발자 값 그대로 / 편집 뒤: 시험 바이너리 전용 임시 상자 | `a210_isolated_home_test.go` `a210RunBoxed` | 상자 생성 · env 설정 실패면 시험을 돌리지 않고 rc 1 |

불변식: 이 패키지 시험이 기본 경로(`journal.DataDir` · `config.DefaultPaths`)로 떨어져도 개발자 HOME 아래(운영 audit · 운영 journal · credential)에 닿지 않는다.

## Branches and early returns

편집 전 AST 열거: 분기 0, early return 0. 순서 고정 호출 둘 — `i18n.SetLang("en")` 뒤 `os.Exit(m.Run())`.

편집: `os.Exit(m.Run())` 를 `os.Exit(a210RunBoxed(m))` 로 바꿈. `i18n.SetLang("en")` 은 그대로 먼저 돔. TestMain 자신에는 분기가 생기지 않음
(분기는 새 함수 `a210RunBoxed` · `a210PinGoDirs` 안 — 새 함수라 gate ⑤ 대상 아님).

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `i18n.SetLang` | 도움말 규약 시험이 영어 문구를 단언 | 반환값 없음 | AST :17 |
| `m.Run` → (편집 뒤) `a210RunBoxed(m)` | 시험 실행 → 상자 안에서 시험 실행 | 상자 준비 실패는 rc 1, 그 밖은 `m.Run` 의 rc 그대로 | 편집 전 AST :18(`m.Run`) → 편집 뒤 `ast.json` :19(`a210RunBoxed`), 새 파일 |
| `os.Exit` | rc 전달 | — | 편집 전 AST :18 → 편집 뒤 `ast.json` :19 |

## State mutations and fallbacks

- 편집 전: 없음.
- 편집 뒤: 프로세스 env 4 개를 상자 경로로, 표지 `A210_TEST_HOME_BOX` 를 상자 경로로 설정(자식 프로세스는 표지를 검증한 뒤 부모 상자를 재사용, 검증 실패면 rc 1), 자격 증명 · 언어 env(`testenv.Env` 나머지)는 빈 값, `GOCACHE` · `GOENV` 가 비어 있으면 go 기본 계산식 값으로 채움. 시험 끝에 상자를 지움.
  `t.Setenv` 를 쓰는 시험은 시험마다 그 위에 덮였다가 이 값으로 돌아옴.

## Safety conclusion

- Safe edit boundary: 시험 바이너리 시작 지점만. 생산 코드 · 생산 기본 경로는 그대로.
- High-risk impact: no — 주문 · 손절 · 원장 경로 아님. 시험이 운영 파일에 쓰던 부작용을 없앰.
