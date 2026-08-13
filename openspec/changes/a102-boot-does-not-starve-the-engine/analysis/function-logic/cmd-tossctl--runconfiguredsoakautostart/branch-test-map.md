# Branch Test Map: `runConfiguredSoakAutostart`

> **이 change는 이 함수를 편집하지 않는다** (D7). 아래는 현재 HEAD 기준 커버리지 실측이고,
> a102 이후에도 **기존 커버리지 그대로**다.

Source: `cmd/tossctl/soakautostart.go` (87-117). AST 기준 branches **7** / returns 6.
source SHA-256: `937b0c68762523ff02a39f0554371572c636308c431d9d19e42a7b7f66adab1c`

## 측정값

`go test ./cmd/tossctl -count=1 -covermode=set -coverprofile` (526건 통과, a102 편집 전).

| Branch | 위치 | 본문 실행 | 근거 블록 | 지는 테스트 | a102의 편집 |
|---|---|---|---|---|---|
| B1 | `:91` `load == nil` | yes | `91.17` count=**1** | `TestConfiguredSoakAutostartWithoutLoadWiringDoesNothing` | **없음** |
| B2 | `:95` 읽기 오류 | yes | `95.16` count=**1** | `TestConfiguredSoakAutostartReadFailureIsFailClosed` | **없음** |
| B3 | `:98` `!on` | yes | `98.9` count=**1** | `TestConfiguredSoakAutostartOffDoesNotStart` | **없음** |
| B4 | `:101` `start == nil` | yes | `101.18` count=**1** | `TestConfiguredSoakAutostartWithoutStartWiringIsVisible` | **없음** |
| B5 | `:106` `start()` 오류 | yes | `106.16` count=**1** | `TestConfiguredSoakAutostartSurvivesAStartFailure` | **없음** |
| B6 | `:108` 실패 note 있음 | yes | `108.36` count=**1** | `TestConfiguredSoakAutostartKeepsAFailedStartsOwnWords` | **없음** |
| B7 | `:113` 성공 note 공백 | yes | `113.35` count=**1** | `TestConfiguredSoakAutostartSaysSomethingWhenTheSeamIsSilent` | **없음** |

성공 경로(`:116`)는 `TestConfiguredSoakAutostartOnStartsExactlyOnce`가 진다.

**7개 분기 전부 본문 실행됨 — a101이 100.0%로 만들어 둔 자리다.**

## a102가 이 표에 더하는 것

없다. a102는 이 함수에 넘기는 **`start` 인자**만 바꾼다(대기로 감싼 closure). 그 closure의
분기는 `soakStartAfterEngineReady`·`awaitEngineReady`의 branch test map이 진다.

## 산출물 근거

- 분기 열거: `ast.json` (branches 7, returns 6) — `go run ./tools/logic-map`
- 커버리지: `go test ./cmd/tossctl -count=1 -covermode=set -coverprofile` exit 0
- 호출자 전수: `rg -n 'runConfiguredSoakAutostart'` → 선언 `soakautostart.go:87` ·
  호출 `console.go:378` · 테스트 8건
