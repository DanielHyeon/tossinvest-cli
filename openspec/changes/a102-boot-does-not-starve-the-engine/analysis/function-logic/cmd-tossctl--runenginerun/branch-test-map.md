# Branch Test Map: `runEngineRun`

Source: `cmd/tossctl/engine.go` (183-301). AST 기준 branches **18** / returns 14.

## 커버리지는 주장이 아니라 측정값이다

`go test ./cmd/tossctl -count=1 -covermode=set -coverprofile`을 **편집 전과 후에 각각** 돌려
`runEngineRun`의 블록 카운트를 잘라 읽었다. 분기의 *조건*이 아니라 **분기 본문 블록의 실행
여부**를 셌다.

| | 통과 | 위치 | 블록 | source SHA-256 |
|---|---|---|---|---|
| 편집 전 | 526건 | `:183-296` | 37개 중 **21개** 실행 | `f13e36b35e08…` |
| **편집 후** | **550건** | **`:183-301`** | **39개 중 22개** 실행 | `8ad1cc88b9e0…` |

늘어난 두 블록은 `marker.Ready` closure(`:256`)와 그 뒤로 밀린 좌표다. **어느 분기의
조건도 바뀌지 않았고, 미측정 분기가 측정으로 넘어가지도 않았다** — 아래 표의 좌표만 밀렸다.

| Branch | 위치 | 본문 실행 | 근거 블록 (편집 전) | 지는 테스트 (편집 전) |
|---|---|---|---|---|
| B1 | `:185` `ctx == nil` | **no** | `185.16,187.3` count=**0** | 없음 — cobra가 항상 ctx를 준다 |
| B2 | `:191` 저널 경로 오류 | **no** | `191.16,193.3` count=**0** | 없음 |
| B3 | `:197` flock 실패 | yes | `197.16,199.3` count=**1** | `TestTheJournalDirectoryIsLockedBeforeAnythingIsAssembled` |
| B4 | `:207` 조립 오류 | yes | `207.16,208.67` count=**1** | `TestAnUnmetInterlockIsEnumerated` |
| B5 | `:208` 인터록 절 있음 | yes | `208.67,210.35` count=**1** | 같음 |
| B6 | `:210` 절 열거 | yes | `210.35,212.5` count=**1** | 같음 |
| B7 | `:219` 게이트 OFF | yes | `219.31,221.3` count=**1** | `TestTheLockIsReleasedWhenTheCommandReturns` |
| B8 | `:229` verify lock 경로 해석됨 | yes | `229.63,230.81` count=**1** | `TestAFreshVerifyRunLockRefusesTheStart` |
| B9 | `:230` verify 진행 중 | yes | `230.81,234.4` count=**1** | 같음 |
| B10 | `:241` **마커를 못 썼다** | **no** | `241.17,245.3` count=**0** | 없음 |
| B11 | `:245` else — 마커 정상 | yes | `245.8,248.3` count=**1** | `TestTheMarkerIsHeldWhileTheLoopsRunAndRemovedAfter` |
| B12 | `:257` 루프 조립 오류 | yes | `257.16,259.3` count=**1** | `TestTheLockIsReleasedWhenTheCommandReturns` |
| B13 | `:261` 정책 명령 서비스 오류 | **no** | `261.16,263.3` count=**0** | 없음 |
| B14 | `:265` 정책 control 서버 오류 | **no** | `265.16,267.3` count=**0** | 없음 |
| B15 | `:270` 정책 runtime 서버 오류 | **no** | `270.16,272.3` count=**0** | 없음 |
| B16 | `:275` 전략 projection 오류 | **no** | `275.16,277.3` count=**0** | 없음 |
| B17 | `:283` alert operations 오류 | **no** | `283.16,285.3` count=**0** | 없음 |
| B18 | `:287` alert control 서버 오류 | **no** | `287.16,289.3` count=**0** | 없음 |

**편집 전·후 측정 동일: 18개 분기 중 10개 본문 실행, 8개 미실행.**

> ⚠ **a102가 편집하는 두 줄 중 하나(`:239`)의 실패 분기 B10이 편집 후에도 미측정이다.** a102는 그것을
> 새로 메우지 않는다 — 마커 쓰기를 실패시키려면 이 함수 안에서 파일시스템을 막아야 하고,
> 그것은 이 change의 범위가 아니다. 대신 **B10이 지켜야 할 계약**(실패한 Hold의 핸들은
> `Ready()`·`Release()`가 아무 일도 하지 않는다)을 `internal/enginelock` 쪽 단위 테스트로
> 고정했다. **침묵하지 않고 이름을 붙여 남긴다.**
>
> B13~B18의 공백은 **물려받은 것**이다(control-plane 서버 기동 실패 6종). a102는 그 줄들을
> 편집하지 않는다.

## 산출물 근거

- 분기·이탈 열거: `ast.json` (branches 18, returns 14) — `go run ./tools/logic-map`
- 커버리지: `go test ./cmd/tossctl -count=1 -covermode=set -coverprofile` exit 0 ·
  **526건 통과**(편집 전) → **550건 통과**(편집 후)
- seam 전수: `rg -n 'engineRuntimeFactory'` → 선언 `engine.go:354` · 호출 `engine.go:256` ·
  테스트 교체 `engine_test.go:61-66`
