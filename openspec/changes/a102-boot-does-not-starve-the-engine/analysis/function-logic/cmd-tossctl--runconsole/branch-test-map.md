# Branch Test Map: `runConsole`

Source: `cmd/tossctl/console.go` (211-539). AST 기준 branches **43** / returns 21.

## 커버리지는 주장이 아니라 측정값이다

`go test ./cmd/tossctl -count=1 -covermode=set -coverprofile`을 **판마다** 돌려 블록
카운트를 잘라 읽었다. 분기의 *조건*이 아니라 **분기 본문 블록의 실행 여부**를 셌다
(같은 줄에서 시작하는 블록 중 시작 column이 가장 큰 것이 본문이다 — a101의 규칙 그대로).

| | 통과 | 위치 | 분기 | return | `go` | 블록 | source SHA-256 |
|---|---|---|---|---|---|---|---|
| 1판 (편집 전) | 526건 | `:211-529` | 44 | 21 | 0 | 84개 중 **0개** 실행 | `08562e7541dd…` |
| 2판 (`6cd643ca`) | 552건 | `:212-545` | 44 | 22 | 1 | 87개 중 **0개** 실행 | `949b8a26cbc6…` |
| **3판 (§3.9, 이 문서)** | **564건** | **`:211-539`** | **43** | **21** | **0** | **84개 중 0개** 실행 | **`ee2c242c2804…`** |

**3판이 분기를 하나 더 줄였다.** 2판은 goroutine 안에 `if note != ""`를 남겨 뒀는데,
§3.9(D7c)가 부팅 블록 전체를 `startSoakAutostartAsync`로 옮기면서 그 판정도 함께 나갔다.
`go` 문도 이 함수에서 사라졌다 — 비동기는 이제 그 함수의 성질이고, **실행으로 단언된다**
(`TestTheBootSurveyNeverBlocksTheConsole`). 편집 전(1판)과 비교하면 이 change가 이 함수에
남긴 것은 **분기 하나 감소**뿐이다.

커버리지는 세 번 다 **0.0%**다 — a101이 세 번 측정해 확인한 사실 그대로다. 이 패키지의
테스트는 `console.ListenAndServe`와 각 seam을 직접 구동하고, **콘솔 조립 함수 자체는 어느
테스트도 부르지 않는다.**

| Branch | 위치 | 본문 실행 | 근거 블록 (3판) |
|---|---|---|---|
| B1 | `:213` if | **no** | `213.16` count=**0** |
| B2 | `:222` if | **no** | `222.16` count=**0** |
| B3 | `:227` if | **no** | `227.16` count=**0** |
| B4 | `:231` if | **no** | `231.16` count=**0** |
| B5 | `:235` if | **no** | `235.16` count=**0** |
| B6 | `:239` if | **no** | `239.16` count=**0** |
| B7 | `:243` if | **no** | `243.16` count=**0** |
| B8 | `:248` if | **no** | `248.16` count=**0** |
| B9 | `:256` if | **no** | `256.23` count=**0** |
| B10 | `:258` if | **no** | `258.17` count=**0** |
| B11 | `:261` else | **no** | `261.9` count=**0** |
| B12 | `:265` if | **no** | `265.17` count=**0** |
| B13 | `:268` else | **no** | `268.9` count=**0** |
| B14 | `:279` if | **no** | `279.54` count=**0** |
| B15 | `:282` else | **no** | `282.8` count=**0** |
| B16 | `:290` if | **no** | `290.42` count=**0** |
| B17 | `:293` else | **no** | `293.59` count=**0** |
| B18 | `:293` if | **no** | `293.59` count=**0** |
| B19 | `:295` else | **no** | `295.8` count=**0** |
| B20 | `:297` if | **no** | `297.18` count=**0** |
| B21 | `:305` else | **no** | `305.9` count=**0** |
| B22 | `:299` if | **no** | `299.59` count=**0** |
| B23 | `:301` else | **no** | `301.10` count=**0** |
| B24 | `:308` if | **no** | `308.22` count=**0** |
| B25 | `:312` if | **no** | `312.19` count=**0** |
| B26 | `:314` else | **no** | `314.10` count=**0** |
| B27 | `:321` if | **no** | `321.21` count=**0** |
| B28 | `:324` if | **no** | `324.18` count=**0** |
| B29 | `:338` if | **no** | `338.23` count=**0** |
| B30 | `:345` if | **no** | `345.26` count=**0** |
| B31 | `:366` if | **no** | `366.21` count=**0** |
| B32 | `:394` if | **no** | `394.21` count=**0** |
| B33 | `:396` if | **no** | `396.60` count=**0** |
| B34 | `:408` else | **no** | `408.49` count=**0** |
| B35 | `:398` if | **no** | `398.22` count=**0** |
| B36 | `:400` else | **no** | `400.10` count=**0** |
| B37 | `:408` if | **no** | `408.49` count=**0** |
| B38 | `:412` if | **no** | `412.64` count=**0** |
| B39 | `:419` else | **no** | `419.49` count=**0** |
| B40 | `:414` if | **no** | `414.22` count=**0** |
| B41 | `:416` else | **no** | `416.10` count=**0** |
| B42 | `:419` if | **no** | `419.49` count=**0** |
| B43 | `:508` if | **no** | `508.23` count=**0** |

**세 판 모두: 전 분기 0개 실행.**

## a102가 여기 놓지 **않은** 것

D7이 요구하는 판정은 넷(엔진 준비 관측 / 엔진 부재 / 상한 초과 / 콘솔 종료)이고, §3.9가
셋을 더했다(시체 마커 판별 / spawn 직렬화 / 비동기 자체). **일곱 다 여기 썼으면 그만큼의
분기가 0.0%로 늘었을 것이다.** 전부 `cmd/tossctl/engineready.go`로 나갔고, `runConsole`에는
**호출 하나**가 남았다.

| 판정 | 어디서 측정되나 | 측정값 (§3.9 후) |
|---|---|---|
| 준비 확인 / 엔진 없음 / 상한 초과 / 콘솔 종료 | `awaitEngineReady` | **100.0%** |
| 시체 마커는 준비가 아니다 (A2 F1) | `readEngineSignal` | **100.0%** |
| 마커 + 프로세스 목록의 합성 (A2 F2/N2) | `consoleEngineReadiness` | **100.0%** |
| verdict → 운영자 노트 문장 | `engineReadyNote` | **100.0%** |
| 대기를 start seam 앞에 붙이는 배선 | `soakStartAfterEngineReady` | **100.0%** |
| 부팅·버튼 spawn 직렬화 (A2 F5) | `spawnOneSurvey` · `guardedSoakRestart` | **100.0%** |
| 대기가 콘솔을 막지 않는다 (A2 F6/N4) | `startSoakAutostartAsync` | **100.0%** |
| 승인 없음 / 읽기 실패 / 배선 없음 / 기동 실패 / 빈 note | `runConfiguredSoakAutostart` (a101, 무편집) | 7개 분기 전부 실행 |
| 이미 돌고 있으면 두기 / 열거 실패는 "없음"이 아님 | `bootSurvey` (a101, 무편집) | 2개 분기 전부 실행 |

## 측정으로 보장되지 않는 것 (숨기지 않는다)

`runConsole`이 0.0%이므로 **이 함수가 그 호출을 한다는 사실 자체**는 실행으로 관측되지 않는다.
2판은 그 자리를 소스 형태 다섯 검사로 메웠고, **A2가 그것을 뚫었다** — `go func() {`도, ctx
인자도, `ListenAndServe`와의 순서도 전부 유지한 채 goroutine을 join해 콘솔을 상한만큼
막는 판본이 다섯 검사를 모두 통과했다(N4).

§3.9는 그래서 형태 검사를 **두 줄로 줄이고** 내용을 실행으로 옮겼다.

| 보장할 수 없는 것 | 어떻게 대신 고정했나 |
|---|---|
| runConsole이 부팅 블록을 위임한다 | `TestTheConsoleDelegatesTheWholeBootSurveyBlock` — 호출 두 줄만 본다 |
| 그 위임이 콘솔을 막지 않는다 | **실행**: `TestTheBootSurveyNeverBlocksTheConsole` (영원히 안 돌아오는 start를 줘도 바깥 호출이 즉시 반환) |
| 버튼과 부팅이 동시에 spawn하지 않는다 | **실행**: `TestTheBootPathAndTheButtonCannotSpawnAtTheSameTime` (첫 경로를 붙잡아 둔 채 둘째가 못 들어오는 것을 관측) |
| (a101) 엔진 autostart보다 뒤 · 실패해도 return하지 않음 · 부팅 경로는 `PrepareSpawn`을 넘기지 않음 | 리뷰 조건 그대로 (a101 `pre-edit-gate.md`) |

**소스 형태 테스트는 실행 증거가 아니다.** 남은 두 줄이 고정하는 것은 "이 두 호출이 여전히
거기 있다"뿐이다. **침묵하지 않고 이름을 붙여 남긴다.**

## 산출물 근거

- 분기·이탈 열거: `ast.json` (branches 43, returns 21, go 0) — `go run ./tools/logic-map`
- 커버리지: `go test ./cmd/tossctl -count=1 -covermode=set -coverprofile` exit 0 ·
  526건(1판) → 552건(2판) → **564건**(3판) · `runConsole` **0.0%** (세 번 다)
