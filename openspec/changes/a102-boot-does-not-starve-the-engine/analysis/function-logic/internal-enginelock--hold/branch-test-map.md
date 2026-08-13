# Branch Test Map: `Hold`

Source: `internal/enginelock/enginelock.go` (277-303). AST 기준 branches **3** / returns 4.

## 커버리지는 주장이 아니라 측정값이다

`go test ./internal/enginelock -count=1 -covermode=set -coverprofile`을 **편집 전과 후에
각각** 돌려 블록 카운트를 잘라 읽었다.

| | 통과 | `Hold` statements | `Hold` 블록 | source SHA-256 |
|---|---|---|---|---|
| 편집 전 (`:178-216`) | 11건 | 80.0% | 14개 중 9개 실행 | `d65deddfd1e6…` |
| **편집 후 (`:277-303`)** | **19건** | **86.7%** | **9개 중 7개 실행** | `8c784e84d88e…` |

블록이 14 → 9로 준 것은 stop 클로저가 `(*Held).Release`로 나갔기 때문이다 — 그 블록들은
사라진 것이 아니라 `Release`(83.3%)·`Ready`(100.0%)·`refresh`(83.3%)로 옮겨 갔다.

| Branch | 위치 | 본문 실행 | 근거 블록 (편집 후) | 지는 테스트 | 편집 전 |
|---|---|---|---|---|---|
| B1 | `:282` 첫 write 실패 | **yes** | `282.46,284.3` count=**1** | `TestAFailedHoldPublishesNothing` | count=0 (없던 커버리지) |
| B2 | `:290` `for {}` | yes | `290.7,291.11` count=**1** | `TestReleasingRemovesTheMarker` | count=1 |
| B3 | `:291` `select` 3-way | **부분** | 갈래표 참조 | 아래 | 같은 상태 |

| B3 갈래 | 근거 블록 (편집 후) | 지는 테스트 |
|---|---|---|
| `<-h.done` `:292` | `292.18,293.11` count=**1** | `TestReleasingRemovesTheMarker` |
| `<-ctx.Done()` `:294` | `294.22,295.11` count=**0** | 없음 (물려받은 공백) |
| `at := <-ticker.C` `:296` | `296.26,297.22` count=**0** | 없음 — **1분 ticker는 테스트가 닿을 수 없다** |

> **ticker 갈래의 공백은 a102가 이름을 붙여 메운 자리다.** 그 갈래가 하는 유일한 일은
> `h.refresh(at)` 한 줄이고, `refresh`는 **83.3%로 측정되며**
> `TestRefreshPreservesTheReadySignal`이 세 번의 갱신을 실제로 돌린다. 즉 "ticker가
> 언제 부르는가"는 여전히 미측정이고, **"ticker가 부르는 것이 무엇을 하는가"는 측정된다.**
> 이 분리가 D4를 테스트 가능하게 만든 편집이다.
>
> `<-ctx.Done()` 갈래는 편집 전과 같이 비어 있다 — a102의 범위가 아니다.
> **침묵하지 않고 이름을 붙여 남긴다.**

## a102가 더한 함수들의 커버리지 (측정값)

| 함수 | 위치 | statements | 지는 테스트 |
|---|---|---|---|
| `(*Held).Release` | `:211` | **83.3%** | `TestReleasingRemovesTheMarker` · `TestAFailedHoldPublishesNothing` |
| `(*Held).Ready` | `:231` | **100.0%** | `TestReadyPublishesTheSignalOnlyWhenItIsCalled` · `TestReadyIsIdempotentAndKeepsTheFirstInstant` · `TestAFailedHoldPublishesNothing` |
| `(*Held).refresh` | `:260` | **83.3%** | `TestRefreshPreservesTheReadySignal` |
| `Status.Ready` | `:350` | **100.0%** | `TestReadyRequiresBothALiveMarkerAndASignal` |

`Release`·`refresh`의 나머지 16.7%는 각각 `os.Remove` 실패 갈래(`217.81,222.4` count=0)와
inert 핸들 갈래에서 오는 write 오류 return(`261.25,263.3` count=0)이다.

## 뮤테이션 정산

design §검증계약 (d)와 자체 뮤테이션 둘을 이 묶음이 진다. 원복은 sha256 동일성으로 증명했다.

| 뮤테이션 | 가한 것 | 죽은 테스트 | 원복 |
|---|---|---|---|
| **(d)** | `refresh`가 쓰기 직전에 `current.ReadyAt = nil` | `TestRefreshPreservesTheReadySignal` | sha `8c784e84d88e` 동일 |
| (g) 자체 | `Status.Ready`에서 `s.Running &&` 제거 | `TestReadyRequiresBothALiveMarkerAndASignal` | sha `8c784e84d88e` 동일 |
| (i) 자체 | `Ready`의 멱등 early return 제거 | `TestReadyIsIdempotentAndKeepsTheFirstInstant` | sha `8c784e84d88e` 동일 |

> **(g)는 처음에 살아남았다.** 파일에서 읽은 Status만 보던 테스트로는 반증되지 않는데,
> `Read`가 stale 파일의 본문을 **파싱하기 전에** 돌아가므로(`if !fresh { return s }`)
> 그 경로의 `Marker`는 언제나 zero이고 `ReadyAt`도 nil이기 때문이다. 주장이 술어에 대한
> 것이므로 술어를 네 조합으로 고정하는 테스트를 새로 써서 죽였다
> (`TestReadyRequiresBothALiveMarkerAndASignal`). **통과는 증거가 아니다.**

## 산출물 근거

- 분기·이탈 열거: `ast.json` (branches 3, returns 4) — `go run ./tools/logic-map`
- 커버리지: `go test ./internal/enginelock -count=1 -covermode=set -coverprofile` exit 0 ·
  **19건 통과** · 패키지 86.6% · `-race` 포함 통과
- 호출자 전수: `rg -n 'enginelock.Hold'` → 프로덕션 1건(`cmd/tossctl/engine.go:239`),
  테스트 8건 (전부 2판 시그니처로 갱신)
