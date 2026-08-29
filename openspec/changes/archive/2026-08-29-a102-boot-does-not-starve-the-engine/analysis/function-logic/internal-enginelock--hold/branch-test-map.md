# Branch Test Map: `Hold`

Source: `internal/enginelock/enginelock.go` (361-389). AST 기준 branches **3** / returns 4.

## 커버리지는 주장이 아니라 측정값이다

`go test ./internal/enginelock -count=1 -covermode=set -coverprofile`을 **편집 전과 후에
각각** 돌려 블록 카운트를 잘라 읽었다.

| | 통과 | `Hold` statements | `Hold` 블록 | source SHA-256 |
|---|---|---|---|---|
| 1판 (`:178-216`) | 11건 | 80.0% | 14개 중 9개 실행 | `d65deddfd1e6…` |
| 2판 (`:277-303`) | 19건 | 86.7% | 9개 중 7개 실행 | `8c784e84d88e…` |
| 3판 (`:296-324`, §3.9) | 24건 | 88.2% | 9개 중 7개 실행 | `9399c1d68e1d…` |
| **4판 (`:361-389`, §3.9c)** | **27건** | **88.2%** | **9개 중 7개 실행** | **`e615b0c66ce3…`** |

블록이 14 → 9로 준 것은 stop 클로저가 `(*Held).Release`로 나갔기 때문이다 — 그 블록들은
사라진 것이 아니라 `Release`(83.3%)·`Ready`(100.0%)·`refresh`(83.3%)로 옮겨 갔다.

| Branch | 위치 | 본문 실행 | 근거 블록 (편집 후) | 지는 테스트 | 편집 전 |
|---|---|---|---|---|---|
| B1 | `:366` 첫 write 실패 | **yes** | `366.46,368.3` count=**1** | `TestAFailedHoldPublishesNothing` | count=0 (없던 커버리지) |
| B2 | `:376` `for {}` | yes | `376.7,377.11` count=**1** | `TestReleasingRemovesTheMarker` | count=1 |
| B3 | `:377` `select` 3-way | **부분** | 갈래표 참조 | 아래 | 같은 상태 |

| B3 갈래 | 근거 블록 (편집 후) | 지는 테스트 |
|---|---|---|
| `<-h.done` `:378` | `378.18,379.11` count=**1** | `TestReleasingRemovesTheMarker` |
| `<-ctx.Done()` `:380` | `380.22,381.11` count=**0** | 없음 (물려받은 공백) |
| `at := <-ticker.C` `:382` | `382.26,383.22` count=**0** | 없음 — **1분 ticker는 테스트가 닿을 수 없다** |

> **ticker 갈래의 공백은 a102가 이름을 붙여 메운 자리다.** 그 갈래가 하는 유일한 일은
> `h.refresh(at)` 한 줄이고, `refresh`는 **83.3%로 측정되며**
> `TestRefreshPreservesTheReadySignal`이 세 번의 갱신을 실제로 돌린다. 즉 "ticker가
> 언제 부르는가"는 여전히 미측정이고, **"ticker가 부르는 것이 무엇을 하는가"는 측정된다.**
> 이 분리가 D4를 테스트 가능하게 만든 편집이다.
>
> `<-ctx.Done()` 갈래는 편집 전과 같이 비어 있다 — a102의 범위가 아니다.
> **침묵하지 않고 이름을 붙여 남긴다.**

## a102가 더한 함수들의 커버리지 (측정값)

| 함수 | 위치 | statements (3판) | 지는 테스트 |
|---|---|---|---|
| `(*Held).Release` | `:275` | **83.3%** | `TestReleasingRemovesTheMarker` · `TestAFailedHoldPublishesNothing` · `TestReleaseIsNotUndoneByARefreshAlreadyInFlight` |
| `(*Held).Ready` | `:308` | **88.9%** | `TestReadyPublishesTheSignalOnlyWhenItIsCalled` · `TestReadyIsIdempotentAndKeepsTheFirstInstant` · `TestAFailedHoldPublishesNothing` |
| `(*Held).refresh` | `:342` | **85.7%** | `TestRefreshPreservesTheReadySignal` · `TestReleaseIsNotUndoneByARefreshAlreadyInFlight` |
| `write` | `:405` | **53.6%** | 별도 묶음 `internal-enginelock--write` |
| `Status.Ready` | `:473` | **100.0%** | `TestReadyRequiresBothALiveMarkerAndASignal` |

`Release`의 나머지는 `os.Remove` 실패 갈래, `Ready`·`refresh`의 나머지는 inert 핸들에서
오는 조기 반환과 `write` 오류 반환이다 — 전부 물려받은 공백이고 §3.9가 넓히지 않았다.
`write`의 53.6%는 그 함수의 여덟 실패 분기가 파일시스템 고장을 요구하기 때문이다
(`internal-enginelock--write/branch-test-map.md`에 이름을 붙여 남겼다).

## 뮤테이션 정산

design §검증계약 (d)와 자체 뮤테이션 둘을 이 묶음이 진다. 원복은 sha256 동일성으로 증명했다.

| 뮤테이션 | 가한 것 | 죽은 테스트 | 원복 |
|---|---|---|---|
| **(d)** | `refresh`가 쓰기 직전에 `ReadyAt = nil` | `TestRefreshPreservesTheReadySignal` · `TestTheMarkerIsNeverReadHalfWritten` | sha `e615b0c66ce3` 동일 |
| (g) 자체 | `Status.Ready`에서 `s.Running &&` 제거 | `TestReadyRequiresBothALiveMarkerAndASignal` | sha `e615b0c66ce3` 동일 |
| (i) 자체 | `Ready`의 멱등 early return 제거 | `TestReadyIsIdempotentAndKeepsTheFirstInstant` | sha `e615b0c66ce3` 동일 |
| **(l)** §3.9 | 원자 교체를 `os.WriteFile`로 되돌린다 (A2 F3) | `TestTheMarkerIsNeverReadHalfWritten` | sha `e615b0c66ce3` 동일 |
| **(m)** §3.9 | `Release`가 `live`를 끄지 않는다 (A2 F4) | `TestReleaseIsNotUndoneByARefreshAlreadyInFlight` · `TestReleaseWinsAgainstConcurrentRefreshes` | sha `e615b0c66ce3` 동일 |

> **(g)는 처음에 살아남았다.** 파일에서 읽은 Status만 보던 테스트로는 반증되지 않는데,
> `Read`가 stale 파일의 본문을 **파싱하기 전에** 돌아가므로(`if !fresh { return s }`)
> 그 경로의 `Marker`는 언제나 zero이고 `ReadyAt`도 nil이기 때문이다. 주장이 술어에 대한
> 것이므로 술어를 네 조합으로 고정하는 테스트를 새로 써서 죽였다
> (`TestReadyRequiresBothALiveMarkerAndASignal`). **통과는 증거가 아니다.**

## 산출물 근거

- 분기·이탈 열거: `ast.json` (branches 3, returns 4) — `go run ./tools/logic-map`
- 커버리지: `go test ./internal/enginelock -count=1 -covermode=set -coverprofile` exit 0 ·
  **27건 통과** · 패키지 79.8% · `-race` 포함 통과 (패키지 수치가 내려간 것은 `write`의
  실패 분기 다섯이 새로 생겼기 때문이다 — 기존 함수의 커버리지는 전부 올랐다)
- 호출자 전수: `rg -n 'enginelock.Hold'` → 프로덕션 1건(`cmd/tossctl/engine.go:239`),
  테스트 8건 (전부 2판 시그니처로 갱신)
