# Branch Test Map: `bootSurvey`

> **이 change는 이 함수를 편집하지 않는다** (D7). 아래는 현재 HEAD 기준 커버리지 실측이고,
> a102 이후에도 **기존 커버리지 그대로**다.

Source: `cmd/tossctl/soakautostart.go` (142-151). AST 기준 branches **2** / returns 3.
source SHA-256: `937b0c68762523ff02a39f0554371572c636308c431d9d19e42a7b7f66adab1c`

## 측정값

`go test ./cmd/tossctl -count=1 -covermode=set -coverprofile` (526건 통과, a102 편집 전).

| Branch | 위치 | 본문 실행 | 근거 블록 | 지는 테스트 | a102의 편집 |
|---|---|---|---|---|---|
| B1 | `:144` `running()` 오류 | yes | `144.16` count=**1** | `TestBootSurveyDoesNotStartWhenItCannotTell` | **없음** |
| B2 | `:147` `len(pids) > 0` | yes | `147.19` count=**1** | `TestBootSurveyLeavesARunningSurveyAlone` · `TestBootSurveyNamesEveryProcessItLeftAlone` | **없음** |

else 경로(`:150` `return start()`)는 `TestBootSurveyStartsOneWhenNoneIsRunning` ·
`TestBootSurveyReportsAFailedStart`가 진다.

**2개 분기 전부 본문 실행됨 — 이 함수는 a101이 100.0%로 만들어 둔 자리다.**

## a102가 이 표에 더하는 것

없다. a102의 대기는 이 함수를 **감싸는** closure 밖에 있고, 이 함수의 입력·분기·반환은
바뀌지 않는다. 대기의 네 verdict는 `awaitEngineReady`의 branch test map이 진다.

## 산출물 근거

- 분기 열거: `ast.json` (branches 2, returns 3) — `go run ./tools/logic-map`
- 커버리지: `go test ./cmd/tossctl -count=1 -covermode=set -coverprofile` exit 0
- 호출자 전수: `rg -n 'bootSurvey\('` → 선언 `soakautostart.go:142` ·
  호출 `console.go:373`(closure) · 테스트 5건
