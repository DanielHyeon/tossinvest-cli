# Branch Test Map: `write`

Source: `internal/enginelock/enginelock.go` (340-382). AST 기준 branches **8** / returns 9.

## 커버리지는 주장이 아니라 측정값이다

`go test ./internal/enginelock -count=1 -covermode=set -coverprofile` (24건 통과 · `write`
**53.6%** statements · 블록 17개 중 9개 실행).

| Branch | 위치 | 본문 실행 | 근거 블록 | 지는 테스트 |
|---|---|---|---|---|
| B1 | `:342` `MkdirAll` 오류 | **yes** | `342.48,344.3` count=**1** | `TestAFailedHoldPublishesNothing` |
| B2 | `:347` 직렬화 오류 | **no** | `347.16,349.3` count=**0** | 없음 — `Marker`는 언제나 직렬화된다 |
| B3 | `:352` `CreateTemp` 오류 | **no** | `352.16,354.3` count=**0** | 없음 |
| B4 | `:356` 임시 파일 쓰기 오류 | **no** | `356.57,360.3` count=**0** | 없음 |
| B5 | `:361` `Close` 오류 | **no** | `361.36,364.3` count=**0** | 없음 |
| B6 | `:365` `Chmod` 오류 | **no** | `365.48,368.3` count=**0** | 없음 |
| B7 | `:373` `Chtimes` 오류 | **no** | `373.51,376.3` count=**0** | 없음 |
| B8 | `:377` `Rename` 오류 | **no** | `377.48,380.3` count=**0** | 없음 |

정상 경로(`:381` `return nil`)는 `340.55,342.48`·`345.2,347.16`·`351.2,352.16`·
`355.2,356.57`·`361.2`·`365.2`·`373.2`·`377.2`·`381.2`가 전부 count=1이다 —
**교체의 여덟 단계가 순서대로 실행된다.**

> ⚠ **여덟 실패 분기는 전부 미측정이다.** 만들려면 파일시스템을 중간에 고장 내야 하고,
> 이 change는 그 주입 지점을 만들지 않는다(YAGNI: 노브 하나가 마커 쓰기 경로에 생긴다).
> B1만 예외로 잡히는데, `MkdirAll`은 경로 상위가 파일일 때 실패하고 그것은 파일로 만들 수
> 있기 때문이다. **침묵하지 않고 이름을 붙여 남긴다.**
>
> 대신 **성공 경로가 가진 성질**은 전부 측정으로 고정했다 — 그것이 A2 F3이 요구한 것이다.

## 이 편집이 지는 성질 (전부 측정)

| 성질 | 지는 테스트 | 측정 |
|---|---|---|
| 독자가 반쪽 파일을 보지 않는다 | `TestTheMarkerIsNeverReadHalfWritten` | 편집 전 **torn=30761 / 52239 reads**, 편집 후 **0** |
| 갱신이 `ready_at`을 잃지 않는다 | `TestRefreshPreservesTheReadySignal` · 같은 위 테스트의 `lost-ready` | 편집 후 0 |
| 파일 모드가 0600이다 | `TestTheMarkerFileKeepsItsMode` | 실측 0600 |
| 임시 파일이 남지 않는다 | `TestAFailedWriteLeavesNoDebris` | `.engine-run-*` 0건 |
| mtime이 주입된 시계의 것이다 | `TestAHeldMarkerReadsAsRunning` · `TestAStaleMarkerIsNotARunningEngine` (기존) | 무회귀 |

## 뮤테이션 정산

| 뮤테이션 | 가한 것 | 죽은 테스트 | 원복 |
|---|---|---|---|
| **(l)** | 원자 교체 앞에 `os.WriteFile`+`Chtimes`를 되돌려 넣는다 (편집 전 형태) | `TestTheMarkerIsNeverReadHalfWritten` | sha `9399c1d68e1d` 동일 |
| **(d)** | `refresh`가 쓰기 직전 `ReadyAt = nil` | `TestRefreshPreservesTheReadySignal` · `TestTheMarkerIsNeverReadHalfWritten` | sha `9399c1d68e1d` 동일 |

## 산출물 근거

- 분기·이탈 열거: `ast.json` (branches 8, returns 9) — `go run ./tools/logic-map`
- 커버리지: `go test ./internal/enginelock -count=1 -covermode=set -coverprofile` exit 0 ·
  **24건 통과** · `-race` 포함 통과
- 호출자 전수: `rg -n 'write\(' internal/enginelock/enginelock.go` → `Hold`(첫 쓰기) ·
  `(*Held).Ready` · `(*Held).refresh` — **뒤의 둘은 `h.mu`를 잡은 채 부른다**
