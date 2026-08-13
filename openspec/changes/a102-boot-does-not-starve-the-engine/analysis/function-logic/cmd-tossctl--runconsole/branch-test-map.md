# Branch Test Map: `runConsole`

Source: `cmd/tossctl/console.go` (212-545). AST 기준 branches **44** / returns 22.

## 커버리지는 주장이 아니라 측정값이다

`go test ./cmd/tossctl -count=1 -covermode=set -coverprofile`을 **편집 전과 후에 각각** 돌려
블록 카운트를 잘라 읽었다. 분기의 *조건*이 아니라 **분기 본문 블록의 실행 여부**를 셌다
(같은 줄에서 시작하는 블록 중 시작 column이 가장 큰 것이 본문이다 — a101의 규칙 그대로).

| | 통과 | 위치 | 분기 | return | 블록 | source SHA-256 |
|---|---|---|---|---|---|---|
| 편집 전 | 526건 | `:211-529` | 44 | 21 | 84개 중 **0개** 실행 | `08562e7541dd…` |
| **편집 후** | **552건** | **`:212-545`** | **44** | **22** | **87개 중 0개** 실행 | `949b8a26cbc6…` |

**분기 수가 44 그대로다.** a101이 세 번 측정해 0.0%를 확인한 사실도 그대로다 — 이 패키지의
테스트는 `console.ListenAndServe`와 각 seam을 직접 구동하고, **콘솔 조립 함수 자체는 어느
테스트도 부르지 않는다.** 늘어난 것은 goroutine 하나(`go` 문 0 → **1**), return 하나,
그리고 그 goroutine이 만드는 블록들뿐이다.

| Branch | 위치 | 본문 실행 | 근거 블록 (편집 후) |
|---|---|---|---|
| B1 | `:214` if | **no** | `214.16` count=**0** |
| B2 | `:223` if | **no** | `223.16` count=**0** |
| B3 | `:228` if | **no** | `228.16` count=**0** |
| B4 | `:232` if | **no** | `232.16` count=**0** |
| B5 | `:236` if | **no** | `236.16` count=**0** |
| B6 | `:240` if | **no** | `240.16` count=**0** |
| B7 | `:244` if | **no** | `244.16` count=**0** |
| B8 | `:249` if | **no** | `249.16` count=**0** |
| B9 | `:257` if | **no** | `257.23` count=**0** |
| B10 | `:259` if | **no** | `259.17` count=**0** |
| B11 | `:262` else | **no** | `262.9` count=**0** |
| B12 | `:266` if | **no** | `266.17` count=**0** |
| B13 | `:269` else | **no** | `269.9` count=**0** |
| B14 | `:280` if | **no** | `280.54` count=**0** |
| B15 | `:283` else | **no** | `283.8` count=**0** |
| B16 | `:291` if | **no** | `291.42` count=**0** |
| B17 | `:294` else | **no** | `294.59` count=**0** |
| B18 | `:294` if | **no** | `294.59` count=**0** |
| B19 | `:296` else | **no** | `296.8` count=**0** |
| B20 | `:298` if | **no** | `298.18` count=**0** |
| B21 | `:306` else | **no** | `306.9` count=**0** |
| B22 | `:300` if | **no** | `300.59` count=**0** |
| B23 | `:302` else | **no** | `302.10` count=**0** |
| B24 | `:309` if | **no** | `309.22` count=**0** |
| B25 | `:313` if | **no** | `313.19` count=**0** |
| B26 | `:315` else | **no** | `315.10` count=**0** |
| B27 | `:322` if | **no** | `322.21` count=**0** |
| B28 | `:325` if | **no** | `325.18` count=**0** |
| B29 | `:339` if | **no** | `339.23` count=**0** |
| B30 | `:346` if | **no** | `346.26` count=**0** |
| B31 | `:367` if | **no** | `367.21` count=**0** |
| B32 | `:393` if | **no** | `393.74` count=**0** ← **a102의 편집 지점 (goroutine 안)** |
| B33 | `:403` if | **no** | `403.21` count=**0** |
| B34 | `:405` if | **no** | `405.60` count=**0** |
| B35 | `:417` else | **no** | `417.49` count=**0** |
| B36 | `:407` if | **no** | `407.22` count=**0** |
| B37 | `:409` else | **no** | `409.10` count=**0** |
| B38 | `:417` if | **no** | `417.49` count=**0** |
| B39 | `:421` if | **no** | `421.64` count=**0** |
| B40 | `:428` else | **no** | `428.49` count=**0** |
| B41 | `:423` if | **no** | `423.22` count=**0** |
| B42 | `:425` else | **no** | `425.10` count=**0** |
| B43 | `:428` if | **no** | `428.49` count=**0** |
| B44 | `:514` if | **no** | `514.23` count=**0** |

**편집 전·후 모두: 44개 중 0개 실행.**

## a102가 여기 놓지 **않은** 것

D7이 요구하는 판정은 넷이다 — 엔진 준비 관측, 엔진 부재 판정, 상한 초과 판정, 콘솔 종료
판정. 넷 다 여기 썼으면 **분기가 4개 늘고 그 넷의 커버리지는 0.0%**였을 것이다. 전부
`awaitEngineReady`로 나갔고, `runConsole`에는 **호출과 출력만** 남았다. 그 결과 분기 수는
44 그대로이고, 늘어난 것은 대기를 화면에서 떼는 `go` 문 하나다.

| 판정 | 어디서 측정되나 | 측정값 (편집 후) |
|---|---|---|
| 준비 확인 / 엔진 없음 / 상한 초과 / 콘솔 종료 | `awaitEngineReady` | **100.0%** |
| verdict → 운영자 노트 문장 | `engineReadyNote` | **100.0%** |
| 대기를 start seam 앞에 붙이는 배선 | `soakStartAfterEngineReady` | **100.0%** |
| 승인 없음 / 읽기 실패 / 배선 없음 / 기동 실패 / 빈 note | `runConfiguredSoakAutostart` (a101, 무편집) | 7개 분기 전부 실행 |
| 이미 돌고 있으면 두기 / 열거 실패는 "없음"이 아님 | `bootSurvey` (a101, 무편집) | 2개 분기 전부 실행 |

## 측정으로 보장되지 않는 것 (숨기지 않는다)

`runConsole`이 0.0%이므로 **이 함수 안의 배선 자체**는 테스트가 아니라 리뷰 조건이다.
a101이 남긴 목록에 a102가 셋을 더했고, 그 셋은 **소스 형태 회귀 테스트**로 고정했다
(`TestTheRestartRecoveryRunsBeforeTheLoops`가 이미 쓰는 기법과 같다).

| 보장할 수 없는 것 | 어떻게 대신 고정했나 |
|---|---|
| soak 자동 시작 블록이 goroutine 안에 있다 (spec: 대기가 콘솔을 막지 않는다) | `TestTheSoakAutostartWaitsOffTheConsolePath` — `go func() {`와 `soakStartAfterEngineReady(` 의 위치 관계, 그리고 `console.ListenAndServe`가 그 뒤에 오는 것 |
| 대기에 넘기는 ctx가 콘솔의 것이다 | 같은 테스트 — goroutine 블록 안에 `soakStartAfterEngineReady(ctx,` 가 있는지 |
| 버튼 경로는 대기 뒤에 있지 않다 | 같은 테스트 — `RestartSoak:` 이후 소스에 `soakStartAfterEngineReady(` 가 없는지 |
| (a101) 엔진 autostart보다 뒤 · 실패해도 return하지 않음 · 부팅 경로는 `PrepareSpawn`을 넘기지 않음 | 리뷰 조건 그대로 (a101 `pre-edit-gate.md`) |

**소스 형태 테스트는 실행 증거가 아니다.** 그것이 고정하는 것은 "이 배선이 여전히 이
모양이다"뿐이고, goroutine이 실제로 콘솔을 막지 않는다는 것은 이 패키지에서 실행으로
관측되지 않는다. **침묵하지 않고 이름을 붙여 남긴다.**

## 산출물 근거

- 분기·이탈 열거: `ast.json` (branches 44, returns 22, go 1) — `go run ./tools/logic-map`
- 커버리지: `go test ./cmd/tossctl -count=1 -covermode=set -coverprofile` exit 0 ·
  **526건 통과**(편집 전) → **552건 통과**(편집 후) · `runConsole` **0.0%** (두 번 다)
