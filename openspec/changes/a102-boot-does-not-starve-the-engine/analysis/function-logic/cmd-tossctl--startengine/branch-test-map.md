# Branch Test Map: `startEngine`

> **이 change는 이 함수를 편집하지 않는다** (D7 · design 「범위 밖」). 아래는 현재 HEAD 기준
> 커버리지 실측이고, a102 이후에도 **기존 커버리지 그대로**다.

Source: `cmd/tossctl/engineproc.go` (177-227). AST 기준 branches **7** / returns 8.
source SHA-256: `502362c0b925e09f57b5253ad92e1ce1bbcfe0d1f601869e76c44ff6f9d13439`

## 측정값

`go test ./cmd/tossctl -count=1 -covermode=set -coverprofile` (526건 통과, a102 편집 전).

| Branch | 위치 | 본문 실행 | 근거 블록 | 지는 테스트 | a102의 편집 |
|---|---|---|---|---|---|
| B1 | `:179` 저널 경로 오류 | **no** | `179.16` count=**0** | 없음(물려받은 공백) | **없음** |
| B2 | `:183` 실행 파일 경로 오류 | **no** | `183.16` count=**0** | 없음(물려받은 공백) | **없음** |
| B3 | `:199` 마커+증거 거절 | yes | `200.45` count=**1** | `TestAFreshMarkerWithALiveProcessStillRefuses` · `TestEnumerationFailureKeepsTheRefusal` | **없음** |
| B4 | `:204` 프로세스 관측됨 | yes | `204.14` count=**1** | `TestStartingIsRefusedWhenAProcessIsAlreadyThere` | **없음** |
| B5 | `:210` spawn 실패 | **no** | `210.16` count=**0** | 없음(물려받은 공백) | **없음** |
| B6 | `:217` probe select | yes | `217.2` count=**1** | `TestStartingSpawnsTheEngineWithThisProfilesConfigDir`(3초 통과 갈래) · `TestARefusedStartReportsTheEnginesOwnLog`(즉시 종료 갈래) | **없음** |
| B7 | `:220` 오류 없이 즉시 종료 | **no** | `220.21` count=**0** | 없음(물려받은 공백) | **없음** |

B6의 두 갈래 근거 블록: `218.25,220.21` count=**1**(`<-wait`) ·
`224.38,224.38` count=**1**(`<-time.After`).

**7개 분기 중 3개 본문 실행, 4개 미실행 — 전부 a102 이전부터의 공백이다.**
a102는 이 함수를 편집하지 않으므로 그 공백을 메우지도, 넓히지도 않는다.
**침묵하지 않고 이름을 붙여 남긴다.**

## a102가 이 표를 근거로 쓴 주장

proposal의 "부팅 자동 시작이 3초 뒤 돌아온다"는 **B6의 `time.After(engineStartProbe)`
갈래**가 근거다 — 그 갈래는 위에서 count=1로 실행이 확인된다
(`TestStartingSpawnsTheEngineWithThisProfilesConfigDir`). 손으로 읽은 주장이 아니다.

## 산출물 근거

- 분기 열거: `ast.json` (branches 7, returns 8) — `go run ./tools/logic-map`
- 커버리지: `go test ./cmd/tossctl -count=1 -covermode=set -coverprofile` exit 0
- 호출자 전수: `rg -n 'startEngine\('` → 선언 `engineproc.go:177` ·
  `console.go:343`(부팅) · `console.go:520`(버튼) · 테스트 8건
