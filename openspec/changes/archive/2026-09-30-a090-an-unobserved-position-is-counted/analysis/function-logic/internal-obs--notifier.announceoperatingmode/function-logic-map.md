# Function Logic Map: `Notifier.AnnounceOperatingMode`

- Source: `internal/obs/mode.go` (`48`–`54`)
- Qualified: `Notifier.AnnounceOperatingMode`
- AST evidence: `ast.json` (`source_sha256` ff7224f12ed7c089… — base `2f698db6`) — branches 1 · returns 2
- Risk scan: `risk-pattern-report.md`
- 작성 시점: **a090 3판 설계 전**(codex 1라운드 N2) 판을 **구현 로트 편집 전(base 재고정 뒤, 2026-09-30)** 에 다시 추출했다.

## a092 이후 바뀐 것 (구현 로트 기록)

설계 단계 번들(옛 base `d3bd1843`)은 **2 분기 · 호출 7** 판이었다 — `Event` 구조체를 함수 안에서 만들고 key 에서 전이 id 를 **뺐다**.
a092(K1, 아카이브 `2026-09-30-a092-…`)가 그 구성을 **이미 순수 함수로 추출**했다: `operatingModeEvent(previous, rec)`
(`internal/obs/record_only.go:169-191`, 비공개) — 이 함수와 `RecordOnly.AnnounceOperatingMode` 가 같이 쓴다. key 도 바뀌었다:
`OperatingModeEventKey(rec)` = `operating_mode:<account>:<mode>:<전이 id>`(`record_only.go:164-166`).

그래서 a090 tasks 3.4 의 「`obs.OperatingModeEvent` 추출」 은 **이 함수를 편집하지 않고** 끝난다 — 새 공개 래퍼
`obs.OperatingModeEvent` 한 줄(새 함수, `operatingModeEvent` 호출)만 더한다. **이 함수는 a090 편집 대상에서 빠진다**(분기·key·`Notify` 무변화가
아니라 **편집 0**). 이 번들은 편집 전 사실의 기록으로 남긴다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `n` | nil 가능 | 배선 | nil 이면 무동작(B1) |
| `previous`·`rec` | 커밋된 전이 | `journal.TransitionOperatingMode` | — |

## Branches and early returns

| Branch | 종류 | 조건 (원문) | 결과 |
|---|---|---|---|
| B1 | if | `:49` `if n == nil {` | **return nil** `:50` |

마지막 **return** `:53` `n.Notify(ctx, operatingModeEvent(previous, rec))`.

## Calls and live bindings

`operatingModeEvent`(순수) · `n.Notify`(동기 경로 — 생산 exit 관측자는 이 함수가 아니라 `RecordOnly.AnnounceOperatingMode` 를 받는다,
`internal/app/engine/exitwiring.go:348-351`).

## State mutations and fallbacks

outbox 행(Notify 경유). 그 밖 없음.

## Safety conclusion

- **Safe edit boundary**: 편집 없음(a092 가 추출 완료). a090 은 `operatingModeEvent` 의 공개 래퍼만 새로 둔다.
- **High-risk impact**: no — a090 에서 이 함수는 편집 0.
