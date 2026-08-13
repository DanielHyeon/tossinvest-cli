# Function Logic Map: `Hold`

- Source: `internal/enginelock/enginelock.go` (296-324)
- AST evidence: `ast.json` — AST 기준 branches **3** / returns 4 / calls 12 / defers 1 / go 1
- Risk scan: `risk-pattern-report.md` (매치 없음)
- source SHA-256: `9399c1d68e1dfad16f6b1334f9875897e8025a88b94beb370354ffb11f318f8d`
- 작성 사유: a102 §3(D4) — 반환을 핸들로 바꾸고 `ReadyAt`을 도입한다. **기존 함수의 내부를
  편집하므로 편집 전에 만들었다.** 엔진 수명주기 신호이므로 High-risk다.

## 세 판

| | 1판 (편집 전) | 2판 (`6cd643ca`) | **3판 (§3.9, 이 문서)** |
|---|---|---|---|
| 위치 | `:178-216` | `:277-303` | **`:296-324`** |
| 시그니처 | `(release func(), err error)` | `(*Held, error)` | **그대로** |
| AST 분기 | 4 | 3 | **3** |
| 이탈 | 5 | 4 | **4** |
| 호출 | 14 | 10 | **12** |
| source SHA-256 | `d65deddfd1e6…` | `8c784e84d88e…` | **`9399c1d68e1d…`** |
| `Hold` 커버리지 | 80.0% | 86.7% | **88.2%** |
| 패키지 통과 | 11건 | 19건 | **24건** |

**3판이 바꾼 것은 두 줄이다** — `h.live = true`를 `h.mu` 아래로 옮겼다(A2 F3·F4).
분기·이탈은 그대로이고 호출이 둘 늘었다(`Lock`/`Unlock`). 몸통의 판정은 불변이다.

**분기가 하나 줄었다.** 1판의 B2(`os.Remove` 실패)는 stop 클로저 안에 있었고, 2판에서는
`(*Held).Release`라는 **이름 있는 메서드**로 나갔다. 함수 밖으로 나간 것이지 사라진 것이
아니다 — 같은 조건, 같은 침묵, 같은 이유(제거 못 한 마커는 StaleAfter로 스스로 늙는다).
1판의 B3(for)·B4(select)는 2판에서 B2·B3로 번호만 밀렸다.

## 이 함수가 하는 일

엔진이 자기에 대해 발행하는 **자문 마커**를 쓰고, 프로세스가 사는 동안 1분마다 다시 쓰고,
정상 종료에 지운다. 배타가 아니다 — 배타는 `engine.lock`의 flock이다(패키지 주석).

2판의 계약: `Hold(ctx, path, now) (*Held, error)`. 핸들은 **항상 non-nil**이므로 호출자가
오류를 보기 전에 `defer h.Release()` 할 수 있다 — 1판의 "release는 항상 non-nil"과 같은
성질이고, 같은 이유(`cmd/tossctl/engine.go`의 6단계는 마커 실패에 return하지 않는다)다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx` | 취소 가능 | `runEngineRun`의 `cmd.Context()` | 취소되면 refresh goroutine이 조용히 끝난다 (마커는 StaleAfter로 늙는다) |
| `path` | 저널 디렉터리의 `engine-run.json` | `enginelock.MarkerPath(dir)` | 쓸 수 없으면 오류 + **inert 핸들**. 기동은 막지 않는다 |
| `now` | 임의 | `clk.Now()` | 파일 mtime이 되고 그것이 freshness의 사실 |
| `binstamp.Self()` | 실패 가능 | 실행 파일 | 오류는 **버린다** — 모르는 빌드는 stale이 아니다 |
| `h.live` | Hold 안에서 한 번만 쓴다 | 첫 write의 성공 | false면 `Release`·`Ready`·`refresh` 전부 no-op |

> **관통 불변식 (2판에서 강화됨)**: 마커의 어떤 실패도 엔진 기동을 막지 않는다.
> **그리고 실패한 Hold는 아무것도 발행하지 않는다** — `live=false`인 핸들의 `Ready`는
> 준비 신호를 쓰지 않고 `Release`는 남의 파일을 지우지 않는다. 1판의
> `return func(){}, werr`가 가지고 있던 의미를 핸들이 그대로 진다.

## Branches and early returns

| Branch | 위치 | Condition | Mutation/side effect | Return/이탈 |
|---|---|---|---|---|
| B1 | `:301` | `write(path, m, now) != nil` | `h.live`가 false로 남는다 | `:302` `return h, werr` — **inert 핸들** |
| B2 | `:311` | `for {}` — refresh 루프 | — | 루프 자체는 이탈이 없다 |
| B3 | `:312` | `select` 3-way: `<-h.done` / `<-ctx.Done()` / `at := <-ticker.C` | 셋째 갈래만 `h.refresh(at)` | `:314`·`:316` goroutine return |

Return 4개: `:302`(B1) · `:314`·`:316`(refresh goroutine) · `:323`(정상 `return h, nil`).

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `os.Getpid` `:297` | 마커의 pid | — | ast.json |
| `now.UTC` `:297` | StartedAt 정규화 | — | ast.json |
| `binstamp.Self` `:298` | 실행 파일 지문 | **오류를 버린다** | ast.json |
| `make(chan struct{})` `:300` | stop 신호 | — | ast.json |
| `write` `:301` | 첫 마커 (**원자 교체** — §3.9 D4c) | 오류는 호출자에게 (거절은 아니다) | ast.json |
| `time.NewTicker(RefreshEvery)` `:309` | 1분 갱신 | `defer ticker.Stop()` `:289` | ast.json |
| `ctx.Done` `:315` | 종료 관측 | — | ast.json |
| **`h.refresh(at)` `:318`** | **갱신 쓰기** | 실패는 침묵 — 다음 tick이 다시 쓴다 | ast.json · **D4의 핵심** |

live binding — 프로덕션 호출자는 **하나**다: `cmd/tossctl/engine.go:239`
(`rg -n 'enginelock.Hold' --glob '!*_test.go'` → 1건). 테스트 호출자 8건은
2판 시그니처에 맞춰 갱신했다(`cmd/tossctl/engineproc_test.go` 4 ·
`internal/console/engineproc_test.go` 1 · `internal/enginelock/enginelock_test.go` 3).

## State mutations and fallbacks

- 프로세스 밖 상태: **파일 하나**(`path`)의 생성·갱신·삭제. 그 외 전역 상태 없음.
- **1판이 값으로 캡처하던 `m`이 2판에서는 `Held.marker`이고 `Held.mu`가 지킨다.**
  `Ready`와 refresh goroutine이 같은 값을 보기 때문이다. 1판에는 그것을 바꾸는 주체가
  없었으므로 뮤텍스가 필요 없었다 — D4가 그 전제를 깼고, 이 문서의 1판이 그것을 미리 적었다.
- **3판은 그 뮤텍스가 파일까지 덮게 했다.** 2판은 잠금 안에서 값을 복사한 뒤 잠금 **밖에서**
  `write`를 불렀고, A2가 그 창을 실측했다(ready_at 디스크 소거 139/3000, 찢어진 읽기
  3617/12259). 3판에서 `Ready`·`refresh`는 `h.mu`를 잡은 채 `write`까지 끝내고,
  `write` 자체가 tmp+rename으로 원자다(`internal-enginelock--write` 묶음).
- `Marker.ReadyAt`은 포인터지만 **가리키는 값을 고쳐 쓰는 코드가 없다.** `Ready`는 새
  포인터를 대입할 뿐이므로 복사본이 들고 있는 옛 포인터는 언제나 안전하다.
- fallback: 첫 쓰기 실패 → inert 핸들 + 오류. refresh 실패 → 침묵. Ready 쓰기 실패 → 침묵
  (그 대가는 콘솔이 안 해도 될 대기를 하는 것이지, 거절이 아니다).

## Safety conclusion

- Safe edit boundary (실제로 한 것):
  1. `Marker`에 `ReadyAt *time.Time` 추가. 관대한 reader 규율은 그대로다 —
     `TestAMarkerWithoutReadyAtStaysReadable`이 옛 빌드의 마커를 고정한다.
  2. 반환을 핸들로 교체하고 **B1의 no-op 의미를 `live` 플래그로 보존**했다
     (`TestAFailedHoldPublishesNothing`).
  3. refresh를 `(*Held).refresh`로 빼서 뮤텍스로 보호된 상태를 읽게 했다
     (`TestRefreshPreservesTheReadySignal` — 검증계약 (d)).
  4. **§3.9**: `h.live`의 쓰기를 `h.mu` 아래로 옮겼다. 읽는 쪽(`Release`·`Ready`·`refresh`)이
     전부 잠금 아래로 들어갔으므로 쓰는 쪽도 같아야 한다.
- High-risk impact: **yes** (엔진 수명주기 신호). 방향은 보수적이다 — 이 편집이 잘못돼도
  최악은 "콘솔이 안 해도 될 대기를 상한까지 한 뒤 서베이를 시작한다"이고, 그것은 오늘의
  동작(즉시 시작)보다 늦을 뿐 위험하지 않다.
- 남은 공백: **`Release`의 `os.Remove` 실패 갈래는 여전히 count=0**이다.
  1판의 B2와 같은 공백을 그대로 물려받았고 a102는 그것을 메우지 않는다 —
  **침묵하지 않고 이름을 붙여 남긴다.**
